package constmetrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Ref struct {
	idx          int
	desc         *prometheus.Desc
	labelValue   string
	latestMetric prometheus.Metric // is nil until first Observe() call
}

type Collector struct {
	refs []*Ref
	mu   sync.Mutex
}

func NewCollector() *Collector {
	return &Collector{
		refs: []*Ref{},
	}
}

func (c *Collector) Register(name string, help string, labelKey string, labelValue string) *Ref {
	c.mu.Lock()
	defer c.mu.Unlock()

	idx := len(c.refs)

	c.refs = append(c.refs, &Ref{
		idx:        idx,
		desc:       prometheus.NewDesc(name, help, []string{labelKey}, nil),
		labelValue: labelValue,
	})

	return c.refs[idx]
}

func (c *Collector) Observe(ref *Ref, value float64, ts time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ref.latestMetric = prometheus.NewMetricWithTimestamp(ts, prometheus.MustNewConstMetric(
		ref.desc,
		prometheus.GaugeValue,
		value,
		ref.labelValue))
}

// contract of prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	// unchecked collector
}

// contract of prometheus.Collector
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, ref := range c.refs {
		// first Observe() not called => no collection
		if ref.latestMetric == nil {
			continue
		}

		ch <- ref.latestMetric
	}
}
