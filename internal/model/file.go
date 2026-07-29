package model

import (
	"crypto/sha256"
	"fmt"
	"time"
)

type FileInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	ETag        string    `json:"etag"`
	CreatedAt   time.Time `json:"created_at"`
	Path        string    `json:"-"`
}

func (f *FileInfo) GenerateETag() {
	h := sha256.New()
	h.Write([]byte(f.ID))
	h.Write([]byte(f.Name))
	h.Write([]byte(fmt.Sprintf("%d", f.Size)))
	f.ETag = fmt.Sprintf("\"%x\"", h.Sum(nil))
}
