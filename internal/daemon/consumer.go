package daemon

import (
	"task/config"
	"task/internal/service"
	"task/pkg/worker_pool"
	"time"

	"context"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/sirupsen/logrus"
)

type TaskConsumer struct {
	consumer pulsar.Consumer
	config   *config.PulsarConfig
	logger   *logrus.Logger
	// 消费者协程池，控制消费并发
	workerPool  *worker_pool.WorkerPool
	taskService service.TaskService
}

func NewTaskConsumer(
	consumer pulsar.Consumer,
	config *config.PulsarConfig,
	logger *logrus.Logger,
	taskService service.TaskService,
) (*TaskConsumer, error) {
	wp := worker_pool.NewWorkerPool(20, 200, 40*time.Second)
	tc := &TaskConsumer{
		consumer:    consumer,
		config:      config,
		logger:      logger,
		workerPool:  wp,
		taskService: taskService,
	}

	// 设置任务处理结果回调函数
	wp.SetCallback(tc.handleTaskResult)
	return tc, nil
}

func (tc *TaskConsumer) Start(ctx context.Context) {
	// 启动协程池
	tc.workerPool.Start(ctx)
	// 启动消息接收循环
	go tc.consumeMessages(ctx)
}

func (tc *TaskConsumer) Stop(ctx context.Context) {
	// 停止协程池
	tc.workerPool.Stop()
	// 关闭消费者
	tc.consumer.Close()
}

func (tc *TaskConsumer) consumeMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := tc.consumer.Receive(ctx)
			if err != nil {
				tc.logger.WithError(err).Error("Failed to consume pulsar message")
				// 避免在错误的情况忙等
				time.Sleep(1 * time.Second)
				continue
			}
			// 处理消息, 用协程池消费
			job := &TaskConsumerJob{
				Message:     msg,
				Consumer:    tc.consumer,
				Logger:      tc.logger,
				TaskService: tc.taskService,
			}

			if !tc.workerPool.Submit(job) {
				tc.logger.Error("Failed to submit consumer task to consumer worker pool")
				tc.consumer.Nack(msg)
			}
		}
	}
}

func (tc *TaskConsumer) processMessage(ctx context.Context, msg pulsar.Message) {
	// TODO 处理消息

	// 确认消息
	tc.consumer.Ack(msg)
}

func (tc *TaskConsumer) handleTaskResult(ctx context.Context, result worker_pool.Result) error {
	if result.Err != nil {
		tc.logger.WithError(result.Err).Error("Failed to handle task consume result")
		// TODO 根据错误类型决定是否重试以及记录更详细的日志
		return result.Err
	}
	return nil
}
