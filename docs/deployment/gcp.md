# GCP Deployment

Provision infrastructure with Terraform (`infrastructure/gcp/`) and deploy
Quantify to GKE with Helm.

---

## Prerequisites

| Tool | Version | Purpose |
| :--- | :--- | :--- |
| gcloud CLI | ≥ 450 | Authentication |
| Terraform | ≥ 1.9 | Infrastructure provisioning |
| kubectl | ≥ 1.26 | Cluster access |
| helm | ≥ 3.12 | Application deployment |

Authenticate and set the project:

```bash
gcloud auth login
gcloud auth application-default login
gcloud config set project <project-id>
```

---

## Terraform setup

### State backend

Use a GCS backend in `infrastructure/gcp/`:

```hcl
terraform {
  backend "gcs" {
    bucket = "quantify-tfstate"
    prefix = "gcp/terraform/state"
  }
}
```

### Variables

Create `terraform.tfvars`:

```hcl
project_id  = "quantify-prod"
environment = "prod"
region      = "us-central1"

# GKE
gke_release_channel = "REGULAR"

# Cloud SQL
db_tier                = "db-custom-2-8192"
db_ha                  = true
db_create_read_replica = true
db_disk_size_gb        = 100

# Memorystore (Redis)
redis_shard_count   = 3
redis_replica_count = 1

# Monitoring
alert_email_endpoints = ["oncall@quantify.com"]
```

All variables: `infrastructure/gcp/variables.tf`.

---

## Provisioning infrastructure

```bash
cd infrastructure/gcp

terraform init
terraform plan  -var-file=terraform.tfvars -out=tfplan
terraform apply tfplan
```

### Resources created

| Module | Resources |
| :--- | :--- |
| `network` | VPC, subnet with secondary ranges (pods/services) |
| `gke` | GKE cluster (Regional), node pools (general + compute), Workload Identity |
| `cloudsql` | Cloud SQL PostgreSQL 15 (HA), read replica |
| `memorystore` | Memorystore Redis cluster (3 shards, 1 replica each) |
| `pubsub` | Pub/Sub topic for async messaging |
| `storage` | GCS bucket for objects |
| `artifact_registry` | Docker repositories (`backend`, `ai-service`) |
| `iam` | Workload Identity service accounts for backend and AI service |
| `monitoring` | Cloud Monitoring alert policies, notification channels |

> **Note**: GCP uses Pub/Sub instead of RabbitMQ. The backend's message broker
> abstraction must be configured for Pub/Sub, or deploy RabbitMQ in-cluster
> via the infrastructure Helm chart.

### Key outputs

```bash
terraform output kubeconfig_command
terraform output cloudsql_connection_name
terraform output cloudsql_private_ip
terraform output redis_discovery_endpoint
terraform output storage_bucket_name
terraform output artifact_registry_urls
```

---

## Configure kubectl

```bash
gcloud container clusters get-credentials \
  $(terraform output -raw gke_cluster_name) \
  --region us-central1 \
  --project <project-id>
```

---

## Deploy with Helm

Follow the [Kubernetes guide](./kubernetes.md) starting at step 1.

Override endpoints for managed services:

```yaml
config:
  DB_HOST: <cloudsql-private-ip>
  REDIS_CLUSTER_ADDRS: <redis-discovery-endpoint>:6379
```

Push images to Artifact Registry:

```bash
gcloud auth configure-docker us-central1-docker.pkg.dev
docker tag quantify-backend:latest \
  us-central1-docker.pkg.dev/<project-id>/quantify/backend:<tag>
docker push us-central1-docker.pkg.dev/<project-id>/quantify/backend:<tag>
```

---

## Workload Identity

The Terraform `iam` module creates service accounts bound to Kubernetes
service accounts via Workload Identity:

| K8s ServiceAccount | GCP SA | Permissions |
| :--- | :--- | :--- |
| `quantify-backend` | `quantify-backend@<project>.iam` | Cloud SQL client, GCS read/write, Pub/Sub |
| `quantify-ai-service` | `quantify-ai-service@<project>.iam` | GCS read |

Annotate the Helm chart service accounts:

```yaml
serviceAccount:
  annotations:
    iam.gke.io/gcp-service-account: quantify-backend@<project-id>.iam.gserviceaccount.com
```

---

## DNS and TLS (Cloud DNS + Managed Certificates)

1. Create a Cloud DNS zone for `quantify.com`.
2. Use GKE managed certificates or cert-manager:

```bash
helm repo add jetstack https://charts.jetstack.io
helm upgrade --install cert-manager jetstack/cert-manager \
  --namespace cert-manager --create-namespace \
  --set installCRDs=true
kubectl apply -f deploy/k8s/cert-manager/cluster-issuers.yaml
```

3. Create a DNS A record pointing at the ingress load balancer IP.

---

## Cost optimization

| Strategy | How |
| :--- | :--- |
| Preemptible/Spot VMs | Add `spot = true` to non-critical node pools |
| Cloud SQL HA off in dev | `db_ha = false` |
| Disable read replicas | `db_create_read_replica = false` in dev |
| Smaller tiers | `db_tier = "db-custom-1-3840"` for dev |
| Disable Pub/Sub | `enable_pubsub = false` if using in-cluster RabbitMQ |
| Autopilot | Consider GKE Autopilot for pay-per-pod billing |

---

## Terraform CI/CD

The `terraform.yml` workflow handles GCP changes under
`infrastructure/gcp/**`. Required secret: `GOOGLE_CREDENTIALS` (service
account JSON key).

---

## Related

- [Kubernetes guide](./kubernetes.md)
- [CI/CD](./ci-cd.md)
- [Monitoring](./monitoring.md)
