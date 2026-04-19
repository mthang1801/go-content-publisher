package contenthttp

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	router.POST("/content", handler.Create)
	router.POST("/content/:id/manual-rewrite", handler.ManualRewrite)
	router.GET("/content/queue", handler.Queue)
	router.GET("/content/recent", handler.Recent)
}
