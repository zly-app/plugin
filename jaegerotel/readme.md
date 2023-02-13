
jaeger on OpenTelemetry

# 示例

```go
package main

import (
	"go.opentelemetry.io/otel"

	"github.com/zly-app/plugin/jaegerotel"
	"github.com/zly-app/zapp"
)

func main() {
	app := zapp.NewApp("test",
		jaegerotel.WithPlugin(),
	)
	defer app.Exit()

	t := otel.Tracer("")
	_, span := t.Start(app.BaseContext(), "testA")
	span.End()
}
```

# 添加配置文件 `configs/default.yml`. 更多配置说明参考[这里](./config.go)

```yaml
plugins:
  jaeger:
    AgentHost: '127.0.0.1' # agent Host
    AgentPort: 6831 # agent port

    Endpoint: '' # 收集器地址, 优先级高于 agent, 如 http://localhost:14268/api/traces
    User: '' # 验证用户名
    Password: '' # 验证密码

    SamplerFraction: 1 # // 采样器采样率, <= 0.0 表示不采样, 1.0 表示总是采样
    SpanQueueSize: 4096 # 待上传的span队列大小. 超出的span会被丢弃
    SpanBatchSize: 1024 # span信息批次发送大小, 存满后一次性发送到jaeger
    BlockOnSpanQueueFull: false # 如果span队列满了, 不会丢弃新的span, 而是阻塞直到有空间. 注意, 开启后如果发生阻塞会影响程序性能.
    AutoRotateTime: 5 # 自动旋转时间(秒), 如果没有达到累计输出批次大小, 在指定时间后也会立即输出
    ExportTimeout: 30 # 上传span超时时间(秒)
```
