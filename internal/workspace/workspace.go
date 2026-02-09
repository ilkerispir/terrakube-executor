package workspace

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ilkerispir/terrakube-executor/internal/model"
)

type Workspace struct {
	Job        *model.TerraformJob
	WorkingDir string
}

func NewWorkspace(job *model.TerraformJob) *Workspace {
	return &Workspace{
		Job: job,
	}
}

func (w *Workspace) Setup() (string, error) {
	// Create temp directory for workspace
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("terrakube-job-%s", w.Job.JobId))
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	w.WorkingDir = tempDir

	// Clone repository
	repoURL := w.Job.Source
	if w.Job.AccessToken != "" && w.Job.VcsType != "PUBLIC" {
		if strings.HasPrefix(w.Job.Source, "https://") {
			repoURL = strings.Replace(w.Job.Source, "https://", fmt.Sprintf("https://oauth2:%s@", w.Job.AccessToken), 1)
		}
	}

	cmdArgs := []string{"clone", "--depth", "1"}
	if w.Job.Branch != "" {
		cmdArgs = append(cmdArgs, "--branch", w.Job.Branch)
	}
	cmdArgs = append(cmdArgs, repoURL, tempDir)

	cloneCmd := exec.Command("git", cmdArgs...)
	cloneCmd.Env = os.Environ()
	// TODO: Add SSH key support if VcsType is SSH

	if output, err := cloneCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git clone failed: %s: %w", string(output), err)
	}

	// Calculate final working directory (if folder is specified)
	finalDir := tempDir
	if w.Job.Folder != "" {
		finalDir = fmt.Sprintf("%s/%s", tempDir, w.Job.Folder)
	}

	return finalDir, nil
}

func (w *Workspace) Cleanup() error {
	if w.WorkingDir != "" {
		return os.RemoveAll(w.WorkingDir)
	}
	return nil
}
