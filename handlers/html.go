package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"fmt"
)

func Html(c *gin.Context){
	tpl := c.Param("tpl")
	user, _, _ := c.Request.BasicAuth()
	fmt.Println(user)
	var html string
	if tpl == "jobs" {
		html = "job.html"
	} else if tpl == "hosts" {
		html = "host.html"
	} else if tpl == "operation_records" {
		html = "operation_record.html"
	} else {
		c.JSON(http.StatusOK, gin.H{"msg": "滚犊子！"})
	}
	c.HTML(http.StatusOK, html, gin.H{"user": user})
}