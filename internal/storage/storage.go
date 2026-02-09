package storage

import (
	"io"
)

type StorageService interface {
	UploadFile(path string, content io.Reader) error
	DownloadFile(path string) (io.ReadCloser, error)
}
