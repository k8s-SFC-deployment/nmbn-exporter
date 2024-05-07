package collector

import (
	"fmt"
	"sync"
	"time"

	"github.com/k8s-SFC-deployment/nmbn-exporter/pkg/config"
	probing "github.com/prometheus-community/pro-bing"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	latency_subsystem = "current_latency"
)

type latencyCollector struct {
	cfg       *config.Config
	latencies []int64
}

func init() {
	registerCollector(latency_subsystem, NewLatencyCollector)
}

func NewLatencyCollector(cfg *config.Config) (Collector, error) {
	latencis := make([]int64, len(cfg.Targets))

	lc := &latencyCollector{
		cfg:       cfg,
		latencies: latencis,
	}
	go lc.registerPing()
	fmt.Printf("[Latency] TARGETS is successfully registered.\n")
	return lc, nil
}

func (lc *latencyCollector) Update(ch chan<- prometheus.Metric) error {
	for idx, latency := range lc.latencies {
		target_ip := lc.cfg.Targets[idx].IP

		metricName := "microseconds"
		labels := prometheus.Labels{"destination": target_ip}
		desc := makeLatencyDesc(metricName, labels)
		ch <- prometheus.MustNewConstMetric(
			desc,
			prometheus.GaugeValue,
			float64(latency),
		)
	}

	return nil
}

func (lc *latencyCollector) registerPing() {
	ticker := time.NewTicker(time.Duration(lc.cfg.PingInterval) * time.Second)
	for range ticker.C {
		var wg sync.WaitGroup
		var pingers []*probing.Pinger
		for _, target := range lc.cfg.Targets {
			pinger, _ := probing.NewPinger(target.IP)
			pingers = append(pingers, pinger)
		}
		latencies := make([]int64, len(lc.cfg.Targets))
		for idx, pinger := range pingers {
			wg.Add(1)

			pinger.Count = 3
			go getLatencyWithPinger(pinger, &wg, idx, latencies)
		}
		wg.Wait()
		copy(lc.latencies, latencies)
	}
}

func getLatencyWithPinger(pinger *probing.Pinger, wg *sync.WaitGroup, idx int, latencies []int64) {
	defer wg.Done()
	pinger.OnFinish = func(statistics *probing.Statistics) {
		latencies[idx] = statistics.AvgRtt.Microseconds()
	}
	pinger.Run()
}

func makeLatencyDesc(metricName string, labels prometheus.Labels) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(namespace, latency_subsystem, metricName), fmt.Sprintf("%s (%s) %s", fullNamespace, latency_subsystem, metricName), nil, labels)
}
