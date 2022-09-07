package main

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	spider "spider/internal/pkg/spider"
)

var runCMD = &cobra.Command{
	Use:   "run",
	Short: "开启爬虫",
	Long:  `开启爬虫。本命令行的时间格式为 1s, 1m2s, 1h`,
	Run:   runCMDFunc,
}

func runCMDFunc(cmd *cobra.Command, args []string) {
	s := spider.NewSpider(spider.NewMysqlTaskHandler(), concurrency, minSleepTime, maxSleepTime, waitForTaskSleepTime)
	logrus.Info("程序已启动")
	s.GoRun()

}

var (
	minSleepTime         time.Duration
	maxSleepTime         time.Duration
	waitForTaskSleepTime time.Duration
	concurrency          int
)

func init() {
	runCMD.Flags().IntVarP(&concurrency, "concurrency", "c", 1, "爬虫并行数")
	runCMD.Flags().DurationVarP(&minSleepTime, "min", "m", time.Second, "两次请求最小间隔时间，下限0.5s")
	runCMD.Flags().DurationVarP(&maxSleepTime, "max", "M", time.Second*2, "两次请求最大间隔时间，上限10s，")
	runCMD.Flags().DurationVarP(&waitForTaskSleepTime, "wait", "w", time.Minute*5, "没有任务时，多久再获取一次任务，范围 1min~1h")

}
