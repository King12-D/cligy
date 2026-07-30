package main

import (
	"fmt"
	"os"

	"github.com/king12-D/cligy/internal/server"
)

func main() {
	cfg := server.DefaultConfig()

	if port := os.Getenv("PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Port)
	}
	if path := os.Getenv("STORAGE_PATH"); path != "" {
		cfg.StoragePath = path
	}
	if bucket := os.Getenv("S3_BUCKET"); bucket != "" {
		cfg.S3Bucket = bucket
	}
	if prefix := os.Getenv("S3_PREFIX"); prefix != "" {
		cfg.S3Prefix = prefix
	}
	if endpoint := os.Getenv("S3_ENDPOINT"); endpoint != "" {
		cfg.S3Endpoint = endpoint
	}
	if region := os.Getenv("S3_REGION"); region != "" {
		cfg.S3Region = region
	}
	if os.Getenv("DEBUG") == "true" {
		cfg.Debug = true
	}

	r, err := server.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	if cfg.S3Bucket != "" {
		fmt.Printf("CDN listening on %s (storage: s3://%s", addr, cfg.S3Bucket)
		if cfg.S3Prefix != "" {
			fmt.Printf("/%s", cfg.S3Prefix)
		}
		fmt.Println(")")
	} else {
		fmt.Printf("CDN listening on %s (storage: %s)\n", addr, cfg.StoragePath)
	}
	if err := r.Run(addr); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
