package daemon

import (
	"context"
	"encoding/json"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/sirupsen/logrus"
	"task/internal/model"
	"task/internal/service"
)

type TaskConsumerJob struct {
	Message     pulsar.Message
	Consumer    pulsar.Consumer
	Logger      *logrus.Logger
	TaskService service.TaskService
}

func (tcj *TaskConsumerJob) Process(ctx context.Context) (interface{}, error) {
	var taskMessage model.TaskMessage
	if err := json.Unmarshal(tcj.Message.Payload(), &taskMessage); err != nil {
		tcj.Logger.WithError(err).Error("Failed to unmarshal task message")
		tcj.Consumer.Ack(tcj.Message)
		return nil, err
	}

	// 处理任务
	modelTask, err := tcj.TaskService.ExecuteTask(ctx, taskMessage)
	if err != nil {
		tcj.Logger.WithError(err).WithField("task", taskMessage)
		// 根据错误类型决定是否Ack或者Nack
		if IsRetryableError(err) {
			tcj.Consumer.Nack(tcj.Message)
		} else {
			tcj.Consumer.Ack(tcj.Message)
		}
		return nil, err
	}

	// 执行成功，确认消息
	tcj.Consumer.Ack(tcj.Message)
	return modelTask, nil
}

func IsRetryableError(err error) bool {
	// 默认不重试
	return false
}
