package spider

import (
	"fmt"
	"testing"
	"time"

	"github.com/parnurzeal/gorequest"
)

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

	res, body, errs := req.Proxy("this_is_proxy").Get("http://dev.kdlapi.com/testproxy").End()
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
