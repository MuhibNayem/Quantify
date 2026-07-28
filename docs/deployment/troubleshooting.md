# Troubleshooting

Common issues and solutions for deployed Quantify environments.

---

## Pod won't start

### ImagePullBackOff

```bash
kubectl describe pod <pod> -n quantify   # check events
```

**Causes and fixes:**

| Cause | Fix |
| :--- | :--- |
| Wrong image tag | Verify `image.tag` in Helm values matches a pushed tag |
| Private registry auth | Create an `imagePullSecret` and set `imagePullSecrets` in values |
| GHCR permissions | Ensure the GHA workflow pushed the image; check `packages: read` |

```bash
# Create pull secret for GHCR
kubectl create secret docker-registry ghcr-pull \
  --namespace quantify \
  --docker-server=ghcr.io \
  --docker-username=<github-user> \
  --docker-password=<pat>
```

### CrashLoopBackOff

```bash
kubectl logs <pod> -n quantify --previous
```

**Common causes:**

| Symptom in logs | Cause | Fix |
| :--- | :--- | :--- |
| `config validation failed` | Missing/default secrets | Set all required secrets (see below) |
| `dial tcp ... connection refused` | Dependency not ready | Check init containers / dependency pods |
| `permission denied` | Read-only filesystem | Ensure `/tmp` emptyDir is mounted |
| `bind: address already in use` | Port conflict | Check service port configuration |

Required secrets that trigger fail-fast:

- `JWT_SECRET`, `REFRESH_TOKEN_SECRET`
- `MINIO_ACCESS_KEY_ID`, `MINIO_SECRET_ACCESS_KEY`
- `SSLCOMMERZ_STORE_PASSWORD`
- `RABBITMQ_URL`

### Init container stuck

```bash
kubectl logs <pod> -n quantify -c wait-for-dependencies
```

The init container blocks until PostgreSQL, Redis, RabbitMQ, and MinIO accept
TCP connections. Verify the infrastructure pods are running:

```bash
kubectl get pods -n quantify -l app.kubernetes.io/instance=quantify-infrastructure
```

---

## Database connection failures

### Symptoms

- `/ready` reports DB failure
- API returns 5xx
- Logs: `dial tcp: lookup quantify-infrastructure-postgresql: no such host`

### Diagnosis

```bash
# Check PostgreSQL pods
kubectl get pods -n quantify -l app.kubernetes.io/name=postgresql

# Check service
kubectl get svc -n quantify | grep postgresql

# Test connectivity from backend pod
kubectl exec -n quantify deploy/quantify-backend -- \
  nc -zv quantify-infrastructure-postgresql 5432

# Check credentials
kubectl get secret quantify-backend -n quantify -o jsonpath='{.data.DB_PASSWORD}' | base64 -d
```

### Fixes

| Issue | Fix |
| :--- | :--- |
| PostgreSQL pod not running | `kubectl describe pod` → check PVC, resources |
| Wrong host in config | Verify `config.DB_HOST` matches the Service name |
| Wrong credentials | Recreate the secret; restart backend pods |
| Connection pool exhausted | Increase `max_connections` in PostgreSQL extended config |
| PVC pending | Check StorageClass exists and has capacity |

---

## Redis cluster issues

### Symptoms

- `/ready` reports Redis failure
- All authenticated requests return 401/500
- Logs: `CLUSTERDOWN` or `connection refused`

### Diagnosis

```bash
kubectl get pods -n quantify -l app.kubernetes.io/name=redis-cluster

kubectl exec -n quantify quantify-infrastructure-redis-cluster-0 -- \
  redis-cli -a <password> cluster info
```

Look for `cluster_state:ok` and `cluster_slots_assigned:16384`.

### Fixes

| Issue | Fix |
| :--- | :--- |
| Cluster not formed | Delete pods to restart; the Bitnami chart auto-initializes |
| `cluster_state:fail` | Scale down to 0, then back up; PVCs preserve data |
| Wrong password | Verify `REDIS_PASSWORD` secret matches chart `redis-cluster.password` |
| OOMKilled | Increase `resources.limits.memory` in values |

> Restarting Redis **logs out all users** (tokens stored in Redis). This is
> expected.

---

## RabbitMQ connection issues

### Symptoms

- `/ready` reports AMQP failure
- Bulk imports/exports stall, report generation halts
- Logs: `AMQP connection failed`

### Diagnosis

```bash
kubectl get pods -n quantify -l app.kubernetes.io/name=rabbitmq

# Check cluster status
kubectl exec -n quantify quantify-infrastructure-rabbitmq-0 -- \
  rabbitmq-diagnostics cluster_status

# List queues
kubectl exec -n quantify quantify-infrastructure-rabbitmq-0 -- \
  rabbitmqctl list_queues name messages consumers
```

### Fixes

| Issue | Fix |
| :--- | :--- |
| Broker pod not running | Check PVC, resources, Erlang cookie secret |
| Cluster partition | Delete the minority partition pod; it rejoins on restart |
| Wrong URL | Verify `RABBITMQ_URL` secret: `amqp://quantify:<pass>@quantify-infrastructure-rabbitmq:5672/` |
| Queues accumulating | Check consumer count; restart backend to re-subscribe |
| Messages lost during outage | Re-trigger affected operations (re-upload CSV, re-request report) |

---

## MinIO / S3 access issues

### Symptoms

- `/ready` reports MinIO failure
- Bulk upload/download fails, report file downloads fail

### Diagnosis

```bash
kubectl get pods -n quantify -l app.kubernetes.io/name=minio

# Test from backend pod
kubectl exec -n quantify deploy/quantify-backend -- \
  wget -qO- http://quantify-infrastructure-minio:9000/minio/health/live
```

### Fixes

| Issue | Fix |
| :--- | :--- |
| MinIO pod not running | Check PVC capacity (50Gi default) |
| Wrong credentials | Verify `MINIO_ACCESS_KEY_ID` / `MINIO_SECRET_ACCESS_KEY` secrets |
| TLS mismatch | Set `MINIO_USE_TLS=true` only for HTTPS endpoints |
| Missing buckets | Create via MinIO console or `mc mb` |
| PVC full | Expand PVC or add lifecycle rules to prune old artifacts |

---

## High latency / timeouts

### Diagnosis

```bash
# Check p95 latency
kubectl exec -n quantify deploy/quantify-backend -- \
  wget -qO- http://localhost:8080/metrics | grep http_request_duration

# Check resource utilization
kubectl top pods -n quantify
kubectl top nodes

# Check HPA
kubectl get hpa -n quantify

# Check for slow queries
kubectl logs -n quantify -l app.kubernetes.io/name=postgresql --tail=200 | grep "duration"
```

### Fixes

| Cause | Fix |
| :--- | :--- |
| CPU throttling | Increase `resources.limits.cpu` or add replicas |
| HPA at max | Increase `autoscaling.maxReplicas` |
| DB slow queries | Add indexes; check `pg_stat_statements` |
| Redis latency | Check cluster health; increase memory |
| Network policy blocking | Verify egress rules allow dependency access |
| Ingress timeout | Increase `proxy-read-timeout` in ingress-nginx values |

---

## OOMKilled pods

### Diagnosis

```bash
kubectl describe pod <pod> -n quantify | grep -A5 "Last State"
# Look for: Reason: OOMKilled
```

### Fixes

| Component | Default limit | Recommended |
| :--- | :--- | :--- |
| backend | 1Gi | 1.5–2Gi for high-traffic |
| ai-service | 1Gi | 2Gi (model loading) |
| postgresql primary | 1Gi | 2Gi |
| redis-cluster | 512Mi | 1Gi |
| rabbitmq | 1Gi | 2Gi |

```yaml
resources:
  limits:
    memory: 2Gi
```

After increasing limits, set matching `requests` (typically 50–75% of limit)
and restart:

```bash
helm upgrade quantify-backend deploy/helm/quantify-backend -n quantify \
  --reuse-values --set resources.limits.memory=2Gi
```

---

## CrashLoopBackOff (general)

### Systematic approach

1. **Read logs**: `kubectl logs <pod> -n quantify --previous`
2. **Check events**: `kubectl describe pod <pod> -n quantify`
3. **Verify config**: `kubectl get configmap quantify-backend -n quantify -o yaml`
4. **Verify secrets**: `kubectl get secret quantify-backend -n quantify -o jsonpath='{.data}' | jq 'keys'`
5. **Check dependencies**: `/ready` endpoint reports which dependency is down
6. **Resource pressure**: `kubectl top pod <pod> -n quantify`

### Common root causes

| Root cause | Log signature | Fix |
| :--- | :--- | :--- |
| Missing secret | `config validation failed` | Populate all required secrets |
| DB unreachable | `dial tcp ... refused` | Fix PostgreSQL (see above) |
| Redis unreachable | `CLUSTERDOWN` | Fix Redis cluster (see above) |
| Bad migration | `migration ... failed` | Roll back: `migrator -cmd down` |
| OOM | (no log, OOMKilled) | Increase memory limits |
| Liveness probe fail | `Killing container` | Tune `probes.liveness` thresholds |

### Liveness probe tuning

If the backend is slow to start under load:

```yaml
probes:
  startup:
    initialDelaySeconds: 30
    periodSeconds: 10
    failureThreshold: 30    # 30 × 10s = 5 min startup window
  liveness:
    periodSeconds: 15
    timeoutSeconds: 5
    failureThreshold: 3
```

---

## Useful diagnostic commands

```bash
# All pods in namespace
kubectl get pods -n quantify -o wide

# Events sorted by time
kubectl get events -n quantify --sort-by='.lastTimestamp'

# Resource usage
kubectl top pods -n quantify --sort-by=memory

# Port-forward for local debugging
kubectl port-forward -n quantify svc/quantify-backend 8080:8080

# Exec into backend pod
kubectl exec -it -n quantify deploy/quantify-backend -- sh

# Check DNS resolution
kubectl exec -n quantify deploy/quantify-backend -- nslookup quantify-infrastructure-postgresql

# Helm release status
helm list -n quantify
helm status quantify-backend -n quantify
```

---

## Related

- [Operations runbook](../runbook.md)
- [Monitoring](./monitoring.md)
- [Security](./security.md)
