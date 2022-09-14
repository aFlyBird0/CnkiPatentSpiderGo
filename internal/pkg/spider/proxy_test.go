package spider

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/parnurzeal/gorequest"
)

func TestProxy(t *testing.T) {
	// 用户名密码, 若已添加白名单则不需要添加
	username := "t16306616951929"
	password := "r5939x1j"

	// 隧道服务器
	proxy_raw := "e344.kdltps.com:15818"
	proxy_str := fmt.Sprintf("http://%s:%s@%s", username, password, proxy_raw)
	proxy, err := url.Parse(proxy_str)

	// 目标网页
	page_url := "http://dev.kdlapi.com/testproxy"

	//  请求目标网页
	client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxy)}}
	req, _ := http.NewRequest("GET", page_url, nil)
	req.Header.Add("Accept-Encoding", "gzip") //使用gzip压缩传输数据让访问更快
	res, err := client.Do(req)

	if err != nil {
		// 请求发生异常
		fmt.Println(err.Error())
	} else {
		defer res.Body.Close() //保证最后关闭Body

		fmt.Println("status code:", res.StatusCode) // 获取状态码

		// 有gzip压缩时,需要解压缩读取返回内容
		if res.Header.Get("Content-Encoding") == "gzip" {
			reader, _ := gzip.NewReader(res.Body) // gzip解压缩
			defer reader.Close()
			io.Copy(os.Stdout, reader)
		}

		// 无gzip压缩, 读取返回内容
		body, _ := ioutil.ReadAll(res.Body)
		fmt.Println(string(body))
	}
}

func TestProxyGoRequest(t *testing.T) {
	for i := 0; i < 20; i++ {
		doRequest()
		time.Sleep(50 * time.Millisecond)
	}
}

func doRequest() {
	// 和上面的相同，不过使用 gorequest 库
	req := gorequest.New()
	// 用户名密码, 若已添加白名单则不需要添加

	res, body, errs := req.Proxy(getProxyStr()).Get("http://dev.kdlapi.com/testproxy").End()
	if len(errs) > 0 {
		fmt.Println(errs)
		return
	}
	if res.StatusCode != 200 {
		fmt.Println(res.Status)
		return
	}
	fmt.Println(body)
}
