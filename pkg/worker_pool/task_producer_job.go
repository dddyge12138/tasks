package worker_pool

import (
	"context"
	"task/internal/model"
)

// TaskProducerJob实现了Job接口，用来处理(生产者的任务对象)
type TaskProducerJob struct {
	Task *model.Task
}

func (tpj *TaskProducerJob) Process(ctx context.Context) (interface{}, error) {
	// 这里是CPU密集型的任务处理逻辑
	// 例如：拆分任务，组装任务等

	select {
	case <-ctx.Done():
		// 超时退出
		return nil, ctx.Err()
	default:
		// TODO 任务处理逻辑

		// 返回处理后的任务
		return tpj.Task, nil
	}
}
