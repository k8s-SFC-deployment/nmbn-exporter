# Network Metric Between Nodes Exporter

Prometheus Exporter for network metrics between nodes with [iptables](https://netfilter.org/projects/iptables/index.html).

## Deployment

```bash
docker build --build-arg ARCH=amd64 -t euidong/nmbn-exporter:0.0.1 .
```

```bash
docker run --net=host --cap-add=NET_ADMIN -v ${PWD}/example:/config -it euidong/nmbn-exporter:0.0.1 ./nmbn-exporter --config.path=/config/config.example.yaml
```
