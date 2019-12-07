package handlers

import (
	"strconv"
	"cronnest/models"
	"net/http"
	"github.com/gin-gonic/gin"
	"fmt"
	db "cronnest/database"
)

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

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data, "total": count})
}


func GetHostJobs(c *gin.Context) {
	host := c.Query("host")
	var jobs []models.Job
	var count int
	query := fmt.Sprintf("hosts @> '[\"%v\"]'::jsonb", host)
	db.DB.Table("job").Where(query).Order("name").Count(&count).Find(&jobs)
	var data []map[string]interface{}
	data = []map[string]interface{} {}
	for _, job := range jobs {
		data = append(data, MakeJobKv(job))
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data, "total": count})
}