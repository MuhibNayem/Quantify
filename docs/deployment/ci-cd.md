# CI/CD Pipelines

GitHub Actions workflows in `.github/workflows/` automate testing, security
scanning, image builds, and deployments.

---

## Workflow overview

| Workflow | Trigger | Purpose |
| :--- | :--- | :--- |
| `ci.yml` | push to `main`/`develop`/`staging`/`phase-*`, PRs | Lint, test, security scan, build + push images |
| `deploy-dev.yml` | push to `develop`, manual | Deploy to dev cluster |
| `deploy-staging.yml` | push to `staging`, manual | Deploy to staging + integration + load tests |
| `deploy-prod.yml` | tag `v*`, manual | Canary deploy to production |
| `security.yml` | daily cron, PRs, manual | gosec, govulncheck, Trivy, npm audit |
| `terraform.yml` | changes to `infrastructure/**`, manual | fmt, validate, plan, apply |
| `go.yml` | (Go-specific checks) | Go module verification |

---

## CI pipeline (`ci.yml`)

```
lint ──┬──► test ──────────────┐
       ├──► integration-test ──┼──► build-and-push
       └──► security-scan ─────┘
```

### Jobs

1. **lint**: `go vet`, `golangci-lint`, `ruff check` (ai-service)
2. **test**: `go test -race` with PostgreSQL service container; `pytest` for
   ai-service. Uploads coverage artifact.
3. **integration-test**: `go test -tags=integration` with testcontainers
4. **security-scan**: `gosec` (SARIF → GitHub Security), `govulncheck`
5. **build-and-push** (push only, after all pass):
   - Docker Buildx with GHA layer caching
   - Push to GHCR (`ghcr.io/<repo>-backend`, `ghcr.io/<repo>-ai-service`)
   - Tags: branch, SHA, semver
   - **Trivy** scan (CRITICAL/HIGH, fails on findings)
   - **SBOM** generation (SPDX)
   - **cosign** image signing (keyless)

---

## Environment protection rules

| Environment | Protection |
| :--- | :--- |
| `dev` | None (auto-deploy on push to `develop`) |
| `staging` | None (auto-deploy on push to `staging`) |
| `production` | GitHub Environment with required reviewers |
| `terraform-*-apply` | GitHub Environment with required reviewers |

Configure in **Settings → Environments**. Recommended production rules:

- Required reviewers: ≥ 2
- Wait timer: 0 (manual approval gates the deploy)
- Deployment branches: `main` and tags only

---

## Deployment flow

### Dev

```
push to develop → deploy-dev.yml
  → helm upgrade (values-dev.yaml)
  → rollout status check
  → smoke tests
  → Slack notification
```

### Staging

```
push to staging → deploy-staging.yml
  → helm upgrade (values-staging.yaml)
  → rollout status check
  → smoke tests
  → integration tests (against staging URL)
  → k6 load tests
  → upload test reports
  → Slack notification
```

### Production (canary)

```
tag v* → deploy-prod.yml
  → GitHub Environment approval (production)
  → resolve image tag
  → deploy canary (1 replica ≈ 5% traffic)
  → monitor error rate for 10 minutes
      ├── > 1% in ≥ 3/10 checks → FAIL → auto-rollback
      └── healthy → promote: 2 → 3 → full replicas
  → smoke tests (https://api.quantify.com)
  → Slack + PagerDuty (on failure)
```

---

## Canary deployment process

The production deploy uses a progressive rollout:

1. **Canary phase**: deploy 1 replica of the new version alongside existing
   replicas. ~5% of traffic hits the canary.
2. **Observation**: query Prometheus every 60s for 10 minutes:
   ```
   sum(rate(http_requests_total{code=~"5.."}[2m]))
   / sum(rate(http_requests_total[2m]))
   ```
3. **Gate**: if the 5xx rate exceeds 1% in 3+ of 10 checks, the deploy fails
   and triggers auto-rollback.
4. **Promotion**: scale to 2, then 3 replicas (30s between steps), then apply
   the full values-prod.yaml (HPA takes over from 3 → 10).

---

## Rollback procedures

### Automatic

On production deploy failure, the `auto-rollback` job runs
`scripts/rollback.sh prod`, which:

1. Finds the previous `deployed` Helm revision
2. Runs `helm rollback` for both backend and ai-service
3. Waits for rollout completion
4. Verifies pod readiness
5. Sends Slack notification

### Manual

```bash
./scripts/rollback.sh <dev|staging|prod>
```

Or directly:

```bash
helm history quantify-backend -n quantify
helm rollback quantify-backend <revision> -n quantify --wait
```

---

## Monitoring deployment health

After any deploy, verify:

```bash
# Rollout status
kubectl rollout status deployment/quantify-backend -n quantify

# Pod health
kubectl get pods -n quantify -l app.kubernetes.io/name=quantify-backend

# Readiness
curl -fsS https://api.quantify.com/ready

# Error rate (Prometheus)
sum(rate(http_requests_total{namespace="quantify",code=~"5.."}[5m]))
/ sum(rate(http_requests_total{namespace="quantify"}[5m]))

# Smoke tests
./scripts/smoke-test.sh https://api.quantify.com
```

Watch Grafana dashboards for elevated latency or error rates for 30 minutes
post-deploy.

---

## Required repository secrets

| Secret | Used by | Purpose |
| :--- | :--- | :--- |
| `KUBE_CONFIG_DEV` | deploy-dev | kubeconfig for dev cluster |
| `KUBE_CONFIG_STAGING` | deploy-staging | kubeconfig for staging cluster |
| `KUBE_CONFIG_PROD` | deploy-prod | kubeconfig for prod cluster |
| `SLACK_WEBHOOK_URL` | all deploys | Slack notifications |
| `PROMETHEUS_URL` | deploy-prod | Canary health checks |
| `PAGERDUTY_ROUTING_KEY` | deploy-prod | Critical failure paging |
| `AWS_ACCESS_KEY_ID` | terraform | AWS provider auth |
| `AWS_SECRET_ACCESS_KEY` | terraform | AWS provider auth |
| `GOOGLE_CREDENTIALS` | terraform | GCP provider auth |
| `ARM_CLIENT_ID` | terraform | Azure provider auth |
| `ARM_CLIENT_SECRET` | terraform | Azure provider auth |
| `ARM_SUBSCRIPTION_ID` | terraform | Azure provider auth |
| `ARM_TENANT_ID` | terraform | Azure provider auth |

---

## Related

- [Kubernetes guide](./kubernetes.md)
- [Monitoring](./monitoring.md)
- [Troubleshooting](./troubleshooting.md)
