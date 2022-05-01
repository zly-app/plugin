package honey

import (
	"fmt"
	"strings"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/service"
	"go.uber.org/zap"

	"github.com/zly-app/plugin/honey/reporter"
)

// 生成上报器
func (h *HoneyPlugin) MakeReporter() {
	if h.conf.Reports == "" {
		return
	}

	names := strings.Split(h.conf.Reports, ",")
	h.reporters = make(map[string]reporter.IReporter, len(names))
	for i := range names {
		r := reporter.MakeReporter(h.c, names[i])
		h.reporters[names[i]] = r
	}
}

// 启动上报器
func (h *HoneyPlugin) StartReporter() {
	for name, r := range h.reporters {
		err := service.WaitRun(h.app, &service.WaitRunOption{
			ServiceType:        core.ServiceType(DefaultPluginType),
			ExitOnErrOfObserve: true,
			RunServiceFn: func() error {
				err := r.Start()
				if err == nil {
					return nil
				}
				return fmt.Errorf("启动Reporter失败, reporter: %s, err: %v", name, err)
			},
		})
		if err != nil {
			h.app.Fatal("启动Reporter失败", zap.String("reporter", name), zap.Error(err))
		}
	}
}

// 关闭上报器
func (h *HoneyPlugin) CloseReporter() {
	for name, r := range h.reporters {
		err := r.Close()
		if err != nil {
			h.app.Error("关闭Reporter失败", zap.String("Reporter", name), zap.Error(err))
		}
	}
}
