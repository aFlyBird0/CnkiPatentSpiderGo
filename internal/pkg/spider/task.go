package spider

import (
	"errors"
	"math/rand"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"spider/db"
)

const maxTaskBatch = 200 // 最大一次从数据库中查询的 task 数量

var ErrTaskAllFinished = errors.New("所有任务已完成，等待新任务中，已切换到超低频爬取模式")

type TaskHandler interface {
	RandomTask() (Task, error)
	RandomBatchTasks(num int) ([]Task, error) // 随机获取至多 num 个任务，返回的任务数量 <= num
	SavePatent(taskID uint, patent *Patent) error
}

// Task 是任务库
// 实际运行可能需要给 deleted_at 和 finish 加个联合索引
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

func NewMysqlTaskHandler() TaskHandler {
	if err := db.GetDB().AutoMigrate(&Task{}, &Patent{}); err != nil {
		logrus.Fatal(err)
	}
	return &MysqlTaskHandler{}
}

func (th *MysqlTaskHandler) listTasks() (tasks []Task, err error) {
	// 寻找未完成的任务
	err = db.GetDB().Find(&tasks, "finish = ?", false).Limit(maxTaskBatch).Error
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, ErrTaskAllFinished
	}
	return tasks, nil
}

func (th *MysqlTaskHandler) RandomTask() (task Task, err error) {
	tasks, err := th.listTasks()
	if err != nil {
		return task, err
	}

	// 随机选择一个任务
	randIndex := rand.Int() % len(tasks)
	task = tasks[randIndex]

	// 更新被爬取次数
	if err := db.GetDB().Model(&Task{}).Where("id = ?", task.ID).Update("crawl_count", task.CrawlCount+1).Error; err != nil {
		logrus.Errorf("更新被爬取次数失败: %v", err)
	}
	//fmt.Printf("task: %+v\n", task)
	return task, nil
}

func (th *MysqlTaskHandler) RandomBatchTasks(num int) ([]Task, error) {
	tasks, err := th.listTasks()
	if err != nil {
		return nil, err
	}

	// 如果现在剩 task 的总数小于 num，直接返回
	if len(tasks) < num {
		return tasks, nil
	}

	// 随机获取至多 num 个任务
	// 生成 num 个随机数，范围是 tasks 的下标
	var tasksReturn []Task
	for i := 0; i < num; i++ {
		randIndex := rand.Int() % len(tasks)
		tasksReturn = append(tasksReturn, tasks[randIndex])
	}

	return tasksReturn, nil
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
