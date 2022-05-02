package honey

import (
	"fmt"
	"strings"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/service"
	"go.uber.org/zap"

	"github.com/zly-app/honey/output"
)

// 生成输出器
func (h *HoneyPlugin) MakeOutput() {
	if h.conf.Outputs == "" {
		return
	}

	names := strings.Split(h.conf.Outputs, ",")
	h.outputs = make(map[string]output.IOutput, len(names))
	for i := range names {
		out := output.MakeOutput(h, names[i])
		h.outputs[names[i]] = out
	}
}

// 解析输出设备配置数据到结构中
func (h *HoneyPlugin) ParseOutputConf(name string, conf interface{}, ignoreNotSet ...bool) error {
	key := fmt.Sprintf("plugins.honey.%s", name)
	return h.app.GetConfig().Parse(key, conf, ignoreNotSet...)
}

// 启动输出设备
func (h *HoneyPlugin) StartOutput() {
	for name, r := range h.outputs {
		err := service.WaitRun(h.app, &service.WaitRunOption{
			ServiceType:        core.ServiceType(DefaultPluginType),
			ExitOnErrOfObserve: true,
			RunServiceFn: func() error {
				err := r.Start()
				if err == nil {
					return nil
				}
				return fmt.Errorf("启动Output失败, Output: %s, err: %v", name, err)
			},
		})
		if err != nil {
			h.app.Fatal("启动Output失败", zap.String("Output", name), zap.Error(err))
		}
	}
}

// 关闭输出设备
func (h *HoneyPlugin) CloseOutput() {
	for name, r := range h.outputs {
		err := r.Close()
		if err != nil {
			h.app.Error("关闭Output失败", zap.String("Output", name), zap.Error(err))
		}
	}
}
