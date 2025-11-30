package metrics

import (
	"context"
	"time"

	"github.com/zly-app/zapp/component/metrics"
	"github.com/zly-app/zapp/pkg/utils"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/atomic"
)

type counterCli struct {
	name       string
	constLabel metric.MeasurementOption
	counter    metric.Float64Counter
}

func (c *counterCli) Inc(labels metrics.Labels, exemplar metrics.Labels) {
	ctx := extractContext(exemplar)
	c.counter.Add(ctx, 1, c.constLabel, genLabels(labels))
}

func (c *counterCli) Add(v float64, labels metrics.Labels, exemplar metrics.Labels) {
	ctx := extractContext(exemplar)
	c.counter.Add(ctx, v, c.constLabel, genLabels(labels))
}

type gaugeCli struct {
	name       string
	v          *atomic.Float64
	constLabel metric.MeasurementOption
	gauge      metric.Float64Gauge
}

func (g *gaugeCli) Set(v float64, labels metrics.Labels) {
	g.v.Store(v)
	g.gauge.Record(context.Background(), v, g.constLabel, genLabels(labels))
}
func (g *gaugeCli) Inc(labels metrics.Labels) {
	nv := g.v.Add(1)
	g.gauge.Record(context.Background(), nv, g.constLabel, genLabels(labels))
}
func (g *gaugeCli) Dec(labels metrics.Labels) {
	nv := g.v.Sub(1)
	g.gauge.Record(context.Background(), nv, g.constLabel, genLabels(labels))
}
func (g *gaugeCli) Add(v float64, labels metrics.Labels) {
	nv := g.v.Add(v)
	g.gauge.Record(context.Background(), nv, g.constLabel, genLabels(labels))
}
func (g *gaugeCli) Sub(v float64, labels metrics.Labels) {
	nv := g.v.Sub(v)
	g.gauge.Record(context.Background(), nv, g.constLabel, genLabels(labels))
}
func (g *gaugeCli) SetToCurrentTime(labels metrics.Labels) {
	t := float64(time.Now().Unix())
	g.v.Store(t)
	g.gauge.Record(context.Background(), t, g.constLabel, genLabels(labels))
}

type histogramCli struct {
	name       string
	constLabel metric.MeasurementOption
	histogram  metric.Float64Histogram
}

func (h *histogramCli) Observe(v float64, labels metrics.Labels, exemplar metrics.Labels) {
	ctx := extractContext(exemplar)
	h.histogram.Record(ctx, v, h.constLabel, genLabels(labels))
}

func extractContext(exemplar metrics.Labels) context.Context {
	if exemplar == nil {
		return context.Background()
	}

	ctx, _ := utils.Otel.GetSpanWithMap(context.Background(), exemplar)
	return ctx
}
