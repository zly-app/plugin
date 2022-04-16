package honey

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var defHoney *HoneyPlugin
var once sync.Once

type HoneyPlugin struct {
	conf       *Config
	ctxCancel  context.CancelFunc
	logCache   chan []byte // 日志总缓存
	batchCache chan []byte // 批次缓存
	state      int32       // 启动状态 0未启动, 1已启动
}

func (h *HoneyPlugin) isStart() bool {
	return atomic.LoadInt32(&h.state) == 1
}

func (h *HoneyPlugin) Inject(a ...interface{}) {}
func (h *HoneyPlugin) Start() error {
	if h.conf.Service == "" {
		h.conf.Service = zapp.App().Name()
	}
	atomic.StoreInt32(&h.state, 1)

	go h.start()
	return nil
}
func (h *HoneyPlugin) start() {
	ctx, cancel := context.WithCancel(context.Background())
	h.ctxCancel = cancel

	counter := int32(0)
	t := time.NewTicker(time.Second)

	for {
		select {
		case <-ctx.Done():
			t.Stop()
			h.rotateAndReport(true) // 最后一次轮转上报
			return
		case log := <-h.logCache:
			h.batchCache <- log
			// 检查轮转
			if len(h.batchCache) >= h.conf.BatchLen {
				atomic.StoreInt32(&counter, 0)
				h.rotateAndReport(false) // 轮转和上报
			}
		case <-t.C:
			v := atomic.AddInt32(&counter, 1)
			if int(v) >= h.conf.WaitReport {
				atomic.StoreInt32(&counter, 0)
				h.rotateAndReport(false) // 轮转和上报
			}
		}
	}
}

func (h *HoneyPlugin) Close() error {
	atomic.StoreInt32(&h.state, 0)
	h.ctxCancel()
	return nil
}

// 保存日志
func (h *HoneyPlugin) saveLog(log []byte) {
	h.logCache <- log
}

// 轮转和上报
func (h *HoneyPlugin) rotateAndReport(block bool) {
	// 以下代码是单线程的
	cacheSize := len(h.batchCache)
	if cacheSize == 0 {
		return
	}

	data := make([][]byte, cacheSize)
	for i := 0; i < cacheSize; i++ {
		data[i] = <-h.batchCache
	}

	if !block {
		go h.report(data)
	} else {
		h.report(data)
	}
}

// 上报
func (h *HoneyPlugin) report(data [][]byte) {

}

type LogData struct {
	T       int64  `json:"t"`
	Level   string `json:"level"`
	Msg     string `json:"msg,omitempty"`
	Fields  string `json:"fields,omitempty"`
	Line    string `json:"line,omitempty"`
	TraceID string `json:"trace_id,omitempty"`
}

func newHoneyPlugin() *HoneyPlugin {
	once.Do(func() {
		conf := newConfig()

		// 解析配置
		err := config.Conf.ParsePluginConfig(nowPluginType, conf)
		if err == nil {
			err = conf.Check()
		}
		if err != nil {
			logger.Log.Panic("honey配置错误", zap.String("nowPluginType", string(nowPluginType)), zap.Error(err))
		}
		defHoney = &HoneyPlugin{
			conf:       conf,
			logCache:   make(chan []byte, conf.CacheLen),
			batchCache: make(chan []byte, conf.BatchLen),
		}
	})
	return defHoney
}

func (h *HoneyPlugin) Interceptor(ent *zapcore.Entry, fields []zapcore.Field) (cancel bool) {
	// 解析fields
	enc := zapcore.NewMapObjectEncoder()
	for i := range fields {
		fields[i].AddTo(enc)
	}

	// 提取traceID
	traceID := ""
	if v, ok := enc.Fields["traceID"]; ok {
		traceID = fmt.Sprint(v)
		delete(enc.Fields, "traceID")
	}

	data := LogData{
		T:       ent.Time.UnixNano() / 1e3,
		Level:   ent.Level.String(),
		Msg:     ent.Message,
		Fields:  "",
		Line:    "",
		TraceID: traceID,
	}

	// 序列化fields
	if len(enc.Fields) > 1 {
		var fieldsBuff bytes.Buffer
		for k, v := range enc.Fields {
			fieldsBuff.WriteByte(',')
			fieldsBuff.WriteString(k)
			fieldsBuff.WriteByte('=')
			fieldsBuff.WriteString(fmt.Sprint(v))
		}
		data.Fields = string(fieldsBuff.Bytes()[1:])
	}

	bs, _ := jsoniter.Marshal(data)
	h.saveLog(bs)
	return h.conf.StopLogOutput && h.isStart() // 设置了拦截并且在插件启动后才允许拦截
}
