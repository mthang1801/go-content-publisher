package contenthttp

import (
	"net/http"
	"strconv"

	"go-content-bot/internal/content/application"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

type createContentRequest struct {
	Text   string `json:"text" binding:"required"`
	Author string `json:"author"`
}

type manualRewriteRequest struct {
	RewrittenText string `json:"rewritten_text" binding:"required"`
}

func (h *Handler) Create(c *gin.Context) {
	var request createContentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	item, err := h.service.EnqueueManual(c.Request.Context(), application.EnqueueManualInput{
		Text:   request.Text,
		Author: request.Author,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create content item"})
		return
	}

	c.JSON(http.StatusCreated, toContentItemResponse(item))
}

func (h *Handler) ManualRewrite(c *gin.Context) {
	var request manualRewriteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	item, err := h.service.SetManualRewrite(c.Request.Context(), c.Param("id"), request.RewrittenText)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to apply manual rewrite"})
		return
	}

	c.JSON(http.StatusOK, toContentItemResponse(item))
}

func (h *Handler) Queue(c *gin.Context) {
	items, err := h.service.ListQueue(c.Request.Context(), parseLimit(c, 50))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list queue"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": toContentItemResponses(items)})
}

func (h *Handler) Recent(c *gin.Context) {
	items, err := h.service.ListRecent(c.Request.Context(), parseLimit(c, 20))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list recent content"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": toContentItemResponses(items)})
}

func parseLimit(c *gin.Context, fallback int) int {
	raw := c.Query("limit")
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
