package storage

import (
	"io"

	"github.com/king12-D/cligy/internal/model"
)

type Storage interface {
	Save(name string, contentType string, reader io.Reader) (*model.FileInfo, error)
	Get(id string) (io.ReadCloser, *model.FileInfo, error)
	Delete(id string) error
	List() ([]*model.FileInfo, error)
}
