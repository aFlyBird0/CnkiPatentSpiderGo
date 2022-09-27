package spider

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

const DefaultTasksChanCap = 50 // task 队列容量

type WorkerPool struct {
	th          TaskHandler // 获取与更新任务的 handler
	workerNum   int         // worker 数量
	taskBatch   int         // 每批次获取的任务数量
	taskChanCap int         // task 任务池容量

	workerFunc           func(task *Task) error // worker 睡眠函数
	workerSleepFunc      func()
	taskHandlerSleepFunc func()
	tasksChan            chan Task
}

func NewWorkerPool(th TaskHandler, workerNum int, taskBatch, taskChanCap int, workerFunc func(task *Task) error,
	workerSleepFunc func(), taskHandlerSleepFunc func()) *WorkerPool {
	wp := &WorkerPool{
		th:                   th,
		workerNum:            workerNum,
		taskBatch:            taskBatch,
		workerFunc:           workerFunc,
		workerSleepFunc:      workerSleepFunc,
		taskHandlerSleepFunc: taskHandlerSleepFunc,
	}
	if taskChanCap <= 0 {
		taskChanCap = DefaultTasksChanCap
	}
	wp.tasksChan = make(chan Task, taskChanCap)

	return wp
}

// AddTasks 不断地增加任务
func (wp *WorkerPool) AddTasks(th TaskHandler, num int) error {
	tasks, err := th.RandomBatchTasks(num)
	if err != nil {
		// 如果获取到的任务是空，则 sleep 一段时间
		if errors.Is(err, ErrTaskAllFinished) {
			wp.taskHandlerSleepFunc()
		}
		return err
	}
	for _, task := range tasks {
		//logrus.Info("任务入队: ", task)
		// 如果任务过多会自动阻塞
		wp.tasksChan <- task
	}
	logrus.Info("当前批次任务已全部入队")
	return nil
}

// GetTask 获取任务
func (wp *WorkerPool) GetTask() Task {
	return <-wp.tasksChan
}

func (wp *WorkerPool) Run() {
	continuousErrCount := 0
	// 1. 不断往队列中塞任务
	go func() {
		for {
			if continuousErrCount > 600 {
				logrus.Error("连续600次错误，退出")
				os.Exit(1)
			}
			logrus.Info("获取下一批次任务")
			if err := wp.AddTasks(wp.th, wp.taskBatch); err != nil {
				logrus.Error("获取任务失败: ", err)
				continuousErrCount += 1
				time.Sleep(time.Second)
				continue
			}
			continuousErrCount = 0
		}
	}()

	// 2. 启动爬虫
	for i := 0; i < wp.workerNum; i++ {
		go func() {
			for {
				// 获取任务，如果任务过少会自动阻塞
				task := wp.GetTask()
				// 自动睡眠一段时间
				wp.workerSleepFunc()
				logrus.Infof("开始爬取，任务: %v", task)
				// 执行爬虫任务
				if err := wp.workerFunc(&task); err != nil {
					logrus.Error("爬取任务失败: ", err)
					continue
				}
			}
		}()
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	// 这里不能加 syscall.SIGHUP，否则会导致终端连接断开后，程序退出（哪怕是后台运行）
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	//logrus.Info("等待退出信号")
	<-done
	logrus.Info("检测到退出信号，退出")
}
