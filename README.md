# Data-sidecar

This sidecar is configured to run alongside prometheus and aggregate data from it into adaptive thresholds and alerts which are made available to scrape back into prometheus.

## Development

Development requires the [go tools](https://golang.org/doc/install) to be installed and a configured [golang workspace](https://golang.org/doc/code.html#Workspaces).

The `vendor` directory is also managed by the `dep` tool. Please read [the dep README.md](https://github.com/golang/dep/blob/master/README.md)

In non-bash terms, the sidecar is designed to be developed within the gopath and to exclusively rely on the `vendor` subdirectory for dependencies.

### Running on a platform that doesn't support the makefile:

```bash
cd $GOPATH/github.com/open-fresh/data-sidecar
go get
go test ./... -cover
go build
```

### Running
To run the sidecar, run the following command:

```bash
$ ./data-sidecar
```

There are a lot of commandline flags that the sidecar supports.

#### Options
```
Usage of C:\Users\bonch05\go\src\github.com\Fresh-Tracks\data-sidecar\data-sidecar.exe:
  -cleanup int
        time after which a missing series may be garbage collected (seconds) (default 300)
  -lookback int
        empirical lookback window (minutes) (default 60)
  -port int
        port on which to expose metrics (default 8077)
  -prom string
        which prometheus to scrape (default "http://localhost:9090")
  -resolution int
        range query resolution (seconds) (default 10)
```

#### Endpoints

* `/metrics` is the p8s exposition format metrics endpoint. It gives both the sidecar's metrics and all the computed metrics.
* `/dump` dump is essentially `\known`+`\dump` for everything at once. Gives the entire state of the data in the sidecar.
* `/score` takes `data`, a json array representing a time-series, `anomalies` an omitted-or-anything (anything is true) to return anomalies only, and `last` which takes the same idea as `anomalies`  as input. Returns a description of all sidecar outputs (including anomalies or not) at each point of the  time series (or just the last one) according to the sidecar.

## Deployment

This is designed to work nicely in a container, but to get it to compile to the container you need a few special options set on the compiler, so use the `make image` command. This container is built `FROM scratch` so, be forewarned that the container will actively resist debugging. The makefile will handle all of this and, in practice, being a go binary, it can just be executed on whatever platform it was compiled for.

## Configuring target metrics

The FreshTracks Sidecar only analyzes series that contain the label `{ft_target="true"}`.

We recommend analyzing the default [cadvisor container metrics](https://github.com/google/cadvisor) for Kubernetes cluster monitoring.
These metrics are enabled by default in Kubernetes clusters and can be automatically labeled for Sidecar analysis with the following Prometheus config found in `prometheus.yml`:

```yaml
metric_relabel_configs:
- source_labels: [__name__]
  regex: ^container_.+$
  target_label: ft_target
  replacement: true
```

Note that the Sidecar process must be re-started after the Prometheus configuration is changed and metrics using the new labeling rules have been ingested.

### Alternative target metrics

Other metrics can be targeted for Sidecar analysis with this `prometheus.yml` config:

```yaml
metric_relabel_configs:
- source_labels: [__name__]
  regex: ^(metric_name_to_analyze|another_metric_name_to_analyze|...)$
  target_label: ft_target
  replacement: true
```

## Generated Metrics

### Threshold metrics

The threshold metrics are generated as follows:

```yaml
ft_[threshold_type_name]_[origin_metric_name]{
  ...origin_metric_labels
} = Gauge
```

### Anomaly metrics

The anomaly metrics are generated as follows:

```yaml
ft_anomalies {
  ...origin_metric_labels,
  ft_metric="origin_metric_name",
  ft_model="model/rule_name"
} = Gauge
```

### Timing metrics

The timing metrics are generated as follows:

```yaml
ft_model_duration_summary {
  ft_model="model/rule_name"
} = Summary
```

## Kinds of analysis

### Adaptive Thresholds
Adaptive Thresholds use time series to predict acceptable bounds on the current series. We provide limited-lookback mean and standard deviation highways, but you're welcome and encouraged to replace them with whatever you like most!

### Anomalies
These may not be visualizable but communicate about whether or not the the series is behaving as expected. We are using some of the [Nelson Rules](https://en.wikipedia.org/wiki/Nelson_rules). These are raw material for alerts, but are probably not alertworthy on their own.
