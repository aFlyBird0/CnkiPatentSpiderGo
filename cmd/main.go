package main

import (
	"io"
	"os"

	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"spider/internal/pkg/spider"
)

var (
	isDebug bool
	rootCMD = &cobra.Command{
		Use:        "spider",
		Short:      `知网专利分布式爬虫`,
		Long:       `知网专利分布式爬虫`,
		SuggestFor: []string{"run"},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initLog()
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCMD.PersistentFlags().BoolVarP(&isDebug, "debug", "", false, "debug level log")
	rootCMD.AddCommand(runCMD)
}

func initConfig() {
	if err := viper.BindPFlags(rootCMD.Flags()); err != nil {
		logrus.Fatal(err)
	}
}

func initLog() {
	if isDebug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Infof("Log level is: %s.", logrus.GetLevel())
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	if err := os.MkdirAll(spider.LogDir, os.ModePerm); err != nil {
		logrus.Fatalf("创建日志目录失败: %v", err)
	}

	infoLogFile, err := os.OpenFile(spider.LogInfoFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Fatalf("打开Info日志文件失败: %v", err)
	}
	errorLogFile, err := os.OpenFile(spider.LogErrorFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Fatalf("打开Error日志文件失败: %v", err)
	}

	// 同时输出 info 级别的日志到文件和控制台
	infoWriters := io.MultiWriter(os.Stdout, infoLogFile)
	// 同时输出 error 级别的日志到文件和控制台
	errorWriters := io.MultiWriter(os.Stdout, errorLogFile)

	// 为不同级别设置不同的输出目标
	lfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.ErrorLevel: errorWriters,
		logrus.InfoLevel:  infoWriters,
	}, nil)
	logrus.AddHook(lfHook)
}

func main() {
	err := rootCMD.Execute()
	if err != nil {
		os.Exit(1)
	}
}
