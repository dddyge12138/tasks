package worker_pool

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"math/rand"
	"strconv"
	"task/internal/model"
	"task/pkg/Constants"
	"task/pkg/logger"
	"task/pkg/pulsar_queue"
	redis_db "task/pkg/redis"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/robfig/cron/v3"
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

func InitProducer() {
	producer, err := pulsar_queue.NewProducer(pulsar.ProducerOptions{
		Topic: "tasks",
	})
	if err != nil {
		logger.Logger.WithError(err).Fatal("连接Pulsar消息队列失败")
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
		resCmd := redis_db.RedisDb.Set(
			ctx,
			fmt.Sprintf(Constants.TaskVersionKey, strconv.FormatInt(tpj.Task.TaskId, 10)),
			tpj.Task.Version,
			24*time.Hour,
		)
		if err := resCmd.Err(); err != nil {
			return tpj.Task, err
		}
		taskMessage, _ := json.Marshal(TaskMessage{
			TaskId:  tpj.Task.TaskId,
			Params:  tpj.Task.Params,
			Version: tpj.Task.Version,
		})
		msgReceipt, err := Producer.Send(ctx, &pulsar.ProducerMessage{
			Payload: taskMessage,
		})
		if err != nil {
			logger.Logger.WithError(err).Error("投递消息失败")
			return nil, err
		}
		logger.Logger.WithFields(logrus.Fields{
			"msg_receipt": msgReceipt,
			"task":        tpj.Task,
		}).Info("生产者投递消息成功")

		tpj.Task.CronTaskIds = []int64{rand.Int63()}
		if len(tpj.Task.Cron) == 0 {
			return tpj.Task, nil
		}
		tpj.Task.Version++
		schedule, err := Parser.Parse(tpj.Task.Cron)
		if err != nil {
			return nil, err
		}
		tpj.Task.NextPendingTime = schedule.Next(time.Unix(tpj.Task.NextPendingTime, 0)).Unix()

		// 返回处理后的任务
		return tpj.Task, nil
	}
}
