package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/king12-D/cligy/internal/cache"
	"github.com/king12-D/cligy/internal/handler"
	"github.com/king12-D/cligy/internal/storage"
)

type Config struct {
	Host        string
	Port        int
	StoragePath string
	CacheSize   int
	Debug       bool
}

func New(cfg Config) (*gin.Engine, error) {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	store, err := storage.NewLocalStorage(cfg.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("init storage: %w", err)
	}

	ch := cache.NewMemoryCache()
	h := handler.New(store, ch)

	r := gin.Default()

	r.GET("/health", h.Health)

	api := r.Group("/api")
	{
		api.POST("/upload", h.Upload)
		api.GET("/files", h.List)
		api.GET("/files/:id", h.Serve)
		api.GET("/files/:id/download", h.Download)
		api.DELETE("/files/:id", h.Delete)
	}

	r.GET("/raw/:id", h.Serve)

	return r, nil
}

func DefaultConfig() Config {
	return Config{
		Host:        "0.0.0.0",
		Port:        8080,
		StoragePath: "./data",
		Debug:       false,
	}
}
