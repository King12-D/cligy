package handler

import (
	"github.com/king12-D/cligy/internal/cache"
	"github.com/king12-D/cligy/internal/storage"
)

type Handler struct {
	storage storage.Storage
	cache   cache.Cache
}

func New(storage storage.Storage, cache cache.Cache) *Handler {
	return &Handler{storage: storage, cache: cache}
}
