package config

import (
	"github.com/zly-app/zapp"

	"github.com/zly-app/honey/pkg/utils"
)

const (
	// 默认环境名
	DefaultEnv = "dev"
	// 默认的实例名
	DefaultInstance = "default"
	// 停止原有的日志输出
	DefaultStopLogOutput = true

	// 默认批次大小
	DefaultLogBatchSize = 10000
	// 默认自动旋转时间(秒)
	DefaultAutoRotateTime = 3
	// 默认最大旋转线程数
	DefaultMaxRotateThreadNum = 10

	// 默认上报者
	DefaultReports = "std"
)

type Config struct {
	Env           string // 上报时标示的环境名
	Service       string // 上报时标示的服务名, 如果为空则使用app名
	Instance      string // 上报时标示的实例名, 如果为空则使用本地ip
	StopLogOutput bool   // 停止原有的日志输出

	LogBatchSize       int // 日志批次大小, 累计达到这个大小立即上报一次日志, 不用等待时间
	AutoRotateTime     int // 自动旋转时间(秒), 如果没有达到累计上报批次大小, 在指定时间后也会立即上报
	MaxRotateThreadNum int // 最大旋转线程数, 表示同时允许多少批次发送到输出设备

	Reports string // 上报者, 支持 std, http
}

func NewConfig() *Config {
	return &Config{
		StopLogOutput: DefaultStopLogOutput,
	}
}

func (conf *Config) Check() error {
	if conf.Env == "" {
		conf.Env = DefaultEnv
	}
	if conf.Service == "" {
		conf.Service = zapp.App().Name()
	}
	if conf.Instance == "" {
		conf.Instance = utils.GetInstance(DefaultInstance)
	}

	if conf.LogBatchSize < 1 {
		conf.LogBatchSize = DefaultLogBatchSize
	}
	if conf.AutoRotateTime < 1 {
		conf.AutoRotateTime = DefaultAutoRotateTime
	}
	if conf.MaxRotateThreadNum < 1 {
		conf.MaxRotateThreadNum = DefaultMaxRotateThreadNum
	}

	if conf.Reports == "" {
		conf.Reports = DefaultReports
	}
	return nil
}
