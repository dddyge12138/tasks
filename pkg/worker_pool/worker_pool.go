package worker_pool

import (
	"context"
	"log"
	"sync"
	"time"
)

type Job interface {
	// Process执行实际的工作，并返回结果或错误
	Process(ctx context.Context) (interface{}, error)
}

// Result 表示任务处理的结果
type Result struct {
	Value interface{}
	Err   error
	Job   Job
}

// 表示一个协程池
type WorkerPool struct {
	workerCount int
	jobQueue    chan Job
	results     chan Result
	done        chan struct{} // 用于通知协程池关闭
	wg          sync.WaitGroup
	timeout     time.Duration      // 单个任务的超时时间
	callback    func(Result) error // 结果处理的回调函数
}

func NewWorkerPool(workerCount int, queueSize int, timeout time.Duration) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		jobQueue:    make(chan Job, queueSize),
		results:     make(chan Result, queueSize),
		done:        make(chan struct{}),
		wg:          sync.WaitGroup{},
		timeout:     timeout,
	}
}

func (wp *WorkerPool) SetCallback(callback func(Result) error) {
	wp.callback = callback
}

func (wp *WorkerPool) Submit(job Job) bool {
	select {
	case wp.jobQueue <- job:
		return true
	case <-wp.done:
		return false
	default:
		// 队列已满的处理, 拒绝任务
		return false
	}
}

// worker是工作协程的主要工作逻辑
func (wp *WorkerPool) worker(ctx context.Context, id int) {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.done:
			return
		case <-ctx.Done():
			return
		case job, ok := <-wp.jobQueue:
			if !ok {
				return
			}
			// 为每一个任务创建一个超时上下文
			jobCtx, cancel := context.WithTimeout(ctx, wp.timeout)
			value, err := job.Process(jobCtx)
			if err != nil {
				log.Printf("worker协程处理任务得到错误, err_msg:%s\n", err.Error())
				cancel()
				continue
			}
			cancel()
			// 将结果发送到结果channel
			select {
			case wp.results <- Result{Value: value, Err: err, Job: job}:
			case <-wp.done:
				return
			case <-jobCtx.Done():
				return
			}
		}
	}
}

// 处理结果队列中的结果
func (wp *WorkerPool) processResults(ctx context.Context) {
	for {
		select {
		case result, ok := <-wp.results:
			if !ok {
				return
			}
			if wp.callback != nil {
				if err := wp.callback(result); err != nil {
					// 这里可以添加错误处理逻辑
					// 例如记录日志或者重试
					log.Printf("Error processing result: %+v, err %s", result, err.Error())
				}
			}
		case <-wp.done:
			return
			// 这里是个bug！！这里不应该监听ctx的退出信号，因为这个结果处理协程的生命周期应该随着协程池结束
			// 引以为戒！
			//case <-ctx.Done():
			//	fmt.Println("接受结果出错")
			//	return
		}
	}
}

// Start启动协程池
func (wp *WorkerPool) Start(ctx context.Context) {
	// 启动工作协程
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}

	// 如果设置了回调函数，就启动结果处理协程
	if wp.callback != nil {
		go wp.processResults(ctx)
	}
}

// 停止协程池，当task没有生产完成的时候，等到下一个周期时会再次进入生产队列
func (wp *WorkerPool) Stop() {
	close(wp.done)
	wp.wg.Wait()
	close(wp.jobQueue)
	close(wp.results)
}
