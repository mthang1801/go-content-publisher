package sourcehttp

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.GET("/sources", handler.List)
	router.GET("/sources/report", handler.Report)
	router.POST("/sources", handler.Create)
	router.PATCH("/sources/:type/:handle", handler.UpdateMetadata)
	router.DELETE("/sources/:type/:handle", handler.Delete)
}
