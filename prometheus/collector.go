package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/zly-app/zapp/component/metrics"
	"github.com/zly-app/zapp/log"
)

type counterCli struct {
	name    string
	counter *prometheus.CounterVec
}

func (c *counterCli) Inc(labels metrics.Labels, exemplar metrics.Labels) {
	c.Add(1, labels, exemplar)
}

func (c *counterCli) Add(v float64, labels metrics.Labels, exemplar metrics.Labels) {
	counter, err := c.counter.GetMetricWith(labels)
	if err != nil {
		log.Error("获取 metrics Counter 计数器失败", zap.String("name", c.name), zap.Error(err))
		return
	}
	if exemplar != nil {
		if coll, ok := counter.(prometheus.ExemplarAdder); ok {
			coll.AddWithExemplar(v, exemplar)
			return
		}
		log.Error("metrics Counter 计数器无法报告 Exemplar", zap.String("name", c.name), zap.Error(err))
	}
	counter.Add(v)
}

type gaugeCli struct {
	name  string
	gauge *prometheus.GaugeVec
}

func (g *gaugeCli) Set(v float64, labels metrics.Labels) {
	gauge, err := g.gauge.GetMetricWith(labels)
	if err != nil {
		log.Error("获取 metrics Gauge 计量器失败", zap.String("name", g.name), zap.Error(err))
		return
	}
	gauge.Set(v)
}
func (g *gaugeCli) Inc(labels metrics.Labels) {
	gauge, err := g.gauge.GetMetricWith(labels)
	if err != nil {
		log.Error("获取 metrics Gauge 计量器失败", zap.String("name", g.name), zap.Error(err))
		return
	}
	gauge.Inc()
}
func (g *gaugeCli) Dec(labels metrics.Labels) {
	gauge, err := g.gauge.GetMetricWith(labels)
	if err != nil {
		log.Error("获取 metrics Gauge 计量器失败", zap.String("name", g.name), zap.Error(err))
		return
	}
	gauge.Dec()
}
func (g *gaugeCli) Add(v float64, labels metrics.Labels) {
	gauge, err := g.gauge.GetMetricWith(labels)
	if err != nil {
		log.Error("获取 metrics Gauge 计量器失败", zap.String("name", g.name), zap.Error(err))
		return
	}
	gauge.Add(v)
}
func (g *gaugeCli) Sub(v float64, labels metrics.Labels) {
	gauge, err := g.gauge.GetMetricWith(labels)
	if err != nil {
		log.Error("获取 metrics Gauge 计量器失败", zap.String("name", g.name), zap.Error(err))
		return
	}
	gauge.Sub(v)
}
func (g *gaugeCli) SetToCurrentTime(labels metrics.Labels) {
	gauge, err := g.gauge.GetMetricWith(labels)
	if err != nil {
		log.Error("获取 metrics Gauge 计量器失败", zap.String("name", g.name), zap.Error(err))
		return
	}
	gauge.SetToCurrentTime()
}

type histogramCli struct {
	name      string
	histogram *prometheus.HistogramVec
}

func (h *histogramCli) Observe(v float64, labels metrics.Labels, exemplar metrics.Labels) {
	histogram, err := h.histogram.GetMetricWith(labels)
	if err != nil {
		log.Error("获取 metrics Histogram 直方图失败", zap.String("name", h.name), zap.Error(err))
		return
	}
	if exemplar != nil {
		if coll, ok := histogram.(prometheus.ExemplarObserver); ok {
			coll.ObserveWithExemplar(v, exemplar)
			return
		}
		log.Error("metrics Histogram 直方图无法报告 Exemplar", zap.String("name", h.name), zap.Error(err))
	}
	histogram.Observe(v)
}

type summaryCli struct {
	name    string
	summary *prometheus.SummaryVec
}

func (s *summaryCli) Observe(v float64, labels metrics.Labels, exemplar metrics.Labels) {
	summary, err := s.summary.GetMetricWith(labels)
	if err != nil {
		log.Error("获取 metrics Summary 汇总失败", zap.String("name", s.name), zap.Error(err))
		return
	}
	if exemplar != nil {
		if coll, ok := summary.(prometheus.ExemplarObserver); ok {
			coll.ObserveWithExemplar(v, exemplar)
			return
		}
		log.Error("metrics Summary 汇总无法报告 Exemplar", zap.String("name", s.name), zap.Error(err))
	}
	summary.Observe(v)
}
