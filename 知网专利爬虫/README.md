# 知网专利爬虫

本爬虫为分布式， 可在不同的机器上运行（不需要额外配置，自动分配任务），以及可同一机器同时运行多次，可以随时停止。

## 使用方法

在 `bin` 目录下找到适合自己操作系统和芯片架构的二进制文件，

1. 给二进制文件赋予执行权限：`chmod a+x 二进制文件名`
2. 运行： `./二进制文件名 run`

如，MacOS 的 M1 芯片架构，运行：

```bash
chmod a+x spider_darwin_arm64
./spider_darwin_arm64 run`
```

### 二进制文件选择

一般来说：

* Windows 选 `spider_windows_amd64.exe`
* Linux 大多是 `spider_linux_amd64`。
* MacO
  * 英特尔芯片：`spider_darwin_amd64`
  * 苹果芯片：`spider_darwin_arm64`

### 运行说明

把程序放到后台运行，不要关掉。当然也可以使用 `nohup` 一类的命令在后台运行。

还有些参数可以设置，可运行 `./二进制文件名 run --help` 查看。一般情况下不需要设置。

例如设置爬虫两次请求最小间隔时间为 2 秒，设置并发数为2：`./二进制文件名 run --min=2 -c=2`。

## 注意事项

本爬虫无内置网络代理，请自行控制爬取速度（详见参数设置），以免被知网封禁。建议不要所有人都用校园网来爬。

## 并发介绍

* 可以用 `--concurrency=2` 或  `-c=2` 来设置同一个进程的爬虫并发数
* 也可以多次运行程序使用多进程
* 也可以分发到多个电脑来

## 数据存储与返还

* 所有爬到的数据都会在存到 `MySQL` 里。
* 同时还会在本地生成一个 `data` 文件夹，用于记录 html 原始文件还有 log。把 `data` 文件夹打包返还即可

## 源代码

代码已开源，地址 [CnkiPatentSpiderGo](https://github.com/aFlyBird0/CnkiPatentSpiderGo), 欢迎 Star、Fork、PR。
