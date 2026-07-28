# Log Shipping

This document describes how to ship application logs to a centralized logging stack.

## Architecture

```
App (JSON stdout) -> Log Shipper -> Loki -> Grafana
```

The backend emits structured JSON logs to stdout via logrus. A log shipper
daemon collects these logs and forwards them to Grafana Loki for storage and
querying.

## Option 1: Vector + Loki

Install [Vector](https://vector.dev/) on each host or run it as a sidecar.

`vector.toml`:

```toml
[sources.app_logs]
type = "docker_logs"
include_containers = ["quantify-backend"]

[transforms.parse]
type = "remap"
inputs = ["app_logs"]
source = '''
. = parse_json!(.message)
'''

[sinks.loki]
type = "loki"
inputs = ["parse"]
endpoint = "http://loki:3100"
encoding.codec = "json"

[sinks.loki.labels]
service = "quantify-backend"
env = "{{ env }}"
```

Run:

```bash
vector --config vector.toml
```

## Option 2: Fluent Bit + Loki

Install [Fluent Bit](https://fluentbit.io/) as a DaemonSet or sidecar.

`fluent-bit.conf`:

```ini
[SERVICE]
    Flush        5
    Daemon       Off
    Log_Level    info
    Parsers_File parsers.conf

[INPUT]
    Name              tail
    Path              /var/log/containers/quantify-backend*.log
    Parser            docker
    Tag               app.*
    Mem_Buf_Limit     5MB
    Skip_Long_Lines   On

[FILTER]
    Name    modify
    Match   app.*
    Add     service quantify-backend

[OUTPUT]
    Name        loki
    Match       app.*
    Host        loki
    Port        3100
    Labels      service=quantify-backend
    Line_Format json
```

`parsers.conf`:

```ini
[PARSER]
    Name        docker
    Format      json
    Time_Key    time
    Time_Format %Y-%m-%dT%H:%M:%S.%L
```

Run:

```bash
fluent-bit -c fluent-bit.conf
```

## Grafana Dashboard

1. Add Loki as a data source in Grafana (`http://loki:3100`).
2. Use LogQL to query:
   - `{service="quantify-backend"}` - all backend logs
   - `{service="quantify-backend"} |= "error"` - errors only
   - `{service="quantify-backend"} | json | level="error"` - structured error filter

## Kubernetes Deployment

For Kubernetes, deploy Vector or Fluent Bit as a DaemonSet with host log volume
mounts. Example Helm charts:

- Vector: `helm install vector https://helm.vector.dev --namespace logging`
- Fluent Bit: `helm install fluent-bit fluent/fluent-bit --namespace logging`

Configure the output to point to your Loki service within the cluster.
