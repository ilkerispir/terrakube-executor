package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type GCPStorageService struct {
	client     *storage.Client
	bucketName string
}

func NewGCPStorageService() (*GCPStorageService, error) {
	bucketName := os.Getenv("GCP_STORAGE_BUCKET")
	credentials := os.Getenv("GCP_SERVICE_ACCOUNT_KEY") // Or handle via ADC

	log.Printf("Initializing GCP Storage Service")
	log.Printf("Bucket: %s", bucketName)

	opts := []option.ClientOption{}
	if credentials != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(credentials)))
	}

	client, err := storage.NewClient(context.TODO(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP client: %v", err)
	}

	return &GCPStorageService{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (s *GCPStorageService) UploadFile(path string, content io.Reader) error {
	w := s.client.Bucket(s.bucketName).Object(path).NewWriter(context.TODO())
	if _, err := io.Copy(w, content); err != nil {
		w.Close()
		return fmt.Errorf("failed to upload file to GCP: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close GCP writer: %w", err)
	}
	return nil
}

func (s *GCPStorageService) DownloadFile(path string) (io.ReadCloser, error) {
	r, err := s.client.Bucket(s.bucketName).Object(path).NewReader(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to download file from GCP: %w", err)
	}
	return r, nil
}
