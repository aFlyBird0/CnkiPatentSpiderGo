package spider

import "math/rand"

type FakeTaskHandler struct {
	CallNumOfRandomBatchTasks int
}

func NewFakeTaskHandler() *FakeTaskHandler {
	return &FakeTaskHandler{}
}

func (f *FakeTaskHandler) RandomTask() (Task, error) {
	randID := rand.Int() % 10000
	task := Task{}
	task.ID = uint(randID)
	return task, nil
}

func (f *FakeTaskHandler) RandomBatchTasks(num int) ([]Task, error) {
	// 模拟第6到7次请求的时候，无任务，在第8次请求的时候又有任务了
	if f.CallNumOfRandomBatchTasks > 5 && f.CallNumOfRandomBatchTasks < 8 {
		return nil, ErrTaskAllFinished
	}
	var tasks []Task
	for i := 0; i < num; i++ {
		task, _ := f.RandomTask()
		tasks = append(tasks, task)
	}
	f.CallNumOfRandomBatchTasks += 1
	return tasks, nil
}

func (f *FakeTaskHandler) SavePatent(_ uint, _ *Patent) error {
	return nil
}
