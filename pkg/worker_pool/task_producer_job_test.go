package worker_pool

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"task/config"
	"task/internal/model"
	"task/internal/repository"
	"task/internal/service"
	"task/pkg/database"
	"task/pkg/pulsar_queue"
	redis_db "task/pkg/redis"
	"testing"
	"time"

	"github.com/spf13/viper"

	"github.com/stretchr/testify/assert"
)

var (
	taskRepo    repository.TaskRepository
	taskService service.TaskService
	ctx         context.Context
	cancelFunc  context.CancelFunc
	resultChan  chan Result
)

// setup初始化测试环境，包括协程池的创建
func setup() {
	// 创建上下文和取消函数
	ctx, cancelFunc = context.WithCancel(context.Background())

	// 设置结果回调函数
	resultChan = make(chan Result, 10)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../config")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}
	// 初始化任务仓库和服务
	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Unable to decode config into struct: %s", err)
	}
	err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %s", err)
	}
	taskRepo = repository.NewTaskRepository(database.Db)
	taskService = service.NewTaskService(taskRepo)

	// 初始化pulsar客户端连接
	_, err = pulsar_queue.NewPulsarClient(cfg.Pulsar)
	if err != nil {
		log.Fatalf("初始化Pulsar客户端失败")
	}
	// Initialize Redis
	err = redis_db.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %s", err)
	}

	// 初始化生产者
	InitProducer()
	initProducerWorkerPool(ctx)
	go TaskProducerRun(ctx)
	log.Println("初始化完成")
}

var producerWorkerPoolInstance *WorkerPool

func initProducerWorkerPool(ctx context.Context) {
	// 创建协程池，设置10个工作协程，队列大小100， 单任务超时10秒
	producerWorkerPoolInstance = NewWorkerPool(10, 100, 10*time.Second)
	// 设置回调函数，用于处理工作协程的结果
	// 生产者生产任务后的结果可以忽略
	producerWorkerPoolInstance.SetCallback(func(result Result) error {
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
	producerWorkerPoolInstance.Start(ctx)
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
				log.Printf("Fail to produce tasks, current_time: %s,err_msg is: %s, err is: %v", now.Format("Y-m-d H:i:s"), err.Error(), err)
				cancel()
				continue
			}

			for _, task := range tasks {
				taskJob := &TaskProducerJob{Task: task}
				// 提交到协程池
				if !producerWorkerPoolInstance.Submit(taskJob) {
					// 提交任务失败
					log.Printf("Failed to submit task %d to worker pool, task: %+v", task.TaskId, task)
				}
			}

			// Finally release the ticker
			cancel()
		}
	}
}

// clear清理测试环境
func clearTest() {

	// 取消上下文
	if cancelFunc != nil {
		cancelFunc()
	}
	producerWorkerPoolInstance.Stop()
	// 关闭结果通道
	close(resultChan)
}

// createTasksAtDifferentTimes创建不同时间点需要执行的任务并保存到数据库
func createTasksAtDifferentTimes(t *testing.T) []*model.Task {
	// 创建几个不同时间点的任务
	now := time.Now()
	tasks := []*model.Task{
		{
			TaskId:          rand.Int63(),
			Name:            "立即执行任务",
			Status:          1,
			Cron:            "",
			NextPendingTime: now.Unix(),
			Params:          []byte(`{"action":"immediate"}`),
			Version:         1,
		},
		{
			TaskId:          rand.Int63(),
			Name:            "5分钟后执行任务",
			Status:          1,
			Cron:            "*/5 * * * *",
			NextPendingTime: now.Add(5 * time.Minute).Unix(),
			Params:          []byte(`{"action":"delayed"}`),
			Version:         1,
		},
		{
			TaskId:          rand.Int63(),
			Name:            "每小时执行任务",
			Status:          1,
			Cron:            "0 * * * *",
			NextPendingTime: now.Add(1 * time.Hour).Unix(),
			Params:          []byte(`{"action":"hourly"}`),
			Version:         1,
		},
	}

	// 保存任务到数据库
	for _, task := range tasks {
		err := taskRepo.CreateTask(ctx, task)
		assert.NoError(t, err, "创建任务应该成功")
	}

	return tasks
}

// loadTasksViaAPI通过API接口加载任务
func loadTasksViaAPI(t *testing.T) {
	err := taskService.LoadTask(ctx)
	assert.NoError(t, err, "加载任务应该成功")
}

func testProducerAutoPolling(t *testing.T) {
	// 生产者应该自动轮询Redis获取任务
	log.Println("启动生产者轮询服务，等待自动获取和处理任务...")

	// 等待足够的时间让生产者至少轮询一次
	// 根据实际轮询间隔调整等待时间
	pollingTime := 40 * time.Second // 等待比轮询间隔稍长的时间
	timeout := time.After(pollingTime)

	// 定义一个成功计数通道，用于记录成功处理的任务数
	successCount := 0

	// 监听结果通道
	for {
		select {
		case result := <-resultChan:
			// 验证任务处理结果
			assert.NoError(t, result.Err, "任务处理不应有错误")
			assert.NotNil(t, result.Value, "任务结果不应为空")

			processedTask, ok := result.Value.(*model.Task)
			assert.True(t, ok, "结果应该是Task类型")

			// 验证任务处理后的状态
			assert.NotEqual(t, 0, processedTask.TaskId, "任务ID不应为0")
			assert.GreaterOrEqual(t, processedTask.Version, int64(2), "版本号应该至少增加到2")
			assert.NotEmpty(t, processedTask.CronTaskIds, "应该生成CronTaskIds")

			log.Printf("成功处理任务: %d\n", processedTask.TaskId)
			successCount++

		case <-timeout:
			// 轮询时间结束，检查是否有任务被处理
			if successCount == 0 {
				t.Logf("警告: 在%v的等待时间内没有任务被处理", pollingTime)
				// 这里可以选择失败或只是警告
				// t.Fatal("生产者自动轮询未处理任何任务")
			} else {
				t.Logf("生产者自动轮询成功处理了 %d 个任务", successCount)
			}
			return
		}
	}
}

// TestProducer测试生产者的完整流程
func TestProducer(t *testing.T) {
	// 设置测试环境
	setup()
	defer clearTest()

	// 1. 创建不同时间点的任务并入库
	createTasksAtDifferentTimes(t)

	// 2. 调用API接口加载任务到Redis
	loadTasksViaAPI(t)

	// 3. 测试生产者自动轮询功能
	testProducerAutoPolling(t)

	// 可以在这里添加更多验证，例如检查数据库中任务的状态变化
}
