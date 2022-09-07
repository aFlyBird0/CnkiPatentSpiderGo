package spider

import (
	"errors"
	"math/rand"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"spider/db"
)

var ErrTaskAllFinished = errors.New("所有任务已完成，等待新任务中，已切换到超低频爬取模式")

type TaskHandler interface {
	GetTask() (taskID uint, publicCode string, date string, code string, err error)
	SavePatent(taskID uint, patent *Patent) error
}

type Task struct {
	gorm.Model
	PublicCode string `gorm:"index:idx_public_code,unique"` // 公开号
	Date       string // 日期
	Code       string // 学科代码
	Finish     bool   `gorm:"default:0"` // 是否已经完成
	CrawlCount int    `gorm:"default:0"` // 总计被爬取的次数
}

type MysqlTaskHandler struct {
}

func init() {
	if err := db.GetDB().AutoMigrate(&Task{}, &Patent{}); err != nil {
		logrus.Fatal(err)
	}
}

func NewMysqlTaskHandler() TaskHandler {
	return &MysqlTaskHandler{}
}

func (th *MysqlTaskHandler) GetTask() (taskID uint, publicCode string, date string, code string, err error) {
	var tasks []Task
	// 寻找未完成的且被爬取次数较小的任务
	err = db.GetDB().Find(&tasks, "finish = ?", false).Limit(200).Error
	if err != nil {
		return 0, "", "", "", err
	}
	if len(tasks) == 0 {
		return 0, "", "", "", ErrTaskAllFinished
	}

	// 随机选择一个任务
	randIndex := rand.Int() % len(tasks)
	task := tasks[randIndex]

	// 更新被爬取次数
	if err := db.GetDB().Model(&Task{}).Where("id = ?", task.ID).Update("crawl_count", task.CrawlCount+1).Error; err != nil {
		logrus.Errorf("更新被爬取次数失败: %v", err)
	}
	//fmt.Printf("task: %+v\n", task)
	return task.ID, task.PublicCode, task.Date, task.Code, nil
}

func (th *MysqlTaskHandler) SavePatent(taskID uint, patent *Patent) error {
	// 保存专利
	err := db.GetDB().
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "apply_publication_no"}, {Name: "auth_publication_no"}},
			DoNothing: true,
		}).
		Create(patent).Error
	// 更新任务状态
	if err == nil {
		return db.GetDB().Model(&Task{}).Where("id = ?", taskID).Update("finish", true).Error
	}

	return err
}
