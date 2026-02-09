package storage

import (
	"fmt"
	"io"
)

// Factory to create storage service
func NewStorageService(storageType string) (StorageService, error) {
	switch storageType {
	case "AWS":
		return NewAWSStorageService()
	case "AZURE":
		return NewAzureStorageService()
	case "GCP":
		return NewGCPStorageService()
	default:
		return nil, fmt.Errorf("unknown storage type: %s", storageType)
	}
}

// Mock implementation for LOCAL (if needed) or when storage is Nop
type NopStorageService struct{}

func (s *NopStorageService) UploadFile(path string, content io.Reader) error { return nil }
func (s *NopStorageService) DownloadFile(path string) (io.ReadCloser, error) { return nil, nil }
