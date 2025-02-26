package router

import (
	"github.com/gin-gonic/gin"
	"task/api/handler"
)

func RegisterRoutes(r *gin.Engine, taskHandler *handler.TaskHandler) {
	v1 := r.Group("/api/v1")
	{
		tasks := v1.Group("/tasks")
		{
			tasks.POST("", taskHandler.CreateTask)
		}
	}
}
