# zapp日志收集插件

用于收集zapp项目的日志, 并输出到 honey 或其它地方

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

```yaml
plugins:
  honey: # honey 配置
    Env = 'dev' # 输出时标示的环境名
    #App = '' # 输出时标示的app名, 如果为空则使用默认名
    #Instance = '' # 输出时标示的实例名, 如果为空则使用本地ip
    #StopLogOutput = true # 停止原有的日志输出
    #LogBatchSize = 10000 # 日志批次大小, 累计达到这个大小立即输出一次日志, 不用等待时间
    #AutoRotateTime = 3 # 自动旋转时间(秒), 如果没有达到累计输出批次大小, 在指定时间后也会立即输出
    #MaxRotateThreadNum = 10 # 最大旋转线程数, 表示同时允许多少批次发送到输出设备
    Outputs = 'std' # 输出设备列表, 多个输出设备用半角逗号`,`分隔, 支持 std, honey-http, loki-http

    honey-http: # honey-http 输出器
      #Disable: false # 是否关闭
      #PushAddress: http://127.0.0.1:8080/push # push地址, 示例: http://127.0.0.1:8080/push
      #Compress: zstd # 压缩器名, 可选 raw, gzip, zstd
      #Serializer: msgpack # 序列化器名, 可选 msgpack, json
      #AuthToken: '' # 验证token, 如果设置, 客户端请求header必须带上 token={AuthToken}, 如 token=myAuthToken
      #ReqTimeout: 5 # 请求超时, 单位秒
      #RetryCount: 2 # 请求失败重试次数, 0表示禁用
      #RetryIntervalMs: 2000 # 请求失败重试间隔毫秒数
      #ProxyAddress = '' # 代理地址. 支持 http, https, socks5, socks5h. 示例: socks5://127.0.0.1:1080 socks5://user:pwd@127.0.0.1:1080

    loki-http: # loki-http 输出器
      #Disable = false # 关闭
      #PushAddress = 'http://127.0.0.1:3100/loki/api/v1/push' # push地址, 示例: http://127.0.0.1:3100/loki/api/v1/push
      #EnableCompress = true # 是否启用压缩
      #ReqTimeout = 5 # 请求超时, 单位秒
      #RetryCount = 2 # 请求失败重试次数, 0表示禁用
      #RetryIntervalMs = 2000 # 请求失败重试间隔毫秒数
      #ProxyAddress = '' # 代理地址. 支持 http, https, socks5, socks5h. 示例: socks5://127.0.0.1:1080
      #ProxyUser = '' # 代理用户名
      #ProxyPasswd = '' # 代理用户密码
```
