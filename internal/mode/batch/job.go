package batch

import (
	"log"

	"github.com/ilkerispir/terrakube-executor/internal/core"
	"github.com/ilkerispir/terrakube-executor/internal/model"
)

func AdjustAndExecute(job *model.TerraformJob, processor *core.JobProcessor) {
	log.Printf("Starting Batch Execution for Job %s", job.JobId)
	if err := processor.ProcessJob(job); err != nil {
		log.Fatalf("Job execution failed: %v", err)
	}
	log.Println("Batch execution finished")
}
