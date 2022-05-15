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
	isInit           bool  // 是否完成初始化
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
	h.isInit = true
}

// 是否拦截
func (h *HoneyPlugin) isInterceptor() bool {
	return atomic.LoadInt32(&h.interceptorState) == 1
}

func (h *HoneyPlugin) Start() {
	if !h.isInit {
		return
	}

	// 启动输出设备
	h.MakeOutput()
	h.StartOutput()
}

func (h *HoneyPlugin) OnAppStart() {
	atomic.StoreInt32(&h.interceptorState, 1)
}

func (h *HoneyPlugin) BeforeAfterClose() {
	atomic.StoreInt32(&h.interceptorState, 0)

	if !h.isInit {
		return
	}

	// 立即旋转
	h.rotate.Rotate()
}

func (h *HoneyPlugin) AfterClose() {
	if !h.isInit {
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
