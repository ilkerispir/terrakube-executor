package core

import (
	"fmt"
	"log"
	"os"

	"github.com/terrakube-io/terrakube/executor-go/internal/config"
	"github.com/terrakube-io/terrakube/executor-go/internal/logs"
	"github.com/terrakube-io/terrakube/executor-go/internal/model"
	"github.com/terrakube-io/terrakube/executor-go/internal/script"
	"github.com/terrakube-io/terrakube/executor-go/internal/status"
	"github.com/terrakube-io/terrakube/executor-go/internal/storage"
	"github.com/terrakube-io/terrakube/executor-go/internal/terraform"
	"github.com/terrakube-io/terrakube/executor-go/internal/workspace"
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
