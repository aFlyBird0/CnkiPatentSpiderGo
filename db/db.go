package db

import (
	_ "embed"
	"fmt"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

// 数据库连接 DSN ，用这种方式简单的来防止因开源带来的密码泄露
// 也能在分发的时候直接把账号密码打进二进制文件
//go:embed dsn.txt
var dsn string

var once sync.Once

func GetDB() *gorm.DB {
	once.Do(func() {
		var err error
		db, err = gorm.Open(mysql.Open(strings.TrimSpace(dsn)), &gorm.Config{})
		//db, err = gorm.Open(sqlite.Open("patent_test.db"), &gorm.Config{})	// sqlite3 for test
		if err != nil {
			logrus.Fatal(fmt.Errorf("数据库连接失败: %w", err))
		}
	})
	return db
}
