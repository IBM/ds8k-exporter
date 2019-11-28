# IBM System Storage DS8000 Prometheus Exporter

This [Prometheus](https://prometheus.io) [Exporter](https://prometheus.io/docs/instrumenting/exporters)
collects metrics from [IBM System Storage DS8000](ihttps://www.ibm.com/support/knowledgecenter/en/HW213_7.2.0/com.ibm.storage.ssic.help.doc/f2c_intro_1t1tqx.html).

## Usage

| Flag | Description | Default Value |
| --- | --- | --- |
| --config.file | Path to configuration file | ds8k.yaml |
| --web.telemetry-path | Path under which to expose metrics | /metrics |
| --web.listen-address | Address on which to expose metrics and web interface | :9710 |
| --web.disable-exporter-metrics | Exclude metrics about the exporter itself (promhttp_*, process_*, go_*) | false |
| --collector.name | Collector are enabled, the name means name of CLI Command | By default enabled collectors: system, pool,volume,performance. |
| --no-collector.name | Collectors that are enabled by default can be disabled, the name means name of CLI Command | By default disabled collectors: . |

## Building and running
* Prerequisites:
    * Go compiler
* Building
    * Binary
        ```
        export GOPATH=your_gopath
        cd your_gopath
        git clone git@github.ibm.com:ZaaS/ds8k-exporter.git
        cd ds8k-exporter
        go build
        go install (Recommended. This step will copy ds8k-exporter binary package to GOPATH/bin. It's will be connvenient to copy it to Monitoring docker image)
        ```
    * Docker image
        ``` docker build -t ds8k-exporter . ```
* Running:
    * Run locally
        ```./ds8k-exporter --config.file=/etc/ds8k-exporter/ds8k.yaml --location="America/New_York"```

    * Run as docker image
        ```docker run -it -d -p 9710:9710 -v /etc/ds8k-exporter/ds8k.yaml:/etc/ds8k-exporter/ds8k.yaml --name ds8k-exporter ds8k-exporter --config.file=/etc/ds8k-exporter/ds8k.yaml --location="America/New_York"```
    * Visit http://localhost:9710/metrics

    > The examples of location:
       * America/New_York (for Ashburn Data Center)
       * Europe//Paris (for Frankfurt Data Center)
       * Australia/Sydney (for Sydney Data Center)
       * America/Chicago (for Dallas Data Center)

## Configuration
The ds8k-exporter reads from ds8k.yaml config file by default. Edit your config YAML file, Enter the IP address of the storage device, your username and your password there.
```
targets:
  - ipAddress: IP address
    userid: user
    password: password
```

## Exported Metrics

| CLI Command | Description | Default | Metrics |
| --- | --- | --- | --- |
| - | Metrics from the exporter itself. | Enabled | [List](docs/exporter_metrics.md) |
| system | Displays all systems data. | Enabled | [List](docs/system_metrics.md) |
| pool | Displays all pools data. | Enabled | [List](docs/pool_metrics.md) |
| volume |  Displays volumes data. | Enabled | [List](docs/volume_metrics.md) |
| performance | Displays performance summary.| Enabled | [List](docs/performance_metrics.md) |

## References
* [IBM DS8K RESTful API](https://www-01.ibm.com/support/docview.wss?uid=ssg1S7005173&aid=1)
* [CLI User's Guide](http://www-01.ibm.com/support/docview.wss?uid=ssg1S7002620&aid=1)
* [Monitoring IBM DS8000](https://www.ibm.com/support/knowledgecenter/en/ST5GLJ_8.5.1/com.ibm.storage.ssic.help.doc/ds8_monitoring_uui75.html)

