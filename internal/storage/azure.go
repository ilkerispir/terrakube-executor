package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

type AzureStorageService struct {
	client        *azblob.Client
	containerName string
}

func NewAzureStorageService() (*AzureStorageService, error) {
	accountName := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	accountKey := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
	containerName := os.Getenv("AZURE_STORAGE_CONTAINER_NAME") // Usually "tfstate" ?

	if containerName == "" {
		containerName = "tfstate" // Default? Or maybe "terrakube"
	}

	log.Printf("Initializing Azure Storage Service")
	log.Printf("Account: %s", accountName)
	log.Printf("Container: %s", containerName)

	cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return nil, fmt.Errorf("invalid azure credentials: %v", err)
	}

	serviceClient, err := azblob.NewClientWithSharedKeyCredential(fmt.Sprintf("https://%s.blob.core.windows.net/", accountName), cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure client: %v", err)
	}

	return &AzureStorageService{
		client:        serviceClient,
		containerName: containerName,
	}, nil
}

func (s *AzureStorageService) UploadFile(path string, content io.Reader) error {
	_, err := s.client.UploadStream(context.TODO(), s.containerName, path, content, nil)
	if err != nil {
		return fmt.Errorf("failed to upload file to Azure: %w", err)
	}
	return nil
}

func (s *AzureStorageService) DownloadFile(path string) (io.ReadCloser, error) {
	resp, err := s.client.DownloadStream(context.TODO(), s.containerName, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from Azure: %w", err)
	}
	return resp.Body, nil
}
