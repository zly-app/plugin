package honey

import (
	"time"

	"github.com/zly-app/zapp/component/gpool"

	"github.com/zly-app/honey/log_data"
	"github.com/zly-app/honey/pkg/rotate"
)

// 生成旋转组
func (h *HoneyPlugin) MakeRotateGroup() {
	h.rotateGPool = gpool.NewGPool(&gpool.GPoolConfig{
		ThreadCount: h.conf.MaxRotateThreadNum,
	})

	opts := []rotate.Option{
		rotate.WithBatchSize(h.conf.LogBatchSize),
		rotate.WithAutoRotateTime(time.Duration(h.conf.AutoRotateTime) * time.Second),
	}
	callback := func(values []interface{}) {
		h.rotateGPool.Go(func() error {
			h.RotateCallback(values)
			return nil
		}, nil)
	}
	h.rotate = rotate.NewRotate(callback, opts...)
}

// 旋转器回调
func (h *HoneyPlugin) RotateCallback(a []interface{}) {
	data := make([]*log_data.LogData, len(a))
	for i, v := range a {
		data[i] = v.(*log_data.LogData)
	}

	env, service, instance := h.conf.Env, h.conf.App, h.conf.Instance

	// 输出
	for _, r := range h.outputs {
		r.Out(env, service, instance, data)
	}
}
