package handlers

import (
	"strconv"
	"github.com/gin-gonic/gin"
	"cronnest/models"
	"net/http"
	"fmt"
	db "cronnest/database"
)


func GetOperationRecords(c *gin.Context) {
	search := c.DefaultQuery("search", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	limit := pageSize
	offset := (page - 1) * pageSize

	var operationRecords []models.OperationRecord
	var count int64
	if search != "" {
		search = fmt.Sprintf("%%%v%%", search)
		db.DB.Table("operation_records").Where("resource_label LIKE ?", search).Count(&count)
		db.DB.Table("operation_records").Where("resource_label LIKE ?", search).
			Limit(limit).Offset(offset).Order("-id").Find(&operationRecords)
	} else {
		db.DB.Table("operation_records").Count(&count)
		db.DB.Table("operation_records").
			Limit(limit).Offset(offset).Order("-id").Find(&operationRecords)
	}

	data := []map[string]interface{}{}
	for _, record := range operationRecords {
		data = append(data, MakeOperationRecordKv(record))
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data, "total": count})
}
