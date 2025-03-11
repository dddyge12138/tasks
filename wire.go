//go:build wireinject
// +build wireinject

package main

import (
	"task/api/handler"
	"task/internal/repository"
	"task/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet 定义provider的集合
var taskProviderSet = wire.NewSet(
	repository.NewTaskRepository,
	service.NewTaskService,
	handler.NewTaskHandler,
)

// InitTaskHandler 初始化TaskHandler
func InitTaskHandler(db *gorm.DB) (*handler.TaskHandler, error) {
	wire.Build(taskProviderSet)
	return nil, nil
}
