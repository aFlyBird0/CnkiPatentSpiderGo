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
	Long:  `开启爬虫`,
	Run:   runCMDFunc,
}

func runCMDFunc(cmd *cobra.Command, args []string) {
	s := spider.NewSpider(spider.NewMysqlTaskHandler(), minSleepTime, maxSleepTime)
	logrus.Info("开启爬虫")
	s.Run()

}

var (
	minSleepTime time.Duration
	maxSleepTime time.Duration
)

func init() {
	runCMD.Flags().DurationVarP(&minSleepTime, "min", "m", time.Second/2, "最小睡眠时间")
	runCMD.Flags().DurationVarP(&maxSleepTime, "max", "M", time.Second*2, "最大睡眠时间")
}
