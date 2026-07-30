package server

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/king12-D/cligy/internal/cache"
	"github.com/king12-D/cligy/internal/handler"
	"github.com/king12-D/cligy/internal/storage"
)

type Config struct {
	Host  string
	Port  int
	Debug bool

	// local storage
	StoragePath string

	// s3 storage
	S3Bucket   string
	S3Prefix   string
	S3Endpoint string
	S3Region   string
}

func New(cfg Config) (*gin.Engine, error) {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	var store storage.Storage
	var err error

	if cfg.S3Bucket != "" {
		store, err = newS3Storage(cfg)
	} else {
		store, err = storage.NewLocalStorage(cfg.StoragePath)
	}
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

func newS3Storage(cfg Config) (*storage.S3Storage, error) {
	opts := []func(*config.LoadOptions) error{}
	if cfg.S3Region != "" {
		opts = append(opts, config.WithRegion(cfg.S3Region))
	}
	awsCfg, err := config.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	var s3Client *s3.Client
	if cfg.S3Endpoint != "" {
		s3Client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.S3Endpoint)
			o.UsePathStyle = true
		})
	} else {
		s3Client = s3.NewFromConfig(awsCfg)
	}

	return storage.NewS3Storage(s3Client, cfg.S3Bucket, cfg.S3Prefix), nil
}

func DefaultConfig() Config {
	return Config{
		Host:        "0.0.0.0",
		Port:        8080,
		StoragePath: "./data",
		Debug:       false,
	}
}
