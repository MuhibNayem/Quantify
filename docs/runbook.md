# Operations Runbook

Operational procedures for the Quantify backend: health checking, incident
response, deployment, and rollback. For architecture see
[architecture.md](./architecture.md); for secrets and key rotation see
[security.md §4](./security.md).

---

## 0. Health & readiness

| Endpoint | Purpose | Fail action |
| :--- | :--- | :--- |
| `GET /health` | Liveness. Returns 200 if the process is up. | Restart the container if down. |
| `GET /ready` | Readiness. Checks **Postgres, Redis, RabbitMQ, and MinIO**. | Investigate the failing dependency (sections below). |
| `GET /metrics` | Prometheus scrape target. | — |

Use `/ready` to decide which dependency is degraded — it reports the failing
component. The Docker Compose stack also defines healthchecks for Postgres
(`pg_isready`), Redis (`redis-cli ping`), and RabbitMQ
(`rabbitmq-diagnostics ping`).

Default local ports: backend `8080`, Postgres `5432`, Redis `6379–6381`
(+ cluster bus `16379–16381`), RabbitMQ `5672` / management `15672`, MinIO
`9000` / console `9001`, Prometheus `9092`, Grafana `3001`.

---

## 1. Common incidents

### 1.1 PostgreSQL down

**Symptoms**: `/ready` reports DB failure; API returns 5xx; consumers stall;
login fails.

**Diagnosis**
```bash
docker compose ps db
docker compose logs --tail=100 db
# From the backend container:
pg_isready -h db -U user -d inventory_db
```

**Recovery**
1. Restart the database: `docker compose up -d db` (it is `restart: always`).
2. Wait for the healthcheck to pass (`pg_isready`).
3. The backend reconnects automatically; verify `GET /ready` is healthy.
4. If the backend holds stale connections and does not recover, restart it:
   `docker compose restart backend`.
5. Check for crashed transactions: any in-flight checkout/return/transfer
   transactions are rolled back by Postgres on disconnect — no manual cleanup
   is needed, but confirm no jobs are stuck (see §1.5).

**Data volume**: `postgres_data`. Back this up; do not delete it.

---

### 1.2 Redis cluster down

**Symptoms**: `/ready` reports Redis failure; **all authenticated requests fail
with 401/500** (the access-token whitelist lives in Redis); permission checks
fail; report caching degrades.

**Diagnosis**
```bash
docker compose ps redis1 redis2 redis3
docker compose logs --tail=100 redis1
docker compose exec redis1 redis-cli -a redispass ping
docker compose exec redis1 redis-cli -a redispass cluster info
```

**Recovery**
1. Restart the nodes: `docker compose up -d redis1 redis2 redis3`.
2. If the cluster is broken (`cluster_state:fail`), re-run the initializer:
   `docker compose up -d redis-init` (it recreates the cluster with
   `--cluster-replicas 0`).
3. Verify `cluster info` shows `cluster_state:ok`.
4. Confirm `GET /ready` is healthy.

**Impact notes**
- Restarting Redis **logs out every user**: access/refresh tokens and CSRF
  tokens are stored in Redis and are lost. This is expected — users re-login.
- The permission cache (60s TTL) rebuilds on demand; no action needed.
- Report caches rebuild lazily.

---

### 1.3 RabbitMQ down

**Symptoms**: `/ready` reports AMQP failure; **async work stops** — bulk
imports/exports stall, report generation halts, alert fan-out stops. The
synchronous API (checkout, returns) still works because it does not depend on
the broker for its core transaction.

**Diagnosis**
```bash
docker compose ps rabbitmq
docker compose logs --tail=100 rabbitmq
# Management UI: http://localhost:15672 (user/password)
docker compose exec rabbitmq rabbitmq-diagnostics ping
docker compose exec rabbitmq rabbitmqctl list_queues
```

**Recovery**
1. Restart: `docker compose up -d rabbitmq` (`restart: always`).
2. Wait for the healthcheck (`rabbitmq-diagnostics ping`).
3. The backend's consumers reconnect and resume. If they do not, restart the
   backend: `docker compose restart backend`.
4. Messages published while the broker was down may have been lost if the
   backend could not reach it. Check for stuck jobs (see §1.5) and re-trigger
   the affected operations (e.g. re-upload a bulk import, re-request a report).

**Note**: publishing failures during an HTTP request return a 5xx to the client
(e.g. bulk upload, report export). The client should retry.

---

### 1.4 MinIO down

**Symptoms**: `/ready` reports MinIO failure; bulk **uploads/downloads** fail;
report **file downloads** fail; export jobs cannot persist their output. The
core API and DB writes are unaffected.

**Diagnosis**
```bash
docker compose ps minio
docker compose logs --tail=100 minio
# Console: http://localhost:9001
```

**Recovery**
1. Restart: `docker compose up -d minio`.
2. Verify the console loads and the buckets exist (`bulk-imports`,
   `bulk-exports`, and the reports bucket from `MINIO_BUCKET_NAME`).
3. Confirm `GET /ready` is healthy.
4. Re-run any bulk/report jobs that failed during the outage (see §1.5).

**Data volume**: `minio_data`. Back this up — it holds uploaded import files and
generated report artifacts.

**Config check**: ensure `MINIO_USE_TLS` matches the endpoint scheme
(`true` only for HTTPS endpoints).

---

### 1.5 Alert / job backlog

**Symptoms**: alerts not delivered; bulk jobs stuck in `QUEUED` or
`PENDING_CONFIRMATION`; report jobs not completing.

**Diagnosis**
- List recent bulk/report jobs: `GET /api/v1/bulk/jobs` (requires
  `bulk.import`). Inspect `status`, `retryCount`, `lastError`,
  `lastAttemptAt`.
- Confirm the broker is up (§1.3) and consumers are running
  (`docker compose logs --tail=200 backend | grep -i consumer`).
- The alert sweep runs on cron `@every 5m`; the daily sales summary runs
  `@daily`. Check backend logs for cron errors.

**Recovery**
1. Resolve the underlying dependency first (Redis/RabbitMQ/MinIO above).
2. Restart the backend to re-establish consumer subscriptions if needed.
3. **Bulk import stuck in `PENDING_CONFIRMATION`**: call
   `POST /api/v1/bulk/products/import/:jobId/confirm` again — confirm is
   idempotent (`ON CONFLICT DO NOTHING`), so re-running is safe.
4. **Job exhausted retries** (`retryCount >= maxRetries`, default 3): re-trigger
   the originating action (re-upload the CSV, re-request the report).
5. **Manual alert sweep**: `POST /api/v1/alerts/check` triggers an immediate
   `CheckAndTriggerAlerts` run.
6. Per-product alert panics are isolated (`recover()` per product) and logged
   with `product_id`; a single bad product will not block the sweep.

---

## 2. Deploy procedure

The reference target is Docker Compose. Adapt the build/push steps for your
orchestrator.

### 2.1 Pre-flight

1. Confirm the working tree is clean and migrations are present under
   `backend/migrations/`.
2. Ensure all required secrets are set to non-default values
   (`JWT_SECRET`, `REFRESH_TOKEN_SECRET`, `MINIO_*`, `SSLCOMMERZ_STORE_PASSWORD`,
   `RABBITMQ_URL`). The backend **fails fast** otherwise (see
   [security.md §3](./security.md)).
3. Set `APP_ENV` appropriately. In production/staging, GORM `AutoMigrate` is
   **skipped** — schema changes only come from the migrator.

### 2.2 Apply migrations

Run the migrator **before** rolling the new image:
```bash
export DATABASE_URL="postgres://user:password@db:5432/inventory_db?sslmode=disable"
cd backend
go run ./cmd/migrator -cmd version -dsn "$DATABASE_URL"   # current version
go run ./cmd/migrator -cmd up      -dsn "$DATABASE_URL"   # apply
go run ./cmd/migrator -cmd version -dsn "$DATABASE_URL"   # confirm
```

### 2.3 Build & roll out

```bash
docker compose build backend
docker compose up -d backend
```

For zero-downtime behind a load balancer, drain the old instance, then stop it.
The backend shuts down gracefully: it cancels consumer subscriptions, drains the
HTTP server (10s timeout), stops cron, flushes traces, and closes DB/Redis/AMQP.

### 2.4 Verify

1. `GET /health` → 200 and `GET /ready` → healthy (DB, Redis, AMQP, MinIO).
2. Smoke-test a login and a permission-guarded endpoint.
3. Check logs for consumer startup and cron registration.
4. Confirm `/metrics` is scraping in Prometheus and dashboards are green.

---

## 3. Rollback procedure

Roll back the **image first**, then the **schema only if required**.

### 3.1 Roll back the application

```bash
# Re-deploy the previous known-good image/tag
docker compose up -d --force-recreate backend
# Verify
curl -fsS http://localhost:8080/ready
```

Because shutdown is graceful, in-flight transactions are rolled back cleanly and
consumers reconnect to the broker on startup.

### 3.2 Decide on the schema

- If the migration was **additive only** (new nullable columns/tables), you can
  usually leave the schema as-is — the old code ignores unknown columns.
- If the migration was **destructive or incompatible** with the old code, roll
  it back:
  ```bash
  go run ./cmd/migrator -cmd down -dsn "$DATABASE_URL"   # revert one version
  go run ./cmd/migrator -cmd version -dsn "$DATABASE_URL"
  ```
  > `down` reverts **one** version per invocation. Repeat as needed. Inspect the
  > corresponding `*.down.sql` before running it in production.

### 3.3 Post-rollback

1. Confirm `/ready` is healthy and the smoke test passes.
2. Watch logs and Grafana for elevated error rates.
3. If Redis was restarted during the incident, expect users to re-login.
4. Re-trigger any jobs lost during the window (§1.5).
5. Record the incident timeline and root cause.

---

## 4. Quick reference

| Task | Command |
| :--- | :--- |
| Full stack up | `docker compose up -d --build` |
| Backend logs | `docker compose logs -f backend` |
| Restart backend | `docker compose restart backend` |
| Migration status | `go run ./cmd/migrator -cmd version -dsn "$DATABASE_URL"` |
| Apply migrations | `go run ./cmd/migrator -cmd up -dsn "$DATABASE_URL"` |
| Revert one migration | `go run ./cmd/migrator -cmd down -dsn "$DATABASE_URL"` |
| Force alert sweep | `POST /api/v1/alerts/check` |
| Readiness probe | `curl -fsS http://localhost:8080/ready` |

---

## 5. Cloud operations

### 5.1 Scaling up/down

**Kubernetes (HPA)**

```bash
kubectl get hpa -n quantify
kubectl describe hpa quantify-backend -n quantify
```

The backend HPA scales 3 → 10 replicas based on CPU (70%), memory (80%), and
custom `http_requests_per_second` (target 100/pod). To adjust:

```bash
helm upgrade quantify-backend deploy/helm/quantify-backend -n quantify \
  --reuse-values \
  --set autoscaling.maxReplicas=15
```

**Manual override**

```bash
kubectl scale deployment quantify-backend -n quantify --replicas=5
```

> Disable HPA first if you need a fixed replica count:
> `--set autoscaling.enabled=false`

**Cloud-managed services**

| Cloud | Database | Redis | Broker |
| :--- | :--- | :--- | :--- |
| AWS | Modify RDS instance class via Terraform (`db_instance_class`) | Modify ElastiCache node type (`redis_node_type`) | Modify Amazon MQ instance (`mq_instance_type`) |
| GCP | Change Cloud SQL tier (`db_tier`) | Change Memorystore shard count (`redis_shard_count`) | Scale Pub/Sub throughput (automatic) |
| Azure | Change PostgreSQL SKU (`db_sku_name`) | Change Redis capacity (`redis_sku_capacity`) | Scale Service Bus tier |

Apply via `terraform plan` / `terraform apply` in the relevant
`infrastructure/<cloud>/` directory.

---

### 5.2 Database failover

**AWS RDS (Multi-AZ)**

Failover is automatic. To trigger manually:

```bash
aws rds failover-db-cluster --db-cluster-identifier quantify-prod-postgres
```

Monitor: `aws rds describe-db-instances --query 'DBInstances[0].DBInstanceStatus'`

**GCP Cloud SQL (HA)**

Automatic regional failover. To test:

```bash
gcloud sql instances failover quantify-prod-cloudsql
```

**Azure PostgreSQL (zone-redundant HA)**

Automatic. To trigger:

```bash
az postgres flexible-server restart \
  --resource-group quantify-prod \
  --name quantify-prod-pg
```

**Post-failover checklist**

1. Verify `/ready` reports healthy DB.
2. Check backend logs for reconnection (automatic via GORM).
3. Confirm no stuck transactions: `SELECT * FROM pg_stat_activity WHERE state = 'idle in transaction';`
4. In-flight transactions are rolled back by PostgreSQL on disconnect.

---

### 5.3 Redis cluster recovery

**In-cluster (Bitnami)**

```bash
kubectl exec -n quantify quantify-infrastructure-redis-cluster-0 -- \
  redis-cli -a <password> cluster info
```

If `cluster_state:fail`:

```bash
kubectl scale statefulset quantify-infrastructure-redis-cluster -n quantify --replicas=0
kubectl scale statefulset quantify-infrastructure-redis-cluster -n quantify --replicas=6
```

**AWS ElastiCache**

```bash
aws elasticache describe-cache-clusters --cluster-id quantify-prod-redis
```

ElastiCache handles node replacement automatically. For shard rebalancing:

```bash
aws elasticache rebalance-slots --replication-group-id quantify-prod-redis
```

**Impact**: Redis restart logs out all users (tokens stored in Redis).
Permission and report caches rebuild lazily.

---

### 5.4 RabbitMQ cluster recovery

**In-cluster (Bitnami)**

```bash
kubectl exec -n quantify quantify-infrastructure-rabbitmq-0 -- \
  rabbitmq-diagnostics cluster_status
```

For a full cluster restart:

```bash
kubectl scale statefulset quantify-infrastructure-rabbitmq -n quantify --replicas=0
kubectl scale statefulset quantify-infrastructure-rabbitmq -n quantify --replicas=3
```

**AWS Amazon MQ**

Amazon MQ manages broker HA. Check status:

```bash
aws mq describe-broker --broker-id <broker-id> --query 'BrokerState'
```

**Post-recovery**: restart the backend to re-establish consumer subscriptions.
Messages published during the outage may be lost — re-trigger affected jobs.

---

### 5.5 Certificate renewal

cert-manager auto-renews Let's Encrypt certificates 30 days before expiry.

**Check certificate status**

```bash
kubectl get certificates -n quantify
kubectl describe certificate quantify-backend-prod-tls -n quantify
```

**Force renewal**

```bash
kubectl cert-manager renew quantify-backend-prod-tls -n quantify
```

**Verify**

```bash
echo | openssl s_client -connect api.quantify.com:443 2>/dev/null | \
  openssl x509 -noout -dates
```

If cert-manager is failing, check the ClusterIssuer and ACME challenge:

```bash
kubectl describe clusterissuer letsencrypt-prod
kubectl get challenges -A
kubectl get orders -A
```

---

### 5.6 Terraform drift detection

Run a plan to detect drift (no changes applied):

```bash
cd infrastructure/aws   # or gcp/ azure/
terraform plan -detailed-exitcode
```

Exit codes: `0` = no drift, `1` = error, `2` = drift detected.

**Scheduled drift detection** (recommended):

```bash
# Add to cron or CI schedule
terraform plan -detailed-exitcode -out=/dev/null || \
  echo "DRIFT DETECTED — investigate"
```

**Remediation**

- If drift is intentional (manual console change), import the resource:
  `terraform import <resource> <id>`
- If drift is unintentional, re-apply: `terraform apply`
- Review `terraform state show <resource>` before any action.

---

### 5.7 Cost anomaly response

**Detection**

| Cloud | Tool |
| :--- | :--- |
| AWS | Cost Explorer + Budgets (alert at 80% of budget) |
| GCP | Cost Management + Budget alerts |
| Azure | Cost Management + Budget alerts |

**Immediate response**

1. Identify the anomalous resource (compute, storage, data transfer).
2. Check for runaway scaling: `kubectl get hpa -n quantify`
3. Check for orphaned resources: `terraform plan` (destroyed resources show as drift).
4. Review recent deploys that may have increased resource requests.

**Common causes**

| Cause | Fix |
| :--- | :--- |
| HPA scaled to max and stayed | Investigate traffic spike; adjust `maxReplicas` |
| Orphaned PVCs / LBs | `terraform destroy -target=<resource>` or manual cleanup |
| NAT gateway data transfer | Reduce egress; use VPC endpoints for AWS services |
| Log storage growth | Reduce Loki/Prometheus retention; add lifecycle rules |
| Over-provisioned instances | Right-size via Terraform variables |

---

## 6. Related documents

- [architecture.md](./architecture.md)
- [security.md](./security.md)
- [business_logic.md](./business_logic.md)
- [deployment/README.md](./deployment/README.md)
- [deployment/monitoring.md](./deployment/monitoring.md)
- [deployment/troubleshooting.md](./deployment/troubleshooting.md)
- `backend/CHANGELOG.md`
