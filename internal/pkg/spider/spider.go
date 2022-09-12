package spider

import (
	_ "embed"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
	"go.uber.org/multierr"
)

const patentPrefix = "https://kns.cnki.net/kcms/detail/detail.aspx?dbcode=SCPD&filename=%s"

type Spider struct {
	th                   TaskHandler
	minSleepTime         time.Duration // 两次爬取之间的最小睡眠时间
	maxSleepTime         time.Duration // 两次爬取之间的最大睡眠时间
	waitForTaskSleepTime time.Duration // 等待任务时的睡眠时间
	concurrency          int           // 并发数
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewSpider(th TaskHandler, concurrency int, minSleepTime, maxSleepTime, waitForTaskSleepTime time.Duration) *Spider {
	// 校验与修正参数
	if concurrency < 1 {
		logrus.Info("并发数不能小于 1，已自动设置为 1")
		concurrency = 1
	}
	if concurrency > 32 {
		logrus.Info("并发数不能大于 32，已自动设置为 32")
		concurrency = 32
	}
	if minSleepTime > maxSleepTime {
		logrus.Fatal("最小睡眠时间不能大于最大睡眠时间")
	}
	if minSleepTime < time.Millisecond*100 {
		logrus.Info("最小睡眠时间不能小于 100 毫秒，已自动设置为 100 毫秒")
		minSleepTime = time.Millisecond * 100
	}
	if maxSleepTime > time.Second*10 {
		logrus.Info("最大睡眠时间不能大于 10 秒，已自动设置为 10 秒")
		maxSleepTime = time.Second * 10
	}
	if waitForTaskSleepTime < time.Minute {
		logrus.Info("等待任务时的睡眠时间不能小于1分钟，已自动设置为 1 分钟")
		waitForTaskSleepTime = time.Minute
	}
	if waitForTaskSleepTime > time.Hour {
		logrus.Info("等待任务时的睡眠时间不能大于1小时，已自动设置为 1 小时")
		waitForTaskSleepTime = time.Hour
	}
	return &Spider{
		th:                   th,
		concurrency:          concurrency,
		minSleepTime:         minSleepTime,
		maxSleepTime:         maxSleepTime,
		waitForTaskSleepTime: waitForTaskSleepTime,
	}
}

func (s *Spider) GoRun() {
	logrus.Infof("并发数为 %d", s.concurrency)
	wp := NewWorkerPool(s.th, s.concurrency, 10, s.Run, s.RandomSleep, s.WaitForTask)
	wp.Run()
}

func (s *Spider) Run(task *Task) error {
	// 解析专利内容
	patent, err := s.ParseContent(task.Date, task.Code, task.PublicCode)
	if err != nil {
		return err
	}
	// 校验合法性
	if !patent.Validate() {
		logrus.Error("数据不合法")
		logrus.Infof("patent: %+v\n", patent)
		return err
	}
	// 保存到数据库
	save := func() {
		patent.RemoveAllBlank()
		logrus.Infof("保存专利到数据库中: %+v", patent)
		if err := s.th.SavePatent(task.ID, patent); err != nil {
			logrus.Error(err)
		}
	}
	go save()
	return nil
}

func (s *Spider) ParseContent(date, code, publicCode string) (patent *Patent, err error) {
	url := getPatentURL(publicCode)
	logrus.Debugf("开始解析 %s %s %s", date, code, url)

	// 请求专利内容 html
	body, err := s.GetHtml(url)
	if err != nil {
		return nil, err
	}

	// 保存 html
	go s.SaveHtml(body, date, code, publicCode)

	patent = &Patent{}
	patent.NaviCode = code
	if len(date) >= 4 {
		patent.Year = date[0:4]
	}
	patent.Url = url

	// 解析 html
	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	if titleNode, err := htmlquery.Query(doc, "//h1//text()"); err != nil {
		return nil, err
	} else if titleNode != nil {
		patent.Title = strings.TrimSpace(htmlquery.InnerText(titleNode))
	}

	// 有的是在row下，有的是在row的row1和row2下，这么写效率最高
	rows, err := htmlquery.QueryAll(doc, "//div[@class='row'] | //div[@class='row-1'] | //div[@class='row-2']")
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		// 获取 key, 形如"申请号："
		key, err := htmlquery.Query(row, "./span[@class='rowtit']/text() | ./span[@class='rowtit2']/text()")
		if err != nil {
			return nil, err
		}
		if key == nil {
			continue
		}
		keyString := strings.TrimSpace(htmlquery.InnerText(key))
		if keyString == "" {
			continue
		}

		// 获取 value
		valueList, err := htmlquery.QueryAll(row, "./p[@class='funds']//text()")
		if err != nil {
			return nil, err
		}
		var valueString string
		for _, value := range valueList {
			valueString += strings.TrimSpace(htmlquery.InnerText(value))
		}

		// 根据 key, value 填充 patent
		patent.FillRowFields(keyString, valueString)

	}

	// 摘要
	abstract, err := htmlquery.Query(doc, "//div[@class='abstract-text']/text()")
	if err != nil {
		return nil, err
	}
	if abstract != nil {
		patent.Abstract = strings.TrimSpace(htmlquery.InnerText(abstract))
	}

	// 主权项
	sovereignty, err := htmlquery.Query(doc, "//div[@class='claim-text']/text()")
	if err != nil {
		return nil, err
	}
	if sovereignty != nil {
		patent.Sovereignty = strings.TrimSpace(htmlquery.InnerText(sovereignty))
	}

	// 融合申请公开号与授权公开号
	// 注：其实这个号就是 publicCode，但是有的是申请公开号，有的是授权公开号
	if patent.ApplyPublicationNo != "" {
		patent.PublicationNo = patent.ApplyPublicationNo
	}
	if patent.AuthPublicationNo != "" {
		patent.PublicationNo = patent.AuthPublicationNo
	}
	if patent.PublicationNo != publicCode {
		return nil, fmt.Errorf("融合申请公开号与授权公开号后，与任务中的公开号匹配失败: "+
			"日期：%s，学科分类%s，任务中的公开号%s，申请公开号：%s ，授权公开号：%s，融合后的公开号：%s",
			date, code, publicCode, patent.ApplyPublicationNo, patent.AuthPublicationNo, patent.PublicationNo)
	}

	return patent, nil
}

func (s *Spider) GetHtml(url string) (string, error) {
	res, body, errs := gorequest.New().Get(url).End()
	if len(errs) > 0 {
		return "", multierr.Combine(errs...)
	}
	if res.StatusCode != 200 {
		return "", fmt.Errorf("请求失败: %s", url)
	}
	return body, nil
}

func (s *Spider) SaveHtml(body, date, code, publicCode string) {
	if err := os.MkdirAll(filepath.Join(HtmlDir, date, code), os.ModePerm); err != nil {
		logrus.Error(err)
		return
	}
	if err := os.WriteFile(filepath.Join(HtmlDir, date, code, publicCode+".html"), []byte(body), os.ModePerm); err != nil {
		logrus.Error(err)
		return
	}
}

func (s *Spider) RandomSleep() {
	// 随机睡眠 minSleepTime ~ maxSleepTime
	sleepTime := time.Duration(rand.Int63n(int64(s.maxSleepTime-s.minSleepTime))) + s.minSleepTime
	time.Sleep(sleepTime)
}

func (s *Spider) WaitForTask() {
	logrus.Info("没有任务，等待 " + s.waitForTaskSleepTime.String())
	time.Sleep(s.waitForTaskSleepTime)
}

func getPatentURL(publicCode string) string {
	return fmt.Sprintf(patentPrefix, publicCode)
}

// 获取被 "," 分隔的发明人，返回前四个
func getFirstFourAuthor(inventors string) (first, second, third, fourth string) {
	inventorsList := strings.Split(inventors, ";")
	length := len(inventorsList)
	if length >= 1 {
		first = inventorsList[0]
	}
	if length >= 2 {
		second = inventorsList[1]
	}
	if length >= 3 {
		third = inventorsList[2]
	}
	if length >= 4 {
		fourth = inventorsList[3]
	}

	return
}
