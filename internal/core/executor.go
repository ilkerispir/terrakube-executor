package core

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/ilkerispir/terrakube-executor/internal/config"
	"github.com/ilkerispir/terrakube-executor/internal/logs"
	"github.com/ilkerispir/terrakube-executor/internal/model"
	"github.com/ilkerispir/terrakube-executor/internal/script"
	"github.com/ilkerispir/terrakube-executor/internal/status"
	"github.com/ilkerispir/terrakube-executor/internal/storage"
	"github.com/ilkerispir/terrakube-executor/internal/terraform"
	"github.com/ilkerispir/terrakube-executor/internal/workspace"
)

type JobProcessor struct {
	Status         status.StatusService
	Config         *config.Config
	Storage        storage.StorageService
	VersionManager *terraform.VersionManager
}

func NewJobProcessor(cfg *config.Config, status status.StatusService, storage storage.StorageService) *JobProcessor {
	return &JobProcessor{
		Config:         cfg,
		Status:         status,
		Storage:        storage,
		VersionManager: terraform.NewVersionManager(),
	}
}

func (p *JobProcessor) generateTerraformCredentials(job *model.TerraformJob, workingDir string) error {
	if job.AccessToken == "" {
		return nil
	}

	credentialsContent := ""

	if p.Config.TerrakubeRegistryDomain != "" {
		credentialsContent += fmt.Sprintf(`credentials "%s" {
  token = "%s"
}
`, p.Config.TerrakubeRegistryDomain, job.AccessToken)
	}

	if p.Config.TerrakubeApiUrl != "" {
		parsedUrl, err := url.Parse(p.Config.TerrakubeApiUrl)
		if err == nil && parsedUrl.Hostname() != "" {
			// Do not duplicate if same domain
			if parsedUrl.Hostname() != p.Config.TerrakubeRegistryDomain {
				credentialsContent += fmt.Sprintf(`credentials "%s" {
  token = "%s"
}
`, parsedUrl.Hostname(), job.AccessToken)
			}
		}
	}

	if credentialsContent == "" {
		return nil
	}

	rcPath := filepath.Join(workingDir, ".terraformrc")
	return os.WriteFile(rcPath, []byte(credentialsContent), 0644)
}

func (p *JobProcessor) generateBackendOverride(job *model.TerraformJob, workingDir string) error {
	log.Printf("generateBackendOverride checking API URL: TerrakubeApiUrl=%s", p.Config.TerrakubeApiUrl)
	if p.Config.TerrakubeApiUrl == "" {
		return nil
	}

	parsedUrl, err := url.Parse(p.Config.TerrakubeApiUrl)
	if err != nil {
		return fmt.Errorf("invalid TerrakubeApiUrl: %v", err)
	}
	hostname := parsedUrl.Hostname()

	overrideContent := fmt.Sprintf(`terraform {
  backend "remote" {
    hostname     = "%s"
    organization = "%s"
    workspaces {
      name = "%s"
    }
  }
}
`, hostname, job.OrganizationId, job.WorkspaceId)

	overridePath := filepath.Join(workingDir, "terrakube_override.tf")
	return os.WriteFile(overridePath, []byte(overrideContent), 0644)
}

func (p *JobProcessor) ProcessJob(job *model.TerraformJob) error {
	log.Printf("Processing Job: %s", job.JobId)

	// 1. Update Status to Running
	if err := p.Status.SetRunning(job); err != nil {
		log.Printf("Failed to set running status: %v", err)
	}

	// 2. Setup Logging
	var streamer logs.LogStreamer
	if os.Getenv("USE_REDIS_LOGS") == "true" {
		streamer = logs.NewRedisStreamer(os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PASSWORD"), job.JobId, job.StepId)
	} else {
		streamer = &logs.ConsoleStreamer{}
	}
	defer streamer.Close()

	// 3. Setup Workspace
	ws := workspace.NewWorkspace(job)
	workingDir, err := ws.Setup()
	if err != nil {
		p.Status.SetCompleted(job, false, err.Error())
		return fmt.Errorf("failed to setup workspace: %w", err)
	}
	defer ws.Cleanup()

	// 4. Download Pre-existing State/Plan if needed
	// TODO: If APPLY, download PLAN
	// TODO: If PLAN/APPLY/DESTROY, download STATE (if not using remote backend)

	// 5. Execute Command
	var executionErr error
	switch job.Type {
	case "terraformPlan", "terraformApply", "terraformDestroy":
		// Install/Get execution path for the specific version
		execPath, err := p.VersionManager.Install(job.TerraformVersion)
		if err != nil {
			executionErr = fmt.Errorf("failed to install terraform %s: %w", job.TerraformVersion, err)
			break
		}

		if err := p.generateBackendOverride(job, workingDir); err != nil {
			executionErr = fmt.Errorf("failed to generate backend override: %w", err)
			break
		}

		if err := p.generateTerraformCredentials(job, workingDir); err != nil {
			executionErr = fmt.Errorf("failed to generate terraform credentials: %w", err)
			break
		}

		if job.EnvironmentVariables == nil {
			job.EnvironmentVariables = make(map[string]string)
		}
		job.EnvironmentVariables["TF_CLI_CONFIG_FILE"] = filepath.Join(workingDir, ".terraformrc")

		tfExecutor := terraform.NewExecutor(job, workingDir, streamer, execPath)
		executionErr = tfExecutor.Execute()

		// Upload State and Output
		if executionErr == nil {
			p.uploadStateAndOutput(job, workingDir)
		}

	case "customScripts", "approval":
		scriptExecutor := script.NewExecutor(job, workingDir, streamer)
		executionErr = scriptExecutor.Execute()
	default:
		executionErr = fmt.Errorf("unknown job type: %s", job.Type)
	}

	// 6. Update Status to Completed/Failed
	success := executionErr == nil
	output := ""
	if executionErr != nil {
		output = executionErr.Error()
	}

	if err := p.Status.SetCompleted(job, success, output); err != nil {
		log.Printf("Failed to set completed status: %v", err)
	}

	return executionErr
}
