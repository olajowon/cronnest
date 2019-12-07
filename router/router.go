package router

import (
	"cronnest/handlers"
	"github.com/gin-gonic/gin"
	"cronnest/configure"
)


func InitRouter() *gin.Engine {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	authorized := router.Group("/", gin.BasicAuth(configure.Accounts))

	authorized.GET("/html/:tpl/", handlers.Html)

	apiGroup := authorized.Group("/api")
	{
		apiGroup.GET("/jobs/", handlers.GetJobs)
		apiGroup.POST("/jobs/", handlers.CreateJob)
		apiGroup.PUT("/jobs/:pk/", handlers.UpdateJob)
		apiGroup.DELETE("/jobs/:pk/", handlers.DeleteJob)
		apiGroup.GET("/hosts/", handlers.GetHosts)
		apiGroup.GET("/host_jobs/", handlers.GetHostJobs)
	}

	return router
}