package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ilkerispir/terrakube-executor/internal/model"
)

type Config struct {
	Mode                    string
	EphemeralJobData        *model.TerraformJob
	TerrakubeApiUrl         string
	TerrakubeRegistryDomain string
	InternalSecret          string
	StorageType             string
	StorageAccountName      string
	StorageAccountKey       string
}

func getEnvWithFallback(primary, fallback string) string {
	val := os.Getenv(primary)
	if val == "" {
		return os.Getenv(fallback)
	}
	return val
}

func getStorageType() string {
	st := os.Getenv("STORAGE_TYPE")
	if st != "" {
		return st
	}
	tst := os.Getenv("TerraformStateType")
	switch tst {
	case "AwsTerraformStateImpl":
		return "AWS"
	case "AzureTerraformStateImpl":
		return "AZURE"
	case "GcpTerraformStateImpl":
		return "GCP"
	case "LocalTerraformStateImpl", "":
		return "LOCAL"
	}
	return "LOCAL"
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Mode:                    os.Getenv("EXECUTOR_MODE"),
		TerrakubeApiUrl:         getEnvWithFallback("TERRAKUBE_API_URL", "TerrakubeApiUrl"),
		TerrakubeRegistryDomain: getEnvWithFallback("TERRAKUBE_REGISTRY_DOMAIN", "TerrakubeRegistryDomain"),
		InternalSecret:          getEnvWithFallback("INTERNAL_SECRET", "InternalSecret"),
		StorageType:             getStorageType(),
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
