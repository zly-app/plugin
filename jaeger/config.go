package jaeger

import (
	"fmt"

	"github.com/uber/jaeger-client-go"
)

const (
	// 默认api地址
	defaultAddress = "127.0.0.1:6831"

	// 默认采样器类型
	defaultSamplerType = "const"
	// 默认采样器参数
	defaultSamplerParam = 1

	// 默认批次大小
	defSpanBatchSize = 10000
	// 默认自动旋转时间(秒)
	defAutoRotateTime = 3
)

type Config struct {
	Address  string // jaeger地址
	User     string // 验证用户名
	Password string // 验证密码

	SamplerType  string  // 采样器类型, const, remote, probabilistic, ratelimiting
	SamplerParam float64 // 采样器参数

	SpanBatchSize  int // span信息批次发送大小, 存满后一次性发送到jaeger
	AutoRotateTime int // 自动旋转时间(秒), 如果没有达到累计输出批次大小, 在指定时间后也会立即输出
}

func newConfig() *Config {
	return &Config{
		SamplerParam: defaultSamplerParam,
	}
}

func (conf *Config) Check() error {
	if conf.Address == "" {
		conf.Address = defaultAddress
	}

	switch conf.SamplerType {
	case "":
		conf.SamplerType = defaultSamplerType
	case jaeger.SamplerTypeConst:
	case jaeger.SamplerTypeRemote:
	case jaeger.SamplerTypeProbabilistic:
	case jaeger.SamplerTypeRateLimiting:
	default:
		return fmt.Errorf("未定义的SamplerType: %s", conf.SamplerType)
	}

	if conf.SpanBatchSize < 1 {
		conf.SpanBatchSize = defSpanBatchSize
	}
	if conf.AutoRotateTime < 1 {
		conf.AutoRotateTime = defAutoRotateTime
	}
	return nil
}
