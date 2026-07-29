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
	if os.Getenv("DEBUG") == "true" {
		cfg.Debug = true
	}

	r, err := server.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	fmt.Printf("CDN listening on %s\n", addr)
	if err := r.Run(addr); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
