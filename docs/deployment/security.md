# Deployment Security

Security controls for deployed Quantify environments. For application-level
security (threat model, auth, CSRF) see [security.md](../security.md).

---

## Secrets management

### Local development

Secrets live in `.env` files (git-ignored). Never commit them.

### Kubernetes — in-chart secrets (dev/staging)

The Helm chart renders a `Secret` from `values.secrets.*`. Acceptable for
non-production; values files must not be committed with real secrets.

### Kubernetes — External Secrets Operator (production)

Production uses [external-secrets.io](https://external-secrets.io) to sync
secrets from an external backend (Vault, AWS Secrets Manager, GCP Secret
Manager, Azure Key Vault) into Kubernetes Secrets.

```bash
helm repo add external-secrets https://charts.external-secrets.io
helm upgrade --install external-secrets external-secrets/external-secrets \
  --namespace external-secrets --create-namespace \
  --set installCRDs=true
```

Apply the ExternalSecret manifest:

```bash
kubectl apply -f deploy/k8s/external-secrets/backend-external-secret.yaml
```

This syncs these keys from `quantify/backend` in the secret store:

| Secret key | Purpose |
| :--- | :--- |
| `DB_PASSWORD` | PostgreSQL |
| `JWT_SECRET` | Access token signing |
| `REFRESH_TOKEN_SECRET` | Refresh token signing |
| `REDIS_PASSWORD` | Redis cluster auth |
| `RABBITMQ_URL` | Broker connection |
| `MINIO_ACCESS_KEY_ID` / `MINIO_SECRET_ACCESS_KEY` | Object storage |
| `SMTP_USER` / `SMTP_PASS` | Email |
| `SSLCOMMERZ_STORE_ID` / `SSLCOMMERZ_STORE_PASSWORD` | Payment gateway |
| `BKASH_API_KEY` / `BKASH_API_SECRET` / `BKASH_USERNAME` / `BKASH_PASSWORD` | bKash |

Enable in the backend chart:

```yaml
externalSecrets:
  enabled: true
  secretStoreName: vault-backend
  secretStoreKind: ClusterSecretStore
  refreshInterval: 1h
  remoteRef: quantify/backend
```

### Cloud-specific secret backends

| Cloud | Backend | IAM mechanism |
| :--- | :--- | :--- |
| AWS | Secrets Manager | IRSA (IAM Roles for Service Accounts) |
| GCP | Secret Manager | Workload Identity |
| Azure | Key Vault | Workload Identity (managed identity) |

### Secret rotation

See [security.md §4](../security.md) for the key rotation runbook. After
rotating in the external store, the ExternalSecret syncs within
`refreshInterval` (default 1h). Force an immediate sync:

```bash
kubectl annotate externalsecret quantify-backend-secrets -n quantify \
  force-sync=$(date +%s) --overwrite
```

---

## Network policies

The `quantify` namespace enforces `restricted` Pod Security Standards.
Network policies are enabled by default in both charts.

### Backend chart

```yaml
networkPolicy:
  enabled: true
  ingressNamespaces:
    - ingress-nginx
  ingressPodLabels:
    app.kubernetes.io/name: ingress-nginx
  allowAiService: true
```

- Ingress: only from ingress-nginx pods + AI service
- Egress: DNS always allowed; add extra rules as needed

### Infrastructure chart

```yaml
networkPolicy:
  enabled: true
  appNamespaces:
    - quantify
```

Restricts PostgreSQL, Redis, RabbitMQ, and MinIO access to pods in the
`quantify` namespace only.

### Verify

```bash
kubectl get networkpolicies -n quantify
kubectl describe networkpolicy quantify-backend -n quantify
```

---

## Pod security standards

The `quantify` namespace enforces `restricted` PSS:

```yaml
pod-security.kubernetes.io/enforce: restricted
pod-security.kubernetes.io/audit: restricted
pod-security.kubernetes.io/warn: restricted
```

Backend pods run as non-root (UID 10001) with:

```yaml
securityContext:
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  runAsNonRoot: true
  capabilities:
    drop: [ALL]
  seccompProfile:
    type: RuntimeDefault
```

Infrastructure namespaces (`ingress-nginx`, `cert-manager`) use `privileged`
PSS as required by their controllers. `monitoring` uses `baseline`.

---

## Image scanning

### CI pipeline

Every image build runs:

1. **Trivy** scan (CRITICAL + HIGH severity, fails the build)
2. **SBOM** generation (SPDX format, uploaded as artifact)
3. **cosign** keyless signing

### Scheduled scans

`security.yml` runs daily at 06:00 UTC:

- `gosec` (Go static security analysis)
- `govulncheck` (Go vulnerability database)
- `trivy` (container image scan, CRITICAL + HIGH + MEDIUM)
- `npm audit` (frontend dependencies)

Results upload to GitHub Security tab (SARIF).

### Admission control (recommended)

Enforce image signatures and vulnerability gates at admission:

```bash
# Kyverno or OPA Gatekeeper policy example:
# - Require cosign verification for ghcr.io/<org>/*
# - Block images with CRITICAL CVEs
```

---

## Compliance checklist

- [ ] All secrets sourced from external secret manager (no `--set` in prod)
- [ ] Network policies enabled and verified
- [ ] Pods run as non-root with read-only filesystem
- [ ] PSS `restricted` enforced on `quantify` namespace
- [ ] Images scanned (Trivy) and signed (cosign)
- [ ] SBOMs generated and archived
- [ ] TLS termination at ingress (cert-manager or cloud LB)
- [ ] HSTS and security headers enabled (backend middleware)
- [ ] CORS restricted to production origins
- [ ] Rate limiting configured on ingress (100 rps)
- [ ] Audit logging enabled for privileged operations
- [ ] Dependency scanning scheduled (gosec, govulncheck, npm audit)
- [ ] Container registries use private repos with pull secrets
- [ ] RBAC: least-privilege ServiceAccounts per workload
- [ ] Secrets rotated per policy (90 days recommended)

---

## Related

- [Application security](../security.md)
- [Kubernetes guide](./kubernetes.md)
- [CI/CD](./ci-cd.md)
