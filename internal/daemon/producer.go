// 需要在后台持续运行的协程
package daemon

import (
	"context"
	"fmt"
	"google.golang.org/appengine/log"
	"task/internal/model"
	redis_db "task/pkg/redis"
	"time"
)

func Init() {

}

func TaskProducerRun() {
	var tasks []*model.Task
	var err error
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for now := range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		tasks, err = redis_db.GetTasksList(ctx, now.Add(-10*time.Minute).Unix(), now.Add(30*time.Second).Unix())
		if err != nil {
			log.Errorf(ctx, "Fail to produce tasks, current_time: %s,err_msg is: %s, err is: %v", now.Format("Y-m-d H:i:s"), err.Error(), err)
			cancel()
			continue
		}
		// TODO 生产任务并发送到pulsar消息队列
		for _, task := range tasks {
			fmt.Println(task.TaskId)
		}

		// TODO 协程异步更新数据库中的任务状态和拆分出来的子任务的cron_task_id

		// Finally release the ticker
		cancel()
	}
}
