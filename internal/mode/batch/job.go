package batch

import (
	"log"

	"github.com/terrakube-io/terrakube/executor-go/internal/core"
	"github.com/terrakube-io/terrakube/executor-go/internal/model"
)

func AdjustAndExecute(job *model.TerraformJob, processor *core.JobProcessor) {
	log.Printf("Starting Batch Execution for Job %s", job.JobId)
	if err := processor.ProcessJob(job); err != nil {
		log.Fatalf("Job execution failed: %v", err)
	}
	log.Println("Batch execution finished")
}
