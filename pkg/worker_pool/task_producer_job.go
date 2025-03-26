package worker_pool

import (
	"context"
	"encoding/json"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/robfig/cron/v3"
	"log"
	"math/rand"
	"task/internal/model"
	"task/pkg/pulsar_queue"
	"time"
)

// TaskProducerJob实现了Job接口，用来处理(生产者的任务对象)
type TaskProducerJob struct {
	Task *model.Task
}

type TaskMessage struct {
	TaskId  int64  `json:"task_id"`
	Params  []byte `json:"params"`
	Version int64  `json:"version"`
}

var Parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
var Producer pulsar.Producer

func init() {
	producer, err := pulsar_queue.NewProducer(pulsar.ProducerOptions{
		Topic: "tasks",
	})
	if err != nil {
		log.Fatalf("连接Pulsar消息队列失败:%v", err)
	}
	Producer = producer
}

func (tpj *TaskProducerJob) Process(ctx context.Context) (interface{}, error) {
	// 这里是CPU密集型的任务处理逻辑
	// 例如：拆分任务，组装任务等
	select {
	case <-ctx.Done():
		// 超时退出
		return nil, ctx.Err()
	default:
		// 处理成可执行的任务，然后投递到消息队列
		taskMessage, _ := json.Marshal(TaskMessage{
			TaskId:  tpj.Task.TaskId,
			Params:  tpj.Task.Params,
			Version: tpj.Task.Version,
		})
		msgReceipt, err := Producer.Send(ctx, &pulsar.ProducerMessage{
			Payload: taskMessage,
		})
		log.Printf("生产者投递消息成功，消息id结构体:%+v", msgReceipt)

		tpj.Task.CronTaskIds = []int64{rand.Int63()}
		tpj.Task.Version++
		schedule, err := Parser.Parse(tpj.Task.Cron)
		if err != nil {
			return nil, err
		}
		tpj.Task.NextPendingTime = int64(schedule.Next(time.Unix(tpj.Task.NextPendingTime, 0)).Second())

		// 返回处理后的任务
		return tpj.Task, nil
	}
}
