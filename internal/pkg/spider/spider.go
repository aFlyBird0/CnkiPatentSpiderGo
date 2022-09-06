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
	th           TaskHandler
	minSleepTime time.Duration
	maxSleepTime time.Duration
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewSpider(th TaskHandler, minSleepTime, maxSleepTime time.Duration) *Spider {
	// 校验与修正参数
	if minSleepTime > maxSleepTime {
		logrus.Fatal("最小睡眠时间不能大于最大睡眠时间")
	}
	if minSleepTime < time.Millisecond*100 {
		minSleepTime = time.Millisecond * 100
	}
	if maxSleepTime > time.Second*10 {
		maxSleepTime = time.Second * 10
	}
	return &Spider{
		th:           th,
		minSleepTime: minSleepTime,
		maxSleepTime: maxSleepTime,
	}
}

func (s *Spider) Run() {
	continuousErrCount := 0
	for {
		s.RandomSleep()

		if continuousErrCount > 60 {
			logrus.Error("连续60次错误，退出")
			break
		}
		// 从数据库中获取任务
		taskID, publicCode, date, code, err := s.th.GetTask()
		if err != nil {
			logrus.Error(err)
			continuousErrCount++
			continue
		}
		continuousErrCount = 0

		// 解析专利内容
		patent, err := s.ParseContent(date, code, publicCode)
		if err != nil {
			logrus.Error(err)
			continuousErrCount++
			continue
		}
		// 校验合法性
		if !patent.Validate() {
			logrus.Error("数据不合法")
			logrus.Infof("patent: %+v\n", patent)
			continuousErrCount++
			continue
		}
		// 保存到数据库
		save := func() {
			logrus.Infof("保存专利到数据库中: %+v\n", patent)
			if err := s.th.SavePatent(taskID, patent); err != nil {
				logrus.Error(err)
				continuousErrCount++
			}
		}
		go save()
	}
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
	if len(patent.Year) >= 4 {
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
