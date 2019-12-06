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
)

type JobReqData struct {
	Name 			string
	Status 			string
	Description 	string
	Mailto			string
	Spec			string
	Content			string
	Hosts			[]string
	Sysuser			string
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
		db.DB.Table("job").Count(&count).Limit(limit).Offset(offset).Where("name LIKE ?", search).
			Or("description LIKE ?", search).Order("name").Find(&jobs)
	} else {
		db.DB.Table("job").Count(&count).Limit(limit).Offset(offset).Order("name").Find(&jobs)
	}

	var data []map[string]interface{}
	data = []map[string]interface{} {}
	for _, job := range jobs {
		data = append(data, makeJobKv(job))
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data, "total": count})
}


func AddJob(c *gin.Context) {
	rawData, _ := c.GetRawData()
	data := JobReqData{}
	json.Unmarshal(rawData, &data)
	error := cleanedJobData(&data, 0)
	if len(error) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": error})
		return
	}

	transaction := db.DB.Begin()
	hosts, _ := json.Marshal(data.Hosts)
	job := models.Job{Name:data.Name, Status:data.Status, Description:data.Description, Mailto:data.Mailto,
		Spec:data.Spec, Content: data.Content, Hosts:hosts, Sysuser:data.Sysuser,
		Created:time.Now(), Updated:time.Now()}
	result := transaction.Table("job").Create(&job)
	if result.Error != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": result.Error})
		return
	}

	for _, name := range data.Hosts {
		host := models.Host{}
		transaction.Table("host").Where("name = ?", name).First(&host)
		if host.Id == 0 || host.Status != "enabled" {
			if host.Id == 0 {
				host.Name = name
				host.Status = "enabled"
				host.Created = time.Now()
				host.Updated = time.Now()
				result = transaction.Table("host").Create(&host)
			} else {
				host.Status = "enabled"
				host.Updated = time.Now()
				result = transaction.Table("host").Save(&host)
			}
			if result.Error != nil {
				transaction.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": result.Error})
				return
			}
		}
	}
	transaction.Commit()
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": makeJobKv(job)})
}

func UpdateJob(c *gin.Context) {
	pk := c.Param("pk")
	updateJob := models.Job{}
	db.DB.Table("job").Where("id = ?", pk).First(&updateJob)
	if updateJob.Id == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "任务不存在"})
		return
	}

	rawData, _ := c.GetRawData()
	data := JobReqData{}
	json.Unmarshal(rawData, &data)
	error := cleanedJobData(&data, updateJob.Id)
	if len(error) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": error})
		return
	}

	transaction := db.DB.Begin()
	hosts, _ := json.Marshal(data.Hosts)
	updateJob.Name = data.Name
	updateJob.Status = data.Status
	updateJob.Description = data.Description
	updateJob.Mailto = data.Mailto
	updateJob.Spec = data.Spec
	updateJob.Content = data.Content
	updateJob.Hosts = hosts
	updateJob.Sysuser = data.Sysuser
	updateJob.Updated = time.Now()
	result := transaction.Table("job").Save(&updateJob)
	if result.Error != nil {
		transaction.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": result.Error})
		return
	}

	for _, name := range data.Hosts {
		host := models.Host{}
		transaction.Table("host").Where("name = ?", name).First(&host)
		if host.Id == 0 || host.Status != "enabled" {
			if host.Id == 0 {
				host.Name = name
				host.Status = "enabled"
				host.Created = time.Now()
				host.Updated = time.Now()
				result = transaction.Table("host").Create(&host)
			} else {
				host.Status = "enabled"
				host.Updated = time.Now()
				result = transaction.Table("host").Save(&host)
			}
			if result.Error != nil {
				transaction.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": result.Error})
				return
			}
		}
	}

	transaction.Commit()
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": makeJobKv(updateJob)})
}


func cleanedJobData(data *JobReqData, updateJobId int64) map[string]string {
	error := make(map[string]string)
	fmt.Println(updateJobId)
	if data.Name == "" {
		error["name"] = "任务名称是必填项"
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
			db.DB.Table("job").Where("description = ? AND id <> ?", data.Description, updateJobId).First(&job)
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
		error["sysuser"] = "任务系统用户是必填项"
	}
	return error
}

func makeJobKv(job models.Job) map[string]interface{} {
	row := make(map[string]interface{})
	row["id"] = job.Id
	row["name"] = job.Name
	row["status"] = job.Status
	row["description"] = job.Description
	row["mailto"] = job.Mailto
	row["spec"] = job.Spec
	row["content"] = job.Content
	row["log"] = job.Log
	row["output"] = job.Output
	row["created"] = job.Created
	row["updated"] = job.Updated
	row["hosts"] = job.Hosts
	row["sysuser"] = job.Sysuser
	return row
}