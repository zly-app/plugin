package metrics

import (
	"sync"

	zapp_metrics "github.com/zly-app/zapp/component/metrics"
	"github.com/zly-app/zapp/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type clientCli struct {
	meter metric.Meter

	counterCollector       map[string]zapp_metrics.ICounter // 计数器
	counterCollectorLocker sync.RWMutex

	gaugeCollector       map[string]zapp_metrics.IGauge // 计量器
	gaugeCollectorLocker sync.RWMutex

	histogramCollector       map[string]zapp_metrics.IHistogram // 直方图
	histogramCollectorLocker sync.RWMutex
}

func NewClient(meter metric.Meter) zapp_metrics.Client {
	return &clientCli{
		meter:              meter,
		counterCollector:   make(map[string]zapp_metrics.ICounter),
		gaugeCollector:     make(map[string]zapp_metrics.IGauge),
		histogramCollector: make(map[string]zapp_metrics.IHistogram),
	}
}

func (c *clientCli) RegistryCounter(name, help string, constLabels zapp_metrics.Labels, labels ...string) zapp_metrics.ICounter {
	c.counterCollectorLocker.Lock()
	defer c.counterCollectorLocker.Unlock()

	if _, ok := c.counterCollector[name]; ok {
		log.Fatal("Register the metrics Counter repeatedly", zap.String("name", name), zap.String("help", help))
	}

	counter, err := c.meter.Float64Counter(name, metric.WithDescription(help))
	var ret zapp_metrics.ICounter
	if err != nil {
		log.Warn("Register the metrics Counter fail.", zap.String("name", name), zap.String("help", help), zap.Error(err))
		ret = zapp_metrics.DefNoopClient.Counter(name)
	} else {
		ret = &counterCli{
			name:       name,
			constLabel: genLabels(constLabels),
			counter:    counter,
		}
	}

	c.counterCollector[name] = ret
	return ret

}
func (c *clientCli) Counter(name string) zapp_metrics.ICounter {
	c.counterCollectorLocker.RLock()
	defer c.counterCollectorLocker.RUnlock()

	counter, ok := c.counterCollector[name]
	if !ok {
		log.Warn("metrics Counter not found", zap.String("name", name))
		return zapp_metrics.DefNoopClient.Counter(name)
	}
	return counter
}

func (c *clientCli) RegistryGauge(name, help string, constLabels zapp_metrics.Labels, labels ...string) zapp_metrics.IGauge {
	c.gaugeCollectorLocker.Lock()
	defer c.gaugeCollectorLocker.Unlock()

	if _, ok := c.gaugeCollector[name]; ok {
		log.Fatal("Register the metrics Gauge repeatedly", zap.String("name", name), zap.String("help", help))
	}

	gauge, err := c.meter.Float64Gauge(name, metric.WithDescription(help))
	var ret zapp_metrics.IGauge
	if err != nil {
		log.Warn("Register the metrics Gauge fail.", zap.String("name", name), zap.String("help", help), zap.Error(err))
		ret = zapp_metrics.DefNoopClient.Gauge(name)
	} else {
		ret = &gaugeCli{
			name:       name,
			v:          atomic.NewFloat64(0),
			constLabel: genLabels(constLabels),
			gauge:      gauge,
		}
	}

	c.gaugeCollector[name] = ret
	return ret
}
func (c *clientCli) Gauge(name string) zapp_metrics.IGauge {
	c.gaugeCollectorLocker.RLock()
	defer c.gaugeCollectorLocker.RUnlock()

	gauge, ok := c.gaugeCollector[name]
	if !ok {
		log.Warn("metrics Gauge not found", zap.String("name", name))
		return zapp_metrics.DefNoopClient.Gauge(name)
	}
	return gauge
}

func (c *clientCli) RegistryHistogram(name, help string, buckets []float64, constLabels zapp_metrics.Labels, labels ...string) zapp_metrics.IHistogram {
	c.histogramCollectorLocker.Lock()
	defer c.histogramCollectorLocker.Unlock()

	if _, ok := c.histogramCollector[name]; ok {
		log.Fatal("Register the metrics Histogram repeatedly", zap.String("name", name), zap.String("help", help))
	}

	histogram, err := c.meter.Float64Histogram(name, metric.WithDescription(help), metric.WithExplicitBucketBoundaries(buckets...))
	var ret zapp_metrics.IHistogram
	if err != nil {
		log.Warn("Register the metrics Histogram fail.", zap.String("name", name), zap.String("help", help), zap.Error(err))
		ret = zapp_metrics.DefNoopClient.Histogram(name)
	} else {
		ret = &histogramCli{
			name:       name,
			constLabel: genLabels(constLabels),
			histogram:  histogram,
		}
	}

	c.histogramCollector[name] = ret
	return ret
}
func (c *clientCli) Histogram(name string) zapp_metrics.IHistogram {
	c.histogramCollectorLocker.RLock()
	defer c.histogramCollectorLocker.RUnlock()

	histogram, ok := c.histogramCollector[name]
	if !ok {
		log.Warn("metrics Histogram not found", zap.String("name", name))
		return zapp_metrics.DefNoopClient.Histogram(name)
	}
	return histogram
}

func (c *clientCli) RegistrySummary(name, help string, constLabels zapp_metrics.Labels, labels ...string) zapp_metrics.ISummary {
	log.Warn("Register the metrics Summary fail. is nonsupport.", zap.String("name", name), zap.String("help", help))
	summary := zapp_metrics.DefNoopClient.Summary(name)
	return summary
}
func (c *clientCli) Summary(name string) zapp_metrics.ISummary {
	log.Warn("Get metrics Summary fail. is nonsupport.", zap.String("name", name))
	return zapp_metrics.DefNoopClient.Histogram(name)
}

func genLabels(a ...zapp_metrics.Labels) metric.MeasurementOption {
	count := 0
	for i := range a {
		count += len(a[i])
	}

	cl := make([]attribute.KeyValue, 0, count)
	for _, m := range a {
		for k, v := range m {
			cl = append(cl, attribute.String(k, v))
		}
	}
	return metric.WithAttributeSet(attribute.NewSet(cl...))
}
