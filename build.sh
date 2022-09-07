# 切换至 main.go 所在目录
cd cmd
# 下载 gox 以一次性交叉编译所有平台
go get github.com/mitchellh/gox
go install github.com/mitchellh/gox
go mod tidy
# 删除旧二进制文件
rm -rf "../知网专利爬虫/bin/"
# 交叉编译到所有平台, 同时忽略 darwin/386 平台，会报错
#gox -output "../知网专利爬虫/bin/spider_{{.OS}}_{{.Arch}}" -osarch='!darwin/386'
# 交叉编译到几个常用的平台
gox -output "../知网专利爬虫/bin/spider_{{.OS}}_{{.Arch}}" -osarch='darwin/amd64 linux/amd64 windows/amd64'

# 单独打包 MacOS 的 M1 芯片版本
go build -o "../知网专利爬虫/bin/spider_darwin_arm64"
