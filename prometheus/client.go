package prometheus

import (
	"context"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/expfmt"
	"github.com/zlyuancn/zretry"
	"go.uber.org/zap"

	"github.com/zly-app/zapp/component/metrics"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/handler"
	"github.com/zly-app/zapp/logger"
)

type Client struct {
	app  core.IApp
	conf *Config

	counterCollector       map[string]metrics.ICounter // 计数器
	counterCollectorLocker sync.RWMutex

	gaugeCollector       map[string]metrics.IGauge // 计量器
	gaugeCollectorLocker sync.RWMutex

	histogramCollector       map[string]metrics.IHistogram // 直方图
	histogramCollectorLocker sync.RWMutex

	summaryCollector       map[string]metrics.ISummary // 汇总
	summaryCollectorLocker sync.RWMutex

	pullRegistry *prometheus.Registry // pull模式注册器
	pusher       *push.Pusher         // push模式推送器
	remoteWrite  *RemoteWrite
}

func (p *Client) Inject(a ...interface{}) {}
func (p *Client) Start() error {
	p.startPullMode(p.conf)
	p.startPushMode(p.conf)
	p.startRemoteWrite(p.conf)
	return nil
}
func (p *Client) Close() error { return nil }

func NewClient(app core.IApp, conf *Config) *Client {
	conf.Check()

	p := &Client{
		app:                app,
		conf:               conf,
		counterCollector:   make(map[string]metrics.ICounter),
		gaugeCollector:     make(map[string]metrics.IGauge),
		histogramCollector: make(map[string]metrics.IHistogram),
		summaryCollector:   make(map[string]metrics.ISummary),
	}

	if conf.PullBind != "" {
		p.pullRegistry = prometheus.NewRegistry()
	}
	if conf.PushAddress != "" {
		p.pusher = push.New(conf.PushAddress, p.app.Name())
	}
	if conf.WriteAddress != "" {
		p.remoteWrite = NewRemoteWrite(conf.WriteAddress)
	}

	coll := []prometheus.Collector{}
	if p.conf.ProcessCollector {
		coll = append(coll, collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	}
	if p.conf.GoCollector {
		coll = append(coll, collectors.NewGoCollector())
	}
	err := p.registryCollector(coll...)
	if err != nil {
		logger.Fatal("注册默认收集器失败", zap.Error(err))
	}
	return p
}

// 启动pull模式
func (p *Client) startPullMode(conf *Config) {
	if p.pullRegistry == nil {
		return
	}

	p.app.Info("启用 metrics pull模式", zap.String("PullBind", conf.PullBind), zap.String("PullPath", conf.PullPath))

	// 构建server
	handle := promhttp.InstrumentMetricHandler(p.pullRegistry, promhttp.HandlerFor(p.pullRegistry, promhttp.HandlerOpts{EnableOpenMetrics: conf.EnableOpenMetrics}))
	mux := http.NewServeMux()
	mux.Handle(conf.PullPath, handle)
	server := &http.Server{Addr: conf.PullBind, Handler: mux}

	handler.AddHandler(handler.AfterExitHandler, func(app core.IApp, handlerType handler.HandlerType) {
		_ = server.Close()
	})
	// 开始监听
	go func(server *http.Server) {
		if err := server.ListenAndServe(); err != nil {
			logger.Log.Fatal("启动pull模式失败", zap.Error(err))
		}
	}(server)
}

// 启动push模式
func (p *Client) startPushMode(conf *Config) {
	if p.pusher == nil {
		return
	}

	if conf.EnableOpenMetrics {
		p.pusher.Format(expfmt.NewFormat(expfmt.TypeOpenMetrics))
	}
	p.pusher.Grouping("app", p.app.Name())
	p.pusher.Grouping("env", p.app.GetConfig().Config().Frame.Env)
	p.pusher.Grouping("instance", conf.PushInstance)

	var defaultDialer = &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	p.pusher.Client(&http.Client{Transport: &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           defaultDialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}})

	p.app.Info("启用 metrics push 模式", zap.String("PushAddress", conf.PushAddress), zap.String("PushInstance", conf.PushInstance))

	// 开始推送
	done, cancel := context.WithCancel(context.Background())
	handler.AddHandler(handler.AfterExitHandler, func(app core.IApp, handlerType handler.HandlerType) {
		cancel()
	})
	go func(ctx context.Context, conf *Config, pusher *push.Pusher) {
		for {
			t := time.NewTimer(time.Duration(conf.PushTimeInterval) * time.Millisecond)
			select {
			case <-ctx.Done():
				t.Stop()
				p.push(conf, pusher) // 最后一次推送
				return
			case <-t.C:
				p.push(conf, pusher)
			}
		}
	}(done, conf, p.pusher)
}

// 启动RemoteWrite模式
func (p *Client) startRemoteWrite(conf *Config) {
	if p.remoteWrite == nil {
		return
	}

	if conf.EnableOpenMetrics {
		p.remoteWrite.FormatType(expfmt.TypeOpenMetrics)
	}

	labels := make([]string, 0, len(p.app.GetConfig().Config().Frame.Labels))
	for k, v := range p.app.GetConfig().Config().Frame.Labels {
		labels = append(labels, k+"="+v)
	}
	sort.Strings(labels)
	p.remoteWrite.ExtraLabel("app", p.app.Name())
	p.remoteWrite.ExtraLabel("env", p.app.GetConfig().Config().Frame.Env)
	p.remoteWrite.ExtraLabel("instance", conf.WriteInstance)

	var defaultDialer = &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	p.remoteWrite.Client(&http.Client{Transport: &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           defaultDialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}})

	p.app.Info("启用 metrics RemoteWrite 模式", zap.String("Url", conf.WriteAddress))

	// 开始写入
	done, cancel := context.WithCancel(context.Background())
	handler.AddHandler(handler.AfterExitHandler, func(app core.IApp, handlerType handler.HandlerType) {
		cancel()
	})
	go func(ctx context.Context, conf *Config, write *RemoteWrite) {
		for {
			t := time.NewTimer(time.Duration(conf.WriteTimeInterval) * time.Millisecond)
			select {
			case <-ctx.Done():
				t.Stop()
				p.write(conf, write) // 最后一次推送
				return
			case <-t.C:
				p.write(conf, write)
			}
		}
	}(done, conf, p.remoteWrite)
}

// 推送
func (p *Client) push(conf *Config, pusher *push.Pusher) {
	zretry.DoRetry(int(conf.PushRetry+1), time.Duration(conf.PushRetryInterval)*time.Millisecond, pusher.Push,
		func(nowAttemptCount, remainCount int, err error) {
			p.app.Error("metrics push 失败", zap.Error(err))
		},
	)
}

// remote write
func (p *Client) write(conf *Config, write *RemoteWrite) {
	err := write.Collect()
	if err != nil {
		p.app.Error("metrics RemoteWrite Collect 失败", zap.Error(err))
		return
	}
	zretry.DoRetry(int(conf.WriteRetry+1), time.Duration(conf.WriteRetryInterval)*time.Millisecond,
		func() error {
			return write.PushLocal()
		},
		func(nowAttemptCount, remainCount int, err error) {
			p.app.Error("metrics RemoteWrite 失败", zap.Error(err))
		},
	)
}

// 注册收集器
func (p *Client) registryCollector(collector ...prometheus.Collector) error {
	if p.pullRegistry != nil {
		for _, coll := range collector {
			err := p.pullRegistry.Register(coll)
			if err != nil {
				return err
			}
		}
	}

	if p.pusher != nil {
		for _, coll := range collector {
			p.pusher.Collector(coll)
		}
	}
	if p.remoteWrite != nil {
		for _, coll := range collector {
			p.remoteWrite.Collector(coll)
		}
	}
	return nil
}

func (p *Client) RegistryCounter(name, help string, constLabels metrics.Labels, labels ...string) metrics.ICounter {
	p.counterCollectorLocker.Lock()
	defer p.counterCollectorLocker.Unlock()

	if _, ok := p.counterCollector[name]; ok {
		p.app.Fatal("重复注册 metrics Counter 计数器", zap.String("name", name))
	}

	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   "",
		Subsystem:   "",
		Name:        name,
		Help:        help,
		ConstLabels: constLabels,
	}, labels)
	err := p.registryCollector(counter)
	if err != nil {
		p.app.Fatal("注册 metrics Counter 计数器失败", zap.Error(err))
	}

	c := &counterCli{name, counter}
	p.counterCollector[name] = c
	return c
}
func (p *Client) Counter(name string) metrics.ICounter {
	p.counterCollectorLocker.RLock()
	defer p.counterCollectorLocker.RUnlock()

	counter, ok := p.counterCollector[name]
	if !ok {
		p.app.Warn("metrics Counter 计数器不存在", zap.String("name", name))
		return metrics.DefNoopClient.Counter(name)
	}
	return counter
}

func (p *Client) RegistryGauge(name, help string, constLabels metrics.Labels, labels ...string) metrics.IGauge {
	p.gaugeCollectorLocker.Lock()
	defer p.gaugeCollectorLocker.Unlock()

	if _, ok := p.gaugeCollector[name]; ok {
		p.app.Fatal("重复注册 metrics Gauge 计量器")
	}

	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   "",
		Subsystem:   "",
		Name:        name,
		Help:        help,
		ConstLabels: constLabels,
	}, labels)
	err := p.registryCollector(gauge)
	if err != nil {
		p.app.Fatal("注册 metrics Gauge 计量器失败", zap.Error(err))
	}

	g := &gaugeCli{name, gauge}
	p.gaugeCollector[name] = g
	return g
}
func (p *Client) Gauge(name string) metrics.IGauge {
	p.gaugeCollectorLocker.RLock()
	defer p.gaugeCollectorLocker.RUnlock()

	gauge, ok := p.gaugeCollector[name]
	if !ok {
		p.app.Warn("metrics Gauge 计量器不存在", zap.String("name", name))
		return metrics.DefNoopClient.Gauge(name)
	}
	return gauge
}

func (p *Client) RegistryHistogram(name, help string, buckets []float64, constLabels metrics.Labels, labels ...string) metrics.IHistogram {
	p.histogramCollectorLocker.Lock()
	defer p.histogramCollectorLocker.Unlock()

	if _, ok := p.histogramCollector[name]; ok {
		p.app.Fatal("重复注册 metrics Histogram 直方图")
	}

	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   "",
		Subsystem:   "",
		Name:        name,
		Help:        help,
		ConstLabels: constLabels,
		Buckets:     buckets,
	}, labels)
	err := p.registryCollector(histogram)
	if err != nil {
		p.app.Fatal("注册 metrics Histogram 直方图失败", zap.Error(err))
	}

	h := &histogramCli{name, histogram}
	p.histogramCollector[name] = h
	return h
}
func (p *Client) Histogram(name string) metrics.IHistogram {
	p.histogramCollectorLocker.RLock()
	defer p.histogramCollectorLocker.RUnlock()

	histogram, ok := p.histogramCollector[name]
	if !ok {
		p.app.Warn("metrics Histogram 直方图不存在", zap.String("name", name))
		return metrics.DefNoopClient.Histogram(name)
	}
	return histogram
}

func (p *Client) RegistrySummary(name, help string, constLabels metrics.Labels, labels ...string) metrics.ISummary {
	p.summaryCollectorLocker.Lock()
	defer p.summaryCollectorLocker.Unlock()

	if _, ok := p.summaryCollector[name]; ok {
		p.app.Fatal("重复注册 metrics Summary 汇总")
	}

	summary := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:   "",
		Subsystem:   "",
		Name:        name,
		Help:        help,
		ConstLabels: constLabels,
	}, labels)
	err := p.registryCollector(summary)
	if err != nil {
		p.app.Fatal("注册 metrics Summary 汇总失败", zap.Error(err))
	}

	s := &summaryCli{name, summary}
	p.summaryCollector[name] = s
	return s
}
func (p *Client) Summary(name string) metrics.ISummary {
	p.summaryCollectorLocker.RLock()
	defer p.summaryCollectorLocker.RUnlock()

	summary, ok := p.summaryCollector[name]
	if !ok {
		p.app.Warn("metrics Summary 汇总不存在", zap.String("name", name))
		return metrics.DefNoopClient.Summary(name)
	}
	return summary
}
