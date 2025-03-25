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
		// TODO 这里很关键，任务在处理的过程中要监听超时信号, 因为如果任务处理的时间过长，
		// TODO 超时了但是没有进行处理，就会出现父协程通知退出但没有退出的情况
		// 超时退出
		return nil, ctx.Err()
	default:
		// TODO 任务处理逻辑

		// 返回处理后的任务
		return tpj.Task, nil
	}
}
