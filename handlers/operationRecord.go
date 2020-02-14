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
	var count int
	if search != "" {
		fmt.Println("se", string(search))
		search = fmt.Sprintf("%%%v%%", search)
		fmt.Println(search)
		db.DB.Table("operation_records").Where( "resource_label LIKE ?", search).Count(&count).
			Limit(limit).Offset(offset).Order("-id").Find(&operationRecords)
	} else {
		fmt.Println("se1", search)
		res := db.DB.Table("operation_records").Count(&count).
			Limit(limit).Offset(offset).Order("-id").Find(&operationRecords)
		fmt.Println(res.Value)
	}

	var data []map[string]interface{}
	data = []map[string]interface{} {}
	for _, record := range operationRecords {
		data = append(data, MakeOperationRecordKv(record))
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": data, "total": count})
}