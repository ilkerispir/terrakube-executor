package online

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/terrakube-io/terrakube/executor-go/internal/core"
	"github.com/terrakube-io/terrakube/executor-go/internal/model"
)

func StartServer(port string, processor *core.JobProcessor) {
	r := gin.Default()

	r.POST("/api/v1/terraform-rs", func(c *gin.Context) {
		var job model.TerraformJob
		if err := c.ShouldBindJSON(&job); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Process job asynchronously
		// In a real implementation this should be offloaded to a worker pool
		go func() {
			processor.ProcessJob(&job)
		}()

		c.JSON(http.StatusAccepted, job)
	})

	r.Run(":" + port)
}
