package main

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"spider/internal/pkg/spider"
)

var runCMD = &cobra.Command{
	Use:   "run",
	Short: "开启爬虫",
	Long:  `开启爬虫。本命令行的时间格式为 1s, 1m2s, 1h`,
	Run:   runCMDFunc,
}

func runCMDFunc(cmd *cobra.Command, args []string) {
	s := spider.NewSpider(spider.NewMysqlTaskHandler(), concurrency, taskBatch, taskPoolCap, minSleepTime, maxSleepTime, waitForTaskSleepTime, proxy)
	logrus.Info("程序已启动")
	s.GoRun()

}

var (
	minSleepTime         time.Duration
	maxSleepTime         time.Duration
	waitForTaskSleepTime time.Duration
	concurrency          int
	taskBatch            int
	taskPoolCap          int

	proxy string
)

func init() {
	runCMD.Flags().IntVarP(&concurrency, "concurrency", "c", 3, "爬虫并发数")
	runCMD.Flags().IntVarP(&taskBatch, "task-batch", "b", 10, "每批次获取任务的数量")
	runCMD.Flags().IntVarP(&taskPoolCap, "task-pool-cap", "p", 50, "任务池容量")
	runCMD.Flags().DurationVarP(&minSleepTime, "min", "m", time.Second, "两次请求最小间隔时间，下限0.5s")
	runCMD.Flags().DurationVarP(&maxSleepTime, "max", "M", time.Second*2, "两次请求最大间隔时间，上限10s，")
	runCMD.Flags().DurationVarP(&waitForTaskSleepTime, "wait", "w", time.Minute*5, "没有任务时，多久再获取一次任务，范围 1min~1h")
	runCMD.Flags().StringVarP(&proxy, "proxy", "P", "", "隧道代理地址，格式为 http://ip:port:username@password")
}
