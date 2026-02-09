package status

import (
	"fmt"

	"github.com/ilkerispir/terrakube-executor/internal/client"
	"github.com/ilkerispir/terrakube-executor/internal/config"
	"github.com/ilkerispir/terrakube-executor/internal/model"
)

type StatusService interface {
	SetRunning(job *model.TerraformJob) error
	SetCompleted(job *model.TerraformJob, success bool, output string) error
}

type Service struct {
	client *client.TerrakubeClient
}

func NewStatusService(cfg *config.Config) *Service {
	return &Service{
		client: client.NewTerrakubeClient(cfg.TerrakubeApiUrl),
	}
}

func (s *Service) SetRunning(job *model.TerraformJob) error {
	return s.client.UpdateJobStatus(job.OrganizationId, job.JobId, "running", "")
}

func (s *Service) SetCompleted(job *model.TerraformJob, success bool, output string) error {
	status := "completed"
	if !success {
		status = "failed"
	}
	if err := s.client.UpdateStepStatus(job.OrganizationId, job.JobId, job.StepId, status, output); err != nil {
		return fmt.Errorf("failed to update step status: %w", err)
	}
	return s.client.UpdateJobStatus(job.OrganizationId, job.JobId, status, "")
}
