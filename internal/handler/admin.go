package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) List(c *gin.Context) {
	files, err := h.storage.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list files"})
		return
	}
	if files == nil {
		files = []*struct{}{}
	}
	c.JSON(http.StatusOK, gin.H{"files": files})
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id required"})
		return
	}

	if err := h.storage.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete file"})
		return
	}

	h.cache.Delete("meta:" + id)

	c.Status(http.StatusNoContent)
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
