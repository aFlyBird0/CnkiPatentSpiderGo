package spider

import (
	"fmt"
)

func getProxyStr() string {
	// 用户名密码, 若已添加白名单则不需要添加
	// 订单已过期，账号密码可放心暴露
	username := "t16329334483765"
	password := "dm9tzng9"

	// 隧道服务器
	proxy_raw := "q580.kdltps.com:15818"
	proxy_str := fmt.Sprintf("http://%s:%s@%s", username, password, proxy_raw)
	return proxy_str
}
