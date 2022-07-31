
# 示例

```go
package main

import (
	"github.com/opentracing/opentracing-go"

	"github.com/zly-app/plugin/jaeger"
	"github.com/zly-app/zapp"
)

func main() {
	app := zapp.NewApp("test",
		jaeger.WithPlugin(), // 启用 jaeger
	)
	defer app.Exit()

	span := opentracing.StartSpan("testA")
	span.Finish()
}

```

# 添加配置文件 `configs/default.yml`. 更多配置说明参考[这里](./config.go)

```yaml
plugins:
  jaeger:
    Address: 127.0.0.1:6831 # jaeger地址
    User: '' # 验证用户名
    Password: '' # 验证密码
    SamplerType: 'probabilistic' # 采样器类型, const, remote, probabilistic, ratelimiting
    SamplerParam: 1 # 采样器参数
    SpanBatchSize: 10000 # span信息批次发送大小, 存满后一次性发送到jaeger
    AutoRotateTime: 3 # 自动旋转时间(秒), 如果没有达到累计输出批次大小, 在指定时间后也会立即输出
```

# 采样策略

| SamplerType   | SamplerParam | 说明                                                                      |
| ------------- | ------------ | ------------------------------------------------------------------------- |
| const         | 0 或 1       | 采样器始终对所有 tracer 做出相同的决定: 要么全部采样, 要么全部不采样      |
| remote        | 无           | 采样器请求Jaeger代理以获取在当前服务中使用的适当采样策略                  |
| probabilistic | 0.0 ~ 1.0    | 采样器做出随机采样决策, SamplerParam 为采样概率                           |
| ratelimiting  | N            | 采样器一定的恒定速率对tracer进行采样, SamplerParam=2.0, 则限制每秒采集2条 |
