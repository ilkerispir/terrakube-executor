package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/terrakube-io/terrakube/executor-go/internal/model"
)

type Config struct {
	Mode               string
	EphemeralJobData   *model.TerraformJob
	TerrakubeApiUrl    string
	StorageType        string
	StorageAccountName string
	StorageAccountKey  string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Mode:            os.Getenv("EXECUTOR_MODE"),
		TerrakubeApiUrl: os.Getenv("TERRAKUBE_API_URL"),
		StorageType:     os.Getenv("STORAGE_TYPE"),
	}

	if cfg.Mode == "BATCH" {
		jobData := os.Getenv("EPHEMERAL_JOB_DATA")
		if jobData == "" {
			return nil, fmt.Errorf("EXECUTOR_MODE is BATCH but EPHEMERAL_JOB_DATA is empty")
		}

		decodedData, err := base64.StdEncoding.DecodeString(jobData)
		if err != nil {
			return nil, fmt.Errorf("failed to decode EPHEMERAL_JOB_DATA: %v", err)
		}

		var job model.TerraformJob
		if err := json.Unmarshal(decodedData, &job); err != nil {
			return nil, fmt.Errorf("failed to unmarshal EPHEMERAL_JOB_DATA: %v", err)
		}
		cfg.EphemeralJobData = &job
	}

	return cfg, nil
}
