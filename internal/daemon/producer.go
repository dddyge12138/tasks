// 需要在后台持续运行的协程
package daemon

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"task/internal/model"
	"task/internal/repository"
	"task/pkg/database"
	"task/pkg/logger"
	redis_db "task/pkg/redis"
	"task/pkg/worker_pool"
	"time"
)

// ProducerWorkerPool是一个全局的生产者协程池实例
var ProducerWorkerPoolInstance *worker_pool.WorkerPool
var taskRepo repository.TaskRepository

func InitProducerWorkerPool(ctx context.Context) {
	taskRepo = repository.NewTaskRepository(database.Db)
	// 创建协程池，设置10个工作协程，队列大小100， 单任务超时10秒
	ProducerWorkerPoolInstance = worker_pool.NewWorkerPool(10, 100, 10*time.Second)
	// 设置回调函数，用于处理工作协程的结果
	// 生产者生产任务后的结果可以忽略
	ProducerWorkerPoolInstance.SetCallback(func(result worker_pool.Result) error {
		if result.Err != nil {
			// 这里可以处理错误
			return result.Err
		}

		task, ok := result.Value.(*model.Task)
		if !ok {
			return fmt.Errorf("expected *mode.Task but got %T", result.Value)
		}

		// 更新任务状态
		if err := taskRepo.UpdateTaskAfterProduce(ctx, task); err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"task_id": task.TaskId,
				"err":     err,
			}).Error("Fail to update task after produce")
			return err
		}
		// 删除redis槽内的任务以及任务详情。保留redis任务的版本号供下游消费者查询是否执行
		if err := redis_db.RemoveSlotTasks(ctx, task.TaskId); err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"task_id": task.TaskId,
				"err":     err,
			}).Error("Fail to remove task from redis")
			return err
		}

		return nil
	})

	// 启动协程池
	ProducerWorkerPoolInstance.Start(ctx)
}

func TaskProducerRun(ctx context.Context) {
	var tasks []*model.Task
	var err error
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			jobTimeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			tasks, err = redis_db.GetTasksList(jobTimeoutCtx, now.Add(-10*time.Minute).Unix(), now.Add(30*time.Second).Unix())
			if err != nil {
				logger.Logger.WithFields(logrus.Fields{
					"current_time": now.Format("Y-m-d H:i:s"),
					"err":          err.Error(),
				}).Error("Fail to produce tasks")
				cancel()
				continue
			}

			for _, task := range tasks {
				taskJob := &worker_pool.TaskProducerJob{Task: task}
				// 提交到协程池
				if !ProducerWorkerPoolInstance.Submit(taskJob) {
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
