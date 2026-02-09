package online

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ilkerispir/terrakube-executor/internal/core"
	"github.com/ilkerispir/terrakube-executor/internal/model"
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
