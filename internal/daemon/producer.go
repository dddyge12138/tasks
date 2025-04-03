// 需要在后台持续运行的协程
package daemon

import (
	"context"
	"fmt"
	"task/internal/model"
	"task/internal/repository"
	"task/pkg/logger"
	"task/pkg/pulsar_queue"
	redis_db "task/pkg/redis"
	"task/pkg/worker_pool"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"

	"github.com/sirupsen/logrus"
)

type TaskProducer struct {
	taskRepo   repository.TaskRepository
	workerPool *worker_pool.WorkerPool
	logger     *logrus.Logger
	producer   pulsar.Producer
	done       chan struct{}
}

func NewTaskProducer(
	taskRepo repository.TaskRepository,
	logger *logrus.Logger,
) (*TaskProducer, error) {
	wp := worker_pool.NewWorkerPool(10, 100, 10*time.Second)

	producer, err := pulsar_queue.NewProducer(pulsar.ProducerOptions{
		Topic: "tasks",
	})
	if err != nil {
		logger.WithError(err).Fatal("连接Pulsar消息队列失败")
		return nil, err
	}
	tp := &TaskProducer{
		taskRepo:   taskRepo,
		workerPool: wp,
		logger:     logger,
		producer:   producer,
		done:       make(chan struct{}),
	}

	// 设置回调函数
	wp.SetCallback(func(ctx context.Context, result worker_pool.Result) error {
		logger.WithFields(logrus.Fields{
			"result": result,
		}).Info("生产者结果监听接收到结果")
		if result.Err != nil {
			// 这里可以处理错误
			logger.WithError(result.Err).Error("Failed to handle task produce result")
			return result.Err
		}

		task, ok := result.Value.(*model.Task)
		if !ok {
			logger.Errorf("expected *mode.Task but got %T", result.Value)
			return fmt.Errorf("expected *mode.Task but got %T", result.Value)
		}

		// 更新任务状态
		if err := taskRepo.UpdateTaskAfterProduce(ctx, task); err != nil {
			logger.WithFields(logrus.Fields{
				"task_id": task.TaskId,
				"err":     err,
			}).Error("Fail to update task after produce")
			return err
		}
		// 删除redis槽内的任务以及任务详情。保留redis任务的版本号供下游消费者查询是否执行
		if err := redis_db.RemoveSlotTasks(ctx, task.TaskId); err != nil {
			logger.WithFields(logrus.Fields{
				"task_id": task.TaskId,
				"err":     err,
			}).Error("Fail to remove task from redis")
			return err
		}

		return nil
	})
	return tp, nil
}

func (tp *TaskProducer) Start(ctx context.Context) {
	// 启动协程池
	tp.workerPool.Start(ctx)
	go tp.TaskProducerRun()
}

func (tp *TaskProducer) Stop(ctx context.Context) {
	tp.workerPool.Stop()
	tp.done <- struct{}{}
}

func (tp *TaskProducer) TaskProducerRun() {
	var tasks []*model.Task
	var err error
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-tp.done:
			return
		case now := <-ticker.C:
			jobTimeoutCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			tasks, err = redis_db.GetTasksList(jobTimeoutCtx, now.Add(-20*time.Minute).Unix(), now.Add(30*time.Second).Unix())
			if err != nil {
				logger.Logger.WithFields(logrus.Fields{
					"current_time": now.Format("Y-m-d H:i:s"),
					"err":          err.Error(),
				}).Error("Fail to produce tasks")
				cancel()
				continue
			}

			for _, task := range tasks {
				taskJob := &TaskProducerJob{Task: task, TaskProducer: tp}
				// 提交到协程池
				if !tp.workerPool.Submit(taskJob) {
					// 提交任务失败
					logger.Logger.WithFields(logrus.Fields{
						"task_id": task.TaskId,
						"task":    task,
					}).Error("Fail to submit task to worker pool")
				}
			}

			// Finally release the ticker
			cancel()
		}
	}
}

func (tp *TaskProducer) Send(ctx context.Context, pulsarMessage *pulsar.ProducerMessage) (pulsar.MessageID, error) {
	return tp.producer.Send(ctx, pulsarMessage)
}
