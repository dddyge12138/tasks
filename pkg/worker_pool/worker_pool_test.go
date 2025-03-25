package worker_pool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// 模拟任务，实现Job接口
type MockJob struct {
	id       int
	duration time.Duration
	result   interface{}
	err      error
	executed bool
}

func (m *MockJob) Process(ctx context.Context) (interface{}, error) {
	// 检查上下文是否已取消
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// 模拟耗时任务
		if m.duration > 0 {
			select {
			case <-time.After(m.duration):
				m.executed = true
				return m.result, m.err
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		m.executed = true
		return m.result, m.err
	}
}

func TestWorkerPoolBasicFunctionality(t *testing.T) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建工作池：3个工作协程，队列大小为5，任务超时2秒
	wp := NewWorkerPool(3, 10, 2*time.Second)

	// 记录处理的结果数量
	var processedCount int32
	var errorCount int32

	// 设置回调函数
	wp.SetCallback(func(result Result) error {
		if result.Err != nil {
			atomic.AddInt32(&errorCount, 1)
			return result.Err
		}
		atomic.AddInt32(&processedCount, 1)
		return nil
	})

	// 启动工作池
	wp.Start(ctx)

	// 创建10个模拟任务
	jobs := make([]*MockJob, 10)
	for i := 0; i < 10; i++ {
		// 每隔一个任务产生一个错误
		if i%2 == 0 {
			jobs[i] = &MockJob{
				id:       i,
				duration: 100 * time.Millisecond,
				result:   i,
				err:      nil,
			}
		} else {
			jobs[i] = &MockJob{
				id:       i,
				duration: 100 * time.Millisecond,
				result:   nil,
				err:      errors.New("mock error"),
			}
		}
	}

	// 提交任务
	for i, job := range jobs {
		if !wp.Submit(job) {
			t.Errorf("Failed to submit job %d", i)
		}
	}

	// 等待足够的时间让任务完成（由于设置了并发=3，处理10个任务应该需要至少4个处理周期）
	time.Sleep(500 * time.Millisecond)

	// 检查所有任务是否都被执行
	for i, job := range jobs {
		if !job.executed {
			t.Errorf("Job %d was not executed", i)
		}
	}

	// 停止工作池
	wp.Stop()

	// 验证处理的结果
	if atomic.LoadInt32(&processedCount) != 5 {
		t.Errorf("Expected 5 successful jobs, got %d", processedCount)
	}

	if atomic.LoadInt32(&errorCount) != 5 {
		t.Errorf("Expected 5 error jobs, got %d", errorCount)
	}
}

func TestWorkerPoolTimeout(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建工作池：2个工作协程，任务超时仅100毫秒
	wp := NewWorkerPool(2, 5, 100*time.Millisecond)

	var timeoutCount int32
	wp.SetCallback(func(result Result) error {
		if result.Err != nil && errors.Is(result.Err, context.DeadlineExceeded) {
			atomic.AddInt32(&timeoutCount, 1)
		}
		return nil
	})

	wp.Start(ctx)

	// 创建一个耗时任务，会超时
	longJob := &MockJob{
		id:       1,
		duration: 300 * time.Millisecond, // 超过设定的100毫秒超时
		result:   "slow result",
		err:      nil,
	}

	wp.Submit(longJob)

	// 给足够时间让任务超时并被处理
	time.Sleep(400 * time.Millisecond)

	wp.Stop()

	// 验证超时处理
	if atomic.LoadInt32(&timeoutCount) != 1 {
		t.Errorf("Expected 1 timeout, got %d", timeoutCount)
	}
}

func TestWorkerPoolCancelContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// 创建工作池
	wp := NewWorkerPool(2, 5, 5*time.Second)
	// 用于等待任务被取消
	var wg sync.WaitGroup
	wg.Add(1)

	var cancelled bool
	// 设置回调函数以检测取消
	wp.SetCallback(func(result Result) error {
		if errors.Is(result.Err, context.Canceled) {
			cancelled = true
			wg.Done()
		}
		return nil
	})

	// 启动工作池
	longJob := &MockJob{
		id:       1,
		duration: 5 * time.Second, // 设置一个长时间任务
		executed: false,
	}

	wp.Start(ctx)
	// 提交任务
	wp.Submit(longJob)

	// 让任务开始执行
	time.Sleep(100 * time.Millisecond)

	// 取消上下文
	cancel()

	// 等待任务被取消或超时
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 成功取消
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for job to be cancelled")
	}

	wp.Stop()

	if !cancelled {
		t.Error("Job should have been cancelled")
	}
}

func TestWorkerPoolQueueFull(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建一个小队列的工作池
	wp := NewWorkerPool(1, 2, 1*time.Second)

	// 创建缓慢执行的任务，这样队列会填满
	slowJob := func() *MockJob {
		return &MockJob{
			duration: 500 * time.Millisecond,
			result:   "slow",
			err:      nil,
		}
	}

	wp.Start(ctx)

	// 提交任务直到队列满
	// 提交第一个任务，正在执行
	if !wp.Submit(slowJob()) {
		t.Error("First job should be accepted")
	}

	// 等待一点时间确保第一个任务开始执行
	time.Sleep(50 * time.Millisecond)

	// 提交第二个和第三个任务，填满队列
	if !wp.Submit(slowJob()) {
		t.Error("Second job should be accepted")
	}

	if !wp.Submit(slowJob()) {
		t.Error("Third job should be accepted")
	}

	// 第四个任务应该被拒绝，因为队列已满
	if wp.Submit(slowJob()) {
		t.Error("Fourth job should be rejected as queue is full")
	}

	// 等待任务完成
	time.Sleep(1 * time.Second)
	wp.Stop()
}
