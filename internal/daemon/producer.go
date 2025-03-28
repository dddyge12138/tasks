// 需要在后台持续运行的协程
package daemon

import (
	"context"
	"fmt"
	"google.golang.org/appengine/log"
	"task/internal/model"
	redis_db "task/pkg/redis"
	"task/pkg/worker_pool"
	"time"
)

func Init() {

}

// ProducerWorkerPool是一个全局的生产者协程池实例
var ProducerWorkerPoolInstance *worker_pool.WorkerPool

func InitProducerWorkerPool(ctx context.Context) {
	// 创建协程池，设置10个工作协程，队列大小100， 单任务超时10秒
	ProducerWorkerPoolInstance = worker_pool.NewWorkerPool(10, 100, 10*time.Second)
	// 设置回调函数，用于处理工作协程的结果
	// 生产者生产任务后的结果可以忽略
	ProducerWorkerPoolInstance.SetCallback(func(result worker_pool.Result) error {
		if result.Err != nil {
			// 这里可以处理错误
			return result.Err
		}

		_, ok := result.Value.(*model.Task)
		if !ok {
			return fmt.Errorf("expected *mode.Task but got %T", result.Value)
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
				log.Errorf(ctx, "Fail to produce tasks, current_time: %s,err_msg is: %s, err is: %v", now.Format("Y-m-d H:i:s"), err.Error(), err)
				cancel()
				continue
			}

			for _, task := range tasks {
				taskJob := &worker_pool.TaskProducerJob{Task: task}
				// 提交到协程池
				if !ProducerWorkerPoolInstance.Submit(taskJob) {
					// 提交任务失败
					log.Errorf(jobTimeoutCtx, "Failed to submit task %s to worker pool, task: %+v", task.TaskId, task)
				}
			}

			// Finally release the ticker
			cancel()
		}
	}
}
