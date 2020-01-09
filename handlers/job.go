package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	db "cronnest/database"
	"cronnest/models"
	"encoding/json"
	"strings"
	"time"
	"strconv"
	"fmt"
	"regexp"
)

type JobReqData struct {
	Name 		string
	Status 		string
	Description	string
	Mailto		string
	Spec		string
	Content		string
	Hosts		[]string
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
		db.DB.Table("job").Where("name LIKE ?", search).Count(&count).Limit(limit).Offset(offset).
			Or("description LIKE ?", search).Order("name").Find(&jobs)
	} else {
		db.DB.Table("job").Count(&count).Limit(limit).Offset(offset).Order("name").Find(&jobs)
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
	error := cleanedJobData(&reqData, 0)
	if len(error) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400,"msg": "Bad POST data", "error": error})
		return
	}

	transaction := db.DB.Begin()
	hosts, _ := json.Marshal(reqData.Hosts)
	job := models.Job{Name:reqData.Name, Status:reqData.Status, Description:reqData.Description, Mailto:reqData.Mailto,
		Spec:reqData.Spec, Content: reqData.Content, Hosts:hosts, Sysuser:reqData.Sysuser,
		Created:time.Now(), Updated:time.Now()}
	result := transaction.Table("job").Create(&job)
	if result.Error != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": result.Error.Error()})
		return
	}

	for _, name := range reqData.Hosts {
		host := models.Host{}
		transaction.Table("host").Where("address = ?", name).First(&host)
		if host.Id == 0 || host.Status != "enabled" {
			var operationType string
			if host.Id == 0 {
				host.Address = name
				host.Status = "enabled"
				host.Created = time.Now()
				host.Updated = time.Now()
				result = transaction.Table("host").Create(&host)
				operationType = "create"
			} else {
				host.Status = "enabled"
				host.Updated = time.Now()
				result = transaction.Table("host").Save(&host)
				operationType = "update"
			}
			if result.Error != nil {
				transaction.Rollback()
				c.JSON(http.StatusInternalServerError,
					gin.H{"code": 500, "msg": result.Error.Error()})
				return
			}

			hostData := MakeHostKv(host)
			jsonHostData, _ := json.Marshal(hostData)
			operRecord := models.OperationRecord{ResourceType:"host", ResourceId:host.Id, ResourceLabel: host.Address,
				OperationType: operationType, Data:jsonHostData, User: user, Created: time.Now()}
			if result = db.DB.Table("operation_record").Create(&operRecord); result.Error != nil {
				transaction.Rollback()
				c.JSON(http.StatusInternalServerError,
					gin.H{"code": 500, "msg": result.Error.Error()})
				return
			}
		}
	}

	data := MakeJobKv(job)
	jsonData, _ := json.Marshal(data)
	operRecord := models.OperationRecord{ResourceType:"job", ResourceId:job.Id, ResourceLabel: job.Name,
		OperationType: "create", Data:jsonData, User: user, Created: time.Now()}
	if result = db.DB.Table("operation_record").Create(&operRecord); result.Error != nil {
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
	db.DB.Table("job").Where("id = ?", pk).First(&updateJob)
	if updateJob.Id == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "Job not found"})
		return
	}

	rawData, _ := c.GetRawData()
	reqData := JobReqData{}
	json.Unmarshal(rawData, &reqData)
	error := cleanedJobData(&reqData, updateJob.Id)
	if len(error) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Bad PUT data", "error": error})
		return
	}

	transaction := db.DB.Begin()
	hosts, _ := json.Marshal(reqData.Hosts)
	updateJob.Name = reqData.Name
	updateJob.Status = reqData.Status
	updateJob.Description = reqData.Description
	updateJob.Mailto = reqData.Mailto
	updateJob.Spec = reqData.Spec
	updateJob.Content = reqData.Content
	updateJob.Hosts = hosts
	updateJob.Sysuser = reqData.Sysuser
	updateJob.Updated = time.Now()
	result := transaction.Table("job").Save(&updateJob)
	if result.Error != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": result.Error.Error()})
		return
	}

	for _, address := range reqData.Hosts {
		host := models.Host{}
		transaction.Table("host").Where("address = ?", address).First(&host)
		if host.Id == 0 || host.Status != "enabled" {
			var operationType string
			if host.Id == 0 {
				host.Address = address
				host.Status = "enabled"
				host.Created = time.Now()
				host.Updated = time.Now()
				result = transaction.Table("host").Create(&host)
				operationType = "create"
			} else {
				host.Status = "enabled"
				host.Updated = time.Now()
				result = transaction.Table("host").Save(&host)
				operationType = "update"
			}
			if result.Error != nil {
				transaction.Rollback()
				c.JSON(http.StatusInternalServerError,
					gin.H{"code": 500, "msg": fmt.Sprintf("%v", result.Error)})
				return
			}

			hostData := MakeHostKv(host)
			jsonHostData, _ := json.Marshal(hostData)
			operRecord := models.OperationRecord{ResourceType:"host", ResourceId:host.Id, ResourceLabel: host.Address,
				OperationType: operationType, Data:jsonHostData, User: user, Created: time.Now()}
			if result = db.DB.Table("operation_record").Create(&operRecord); result.Error != nil {
				transaction.Rollback()
				c.JSON(http.StatusInternalServerError,
					gin.H{"code": 500, "msg": result.Error.Error()})
				return
			}
		}
	}

	data := MakeJobKv(updateJob)
	jsonData, _ := json.Marshal(data)
	operRecord := models.OperationRecord{ResourceType:"job", ResourceId:updateJob.Id, ResourceLabel: updateJob.Name,
		OperationType: "update", Data:jsonData, User: user, Created: time.Now()}
	if result = db.DB.Table("operation_record").Create(&operRecord); result.Error != nil {
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
	db.DB.Table("job").Where("id = ?", pk).First(&deleteJob)
	if deleteJob.Id == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "Job not found"})
		return
	}

	transaction := db.DB.Begin()
	result := db.DB.Table("job").Delete(&deleteJob)
	if result.Error != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": result.Error.Error()})
		return
	}

	data := MakeJobKv(deleteJob)
	jsonData, _ := json.Marshal(data)
	operRecord := models.OperationRecord{ResourceType:"job", ResourceId:deleteJob.Id, ResourceLabel: deleteJob.Name,
		OperationType: "delete", Data:jsonData, User: user, Created: time.Now()}
	if result = db.DB.Table("operation_record").Create(&operRecord); result.Error != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError,
			gin.H{"code": 500, "msg": result.Error.Error()})
		return
	}

	transaction.Commit()
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data})
}

func cleanedJobData(data *JobReqData, updateJobId int64) map[string]string {
	error := make(map[string]string)
	nameReg, _ := regexp.Compile("^[a-z_]*$")
	if data.Name == "" {
		error["name"] = "任务名称是必填项"
	} else if !nameReg.MatchString(data.Name) {
		error["name"] = "任务名称格式不正确"
	} else {
		job := models.Job{}
		if updateJobId == 0 {
			db.DB.Table("job").Where("name = ?", data.Name).First(&job)
		} else {
			db.DB.Table("job").Where("name = ? AND id <> ?", data.Name, updateJobId).First(&job)
		}
		if job.Id != 0 {
			error["name"] = fmt.Sprintf("任务名称[%v]已存在", data.Name)
		}
	}

	if data.Status != "enabled" && data.Status != "disabled" {
		error["status"] = "任务状态无效"
	}

	if data.Description == "" {
		error["description"] = "任务描述是必填项"
	} else {
		job := models.Job{}
		if updateJobId == 0 {
			db.DB.Table("job").Where("description = ?", data.Description).First(&job)
		} else {
			db.DB.Table("job").Where("description = ? AND id <> ?",
				data.Description, updateJobId).First(&job)
		}
		if job.Id != 0 {
			error["description"] = fmt.Sprintf("任务描述[%v]已存在", data.Description)
		}
	}

	if data.Spec == "" {
		error["spec"] = "任务调度是必填项"
	}

	if data.Content == "" {
		error["content"] = "任务脚本内容是必填项"
	}

	var badHosts []string
	for _, host := range data.Hosts {
		if host == "" {
			badHosts = append(badHosts, host)
		}
	}
	if len(badHosts) > 0 {
		error["hosts"] = fmt.Sprintf("任务主机[%v]无效或无法连接", strings.Join(badHosts, ","))
	}

	if data.Sysuser == "" {
		error["sysuser"] = "任务所属的系统用户是必填项"
	}
	return error
}