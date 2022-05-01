# zapp日志收集插件

用于收集zapp项目的日志, 并上报到 honey 或其它地方

# 示例

```go
package main

import (
	"github.com/zly-app/plugin/honey"
	"github.com/zly-app/zapp"
)

func main() {
	app := zapp.NewApp("test",
		honey.WithPlugin(), // 启用日志收集插件
	)
	defer app.Exit()

	app.Run()
}
```

# 配置

```toml
# honey日志收集插件配置项
[plugins.honey]
# 上报时标示的环境名
Env = 'dev'
# 上报时标示的服务名, 如果为空则使用app名
#Service = ''
# 上报时标示的实例名, 如果为空则使用本地ip
#Instance = ''
# 停止原有的日志输出
#StopLogOutput = true
# 日志批次大小, 累计达到这个大小立即上报一次日志, 不用等待时间
#LogBatchSize = 10000
# 自动旋转时间(秒), 如果没有达到累计上报批次大小, 在指定时间后也会立即上报
#AutoRotateTime = 3
# 最大旋转线程数, 表示同时允许多少批次发送到输出设备
#MaxRotateThreadNum = 10
# 上报者, 支持 std, http
Reports = 'std'

# http上报者配置项
[plugins.honey.http]
# 关闭
#Disable = false
# 上报地址, 示例: http://127.0.0.1:8080/report
#ReportAddress = 'http://127.0.0.1:8080/report'
# 压缩器名
#Compress = 'zstd'
# 序列化器名
#Serializer = 'msgpack'
# 验证token, 如何设置, 请求header必须带上 token={AuthToken}, 如 token=myAuthToken
#AuthToken = ''
# 上报超时, 单位秒
#ReportTimeout = 5
```
