package router

import (
	"cronnest/handlers"
	"github.com/gin-gonic/gin"
)


func InitRouter() *gin.Engine {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.GET("/html/:tpl/", handlers.Html)
	apiGroup := router.Group("/api")
	{
		apiGroup.GET("/jobs/", handlers.GetJobs)
		apiGroup.POST("/jobs/", handlers.AddJob)
		apiGroup.PUT("/jobs/:pk/", handlers.UpdateJob)
		apiGroup.DELETE("/jobs/:pk/", handlers.DeleteJob)
		apiGroup.GET("/hosts/", handlers.GetHosts)
		apiGroup.GET("/host_jobs/", handlers.GetHostJobs)
	}

	return router
}