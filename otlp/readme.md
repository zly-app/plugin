
otlp on OpenTelemetry

# 示例

```go
package main

import (
	"go.opentelemetry.io/otel"

	"github.com/zly-app/plugin/otlp"
	"github.com/zly-app/zapp"
)

func main() {
	app := zapp.NewApp("test",
		otlp.WithPlugin(),
	)
	defer app.Exit()

	t := otel.Tracer("")
	_, span := t.Start(app.BaseContext(), "testA")
	app.Info("TraceID", span.SpanContext().TraceID())
	app.Info("SpanID", span.SpanContext().SpanID())
	span.End()
}
```

# 兼容 OpenTracing

```go
span := opentracing.StartSpan("test")
span.Finish()
```

# 添加配置文件 `configs/default.yml`. 更多配置说明参考[这里](./config.go)

基础配置

```yaml
plugins:
  otlp:
    Trace:
      Addr: '' # 地址, 如 http://localhost:9411
	Metric:
      Addr: '' # 地址, 如 http://localhost:9411
```

完整配置如下

```yaml
plugins:
  otlp:
    Trace:
      Enabled: true # 是否启用
      Addr: '' # 地址, 如 http://localhost:9411
      Gzip: true # 是否启用gzip压缩
      SamplerFraction: 1 # // 采样器采样率, <= 0.0 表示不采样, 1.0 表示总是采样
      SpanQueueSize: 4096 # 待上传的span队列大小. 超出的span会被丢弃
      SpanBatchSize: 1024 # span信息批次发送大小, 存满后一次性发送到收集器
      BlockOnSpanQueueFull: false # 如果span队列满了, 不会丢弃新的span, 而是阻塞直到有空间. 注意, 开启后如果发生阻塞会影响程序性能.
      AutoRotateTime: 5 # 自动旋转时间(秒), 如果没有达到累计输出批次大小, 在指定时间后也会立即输出
      ExportTimeout: 30 # 上传span超时时间(秒)
      Retry: # 重试配置
        Enabled: true # 是否启用
        InitialIntervalSec: 5 # 第一次上传失败的重试间隔秒数
        MaxIntervalSec: 30 # 最大重试间隔秒数
        MaxElapsedTimeSec: 60 # 超过这个秒数后则放弃这一批数据
	Metric:
      Enabled: true # 是否启用
      Addr: '' # 地址, 如 http://localhost:9411
      Gzip: true # 是否启用gzip压缩
      AutoRotateTime: 5 # 自动旋转时间(秒)
      ExportTimeout: 30 # 上传metric超时时间(秒)
      Retry: # 重试配置
        Enabled: true # 是否启用
        InitialIntervalSec: 5 # 第一次上传失败的重试间隔秒数
        MaxIntervalSec: 30 # 最大重试间隔秒数
        MaxElapsedTimeSec: 60 # 超过这个秒数后则放弃这一批数据
```
