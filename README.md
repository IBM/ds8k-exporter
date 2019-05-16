# IBM System Storage DS8000 Prometheus Exporter

This [Prometheus](https://prometheus.io) [Exporter](https://prometheus.io/docs/instrumenting/exporters)
collects metrics from [IBM System Storage DS8000](ihttps://www.ibm.com/support/knowledgecenter/en/HW213_7.2.0/com.ibm.storage.ssic.help.doc/f2c_intro_1t1tqx.html).

## Usage

| Flag | Description | Default Value |
| --- | --- | --- |
| --web.telemetry-path | Path under which to expose metrics | /metrics |
| --web.listen-address | Address on which to expose metrics and web interface | :???? |
| --web.disable-exporter-metrics | Exclude metrics about the exporter itself (promhttp_*, process_*, go_*) | false |
| --collector.name | Collector are enabled, the name means name of CLI Command | By default enabled collectors: . |
| --no-collector.name | Collectors that are enabled by default can be disabled, the name means name of CLI Command | By default disabled collectors: . |
| --ssh.targets | Hosts to scrape | - |
| --ssh.user | Username to use when connecting to DS8K using ssh | - |
| --ssh.passwd | Passwd to use when connecting to DS8K using ssh | - |

## Building and running
* Prerequisites:
    * Go compiler
* Building
    * Binary
        ```
        export GOPATH=your_gopath
        cd your_gopath
        mkdir src
        cd src
        mkdir github.com
        cd github.com
        git clone git@github.ibm.com:ZaaS/ds8k-exporter.git
        cd ds8k-exporter
        go build
        ```
    * Docker image
        ``` docker build -t ds8k-exporter . ```
* Running:
    * Run locally
        ```./ds8k-exporter --ssh.targets=X.X.X.X,X.X.X.X --ssh.user=XXX --ssh.passwd=XXX```

    * Run as docker image
        ```docker run -d -p ????:???? --name ds8k-exporter ds8k-exporter --ssh.targets=X.X.X.X --ssh.user=XX --ssh.passwd=XXXX ```
    * Visit http://localhost:????/metrics

## Exported Metrics

| CLI Command | Description | Default | Metrics |
| --- | --- | --- | --- |
| - | Metrics from the exporter itself. | Enabled | [List](docs/exporter_metrics.md) |
| uptime | Displays length of time the system has been operational. | Enabled | [List](docs/uptime_metrics.md) |

## References
* [IBM DS8K RESTful API](https://www-01.ibm.com/support/docview.wss?uid=ssg1S7005173&aid=1)
* [CLI User's Guide](http://www-01.ibm.com/support/docview.wss?uid=ssg1S7002620&aid=1)
* [Monitoring IBM DS8000](https://www.ibm.com/support/knowledgecenter/en/ST5GLJ_8.5.1/com.ibm.storage.ssic.help.doc/ds8_monitoring_uui75.html)

