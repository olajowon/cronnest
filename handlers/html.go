package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Html(c *gin.Context){
	tpl := c.Param("tpl")
	var html string
	if tpl == "jobs" {
		html = "job.html"
	} else if tpl == "hosts" {
		html = "host.html"
	} else if tpl == "records" {
		html = "record.html"
	} else {
		c.JSON(http.StatusOK, gin.H{"msg": "滚犊子！"})
	}
	c.HTML(http.StatusOK, html, gin.H{})
}