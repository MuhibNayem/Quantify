# Deployment Guide

Quantify supports three deployment targets, from local development to
production-grade cloud infrastructure.

| Target | Guide | Stack |
| :--- | :--- | :--- |
| Local development | [local-development.md](./local-development.md) | Docker Compose |
| Kubernetes (any cluster) | [kubernetes.md](./kubernetes.md) | Helm charts |
| AWS | [aws.md](./aws.md) | Terraform + EKS + Helm |
| GCP | [gcp.md](./gcp.md) | Terraform + GKE + Helm |
| Azure | [azure.md](./azure.md) | Terraform + AKS + Helm |

Supporting guides:

- [CI/CD pipelines](./ci-cd.md)
- [Monitoring & observability](./monitoring.md)
- [Security hardening](./security.md)
- [Troubleshooting](./troubleshooting.md)

---

## Architecture at a glance

```
                     ┌──────────────┐
                     │   Ingress    │
                     │  (NGINX/TLS) │
                     └──────┬───────┘
                            │
              ┌─────────────┼─────────────┐
              ▼                           ▼
     ┌────────────────┐         ┌──────────────────┐
     │    backend     │◄───────►│    ai-service    │
     │  Go/Gin :8080  │  REST   │ Python/FastAPI   │
     └───┬──┬──┬──┬───┘         └──────────────────┘
         │  │  │  │
         │  │  │  └──────────► MinIO / S3 (objects)
         │  │  └─────────────► RabbitMQ (async)
         │  └────────────────► Redis cluster (cache/sessions)
         └───────────────────► PostgreSQL (primary + replicas)
```

The backend is a Go monolith (Gin) backed by PostgreSQL, a Redis cluster,
RabbitMQ, and MinIO. The AI service (Python/FastAPI) provides forecasting and
churn prediction. See [architecture.md](../architecture.md) for details.

---

## Quick start

### Local (5 minutes)

```bash
cp .env.example .env
make dev            # docker compose up -d --build
make migrate        # apply schema migrations
make seed           # insert sample data
```

Backend: http://localhost:8080 &middot; Grafana: http://localhost:3001

Full walkthrough: [local-development.md](./local-development.md)

### Kubernetes (30 minutes)

```bash
kubectl apply -f deploy/k8s/namespaces.yaml

helm dependency update deploy/helm/quantify-infrastructure
helm upgrade --install quantify-infrastructure deploy/helm/quantify-infrastructure \
  -n quantify --wait

helm upgrade --install quantify-backend deploy/helm/quantify-backend \
  -n quantify --wait

helm upgrade --install quantify-ai-service deploy/helm/quantify-ai-service \
  -n quantify --wait
```

Full walkthrough: [kubernetes.md](./kubernetes.md)

### Cloud (2+ hours)

```bash
cd infrastructure/aws        # or gcp/ azure/
terraform init
terraform plan  -var="environment=prod"
terraform apply -var="environment=prod"
```

Then deploy the Helm charts to the provisioned cluster. See the
cloud-specific guides for full instructions.

---

## Environment matrix

| | dev | staging | prod |
| :--- | :--- | :--- | :--- |
| Trigger | push to `develop` | push to `staging` | tag `v*` |
| Approval | none | none | GitHub Environment |
| Strategy | rolling | rolling | canary (5% → 100%) |
| Replicas | 1 | 2 | 3 (HPA → 10) |
| Secrets | in-chart | in-chart | External Secrets (Vault) |
| Load tests | no | yes (k6) | no |

---

## Related documents

- [Operations runbook](../runbook.md)
- [Architecture](../architecture.md)
- [Security](../security.md)
