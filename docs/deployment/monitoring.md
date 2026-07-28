# Monitoring & Observability

Quantify ships with Prometheus metrics, Grafana dashboards, Alertmanager
routing, and Loki log aggregation for both Docker Compose and Kubernetes.

---

## Prometheus metrics reference

The backend exposes `/metrics` (Prometheus text format). Key metrics:

| Metric | Type | Description |
| :--- | :--- | :--- |
| `http_requests_total` | counter | Total HTTP requests by method, path, status |
| `http_request_duration_seconds` | histogram | Request latency (buckets: 0.01–10s) |
| `process_resident_memory_bytes` | gauge | Process RSS |
| `go_goroutines` | gauge | Active goroutines |
| `up` | gauge | Scrape target health (1 = up) |

The AI service also exposes `/metrics` on port 8000.

### Scrape configuration

Docker Compose (`monitoring/prometheus/prometheus.yml`):

```yaml
scrape_configs:
  - job_name: "backend"
    static_configs:
      - targets: ["backend:8080"]
  - job_name: "ai-service"
    static_configs:
      - targets: ["ai-service:8000"]
```

Kubernetes: the backend chart creates a `ServiceMonitor` (15s interval)
picked up by kube-prometheus-stack automatically.

---

## Grafana dashboards

### Docker Compose

Grafana runs at http://localhost:3001 (admin/password). Dashboards and
datasources are auto-provisioned from `grafana/provisioning/`:

- `dashboards/backend-dashboard.json` — backend overview
- `datasources/prometheus.yml` — Prometheus datasource

### Kubernetes

kube-prometheus-stack ships Grafana with sidecar dashboard discovery.
Import the backend dashboard JSON via the Grafana UI or provision it as a
ConfigMap labeled `grafana_dashboard: "1"`.

### Key panels

| Panel | Query |
| :--- | :--- |
| Request rate | `sum(rate(http_requests_total[5m]))` |
| Error rate | `sum(rate(http_requests_total{code=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))` |
| P95 latency | `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))` |
| Memory usage | `process_resident_memory_bytes` |
| Goroutines | `go_goroutines` |

---

## Alert rules and escalation

### Alert rules (`monitoring/prometheus/alert-rules.yml`)

| Alert | Condition | Severity | For |
| :--- | :--- | :--- | :--- |
| `BackendDown` | `up{job="backend"} == 0` | critical | 1m |
| `BackendHighErrorRate` | 5xx rate > 10% | warning | 5m |
| `BackendHighLatency` | p95 > 1s | warning | 5m |
| `PostgresDown` | `pg_up == 0` | critical | 1m |
| `RedisDown` | `redis_up == 0` | critical | 1m |
| `RabbitMQDown` | `rabbitmq_up == 0` | critical | 1m |
| `HighMemoryUsage` | RSS > 400MB | warning | 5m |

### Alertmanager routing (`monitoring/alertmanager/alertmanager.yml`)

```
route:
  group_by: [alertname, severity]
  group_wait: 10s
  repeat_interval: 1h
  routes:
    - severity: critical → receiver: critical
    - severity: warning  → receiver: default
```

Inhibit rules: critical alerts suppress matching warning alerts.

### Escalation (production)

| Severity | Channel | Response time |
| :--- | :--- | :--- |
| critical | PagerDuty (via deploy-prod workflow) + Slack | Immediate |
| warning | Slack | Next business day |

Configure Alertmanager receivers for PagerDuty/Slack in production by
updating `alertmanager.yml` or the kube-prometheus-stack Alertmanager config.

---

## Log aggregation with Loki

### Docker Compose

- **Loki**: http://localhost:3100 (7-day retention, filesystem storage)
- **Promtail**: ships Docker container logs to Loki
- Query in Grafana → Explore → Loki datasource

### Kubernetes

```bash
helm upgrade --install loki grafana/loki-stack \
  --namespace monitoring \
  -f deploy/k8s/monitoring/loki-values.yaml
```

- 30-day retention, 50Gi PVC
- Promtail DaemonSet with JSON pipeline for backend logs (extracts `level`)
- Query: `{app="quantify-backend"} |= "error" | json | level="error"`

### Useful LogQL queries

```logql
# All backend errors in the last hour
{app="quantify-backend"} |= "error" | json | level="error"

# Slow requests (> 1s)
{app="quantify-backend"} | json | duration > 1000

# Consumer activity
{app="quantify-backend"} |= "consumer"

# Payment gateway callbacks
{app="quantify-backend"} |= "/payment/"
```

---

## SLI/SLO definitions

| SLI | Measurement | SLO target |
| :--- | :--- | :--- |
| **Availability** | `1 - (5xx / total)` over 30d | ≥ 99.9% |
| **Latency (p95)** | `histogram_quantile(0.95, ...)` | ≤ 500ms |
| **Latency (p99)** | `histogram_quantile(0.99, ...)` | ≤ 2s |
| **Error budget** | `1 - 0.999 = 0.1%` downtime/month | ~43 min/month |

### Burn rate alerts (recommended)

```yaml
- alert: HighErrorBurnRate
  expr: |
    (sum(rate(http_requests_total{code=~"5.."}[1h]))
    / sum(rate(http_requests_total[1h]))) > 0.001
  for: 5m
  labels:
    severity: critical
```

---

## On-call procedures

### Shift handoff

1. Review open alerts in Alertmanager / PagerDuty.
2. Check Grafana dashboards for anomalies.
3. Note any in-progress deploys or maintenance windows.

### Incident response

1. **Acknowledge** the alert in PagerDuty/Alertmanager.
2. **Assess** severity using the runbook ([runbook.md](../runbook.md)).
3. **Mitigate**: restart services, roll back deploy, scale up.
4. **Communicate**: update Slack #incidents channel.
5. **Resolve**: confirm metrics recover, resolve the alert.
6. **Post-mortem**: document timeline, root cause, action items within 48h.

### Quick diagnostics

```bash
# Service health
curl -fsS https://api.quantify.com/ready

# Pod status
kubectl get pods -n quantify -o wide

# Recent logs
kubectl logs -n quantify -l app.kubernetes.io/name=quantify-backend --tail=100 -f

# HPA status
kubectl get hpa -n quantify

# Node pressure
kubectl top nodes
kubectl top pods -n quantify
```

---

## Related

- [Operations runbook](../runbook.md)
- [CI/CD](./ci-cd.md)
- [Troubleshooting](./troubleshooting.md)
