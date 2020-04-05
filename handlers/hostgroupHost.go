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

type hostgroupHostReqData struct {
	Hosts    []string `json:"hosts"`
}

func GetHostgroupHosts(c *gin.Context) {
	hgId := c.Param("hgId")
	search := c.DefaultQuery("search", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	limit := pageSize
	offset := (page - 1) * pageSize

	hosts := []models.Host{}
	count := 0
	q := db.DB.Table("host").Joins(
		"JOIN hostgroup_host ON hostgroup_host.hostgroup_id=? AND host.id=hostgroup_host.host_id", hgId)
	if search != "" {
		search = fmt.Sprintf("%%%v%%", search)
		q.Where("address LIKE ?", search).Count(&count).Limit(limit).
			Offset(offset).Order("address").Find(&hosts)
	} else {
		q.Count(&count).Limit(limit).Offset(offset).Order("address").Find(&hosts)
	}

	data := []map[string]interface{}{}
	for _, h := range hosts {
		data = append(data, MakeHostKv(h))
	}

	c.JSON(http.StatusOK, gin.H{"data": data, "total": count})
}

func AddHostgroupHosts(c *gin.Context) {
	user, _, _ := c.Request.BasicAuth()
	hgId := c.Param("hgId")
	var hostgroupMdl models.Hostgroup
	db.DB.Table("hostgroup").Where("id=?", hgId).Find(&hostgroupMdl)
	if hostgroupMdl.Id == 0 {
		c.JSON(http.StatusNotFound, gin.H{"msg": "主机组不存在"})
		return
	}

	var reqData hostgroupHostReqData
	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	transaction := db.DB.Begin()

	var addedHostMdls []models.Host
	var createdHostMdls []models.Host
	for _, host := range reqData.Hosts {
		if host != "" {
			var hostMdl models.Host
			transaction.Table("host").Where("address=?", host).First(&hostMdl)
			if hostMdl.Id == 0 {
				hostMdl.Address = host
				hostMdl.CreatedAt = time.Now()
				hostMdl.UpdatedAt = time.Now()
				hostMdl.Status = "enabled"
				result := transaction.Table("host").Save(&hostMdl)
				if result.Error != nil {
					transaction.Rollback()
					msg := fmt.Sprintf("创建主机[%s]失败, %v", hostgroupMdl.Name, host, result.Error)
					lg.Logger.Error(msg)
					c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
					return
				}
				createdHostMdls = append(createdHostMdls, hostMdl)
			}
			addedHostMdls = append(addedHostMdls, hostMdl)

			var hostgroupHostMdl models.HostgroupHost
			transaction.Table("hostgroup_host").Where(
				"hostgroup_id=? AND host_id=?", hgId, hostMdl.Id).Find(&hostgroupHostMdl)
			if hostgroupHostMdl.HostgroupId == 0 {
				hostgroupHostMdl.HostgroupId = hostgroupMdl.Id
				hostgroupHostMdl.HostId = hostMdl.Id
				result := transaction.Table("hostgroup_host").Save(&hostgroupHostMdl)
				if result.Error != nil {
					transaction.Rollback()
					msg := fmt.Sprintf("创建主机组[%s]主机[%s]关系失败, host, %v", hostgroupMdl.Name, result.Error)
					lg.Logger.Error(msg)
					c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
					return
				}
			}
		}
	}

	addedHostData := []map[string]interface{}{}
	createdHostData := []map[string]interface{}{}
	for _, h := range addedHostMdls {
		addedHostData = append(addedHostData, MakeHostKv(h))
	}
	for _, h := range createdHostMdls {
		createdHostData = append(createdHostData, MakeHostKv(h))
	}

	recordData, _ := json.Marshal(map[string]interface{}{
		"add_hosts": addedHostData, "createdHostData": createdHostData})
	operRecord := models.OperationRecord{SourceType: "hostgroup_host", SourceId: hostgroupMdl.Id,
		SourceLabel: hostgroupMdl.Name, OperationType: "add", Data: recordData,
		User: user, CreatedAt: time.Now()}
	if result := transaction.Create(&operRecord); result.Error != nil {
		transaction.Rollback()
		msg := fmt.Sprintf("记录操作失败, %v", result.Error)
		lg.Logger.Error(msg)
		c.JSON(http.StatusInternalServerError,
			gin.H{"code": 500, "msg": msg})
		return
	}

	transaction.Commit()
	c.JSON(http.StatusOK, gin.H{"data": addedHostData})
}


func RemoveHostgroupHosts(c *gin.Context) {
	user, _, _ := c.Request.BasicAuth()
	hgId := c.Param("hgId")
	var hostgroupMdl models.Hostgroup
	db.DB.Table("hostgroup").Where("id=?", hgId).Find(&hostgroupMdl)
	if hostgroupMdl.Id == 0 {
		c.JSON(http.StatusNotFound, gin.H{"msg": "主机组不存在"})
		return
	}

	var reqData hostgroupHostReqData
	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}


	var removedHostMdls []models.Host
	var deletedHostMdls []models.Host
	transaction := db.DB.Begin()
	for _, host := range reqData.Hosts {
		var hostMdl models.Host
		transaction.Table("host").Where("address=?", host).Joins(
			"join hostgroup_host ON hostgroup_host.hostgroup_id=? AND hostgroup_host.host_id=host.id",
			hgId).First(&hostMdl)
		if hostMdl.Id == 0 {
			transaction.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("主机[%s]不存在或者不属于该组", host)})
			return
		}

		removedHostMdls = append(removedHostMdls, hostMdl)

		result := transaction.Table("hostgroup_host").Where(
			"hostgroup_id=? AND host_id=?", hgId, hostMdl.Id).Delete(models.HostgroupHost{})
		if result.Error != nil {
			transaction.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf(
				"移除主机组与主机[%s]关系失败，%v", host, result.Error)})
			return
		}

		var hostOtherHgCount int64
		transaction.Table("hostgroup").Joins(
			"join hostgroup_host ON hostgroup_host.hostgroup_id=hostgroup.id AND hostgroup_host.host_id=?",
			hostMdl.Id).Count(&hostOtherHgCount)

		if hostOtherHgCount > 0 {
			continue
		}
		deletedHostMdls = append(deletedHostMdls, hostMdl)

		result = transaction.Table("host_crontab").Where(
			"host_id=?", hostMdl.Id).Delete(models.HostCrontab{})
		if result.Error != nil {
			transaction.Rollback()
			msg := fmt.Sprintf("删除主机组[%s]的主机[%s]Crontab失败, %v", hostgroupMdl.Name, host, result.Error)
			lg.Logger.Error(msg)
			c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
			return
		}

		result = transaction.Table("host").Where("id=?", hostMdl.Id).Delete(models.Host{})
		if result.Error != nil {
			transaction.Rollback()
			msg := fmt.Sprintf("删除主机组[%s]的主机[%s]失败, %v", hostgroupMdl.Name, host, result.Error)
			lg.Logger.Error(msg)
			c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
			return
		}
	}

	removedHostData := []map[string]interface{}{}
	deletedHostData := []map[string]interface{}{}
	for _, h := range removedHostMdls {
		removedHostData = append(removedHostData, MakeHostKv(h))
	}
	for _, h := range deletedHostMdls {
		deletedHostData = append(deletedHostData, MakeHostKv(h))
	}

	recordData, _ := json.Marshal(map[string]interface{}{
		"removed_hosts": removedHostData, "deleted_hosts": deletedHostData})
	operRecord := models.OperationRecord{SourceType: "hostgroup_host", SourceId: hostgroupMdl.Id,
		SourceLabel: hostgroupMdl.Name, OperationType: "remove", Data: recordData,
		User: user, CreatedAt: time.Now()}
	if result := transaction.Create(&operRecord); result.Error != nil {
		transaction.Rollback()
		msg := fmt.Sprintf("记录操作失败, %v", result.Error)
		lg.Logger.Error(msg)
		c.JSON(http.StatusInternalServerError,
			gin.H{"code": 500, "msg": msg})
		return
	}

	transaction.Commit()
	c.JSON(http.StatusOK, gin.H{"data": removedHostData})
}