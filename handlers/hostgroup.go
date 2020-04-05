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

type reqData struct {
	Name string `json:"name"`
}

func GetHostgroup(c *gin.Context) {
	search := c.DefaultQuery("search", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	limit := pageSize
	offset := (page - 1) * pageSize

	var hostgroupMdls []models.Hostgroup
	var count int
	if search != "" {
		search = fmt.Sprintf("%%%v%%", search)
		db.DB.Table("hostgroup").Where("name LIKE ?", search).Count(&count).Limit(limit).Offset(offset).
			Order("name").Find(&hostgroupMdls)
	} else {
		db.DB.Table("hostgroup").Count(&count).Limit(limit).Offset(offset).Order(
			"name").Find(&hostgroupMdls)
	}

	data := []map[string]interface{}{}
	for _, hg := range hostgroupMdls {
		data = append(data, MakeHostgroupKv(hg))
	}

	c.JSON(http.StatusOK, gin.H{"data": data, "total": count})
}

func CreateHostgroup(c *gin.Context) {
	user, _, _ := c.Request.BasicAuth()

	var reqData reqData
	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	var hostgroupMdl models.Hostgroup
	db.DB.Table("hostgroup").Where("name=?", reqData.Name).Find(&hostgroupMdl)
	if hostgroupMdl.Id > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "主机组已存在"})
		return
	}

	transaction := db.DB.Begin()

	hostgroupMdl = models.Hostgroup{Name: reqData.Name, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	result := transaction.Table("hostgroup").Create(&hostgroupMdl)
	if result.Error != nil {
		transaction.Rollback()
		msg := fmt.Sprintf("创建主机组[%s]失败, %v", hostgroupMdl.Name, result.Error)
		lg.Logger.Error(msg)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
		return
	}

	data := MakeHostgroupKv(hostgroupMdl)

	recordData, _ := json.Marshal(map[string]interface{}{"created_hostgroup": data})
	operRecord := models.OperationRecord{SourceType: "hostgroup", SourceId: hostgroupMdl.Id,
		SourceLabel: hostgroupMdl.Name, OperationType: "create", Data: recordData, User: user, CreatedAt: time.Now()}
	if result = transaction.Create(&operRecord); result.Error != nil {
		transaction.Rollback()
		msg := fmt.Sprintf("记录操作失败, %v", result.Error)
		lg.Logger.Error(msg)
		c.JSON(http.StatusInternalServerError,
			gin.H{"code": 500, "msg": msg})
		return
	}
	transaction.Commit()
	c.JSON(http.StatusCreated, gin.H{"data": data})
}

func UpdateHostgroup(c *gin.Context) {
	user, _, _ := c.Request.BasicAuth()
	hgId := c.Param("hgId")
	var hostgroupMdl models.Hostgroup
	db.DB.Table("hostgroup").Where("id=?", hgId).Find(&hostgroupMdl)
	if hostgroupMdl.Id == 0 {
		c.JSON(http.StatusNotFound, gin.H{"msg": "主机组不存在"})
		return
	}

	var reqData reqData
	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	var hgNameCount int64
	db.DB.Table("hostgroup").Where(
		"name=? AND id!=?", reqData.Name, hgId).Count(&hgNameCount)
	if hgNameCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("主机组名称[%s]已存在", reqData.Name)})
		return
	}

	transaction := db.DB.Begin()
	hostgroupMdl.Name = reqData.Name
	hostgroupMdl.UpdatedAt = time.Now()
	result := transaction.Table("hostgroup").Save(&hostgroupMdl)
	if result.Error != nil {
		transaction.Rollback()
		msg := fmt.Sprintf("修改主机组[%s]失败, %v", hostgroupMdl.Name, result.Error)
		lg.Logger.Error(msg)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
		return
	}

	data := MakeHostgroupKv(hostgroupMdl)

	recordData, _ := json.Marshal(map[string]interface{}{"updated_hostgroup": data})
	operRecord := models.OperationRecord{SourceType: "hostgroup", SourceId: hostgroupMdl.Id,
		SourceLabel: hostgroupMdl.Name, OperationType: "update", Data: recordData, User: user, CreatedAt: time.Now()}
	if result := transaction.Create(&operRecord); result.Error != nil {
		transaction.Rollback()
		msg := fmt.Sprintf("记录操作失败, %v", result.Error)
		lg.Logger.Error(msg)
		c.JSON(http.StatusInternalServerError,
			gin.H{"code": 500, "msg": msg})
		return
	}
	transaction.Commit()
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func DeleteHostgroup(c *gin.Context) {
	user, _, _ := c.Request.BasicAuth()
	hgId := c.Param("hgId")
	var hostgroupMdl models.Hostgroup
	db.DB.Table("hostgroup").Where("id=?", hgId).Find(&hostgroupMdl)
	if hostgroupMdl.Id == 0 {
		c.JSON(http.StatusNotFound, gin.H{"msg": "主机组不存在"})
		return
	}

	transaction := db.DB.Begin()
	var hostMdls []models.Host
	var removedHostMdls []models.Host
	var deletedHostMdls []models.Host
	transaction.Table("host").Joins(
		"join hostgroup_host ON hostgroup_host.host_id=host.id AND hostgroup_host.hostgroup_id=?",
		hgId).Find(&hostMdls)
	for _, hostMdl := range hostMdls {
		var hostOtherHgCount int64
		transaction.Table("hostgroup").Joins(
			"join hostgroup_host ON hostgroup_host.host_id=? AND "+
				"hostgroup_host.hostgroup_id=hostgroup.id AND hostgroup_host.hostgroup_id!=?",
			hostMdl.Id, hgId).Count(&hostOtherHgCount)

		if hostOtherHgCount > 0 {
			removedHostMdls = append(removedHostMdls, hostMdl)
			continue
		}
		deletedHostMdls = append(deletedHostMdls, hostMdl)

		result := transaction.Table("host_crontab").Where(
			"host_id=?", hostMdl.Id).Delete(models.HostCrontab{})
		if result.Error != nil {
			transaction.Rollback()
			msg := fmt.Sprintf(
				"删除主机组[%s]的主机[%s]Crontab失败, %v", hostgroupMdl.Name, hostMdl.Address, result.Error)
			lg.Logger.Error(msg)
			c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
			return
		}

		result = transaction.Table("host").Where("id=?", hostMdl.Id).Delete(models.Host{})
		if result.Error != nil {
			transaction.Rollback()
			msg := fmt.Sprintf(
				"删除主机组[%s]的主机[%s]失败, %v", hostgroupMdl.Name, hostMdl.Address, result.Error)
			lg.Logger.Error(msg)
			c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
			return
		}
	}

	result := transaction.Table("hostgroup").Delete(&hostgroupMdl)
	if result.Error != nil {
		transaction.Rollback()
		msg :=  fmt.Sprintf("删除主机组[%s]失败, %v", hostgroupMdl.Name, result.Error)
		lg.Logger.Error(msg)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
		return
	}

	result = transaction.Table("hostgroup_host").Where(
		"hostgroup_id=?", hostgroupMdl.Id).Delete(&models.HostgroupHost{})
	if result.Error != nil {
		transaction.Rollback()
		msg := fmt.Sprintf(
			"删除主机组[%s]与主机关系失败, %v", hostgroupMdl.Name, result.Error)
		lg.Logger.Error(msg)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
		return
	}

	removedHostData := []map[string]interface{}{}
	for _, hMdl := range removedHostMdls {
		removedHostData = append(removedHostData, MakeHostKv(hMdl))
	}

	deletedHostData := []map[string]interface{}{}
	for _, hMdl := range deletedHostMdls {
		deletedHostData = append(deletedHostData, MakeHostKv(hMdl))
	}

	hgData := MakeHostgroupKv(hostgroupMdl)

	recordData, _ := json.Marshal(map[string]interface{}{
		"deleted_hostgroup": hgData, "removed_hosts": removedHostData, "deleted_hosts": deletedHostData})
	operRecord := models.OperationRecord{SourceType: "hostgroup", SourceId: hostgroupMdl.Id,
		SourceLabel: hostgroupMdl.Name, OperationType: "delete", Data: recordData, User: user, CreatedAt: time.Now()}
	if result := transaction.Create(&operRecord); result.Error != nil {
		transaction.Rollback()
		msg := fmt.Sprintf("记录操作失败, %v", result.Error)
		lg.Logger.Error(msg)
		c.JSON(http.StatusInternalServerError,
			gin.H{"code": 500, "msg": msg})
		return
	}

	transaction.Commit()

	c.JSON(http.StatusOK, gin.H{"data": hgData})
}
