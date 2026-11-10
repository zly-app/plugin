
otlp on OpenTelemetry

上报 trace / metric / log, 同时支持 **OTLP/HTTP(protobuf)** 与 **OTLP/gRPC** 两种传输.

# 传输协议

`Protocol` 的取值与 OpenTelemetry 规范的 `OTEL_EXPORTER_OTLP_PROTOCOL` 一致:

| 值 | 传输 | 默认端口 | URLPath |
| --- | --- | --- | --- |
| `grpc`(默认) | OTLP/gRPC | `https` 443 / `http` 4317 | 忽略 |
| `http/protobuf` | OTLP/HTTP + protobuf | 4318 | 生效 |

- 插件级 `Protocol` 是全局默认值, `Trace` / `Metric` / `Log` 下的 `Protocol` 可逐个覆盖, 为空时继承插件级配置.
- `Addr` 写法(gRPC):
  - `host:port` — 明文, 端口必填
  - `http://host[:port]` — 明文, 端口缺省 4317
  - `https://host[:port]` — TLS(系统根证书), 端口缺省 443
- `Addr` 写法(HTTP): 必须带 `http://` 或 `https://`, 端口缺省 4318.
- 未配置 `Addr` 时取 `http://localhost:4317`(grpc) 或 `http://localhost:4318`(http/protobuf).
- `Gzip` 在 gRPC 下对应 gRPC 的 gzip 压缩器, 在 HTTP 下对应 `Content-Encoding: gzip`.

# 示例

```go
package main

import (
	"github.com/zly-app/plugin/otlp"
	"github.com/zly-app/zapp"
)

func main() {
	app := zapp.NewApp("test",
		otlp.WithPlugin(),
	)
	defer app.Exit()

	t := otlp.Tracer()
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

客户端与服务端(收集器)的协议必须一致, 例如收集器只开了 4317 的 gRPC 端点时, 客户端要配置 `Protocol: grpc`.

基础配置

```yaml
plugins:
  otlp:
    Protocol: 'grpc' # 全局传输协议, 可选 grpc / http/protobuf

    Trace:
      Addr: 'https://otel.zly.m' # 走 TLS 443, 反代到收集器
    Metric:
      Addr: 'https://otel.zly.m'
    Log:
      Addr: 'https://otel.zly.m'
```

完整配置如下

```yaml
plugins:
  otlp:
    Protocol: 'grpc' # 全局传输协议, 可选 grpc / http/protobuf. 各信号的 Protocol 可覆盖它

    Trace:
      Enabled: true # 是否启用
      Protocol: '' # 传输协议, 为空则使用全局 Protocol
      Addr: 'https://otel.zly.m' # gRPC 如 https://host(443) / localhost:4317; HTTP 如 http://host:4318
      URLPath: '/v1/traces' # path, 仅 http/protobuf 生效
      Gzip: true # 是否启用gzip压缩
      SamplerFraction: 1 # 采样器采样率, <= 0.0 表示不采样, 1.0 表示总是采样
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
      Protocol: '' # 传输协议, 为空则使用全局 Protocol
      Addr: 'https://otel.zly.m' # gRPC 如 https://host(443) / localhost:4317; HTTP 如 http://host:4318
      URLPath: '/v1/metrics' # path, 仅 http/protobuf 生效
      Gzip: true # 是否启用gzip压缩
      ProcessCollector: true # 是否启用进程收集器
      GoCollector: true # 是否启用go收集器
      AutoRotateTime: 5 # 自动旋转时间(秒)
      ExportTimeout: 30 # 上传metric超时时间(秒)
      Retry: # 重试配置
        Enabled: true # 是否启用
        InitialIntervalSec: 5 # 第一次上传失败的重试间隔秒数
        MaxIntervalSec: 30 # 最大重试间隔秒数
        MaxElapsedTimeSec: 60 # 超过这个秒数后则放弃这一批数据

    Log:
      Enabled: true # 是否启用
      Protocol: '' # 传输协议, 为空则使用全局 Protocol
      Addr: 'https://otel.zly.m' # gRPC 如 https://host(443) / localhost:4317; HTTP 如 http://host:4318
      URLPath: '/v1/logs' # path, 仅 http/protobuf 生效
      Gzip: true # 是否启用gzip压缩
      SpanQueueSize: 8192 # 待上传的log队列大小. 超出的log会被丢弃
      SpanBatchSize: 2048 # log信息批次发送大小, 存满后一次性发送
      AutoRotateTime: 5 # 自动旋转时间(秒), 如果没有达到累计输出批次大小, 在指定时间后也会立即输出
      ExportTimeout: 30 # 上传log超时时间(秒)
      Retry: # 重试配置
        Enabled: true # 是否启用
        InitialIntervalSec: 5 # 第一次上传失败的重试间隔秒数
        MaxIntervalSec: 30 # 最大重试间隔秒数
        MaxElapsedTimeSec: 60 # 超过这个秒数后则放弃这一批数据
```
