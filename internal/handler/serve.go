package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Serve(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id required"})
		return
	}

	// check cache
	etag := c.GetHeader("If-None-Match")

	cached, ok := h.cache.Get("meta:" + id)
	if ok && string(cached) == etag {
		c.Status(http.StatusNotModified)
		return
	}

	reader, info, err := h.storage.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	defer reader.Close()

	// cache the etag
	h.cache.Set("meta:"+id, []byte(info.ETag))

	c.Header("ETag", info.ETag)
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("Content-Type", info.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", info.Size))
	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, filepath.Base(info.Name)))

	c.Status(http.StatusOK)
	io.Copy(c.Writer, reader)
}

func (h *Handler) Download(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id required"})
		return
	}

	reader, info, err := h.storage.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	defer reader.Close()

	c.Header("Content-Type", info.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", info.Size))
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, info.Name))

	io.Copy(c.Writer, reader)
}

func (h *Handler) ServeDir(c *gin.Context) {
	dir := c.Param("dirpath")
	if dir == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path required"})
		return
	}

	sanitized := filepath.Clean("/" + dir)
	fpath := filepath.Join(".", "www", sanitized)

	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.File(fpath)
}
