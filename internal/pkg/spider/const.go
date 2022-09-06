package spider

import "path/filepath"

const (
	RootDir = "data"
)

var (
	HtmlDir = filepath.Join(RootDir, "html")
	LogDir  = filepath.Join(RootDir, "log")
	//LogFile = filepath.Join(LogDir, "spider.log")
	LogInfoFile  = filepath.Join(LogDir, "info.log")
	LogErrorFile = filepath.Join(LogDir, "error.log")
)
