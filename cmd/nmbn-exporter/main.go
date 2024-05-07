package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/k8s-SFC-deployment/nmbn-exporter/pkg/collector"
	"github.com/k8s-SFC-deployment/nmbn-exporter/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	addr       = kingpin.Flag("web.listen-address", "The Address to listen on for HTTP Requests.").Default("9279").String()
	configFile = kingpin.Flag("config.path", "Path to config file").Default("").String()
)

func main() {
	kingpin.Parse()

	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	c, err := collector.NewNMBNCollector(cfg)
	if err != nil {
		log.Fatal(err)
	}
	prometheus.MustRegister(c)

	http.Handle("/metrics", promhttp.Handler())

	log.Println("Listening on ", ":"+*addr)
	if err := http.ListenAndServe(":"+*addr, nil); err != nil {
		log.Fatal(err)
	}
}

func loadConfig() (*config.Config, error) {
	if *configFile == "" {
		return &config.Config{}, nil
	}
	f, err := os.Open(*configFile)
	if err != nil {
		return nil, fmt.Errorf("cannot load config file: %w", err)
	}
	defer f.Close()

	return config.FromYAML(f)
}
