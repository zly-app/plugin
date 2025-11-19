package metrics

import (
	"context"
	"time"

	"github.com/zly-app/zapp/component/metrics"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/atomic"
)

type counterCli struct {
	name       string
	constAttr  metric.MeasurementOption
	constLabel metric.MeasurementOption
	counter    metric.Float64Counter
}

func (c *counterCli) Inc(labels metrics.Labels, exemplar metrics.Labels) {
	c.Add(1, labels, exemplar)
}

func (c *counterCli) Add(v float64, labels metrics.Labels, exemplar metrics.Labels) {
	c.counter.Add(context.Background(), v, c.constAttr, c.constLabel, genLabels(labels, exemplar))
}

type gaugeCli struct {
	name       string
	v          *atomic.Float64
	constAttr  metric.MeasurementOption
	constLabel metric.MeasurementOption
	gauge      metric.Float64Gauge
}

func (g *gaugeCli) Set(v float64, labels metrics.Labels) {
	g.v.Store(v)
	g.gauge.Record(context.Background(), v, g.constAttr, g.constLabel, genLabels(labels))
}
func (g *gaugeCli) Inc(labels metrics.Labels) {
	nv := g.v.Add(1)
	g.gauge.Record(context.Background(), nv, g.constAttr, g.constLabel, genLabels(labels))
}
func (g *gaugeCli) Dec(labels metrics.Labels) {
	nv := g.v.Sub(1)
	g.gauge.Record(context.Background(), nv, g.constAttr, g.constLabel, genLabels(labels))
}
func (g *gaugeCli) Add(v float64, labels metrics.Labels) {
	nv := g.v.Add(v)
	g.gauge.Record(context.Background(), nv, g.constAttr, g.constLabel, genLabels(labels))
}
func (g *gaugeCli) Sub(v float64, labels metrics.Labels) {
	nv := g.v.Sub(v)
	g.gauge.Record(context.Background(), nv, g.constAttr, g.constLabel, genLabels(labels))
}
func (g *gaugeCli) SetToCurrentTime(labels metrics.Labels) {
	t := float64(time.Now().Unix())
	g.v.Store(t)
	g.gauge.Record(context.Background(), t, g.constAttr, g.constLabel, genLabels(labels))
}

type histogramCli struct {
	name       string
	constAttr  metric.MeasurementOption
	constLabel metric.MeasurementOption
	histogram  metric.Float64Histogram
}

func (h *histogramCli) Observe(v float64, labels metrics.Labels, exemplar metrics.Labels) {
	h.histogram.Record(context.Background(), v, h.constAttr, h.constLabel, genLabels(labels, exemplar))
}
