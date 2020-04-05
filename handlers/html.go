package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Html(c *gin.Context){
	tpl := c.Param("tpl")
	user, _, _ := c.Request.BasicAuth()
	var html string
	if tpl == "hosts" {
		html = "host.html"
	} else if tpl == "operation_records" {
		html = "operation_record.html"
	} else if tpl == "crontab" {
		html = "crontab.html"
	} else {
		c.JSON(http.StatusOK, gin.H{"msg": "页面不存在！"})
	}
	c.HTML(http.StatusOK, html, gin.H{"user": user})
}