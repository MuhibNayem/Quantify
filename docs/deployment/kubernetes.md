# Kubernetes Deployment

Deploy Quantify to any Kubernetes cluster (≥ 1.26) using the Helm charts in
`deploy/helm/`.

---

## Prerequisites

| Tool | Version | Purpose |
| :--- | :--- | :--- |
| kubectl | ≥ 1.26 | Cluster access |
| helm | ≥ 3.12 | Chart deployment |
| jq | any | JSON processing (scripts) |

Cluster access must be configured (`kubectl cluster-info` succeeds).

---

## 1. Create namespaces

```bash
kubectl apply -f deploy/k8s/namespaces.yaml
```

Creates: `quantify` (restricted PSS), `ingress-nginx`, `cert-manager`,
`monitoring`.

---

## 2. Install infrastructure

The `quantify-infrastructure` chart bundles Bitnami subcharts for PostgreSQL
(replication), Redis cluster (6 nodes), RabbitMQ (3-node cluster), and MinIO.

```bash
cd deploy/helm/quantify-infrastructure
helm dependency update
cd -

helm upgrade --install quantify-infrastructure deploy/helm/quantify-infrastructure \
  --namespace quantify \
  --set postgresql.auth.postgresPassword="$(openssl rand -base64 24)" \
  --set postgresql.auth.password="$(openssl rand -base64 24)" \
  --set redis-cluster.password="$(openssl rand -base64 24)" \
  --set rabbitmq.auth.password="$(openssl rand -base64 24)" \
  --set minio.auth.rootPassword="$(openssl rand -base64 24)" \
  --wait --timeout 15m
```

> **Production**: use `existingSecret` references or External Secrets instead
> of `--set` for passwords. See [security.md](./security.md).

Verify:

```bash
kubectl get pods -n quantify -l app.kubernetes.io/instance=quantify-infrastructure
```

---

## 3. Deploy backend

```bash
helm upgrade --install quantify-backend deploy/helm/quantify-backend \
  --namespace quantify \
  --values deploy/helm/quantify-backend/values-dev.yaml \
  --set image.repository=ghcr.io/<org>/quantify-backend \
  --set image.tag=<sha-or-tag> \
  --wait --timeout 5m
```

Environment-specific values files:

| File | Use case |
| :--- | :--- |
| `values.yaml` | Defaults (3 replicas, HPA, network policies) |
| `values-dev.yaml` | Dev cluster (1 replica, debug logging) |
| `values-staging.yaml` | Staging (2 replicas) |
| `values-prod.yaml` | Production (3 replicas, hard anti-affinity, external secrets) |

---

## 4. Deploy AI service

```bash
helm upgrade --install quantify-ai-service deploy/helm/quantify-ai-service \
  --namespace quantify \
  --values deploy/helm/quantify-ai-service/values-dev.yaml \
  --set image.repository=ghcr.io/<org>/quantify-ai-service \
  --set image.tag=<sha-or-tag> \
  --wait --timeout 5m
```

---

## 5. Ingress and TLS

### Install ingress-nginx

```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx --create-namespace \
  -f deploy/k8s/ingress-nginx/values.yaml
```

### Install cert-manager

```bash
helm repo add jetstack https://charts.jetstack.io
helm upgrade --install cert-manager jetstack/cert-manager \
  --namespace cert-manager --create-namespace \
  --set installCRDs=true

kubectl apply -f deploy/k8s/cert-manager/cluster-issuers.yaml
```

This creates `letsencrypt-staging` and `letsencrypt-prod` ClusterIssuers
(HTTP-01 via ingress-nginx). The backend Ingress references
`letsencrypt-prod` by default; update `ingress.tls.issuer` in values to
`letsencrypt-staging` for non-production.

---

## 6. Monitoring

### kube-prometheus-stack

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm upgrade --install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace monitoring --create-namespace \
  -f deploy/k8s/monitoring/kube-prometheus-stack-values.yaml
```

> Change `grafana.adminPassword` in the values file before applying.

### Loki (log aggregation)

```bash
helm repo add grafana https://grafana.github.io/helm-charts
helm upgrade --install loki grafana/loki-stack \
  --namespace monitoring \
  -f deploy/k8s/monitoring/loki-values.yaml
```

### prometheus-adapter (custom HPA metrics)

Bundled with kube-prometheus-stack. To install standalone:

```bash
helm upgrade --install prometheus-adapter prometheus-community/prometheus-adapter \
  --namespace monitoring \
  -f deploy/k8s/monitoring/prometheus-adapter-values.yaml
```

Exposes `http_requests_per_second` as a custom metric for the backend HPA.

---

## 7. Scaling

### Horizontal Pod Autoscaler

Enabled by default in the backend chart:

```yaml
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80
  customMetrics:
    enabled: true
    requestsPerSecondTarget: 100
```

Inspect:

```bash
kubectl get hpa -n quantify
kubectl describe hpa quantify-backend -n quantify
```

### Manual scaling

```bash
kubectl scale deployment quantify-backend -n quantify --replicas=5
```

---

## 8. Upgrading

```bash
helm upgrade --install quantify-backend deploy/helm/quantify-backend \
  --namespace quantify \
  --values deploy/helm/quantify-backend/values-prod.yaml \
  --set image.tag=<new-tag> \
  --wait --timeout 10m
```

Run migrations **before** upgrading if the release includes schema changes:

```bash
kubectl exec -n quantify deploy/quantify-backend -- \
  /app/migrator -cmd up -dsn "$DATABASE_URL"
```

---

## 9. Rollback

### Automatic (CI/CD)

The production pipeline auto-rolls back on canary failure via
`scripts/rollback.sh prod`.

### Manual

```bash
# Find the previous good revision
helm history quantify-backend -n quantify

# Roll back
helm rollback quantify-backend <revision> -n quantify --wait --timeout 5m

# Or use the script
./scripts/rollback.sh prod
```

Verify:

```bash
kubectl rollout status deployment/quantify-backend -n quantify
curl -fsS https://api.quantify.com/ready
```

Schema rollback (only if the migration was destructive):

```bash
kubectl exec -n quantify deploy/quantify-backend -- \
  /app/migrator -cmd down -dsn "$DATABASE_URL"
```

---

## 10. Smoke tests

```bash
./scripts/smoke-test.sh https://api.quantify.com
```

Checks `/health`, `/ready`, `/metrics`, auth flow, product CRUD, and checkout.

---

## Related

- [CI/CD](./ci-cd.md) — automated deployment pipelines
- [Monitoring](./monitoring.md) — dashboards and alerts
- [Security](./security.md) — secrets and network policies
- [Troubleshooting](./troubleshooting.md)
