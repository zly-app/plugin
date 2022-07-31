package honey

import (
	"sync"
	"sync/atomic"

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/zly-app/honey/output"

	"github.com/zly-app/honey/log_data"
	"github.com/zly-app/honey/pkg/rotate"

	"github.com/zly-app/plugin/honey/config"
)

type HoneyPlugin struct {
	app              core.IApp
	conf             *config.Config
	startState       int32 // 启动状态 0未启动, 1已初始化, 2已启动
	interceptorState int32 // 拦截状态 0未启动, 1已启动

	rotate      rotate.IRotator // 旋转器
	rotateGPool core.IGPool     // 用于处理同时旋转的协程池

	outputs map[string]output.IOutput // 输出设备
}

func (h *HoneyPlugin) Init() {
	h.app = zapp.App()
	// 解析配置
	conf := config.NewConfig()
	err := h.app.GetConfig().ParsePluginConfig(DefaultPluginType, conf, true)
	if err == nil {
		err = conf.Check()
	}
	if err != nil {
		h.app.Fatal("honey插件配置错误", zap.String("PluginType", string(DefaultPluginType)), zap.Error(err))
	}
	h.conf = conf

	h.MakeRotateGroup()
	atomic.StoreInt32(&h.startState, 1)
}

// 是否拦截
func (h *HoneyPlugin) isInterceptor() bool {
	return atomic.LoadInt32(&h.interceptorState) == 1
}

func (h *HoneyPlugin) Start() {
	if !atomic.CompareAndSwapInt32(&h.startState, 1, 2) {
		return
	}

	// 启动输出设备
	h.MakeOutput()
	h.StartOutput()
}

func (h *HoneyPlugin) OnAppStart() {
	atomic.StoreInt32(&h.interceptorState, 1)
}

func (h *HoneyPlugin) BeforeClose() {
	atomic.StoreInt32(&h.interceptorState, 0)

	if atomic.LoadInt32(&h.startState) != 2 {
		return
	}

	// 立即旋转
	h.rotate.Rotate()
}

func (h *HoneyPlugin) AfterClose() {
	if !atomic.CompareAndSwapInt32(&h.startState, 2, 1) {
		return
	}

	// 立即旋转
	h.rotate.Rotate()

	// 等待处理
	h.rotateGPool.Wait()

	// 关闭输出设备
	h.CloseOutput()
}

// 日志拦截函数
func (h *HoneyPlugin) LogInterceptorFunc(ent *zapcore.Entry, fields []zapcore.Field) (cancel bool) {
	log := log_data.MakeLogData(ent, fields)
	h.rotate.Add(log)
	if ent.Level == zap.PanicLevel && atomic.LoadInt32(&h.startState) != 2 {
		h.rotate.Rotate()
	}
	if ent.Level == zap.FatalLevel {
		h.rotate.Rotate()
	}
	return h.conf.StopLogOutput && h.isInterceptor()
}

var honey *HoneyPlugin
var once sync.Once

func NewHoneyPlugin() *HoneyPlugin {
	once.Do(func() {
		honey = &HoneyPlugin{}
	})
	return honey
}
