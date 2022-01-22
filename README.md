# mcping

Ping your mincraft servers and save playercounts in prometheus

## Building

```shell
$ go version # installed golang
go version go1.13.4 darwin/amd64
$ go get github.com/lhridder/mcping # install
$ cd mcping
$ go build .
```

## config example

```json
{
  "delay": 60,
  "debug": true,
  "promListen": ":5000",
  "targets": [
    "play.hypixel.net",
    "yourminecraftserver.net"
  ]
}
```

> mcping will lookup SRV record like minecraft clients do.

## Prometheus
### Prometheus configuration:
Example prometheus.yml configuration:
```yaml
scrape_configs:
  - job_name: mcping
    static_configs:
    - targets: ['127.0.0.1:5000']
```

### Metrics:
* mcping_playercount: Number of connected players:
    * **Example response:** `mcping_playercount{host="play.example.net", instance="localhost:5000", job="mcping"} 22`
    * **host:** domain that was pinged.
    * **instance:** what mcping instance supplied this information.
    * **job:** what job was specified in the prometheus configuration.
* mcping_pingdelay: Delay when dialing server:
    * **Example response:** `mcping_pingdelay{host="play.example.net", instance="localhost:5000", job="mcping"} 2.6849949e+07`
    * **host:** domain that was pinged.
    * **instance:** what mcping instance supplied this information.
    * **job:** what job was specified in the prometheus configuration.
