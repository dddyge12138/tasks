package daemon

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
	redis_db "task/pkg/redis"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/robfig/cron/v3"
)

// TaskProducerJob实现了Job接口，用来处理(生产者的任务对象)
type TaskProducerJob struct {
	Task         *model.Task
	TaskProducer *TaskProducer
}

var Parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

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
		cronTaskId := rand.Int63()
		taskMessage, _ := json.Marshal(model.TaskMessage{
			TaskId:      tpj.Task.TaskId,
			Params:      tpj.Task.Params,
			Version:     tpj.Task.Version,
			CronTaskIds: []int64{cronTaskId},
		})
		msgReceipt, err := tpj.TaskProducer.Send(ctx, &pulsar.ProducerMessage{
			Payload: taskMessage,
		})
		if err != nil {
			logger.Logger.WithError(err).Error("投递消息失败")
			return nil, err
		}
		tpj.Task.TaskProduceCount++
		tpj.Task.CronTaskIds = []int64{cronTaskId}
		logger.Logger.WithFields(logrus.Fields{
			"msg_receipt": msgReceipt.String(),
			"task":        tpj.Task,
		}).Info("生产者投递消息成功")

		if len(tpj.Task.Cron) == 0 {
			// 一次性任务执行结束就把IsDeleted置1
			tpj.Task.IsDeleted = 1
			return tpj.Task, nil
		}
		schedule, err := Parser.Parse(tpj.Task.Cron)
		if err != nil {
			return nil, err
		}
		tpj.Task.NextPendingTime = schedule.Next(time.Unix(tpj.Task.NextPendingTime, 0)).Unix()

		// 返回处理后的任务
		return tpj.Task, nil
	}
}
