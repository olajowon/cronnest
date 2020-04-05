package handlers

import (
	"cronnest/models"
	"net/http"
	"github.com/gin-gonic/gin"
	"fmt"
	db "cronnest/database"
)


func GetHostCrontab(c *gin.Context) {
	hId := c.Param("hId")
	data := map[string]interface{} {}
	host := models.Host{}
	db.DB.Table("host").Where(fmt.Sprintf("id=%v", hId)).Find(&host)
	if host.Id > 0 {
		data = UpdateHostCrontabRecord(host)
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}