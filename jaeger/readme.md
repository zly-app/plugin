
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

# 采样策略

| SamplerType   | SamplerParam | 说明                                                                      |
| ------------- | ------------ | ------------------------------------------------------------------------- |
| const         | 0 或 1       | 采样器始终对所有 tracer 做出相同的决定: 要么全部采样, 要么全部不采样      |
| remote        | 无           | 采样器请求Jaeger代理以获取在当前服务中使用的适当采样策略                  |
| probabilistic | 0.0 ~ 1.0    | 采样器做出随机采样决策, SamplerParam 为采样概率                           |
| ratelimiting  | N            | 采样器一定的恒定速率对tracer进行采样, SamplerParam=2.0, 则限制每秒采集2条 |
