package router

import (
	"cronnest/handlers"
	"github.com/gin-gonic/gin"
	"cronnest/configure"
	"os"
	"io"
)


func InitRouter() *gin.Engine {
	f, _ := os.Create(configure.Log["request"])
	gin.DefaultWriter = io.MultiWriter(f)


	router := gin.Default()

	router.LoadHTMLGlob("templates/*")

	router.Static("/static", "./static")

	authorized := router.Group("/", gin.BasicAuth(configure.Accounts))

	authorized.GET("/html/:tpl/", handlers.Html)

	apiGroup := authorized.Group("/api")
	{
		apiGroup.GET("/hostgroups/:hgId/hosts/", handlers.GetHostgroupHosts)
		apiGroup.POST("/hostgroups/:hgId/hosts/", handlers.AddHostgroupHosts)
		apiGroup.DELETE("/hostgroups/:hgId/hosts/", handlers.RemoveHostgroupHosts)

		apiGroup.GET("/hostgroups/", handlers.GetHostgroup)
		apiGroup.POST("/hostgroups/", handlers.CreateHostgroup)
		apiGroup.PUT("/hostgroups/:hgId/", handlers.UpdateHostgroup)
		apiGroup.DELETE("/hostgroups/:hgId/", handlers.DeleteHostgroup)


		apiGroup.GET("/hosts/:hId/crontab/", handlers.GetHostCrontab)
		apiGroup.POST("/hosts/:hId/crontab/job/", handlers.CreateHostCrontabJob)
		apiGroup.PUT("/hosts/:hId/crontab/job/", handlers.UpdateHostCrontabJob)
		apiGroup.DELETE("/hosts/:hId/crontab/job/", handlers.DeleteHostCrontabJob)

		apiGroup.GET("/hosts/", handlers.GetHosts)
		apiGroup.PUT("/hosts/:hId/", handlers.UpdateHost)
		apiGroup.DELETE("/hosts/:hId/", handlers.DeleteHost)


		apiGroup.GET("/operation_records/", handlers.GetOperationRecords)
	}

	return router
}