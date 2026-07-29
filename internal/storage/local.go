package storage

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/king12-D/cligy/internal/model"
)

type LocalStorage struct {
	basePath string
	mu       sync.RWMutex
}

func NewLocalStorage(basePath string) (*LocalStorage, error) {
	abs, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}
	if err := os.MkdirAll(abs, 0755); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(abs, "files"), 0755); err != nil {
		return nil, fmt.Errorf("create files dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(abs, "meta"), 0755); err != nil {
		return nil, fmt.Errorf("create meta dir: %w", err)
	}
	return &LocalStorage{basePath: abs}, nil
}

func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func (s *LocalStorage) filePath(id string) string {
	return filepath.Join(s.basePath, "files", id)
}

func (s *LocalStorage) metaPath(id string) string {
	return filepath.Join(s.basePath, "meta", id+".json")
}

func (s *LocalStorage) Save(name string, contentType string, reader io.Reader) (*model.FileInfo, error) {
	id, err := generateID()
	if err != nil {
		return nil, err
	}

	fpath := s.filePath(id)
	f, err := os.Create(fpath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	size, err := io.Copy(f, reader)
	if err != nil {
		os.Remove(fpath)
		return nil, fmt.Errorf("write file: %w", err)
	}

	info := &model.FileInfo{
		ID:          id,
		Name:        name,
		Size:        size,
		ContentType: contentType,
		Path:        fpath,
	}
	info.GenerateETag()

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.writeMeta(info); err != nil {
		os.Remove(fpath)
		os.Remove(s.metaPath(id))
		return nil, err
	}

	return info, nil
}

func (s *LocalStorage) Get(id string) (io.ReadCloser, *model.FileInfo, error) {
	s.mu.RLock()
	info, err := s.readMeta(id)
	s.mu.RUnlock()
	if err != nil {
		return nil, nil, err
	}

	f, err := os.Open(s.filePath(id))
	if err != nil {
		return nil, nil, fmt.Errorf("open file: %w", err)
	}

	return f, info, nil
}

func (s *LocalStorage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.Remove(s.filePath(id)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete file: %w", err)
	}
	if err := os.Remove(s.metaPath(id)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete meta: %w", err)
	}
	return nil
}

func (s *LocalStorage) List() ([]*model.FileInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(filepath.Join(s.basePath, "meta"))
	if err != nil {
		return nil, fmt.Errorf("list meta: %w", err)
	}

	var files []*model.FileInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		id := e.Name()
		if len(id) < 5 || id[len(id)-5:] != ".json" {
			continue
		}
		id = id[:len(id)-5]
		info, err := s.readMeta(id)
		if err != nil {
			continue
		}
		files = append(files, info)
	}
	return files, nil
}

func (s *LocalStorage) writeMeta(info *model.FileInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("marshal meta: %w", err)
	}
	if err := os.WriteFile(s.metaPath(info.ID), data, 0644); err != nil {
		return fmt.Errorf("write meta: %w", err)
	}
	return nil
}

func (s *LocalStorage) readMeta(id string) (*model.FileInfo, error) {
	data, err := os.ReadFile(s.metaPath(id))
	if err != nil {
		return nil, fmt.Errorf("read meta: %w", err)
	}
	var info model.FileInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("unmarshal meta: %w", err)
	}
	info.Path = s.filePath(id)
	return &info, nil
}
