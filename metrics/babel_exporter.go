package metrics

import (
	"github.com/czerwonk/bird_exporter/client"
	"github.com/czerwonk/bird_exporter/protocol"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type babelDesc struct {
	runningDesc       *prometheus.Desc
	routingMetricDesc *prometheus.Desc
	routeCountDesc    *prometheus.Desc
	sourcesCountDesc  *prometheus.Desc
}

type babelMetricExporter struct {
	descriptions map[string]*babelDesc
	client       client.Client
}

// NewBabelExporter creates a new MetricExporter for Babel metrics
func NewBabelExporter(prefix string, client client.Client) MetricExporter {
	d := make(map[string]*babelDesc)
	d["4"] = getBabelDesc(prefix + "babel4")
	d["6"] = getBabelDesc(prefix + "babel6")

	return &babelMetricExporter{descriptions: d, client: client}
}

func getBabelDesc(prefix string) *babelDesc {
	labels := []string{"name"}

	d := &babelDesc{}
	d.runningDesc = prometheus.NewDesc(prefix+"_running", "State of Babel: 0 = Alone, 1 = Running (Neighbor-Adjacencies established)", labels, nil)

	labels = append(labels, "prefix")
	d.routingMetricDesc = prometheus.NewDesc(prefix+"_routing_metric",
		"Minimum cost of a feasible route to the prefix", labels, nil)
	d.routeCountDesc = prometheus.NewDesc(prefix+"_route_count",
		"Number of feasible routes to the prefix", labels, nil)
	d.sourcesCountDesc = prometheus.NewDesc(prefix+"_source_count",
		"Number babel participants advertising external routes to this prefix", labels, nil)

	return d
}

func (m *babelMetricExporter) Describe(ch chan<- *prometheus.Desc) {
	m.describe("4", ch)
	m.describe("6", ch)
}

func (m *babelMetricExporter) describe(ipVersion string, ch chan<- *prometheus.Desc) {
	d := m.descriptions[ipVersion]
	ch <- d.runningDesc
	ch <- d.routingMetricDesc
	ch <- d.routeCountDesc
	ch <- d.sourcesCountDesc
}

func (m *babelMetricExporter) Export(p *protocol.Protocol, ch chan<- prometheus.Metric, newFormat bool) {
	d := m.descriptions[p.IPVersion]
	ch <- prometheus.MustNewConstMetric(d.runningDesc, prometheus.GaugeValue, p.Attributes["running"], p.Name)

	entries, err := m.client.GetBabelEntries(p)
	if err != nil {
		log.Errorln(err)
		return
	}

	for _, entry := range entries {
		l := []string{p.Name, entry.Prefix}
		ch <- prometheus.MustNewConstMetric(d.routingMetricDesc, prometheus.GaugeValue, float64(entry.Metric), l...)
		ch <- prometheus.MustNewConstMetric(d.routeCountDesc, prometheus.GaugeValue, float64(entry.Routes), l...)
		ch <- prometheus.MustNewConstMetric(d.sourcesCountDesc, prometheus.GaugeValue, float64(entry.Sources), l...)
	}
}
