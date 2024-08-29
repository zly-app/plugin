package prometheus

import (
	"github.com/zly-app/zapp/pkg/utils"
)

const (
	defaultProcessCollector = true
	defaultGoCollector = true

	defaultPullPath = "/metrics"

	defaultPushTimeInterval = 10000
	defaultPushRetry = 2
	defaultPushRetryInterval = 1000

	defaultWriteTimeInterval = 10000
	defaultWriteRetry = 2
	defaultWriteRetryInterval = 1000
)

type Config struct {
	ProcessCollector  bool // 启用进程收集器
	GoCollector       bool // 启用go收集器
	EnableOpenMetrics bool // 启用 OpenMetrics 格式

	PullBind string // pull模式bind地址, 如: ':9100', 如果为空则不启用pull模式
	PullPath string // pull模式拉取路径, 如: '/metrics'

	PushAddress string // push模式 pushGateway地址, 如果为空则不启用push模式, 如: 'http://127.0.0.1:9091'
	/*push模式 instance 标记的值
	  这个值用于区分相同服务的不同实例.
	  如果为空则设为主机名, 如果无法获取主机名则设为app名.
	*/
	PushInstance      string // 实例, 一般为ip或主机名
	PushTimeInterval  int64  // push模式推送时间间隔, 单位毫秒
	PushRetry         uint32 // push模式推送重试次数
	PushRetryInterval int64  // push模式推送重试时间间隔, 单位毫秒

	WriteAddress       string // RemoteWrite 地址, 如果为空则不启用
	WriteInstance      string // 实例, 一般为ip或主机名
	WriteTimeInterval  int64  // RemoteWrite 模式推送时间间隔, 单位毫秒
	WriteRetry         uint32 // RemoteWrite 模式推送重试次数
	WriteRetryInterval int64  // RemoteWrite 模式推送重试时间间隔, 单位毫秒
}

func newConfig() *Config {
	return &Config{
		ProcessCollector: defaultProcessCollector,
		GoCollector:      defaultGoCollector,
		PushRetry:        defaultPushRetry,
	}
}

func (conf *Config) Check() {
	if conf.PullPath == "" {
		conf.PullPath = defaultPullPath
	}

	if conf.PushInstance == "" {
		conf.PushInstance = utils.GetInstance("")
	}
	if conf.PushTimeInterval < 1 {
		conf.PushTimeInterval = defaultPushTimeInterval
	}
	if conf.PushRetryInterval < 1 {
		conf.PushRetryInterval = defaultPushRetryInterval
	}

	if conf.WriteInstance == "" {
		conf.WriteInstance = utils.GetInstance("")
	}
	if conf.WriteTimeInterval < 1 {
		conf.WriteTimeInterval = defaultWriteTimeInterval
	}
	if conf.WriteRetryInterval < 1 {
		conf.WriteRetryInterval = defaultWriteRetryInterval
	}
}
