package sourcehttp

import (
	"errors"
	"net/http"
	"strings"

	"go-content-bot/internal/source/application"
	"go-content-bot/internal/source/domain"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

type createSourceRequest struct {
	Type   string   `json:"type" binding:"required"`
	Handle string   `json:"handle" binding:"required"`
	Name   string   `json:"name" binding:"required"`
	Tags   []string `json:"tags"`
	Topics []string `json:"topics"`
}

type updateSourceMetadataRequest struct {
	Tags   []string `json:"tags"`
	Topics []string `json:"topics"`
}

func (h *Handler) List(c *gin.Context) {
	sources, err := h.service.ListActive(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list sources"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": toSourceResponses(sources)})
}

func (h *Handler) Report(c *gin.Context) {
	sources, err := h.service.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list source report"})
		return
	}

	sourceType := strings.ToLower(strings.TrimSpace(c.Query("type")))
	if sourceType != "" {
		filtered := make([]domain.Source, 0, len(sources))
		for _, source := range sources {
			if string(source.Type) == sourceType {
				filtered = append(filtered, source)
			}
		}
		sources = filtered
	}

	c.JSON(http.StatusOK, gin.H{"items": toSourceResponses(sources)})
}

func (h *Handler) Create(c *gin.Context) {
	var request createSourceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	source, err := h.service.Create(c.Request.Context(), domain.Source{
		Type:   domain.Type(strings.ToLower(strings.TrimSpace(request.Type))),
		Handle: strings.TrimSpace(request.Handle),
		Name:   strings.TrimSpace(request.Name),
		Tags:   request.Tags,
		Topics: request.Topics,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidSource):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid source"})
		case errors.Is(err, domain.ErrSourceAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "source already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create source"})
		}
		return
	}

	c.JSON(http.StatusCreated, toSourceResponse(source))
}

func (h *Handler) UpdateMetadata(c *gin.Context) {
	var request updateSourceMetadataRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	sourceType := domain.Type(strings.ToLower(strings.TrimSpace(c.Param("type"))))
	handle := strings.TrimSpace(c.Param("handle"))
	if handle == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing handle"})
		return
	}

	if err := h.service.UpdateMetadataByHandle(c.Request.Context(), sourceType, handle, request.Tags, request.Topics); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidSource):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid source metadata"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update source metadata"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) Delete(c *gin.Context) {
	sourceType := domain.Type(strings.ToLower(strings.TrimSpace(c.Param("type"))))
	handle := strings.TrimSpace(c.Param("handle"))
	if handle == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing handle"})
		return
	}

	if err := h.service.DeleteByHandle(c.Request.Context(), sourceType, handle); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete source"})
		return
	}

	c.Status(http.StatusNoContent)
}
