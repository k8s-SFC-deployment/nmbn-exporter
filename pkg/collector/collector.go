package collector

import (
	"log"
	"sync"
	"time"

	"github.com/k8s-SFC-deployment/nmbn-exporter/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
)

// Namespace defines the common namespace to be used by all metrics.
const (
	namespace     = "nmbn"
	fullNamespace = "Network Metric Between Nodes"
)

var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
		"nmbn-exporter: Duration of a collector scrape.",
		[]string{"collector"},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_sucess"),
		"nmbn-exporter: Whether a collector succeeded.",
		[]string{"collector"},
		nil,
	)
)

var (
	factories      = make(map[string]func(config *config.Config) (Collector, error))
	collectorState = make(map[string]int)
)

func registerCollector(collector string, f func(cfg *config.Config) (Collector, error)) {
	factories[collector] = f
	collectorState[collector] = 0
}

// Collector is the interface a collector has to implement.
type Collector interface {
	// Get new metrics and expose them via prometheus registry.
	Update(ch chan<- prometheus.Metric) error
}

type NMBNCollector struct {
	Collectors map[string]Collector
}

func NewNMBNCollector(cfg *config.Config) (*NMBNCollector, error) {
	collectors := make(map[string]Collector)
	for k, factory := range factories {
		f, err := factory(cfg)
		if err != nil {
			return nil, err
		}
		collectors[k] = f
	}
	return &NMBNCollector{Collectors: collectors}, nil
}

func (nc *NMBNCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
}

func (nc *NMBNCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(nc.Collectors))
	for name, c := range nc.Collectors {
		go func(name string, c Collector) {
			execute(name, c, ch)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
}

func execute(name string, c Collector, ch chan<- prometheus.Metric) {
	begin := time.Now()
	err := c.Update(ch)
	duration := time.Since(begin)
	var success float64

	if err != nil {
		log.Printf("ERROR: %s collector failed after %fs: %s", name, duration.Seconds(), err.Error())
	}
	success = 1
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name)
}
