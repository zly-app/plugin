package prometheus

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/prometheus/prompb"
)

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type RemoteWrite struct {
	error      error
	url        string
	gatherers  prometheus.Gatherers
	registerer prometheus.Registerer
	interval   int

	client             HTTPDoer
	useBasicAuth       bool
	username, password string

	expfmt     expfmt.Format
	expfmtType expfmt.FormatType
	labels     map[string]string
	snappyBuf  *bytes.Buffer
}

func NewRemoteWrite(url string) *RemoteWrite {
	var (
		reg = prometheus.NewRegistry()
	)

	if !strings.Contains(url, "://") {
		url = "http://" + url
	}
	url = strings.TrimSuffix(url, "/")

	return &RemoteWrite{
		url:        url + RemoteWriteUri,
		gatherers:  prometheus.Gatherers{reg},
		registerer: reg,
		client:     &http.Client{},
		expfmt:     expfmt.NewFormat(expfmt.TypeProtoDelim),
		expfmtType: expfmt.TypeProtoDelim,
		labels:     map[string]string{},
	}
}

func (p *RemoteWrite) Client(client *http.Client) {
	p.client = client
}

func (p *RemoteWrite) FormatType(t expfmt.FormatType) *RemoteWrite {
	p.expfmt = expfmt.NewFormat(t)
	p.expfmtType = t
	return p
}

func (p *RemoteWrite) BasicAuth(username, password string) *RemoteWrite {
	p.useBasicAuth = true
	p.username = username
	p.password = password
	return p
}

func (p *RemoteWrite) Push() error {
	if err := p.Collect(); err != nil {
		return err
	}
	return p.PushLocal()
}

func (p *RemoteWrite) Add() error {
	if err := p.Collect(); err != nil {
		return err
	}
	return p.AddLocal()
}

func (p *RemoteWrite) PushLocal() error {
	return p.push(http.MethodPost)
}

func (p *RemoteWrite) AddLocal() error {
	return p.push(http.MethodPut)
}

func (p *RemoteWrite) Collect() error {
	if p.error != nil {
		return p.error
	}
	wr, err := p.toPromWriteRequest()
	if err != nil {
		return err
	}
	data, err := proto.Marshal(wr)
	if err != nil {
		return fmt.Errorf("unable to marshal protobuf: %v", err)
	}

	p.snappyBuf = &bytes.Buffer{}
	p.snappyBuf.Write(snappy.Encode(nil, data))
	return nil
}

//func (p *RemoteWrite) Gatherer(g prometheus.Gatherer) *RemoteWrite {
//	p.gatherers = append(p.gatherers, g)
//	return p
//}

func (p *RemoteWrite) ExtraLabel(key, values string) *RemoteWrite {
	p.labels[key] = values
	return p
}

func (p *RemoteWrite) Collector(c prometheus.Collector) *RemoteWrite {
	if p.error == nil {
		p.error = p.registerer.Register(c)
	}
	return p
}

func (p *RemoteWrite) push(method string) error {
	if p.snappyBuf == nil || p.snappyBuf.Len() == 0 {
		return fmt.Errorf("empty snappy buf")
	}

	req, err := http.NewRequest(method, p.url, p.snappyBuf)
	if err != nil {
		return err
	}

	if p.useBasicAuth {
		req.SetBasicAuth(p.username, p.password)
	}

	req.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Content-Encoding", "snappy")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 && resp.StatusCode > 299 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d while pushing to %s: %s", resp.StatusCode, p.url, body)
	}

	return nil
}

func (p *RemoteWrite) toPromWriteRequest() (*prompb.WriteRequest, error) {
	mfs, err := p.gatherers.Gather()
	if err != nil {
		return nil, err
	}

	promTs := make([]prompb.TimeSeries, 0, 16)
	t := time.Now().UnixNano() / 1e6
	for _, mf := range mfs {
		for _, m := range mf.Metric {
			mt := m.GetTimestampMs()
			if mt > 0 {
				t = m.GetTimestampMs()
			}
			//t := m.GetTimestampMs()
			samples := make([]prompb.Sample, 0, 1)
			exemplars := make([]prompb.Exemplar, 0, 1)
			var res []prompb.TimeSeries
			if *mf.Type == io_prometheus_client.MetricType_SUMMARY {
				res = p.parseMetricTypeSummary(mf, m, t)
			} else if *mf.Type == io_prometheus_client.MetricType_HISTOGRAM {
				res = p.parseMetricTypeHistogram(mf, m, t)
			} else {
				labels := p.getMetricLabels(mf, m)
				switch *mf.Type {
				case io_prometheus_client.MetricType_COUNTER:
					samples = append(samples, prompb.Sample{Value: m.GetCounter().GetValue(), Timestamp: t})
					e, ok := p.getPrompbExemplar(m.GetCounter().GetExemplar(), t)
					if ok {
						exemplars = append(exemplars, e)
					}
				case io_prometheus_client.MetricType_GAUGE:
					samples = append(samples, prompb.Sample{Value: m.GetCounter().GetValue(), Timestamp: t})
					e, ok := p.getPrompbExemplar(m.GetCounter().GetExemplar(), t)
					if ok {
						exemplars = append(exemplars, e)
					}
				case io_prometheus_client.MetricType_UNTYPED:
					samples = append(samples, prompb.Sample{Value: m.GetUntyped().GetValue(), Timestamp: t})
				}

				res = []prompb.TimeSeries{{Labels: labels, Samples: samples, Exemplars: exemplars}}
			}
			promTs = append(promTs, res...)
		}

	}

	return &prompb.WriteRequest{
		Timeseries: promTs,
	}, nil
}

func (p *RemoteWrite) getPrompbExemplar(e *io_prometheus_client.Exemplar, t int64) (prompb.Exemplar, bool) {
	ret := prompb.Exemplar{}

	if e == nil {
		return ret, false
	}
	if p.expfmtType != expfmt.TypeOpenMetrics {
		return ret, false
	}
	if e.Value != nil {
		ret.Value = *e.Value
	}
	ret.Timestamp = t
	if e.Timestamp != nil {
		ret.Timestamp = e.Timestamp.AsTime().UnixMilli()
	}
	ret.Labels = make([]prompb.Label, 0, len(e.Label))
	for _, l := range e.Label {
		if l != nil && l.Name != nil && l.Value != nil {
			ret.Labels = append(ret.Labels, prompb.Label{
				Name:  *l.Name,
				Value: l.GetValue(),
			})
		}
	}
	return ret, true
}

func (p *RemoteWrite) parseMetricTypeSummary(mf *io_prometheus_client.MetricFamily, m *io_prometheus_client.Metric, t int64) []prompb.TimeSeries {
	sum := p.getMetricTypeSummarySum(mf, m, t)
	count := p.getMetricTypeSummaryCount(mf, m, t)

	var promTs = []prompb.TimeSeries{sum, count}
	for _, q := range m.Summary.Quantile {
		labels := p.getMetricLabels(mf, m)
		labels = append(labels, prompb.Label{Name: "quantile", Value: fmt.Sprintf("%g", q.GetQuantile())})
		samples := []prompb.Sample{{Value: q.GetValue(), Timestamp: t}}
		promTs = append(promTs, prompb.TimeSeries{Labels: labels, Samples: samples})
	}
	return promTs
}

func (p *RemoteWrite) parseMetricTypeHistogram(mf *io_prometheus_client.MetricFamily, m *io_prometheus_client.Metric, t int64) []prompb.TimeSeries {
	sum := p.getMetricTypeHistogramSum(mf, m, t)
	count := p.getMetricTypeHistogramCount(mf, m, t)

	var promTs = []prompb.TimeSeries{sum, count}
	for _, b := range m.GetHistogram().GetBucket() {
		labels := p.getMetricLabels(mf, m)
		labels = append(labels, prompb.Label{Name: "le", Value: fmt.Sprintf("%g", b.GetUpperBound())})
		samples := []prompb.Sample{{Value: float64(b.GetCumulativeCount()), Timestamp: t}}
		promTs = append(promTs, prompb.TimeSeries{Labels: labels, Samples: samples})
	}
	labels := p.getMetricLabels(mf, m)
	labels = append(labels, prompb.Label{Name: "le", Value: fmt.Sprintf("%g", math.Inf(1))})
	samples := []prompb.Sample{{Value: float64(m.GetHistogram().GetSampleCount()), Timestamp: t}}
	promTs = append(promTs, prompb.TimeSeries{Labels: labels, Samples: samples})
	return []prompb.TimeSeries{sum, count}
}

func (p *RemoteWrite) getMetricTypeHistogramSum(mf *io_prometheus_client.MetricFamily, m *io_prometheus_client.Metric, t int64) prompb.TimeSeries {
	return prompb.TimeSeries{Labels: p.getMetricLabelsNameWithWithSuffix(mf, m, "sum"),
		Samples: []prompb.Sample{{Value: m.GetHistogram().GetSampleSum(), Timestamp: t}}}
}

func (p *RemoteWrite) getMetricTypeHistogramCount(mf *io_prometheus_client.MetricFamily, m *io_prometheus_client.Metric, t int64) prompb.TimeSeries {
	return prompb.TimeSeries{Labels: p.getMetricLabelsNameWithWithSuffix(mf, m, "count"),
		Samples: []prompb.Sample{{Value: float64(m.GetHistogram().GetSampleCount()), Timestamp: t}}}
}

func (p *RemoteWrite) getMetricTypeSummarySum(mf *io_prometheus_client.MetricFamily, m *io_prometheus_client.Metric, t int64) prompb.TimeSeries {
	return prompb.TimeSeries{Labels: p.getMetricLabelsNameWithWithSuffix(mf, m, "sum"),
		Samples: []prompb.Sample{{Value: m.GetSummary().GetSampleSum(), Timestamp: t}}}
}

func (p *RemoteWrite) getMetricTypeSummaryCount(mf *io_prometheus_client.MetricFamily, m *io_prometheus_client.Metric, t int64) prompb.TimeSeries {
	return prompb.TimeSeries{Labels: p.getMetricLabelsNameWithWithSuffix(mf, m, "count"),
		Samples: []prompb.Sample{{Value: float64(m.GetSummary().GetSampleCount()), Timestamp: t}}}
}

func (p *RemoteWrite) getMetricLabels(mf *io_prometheus_client.MetricFamily, m *io_prometheus_client.Metric) []prompb.Label {
	return p.getMetricLabelsNameWithWithSuffix(mf, m, "")
}

func (p *RemoteWrite) getMetricLabelsNameWithWithSuffix(mf *io_prometheus_client.MetricFamily, m *io_prometheus_client.Metric, suffix string) []prompb.Label {
	labels := make([]prompb.Label, 0, 1)
	if suffix != "" {
		if strings.HasPrefix(suffix, "_") {
			suffix = "_" + suffix
		}
	}
	labels = append(labels, prompb.Label{Name: "__name__", Value: mf.GetName() + suffix})
	for _, l := range m.GetLabel() {
		labels = append(labels, prompb.Label{Name: *l.Name, Value: *l.Value})
	}
	for k, v := range p.labels {
		labels = append(labels, prompb.Label{Name: k, Value: v})
	}
	return labels
}
