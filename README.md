# Network Metric Between Nodes Exporter

Prometheus Exporter for network metrics between nodes with [iptables](https://github.com/coreos/go-iptables) and [ping](https://github.com/prometheus-community/pro-bing).

## Deployment

```bash
docker build --build-arg ARCH=amd64 -t euidong/nmbn-exporter:0.0.1 .
```

```bash
docker run --net=host --cap-add=NET_ADMIN -v ${PWD}/example:/config -it euidong/nmbn-exporter:0.0.1 ./nmbn-exporter --config.path=/config/config.example.yaml
```

## Trouble shooting

### 1. iptables rules are not removed

when you delete your container directly (unsafely) and also remove at least one target in `cofig.yaml`, you should remove iptables rules manually.

- you can check iptables rules with the following commands.

```bash
sudo iptables -L INPUT
sudo iptables -L OUTPUT
```

- you can remove iptables rules with the following commands.

```bash
sudo iptables -D INPUT -s <ip> -j IP_TRAFFIC_IN
sudo iptables -D OUTPUT -d <ip> -j IP_TRAFFIC_OUT
```
