package spider

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestWorker(t *testing.T) {
	workerFunc := func(task *Task) error {
		logrus.Info("假装正在爬取任务: ", task)
		return nil
	}
	workerSleepFunc := func() {
		time.Sleep(time.Second)
	}
	taskHandlerSleepFunc := func() {
		logrus.Info("未检测到任务，睡眠 10 秒")
		time.Sleep(time.Second * 10)
	}
	wp := NewWorkerPool(NewFakeTaskHandler(), 5, 10, workerFunc, workerSleepFunc, taskHandlerSleepFunc)
	wp.Run()
}
