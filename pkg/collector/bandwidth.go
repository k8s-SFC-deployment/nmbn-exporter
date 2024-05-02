package collector

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/coreos/go-iptables/iptables"
	"github.com/k8s-SFC-deployment/nmbn-exporter/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	bandwidth_subsystem = "bandwidth"
	defaultTable        = "filter"
	inChainName         = "IP_TRAFFIC_IN"
	outChainName        = "IP_TRAFFIC_OUT"
)

type bandwidthCollector struct {
	ipt *iptables.IPTables
	cfg *config.Config

	myIP string

	receives  map[string]*bandwidthMetric
	transmits map[string]*bandwidthMetric
}

type bandwidthMetric struct {
	bytes float64
	pkts  float64
}

func init() {
	registerCollector(bandwidth_subsystem, NewBandwidthCollector)
}

// Get target list and create a iptable chain and rules.
func NewBandwidthCollector(cfg *config.Config) (Collector, error) {
	ipt, err := iptables.New()
	if err != nil {
		return nil, err
	}
	clearIptables(cfg, ipt)
	if err := initIptables(cfg, ipt); err != nil {
		return nil, err
	}

	return &bandwidthCollector{cfg: cfg, ipt: ipt}, nil
}

func initIptables(cfg *config.Config, ipt *iptables.IPTables) error {
	var err error
	if err = ipt.NewChain(defaultTable, inChainName); err != nil {
		return err
	}
	if err = ipt.NewChain(defaultTable, outChainName); err != nil {
		return err
	}
	if err = ipt.Append(defaultTable, inChainName, "-j", "RETURN"); err != nil {
		return err
	}
	if err = ipt.Append(defaultTable, outChainName, "-j", "RETURN"); err != nil {
		return err
	}

	for _, target := range cfg.Targets {
		if err = ipt.Append(defaultTable, "INPUT", "-s", target.IP, "-j", inChainName); err != nil {
			return err
		}
		if err = ipt.Append(defaultTable, "OUTPUT", "-d", target.IP, "-j", outChainName); err != nil {
			return err
		}
		fmt.Printf("TARGET[%s] is successfully registered.\n", target.IP)
	}
	return nil
}

func clearIptables(cfg *config.Config, ipt *iptables.IPTables) {
	for _, target := range cfg.Targets {
		ipt.Delete(defaultTable, "INPUT", "-s", target.IP, "-j", inChainName)
		ipt.Delete(defaultTable, "OUTPUT", "-d", target.IP, "-j", outChainName)
	}

	ipt.ClearChain(defaultTable, inChainName)
	ipt.ClearChain(defaultTable, outChainName)
	ipt.DeleteChain(defaultTable, inChainName)
	ipt.DeleteChain(defaultTable, outChainName)
}

func (bc *bandwidthCollector) updateWithIptables() error {
	receives := make(map[string]*bandwidthMetric)
	transmits := make(map[string]*bandwidthMetric)
	for _, target := range bc.cfg.Targets {
		receives[target.IP] = &bandwidthMetric{}
		transmits[target.IP] = &bandwidthMetric{}
	}

	stats, err := bc.ipt.Stats(defaultTable, "INPUT")
	if err != nil {
		return err
	}
	for _, stat := range stats {
		chainName := stat[2]
		if chainName == inChainName {
			pkg, bytes, s_ip, _ := stat[0], stat[1], stat[7], stat[8]
			s_ip = strings.Split(s_ip, "/")[0]
			receives[s_ip].bytes, err = strconv.ParseFloat(bytes, 64)
			if err != nil {
				return err
			}
			receives[s_ip].pkts, err = strconv.ParseFloat(pkg, 64)
			if err != nil {
				return err
			}
		}
	}

	stats, err = bc.ipt.Stats(defaultTable, "OUTPUT")
	if err != nil {
		return err
	}
	for _, stat := range stats {
		chainName := stat[2]
		if chainName == outChainName {
			pkg, bytes, _, d_ip := stat[0], stat[1], stat[7], stat[8]
			d_ip = strings.Split(d_ip, "/")[0]
			transmits[d_ip].bytes, err = strconv.ParseFloat(bytes, 64)
			if err != nil {
				return err
			}
			transmits[d_ip].pkts, err = strconv.ParseFloat(pkg, 64)
			if err != nil {
				return err
			}
		}
	}

	bc.receives = receives
	bc.transmits = transmits

	return nil
}

func (bc *bandwidthCollector) Update(ch chan<- prometheus.Metric) error {
	// 1. get metric with iptables
	if err := bc.updateWithIptables(); err != nil {
		return err
	}

	for _, target := range bc.cfg.Targets {
		// receiveBytes := bc.receives[target.IP].bytes
		// receivePkts := bc.receives[target.IP].pkts
		// transmitBytes := bc.transmits[target.IP].bytes
		// transmitPkts := bc.transmits[target.IP].pkts

		metrics := []float64{bc.receives[target.IP].bytes, bc.receives[target.IP].pkts, bc.transmits[target.IP].bytes, bc.transmits[target.IP].pkts}
		metricNames := []string{"receive_bytes", "receive_packets", "transmit_bytes", "transmit_packets"}
		labels := []prometheus.Labels{{"source": target.IP}, {"source": target.IP}, {"destination": target.IP}, {"destination": target.IP}}

		for i := range 4 {
			ch <- prometheus.MustNewConstMetric(
				makeDesc(metricNames[i], labels[i]),
				prometheus.CounterValue,
				metrics[i],
			)
		}
	}

	return nil
}

func makeDesc(metricName string, labels prometheus.Labels) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(namespace, bandwidth_subsystem, metricName), fmt.Sprintf("%s (%s) %s", fullNamespace, bandwidth_subsystem, metricName), nil, labels)
}
