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
		db.DB.Table("host").Count(&count).Limit(limit).Offset(offset).Where("name LIKE ?", search).
			Order("name").Find(&hosts)
	} else {
		db.DB.Table("host").Count(&count).Limit(limit).Offset(offset).Order("name").Find(&hosts)
	}

	var data []map[string]interface{}
	data = []map[string]interface{} {}
	for _, host := range hosts {
		row := make(map[string]interface{})
		row["id"] = host.Id
		row["name"] = host.Name
		row["status"] = host.Status
		row["created"] = host.Created
		row["updated"] = host.Updated
		data = append(data, row)
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data, "total": count})
}