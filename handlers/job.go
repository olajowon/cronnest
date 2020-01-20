package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	db "cronnest/database"
	"cronnest/models"
	"encoding/json"
	"time"
	"strconv"
	"fmt"
)

type JobReqData struct {
	Comment 	string
	Status 		string
	Spec		string
	Content		string
	Host		string
	Sysuser		string
}

func GetJobs(c *gin.Context) {
	search := c.DefaultQuery("search", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	limit := pageSize
	offset := (page - 1) * pageSize

	var jobs []models.Job
	var count int
	if search != "" {
		search = fmt.Sprintf("%%%v%%", search)
		db.DB.Find(&jobs).Where("comment LIKE ?", search).Count(&count).
			Order("comment").Limit(limit).Offset(offset)
	} else {
		db.DB.Find(&jobs).Count(&count).Order("comment").Limit(limit).Offset(offset)
	}

	var data []map[string]interface{}
	data = []map[string]interface{} {}
	for _, job := range jobs {
		data = append(data, MakeJobKv(job))
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data, "total": count})
}


func CreateJob(c *gin.Context) {
	user, _, _ := c.Request.BasicAuth()
	rawData, _ := c.GetRawData()
	reqData := JobReqData{}
	json.Unmarshal(rawData, &reqData)
	error := cleanedJobData(&reqData)
	if len(error) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400,"msg": "Bad POST data", "error": error})
		return
	}

	transaction := db.DB.Begin()

	// 创建新 job
	job := models.Job{Comment:reqData.Comment, Status:reqData.Status,
		Spec:reqData.Spec, Content: reqData.Content, Host:reqData.Host, Sysuser:reqData.Sysuser,
		Created:time.Now(), Updated:time.Now()}
	result := transaction.Create(&job)
	if result.Error != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": result.Error.Error()})
		return
	}

	// 更新主机
	host := models.Host{}
	transaction.Where("address = ?", reqData.Host).First(&host)
	if host.Id == 0 {
		host.Address = reqData.Host
		host.Status = "enabled"
		host.Created = time.Now()
		host.Updated = time.Now()
		result = transaction.Create(&host)

		if result.Error != nil {
			transaction.Rollback()
			c.JSON(http.StatusInternalServerError,
				gin.H{"code": 500, "msg": result.Error.Error()})
			return
		}

		hostData := MakeHostKv(host)
		jsonHostData, _ := json.Marshal(hostData)
		operRecord := models.OperationRecord{ResourceType:"host", ResourceId:host.Id, ResourceLabel: host.Address,
			OperationType: "create", Data:jsonHostData, User: user, Created: time.Now()}
		if result = db.DB.Create(&operRecord); result.Error != nil {
			transaction.Rollback()
			c.JSON(http.StatusInternalServerError,
				gin.H{"code": 500, "msg": result.Error.Error()})
			return
		}
	}

	status, msg := DoHostCronJob(job, "update")
	fmt.Println(status, msg)
	if status == false {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError,
			gin.H{"code": 500, "msg": fmt.Sprintf("推送Job至主机[%s]失败：%s", job.Host, msg)})
		return
	}

	// 记录操作
	data := MakeJobKv(job)
	jsonData, _ := json.Marshal(data)
	operRecord := models.OperationRecord{ResourceType:"job", ResourceId:job.Id, ResourceLabel: job.Comment,
		OperationType: "create", Data:jsonData, User: user, Created: time.Now()}
	if result = db.DB.Create(&operRecord); result.Error != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError,
			gin.H{"code": 500, "msg": result.Error.Error()})
		return
	}
	transaction.Commit()
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

func UpdateJob(c *gin.Context) {
	user, _, _ := c.Request.BasicAuth()
	pk := c.Param("pk")
	updateJob := models.Job{}

	db.DB.Where("id = ?", pk).First(&updateJob)
	if updateJob.Id == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "Job not found"})
		return
	}

	oldJob := updateJob

	rawData, _ := c.GetRawData()
	reqData := JobReqData{}
	json.Unmarshal(rawData, &reqData)
	error := cleanedJobData(&reqData)
	if len(error) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Bad PUT data", "error": error})
		return
	}

	transaction := db.DB.Begin()
	updateJob.Comment = reqData.Comment
	updateJob.Status = reqData.Status
	updateJob.Spec = reqData.Spec
	updateJob.Content = reqData.Content
	updateJob.Host = reqData.Host
	updateJob.Sysuser = reqData.Sysuser
	updateJob.Updated = time.Now()
	result := transaction.Save(&updateJob)
	if result.Error != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": result.Error.Error()})
		return
	}

	fmt.Println(oldJob)
	fmt.Println(updateJob)

	status, msg := true, ""

	// 删除原主机JOB
	if oldJob.Host != updateJob.Host {
		rmStatus, rmMsg := DoHostCronJob(oldJob, "remove")
		if rmStatus == false {
			status = rmStatus
			msg = fmt.Sprintf("从原主机[%s]删除Job失败：%s", oldJob.Host, rmMsg)
		}
	}

	// 更新主机JOB
	if status == true {
		upStatus, upMsg := DoHostCronJob(updateJob, "update")
		fmt.Println(status, msg)
		if upStatus == false {
			status = upStatus
			msg = fmt.Sprintf("推送Job至主机[%s]失败：%s", updateJob.Host, upMsg)
		}
	}

	// 如出现异常进行回滚
	if status == false {
		transaction.Rollback()
		rbStatus, rbMsg := DoHostCronJob(oldJob, "update")
		if rbStatus == false {
			msg = fmt.Sprintf("%s\n\n 且回滚主机[%s]Job失败：%s", msg, oldJob.Host, rbMsg)
		} else {
			msg = fmt.Sprintf("%s\n\n 回滚主机[%s]Job成功", msg, oldJob.Host)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": msg})
		return
	}

	host := models.Host{}
	transaction.Where("address = ?", reqData.Host).First(&host)
	if host.Id == 0 {
		host.Address = reqData.Host
		host.Status = "enabled"
		host.Created = time.Now()
		host.Updated = time.Now()
		result = transaction.Create(&host)

		if result.Error != nil {
			transaction.Rollback()
			c.JSON(http.StatusInternalServerError,
				gin.H{"code": 500, "msg": fmt.Sprintf("%v", result.Error)})
			return
		}

		hostData := MakeHostKv(host)
		jsonHostData, _ := json.Marshal(hostData)
		operRecord := models.OperationRecord{ResourceType:"host", ResourceId:host.Id, ResourceLabel: host.Address,
			OperationType: "create", Data:jsonHostData, User: user, Created: time.Now()}
		if result = db.DB.Create(&operRecord); result.Error != nil {
			transaction.Rollback()
			c.JSON(http.StatusInternalServerError,
				gin.H{"code": 500, "msg": result.Error.Error()})
			return
		}
	}

	data := MakeJobKv(updateJob)
	jsonData, _ := json.Marshal(data)
	operRecord := models.OperationRecord{ResourceType:"job", ResourceId:updateJob.Id, ResourceLabel: updateJob.Comment,
		OperationType: "update", Data:jsonData, User: user, Created: time.Now()}
	if result = db.DB.Create(&operRecord); result.Error != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError,
			gin.H{"code": 500, "msg": result.Error.Error()})
		return
	}
	transaction.Commit()
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

func DeleteJob(c *gin.Context) {
	user, _, _ := c.Request.BasicAuth()
	pk := c.Param("pk")
	deleteJob := models.Job{}
	db.DB.Where("id = ?", pk).First(&deleteJob)
	if deleteJob.Id == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "Job not found"})
		return
	}

	transaction := db.DB.Begin()
	status, msg := DoHostCronJob(deleteJob, "remove")
	fmt.Println(status, msg)
	if status == false {
		transaction.Rollback()
		msg = fmt.Sprintf("删除主机Job[%s]失败：%s", deleteJob.Host, msg)

		rbStatus, rbMsg := DoHostCronJob(deleteJob, "update")
		if rbStatus == false {
			msg = fmt.Sprintf("%s\n\n 且回滚主机[%s]Job失败：%s", msg, deleteJob.Host, rbMsg)
		} else {
			msg = fmt.Sprintf("%s\n\n 回滚主机[%s]Job成功", msg, deleteJob.Host)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": msg})
		return
	}

	result := db.DB.Delete(&deleteJob)
	if result.Error != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": result.Error.Error()})
		return
	}

	data := MakeJobKv(deleteJob)
	jsonData, _ := json.Marshal(data)
	operRecord := models.OperationRecord{ResourceType:"job", ResourceId:deleteJob.Id, ResourceLabel: deleteJob.Comment,
		OperationType: "delete", Data:jsonData, User: user, Created: time.Now()}
	if result = db.DB.Create(&operRecord); result.Error != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError,
			gin.H{"code": 500, "msg": result.Error.Error()})
		return
	}

	transaction.Commit()
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

func cleanedJobData(data *JobReqData) map[string]string {
	error := make(map[string]string)
	if data.Comment == "" {
		error["name"] = "任务注释（描述）是必填项"
	}

	if data.Status != "enabled" && data.Status != "disabled" {
		error["status"] = "任务状态无效"
	}


	if data.Spec == "" {
		error["spec"] = "任务调度是必填项"
	}

	if data.Content == "" {
		error["content"] = "任务脚本内容是必填项"
	}

	if data.Host == "" {
		error["host"] = "任务主机地址是必填项"
	}

	if data.Sysuser == "" {
		error["sysuser"] = "任务所属的系统用户是必填项"
	}

	return error
}