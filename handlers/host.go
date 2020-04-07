package handlers

import (
	"strconv"
	"cronnest/models"
	"net/http"
	"github.com/gin-gonic/gin"
	"fmt"
	db "cronnest/database"
	lg "cronnest/logger"
	"time"
	"encoding/json"
)

type hostReqData struct {
	Address string `json:"address"`
}


func GetHosts(c *gin.Context) {
	search := c.DefaultQuery("search", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	limit := pageSize
	offset := (page - 1) * pageSize

	var hosts []models.Host
	var count int
	if search != "" {
		search = fmt.Sprintf("%%%v%%", search)
		db.DB.Table("host").Where("address LIKE ?", search).Count(&count).Limit(limit).Offset(offset).
			Order("address").Find(&hosts)
	} else {
		db.DB.Table("host").Count(&count).Limit(limit).Offset(offset).Order("address").Find(&hosts)
	}

	var data []map[string]interface{}
	data = []map[string]interface{} {}
	for _, host := range hosts {
		data = append(data, MakeHostKv(host))
	}

	c.JSON(http.StatusOK, gin.H{"data": data, "total": count})
}


func UpdateHost(c *gin.Context) {
	user, _, _ := c.Request.BasicAuth()
	hId := c.Param("hId")
	var hostMdl models.Host
	db.DB.Table("host").Where("id=?", hId).First(&hostMdl)
	if hostMdl.Id == 0 {
		c.JSON(http.StatusNotFound, gin.H{"msg": "主机不存在"})
	}

	var reqData hostReqData
	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	var addressCount int64
	db.DB.Table("host").Where(
		"address=? AND id!=?", reqData.Address, hId).Count(&addressCount)
	if addressCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("主机地址[%s]已存在", reqData.Address)})
		return
	}

	transaction := db.DB.Begin()
	hostMdl.Address = reqData.Address
	hostMdl.UpdatedAt = time.Now()
	result := transaction.Table("host").Save(&hostMdl)
	if result.Error != nil {
		transaction.Rollback()
		msg := fmt.Sprintf("修改主机[%s]失败, %v", hostMdl.Address, result.Error)
		lg.Logger.Error(msg)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
		return
	}

	hostData := MakeHostKv(hostMdl)

	jsonData, _ := json.Marshal(map[string]interface{}{"updated_host": hostData})
	operRecord := models.OperationRecord{SourceType:"host", SourceId:hostMdl.Id, SourceLabel: hostMdl.Address,
		OperationType: "update", Data:jsonData, User: user, CreatedAt: time.Now()}
	if result = transaction.Table("operation_record").Create(&operRecord); result.Error != nil {
		transaction.Rollback()
		msg := fmt.Sprintf("记录操作失败, %v", result.Error)
		lg.Logger.Error(msg)
		c.JSON(http.StatusInternalServerError,
			gin.H{"code": 500, "msg": msg})
		return
	}
	transaction.Commit()

	c.JSON(http.StatusOK, gin.H{"data": hostData})
}


func DeleteHost(c *gin.Context) {
	user, _, _ := c.Request.BasicAuth()

	hId := c.Param("hId")
	var hostMdl models.Host
	db.DB.Table("host").Where("id=?", hId).First(&hostMdl)
	if hostMdl.Id == 0 {
		c.JSON(http.StatusNotFound, gin.H{"msg": "主机不存在"})
	}

	transaction := db.DB.Begin()
	result := transaction.Table("hostgroup_host").Where("host_id=?", hId).Delete(models.HostgroupHost{})
	if result.Error != nil {
		transaction.Rollback()
		msg := fmt.Sprintf("删除主机[%s]与组关系失败，%v", hostMdl.Address, result.Error)
		lg.Logger.Error(msg)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
		return
	}
	result = transaction.Table("host_crontab").Where("host_id=?", hId).Delete(models.HostCrontab{})
	if result.Error != nil {
		transaction.Rollback()
		msg := fmt.Sprintf("删除主机[%s]Crontab失败，%v", hostMdl.Address, result.Error)
		lg.Logger.Error(msg)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
		return
	}

	result = transaction.Table("host").Where("id=?", hId).Delete(models.Host{})
	if result.Error != nil {
		transaction.Rollback()
		msg :=fmt.Sprintf("删除主机[%s]失败，%v", hostMdl.Address, result.Error)
		lg.Logger.Error(msg)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
		return
	}

	hostData := MakeHostKv(hostMdl)

	recordData, _ := json.Marshal(map[string]interface{}{"deleted_host": hostData})
	operRecord := models.OperationRecord{SourceType:"host", SourceId:hostMdl.Id, SourceLabel: hostMdl.Address,
		OperationType: "delete", Data:recordData, User: user, CreatedAt: time.Now()}
	if result = transaction.Table("operation_record").Create(&operRecord); result.Error != nil {
		transaction.Rollback()
		msg := fmt.Sprintf("记录操作失败, %v", result.Error)
		lg.Logger.Error(msg)
		c.JSON(http.StatusInternalServerError,
			gin.H{"code": 500, "msg": msg})
		return
	}
	transaction.Commit()

	c.JSON(http.StatusOK, gin.H{"data": hostData})
}