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
)

type Config struct {
	Address  string // jaeger地址
	User     string // 验证用户名
	Password string // 验证密码

	SamplerType  string  // 采样器类型, const, remote, probabilistic, ratelimiting
	SamplerParam float64 // 采样器参数
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
	return nil
}
