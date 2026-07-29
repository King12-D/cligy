package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	info, err := h.storage.Save(header.Filename, contentType, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":           info.ID,
		"name":         info.Name,
		"size":         info.Size,
		"content_type": info.ContentType,
		"etag":         info.ETag,
		"created_at":   info.CreatedAt,
		"url":          c.Request.Host + "/raw/" + info.ID,
	})
}
