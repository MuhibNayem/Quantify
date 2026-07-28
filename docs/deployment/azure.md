# Azure Deployment

Provision infrastructure with Terraform (`infrastructure/azure/`) and deploy
Quantify to AKS with Helm.

---

## Prerequisites

| Tool | Version | Purpose |
| :--- | :--- | :--- |
| Azure CLI | ≥ 2.50 | Authentication |
| Terraform | ≥ 1.9 | Infrastructure provisioning |
| kubectl | ≥ 1.26 | Cluster access |
| helm | ≥ 3.12 | Application deployment |

Authenticate:

```bash
az login
az account set --subscription <subscription-id>
```

---

## Terraform setup

### State backend

Use an Azure Storage backend in `infrastructure/azure/`:

```hcl
terraform {
  backend "azurerm" {
    resource_group_name  = "quantify-tfstate"
    storage_account_name = "quantifytfstate"
    container_name       = "tfstate"
    key                  = "azure/terraform.tfstate"
  }
}
```

### Variables

Create `terraform.tfvars`:

```hcl
environment = "prod"
location    = "eastus"

# AKS
aks_kubernetes_version = "1.30"
aks_private_cluster    = true

# PostgreSQL
db_sku_name            = "GP_Standard_D2ds_v4"
db_storage_size_gb     = 128
db_high_availability   = true
db_create_read_replica = true

# Redis
redis_sku_capacity  = 2
redis_shard_count   = 3
redis_replica_count = 1

# Monitoring
alert_email_endpoints = ["oncall@quantify.com"]
```

All variables: `infrastructure/azure/variables.tf`.

---

## Provisioning infrastructure

```bash
cd infrastructure/azure

terraform init
terraform plan  -var-file=terraform.tfvars -out=tfplan
terraform apply tfplan
```

### Resources created

| Module | Resources |
| :--- | :--- |
| `network` | VNet, subnets (AKS, PostgreSQL, Redis) |
| `aks` | AKS cluster (private), node pools (general + compute), Log Analytics |
| `postgres` | Azure Database for PostgreSQL (Flexible Server, zone-redundant HA), read replica |
| `redis` | Azure Cache for Redis (Premium, clustered) |
| `servicebus` | Azure Service Bus namespace |
| `storage` | Storage account + blob container |
| `acr` | Azure Container Registry |
| `iam` | Workload Identity federated credentials for backend and AI service |
| `monitoring` | Log Analytics workspace, metric alerts, action groups |

> **Note**: Azure uses Service Bus instead of RabbitMQ. The backend's message
> broker abstraction must be configured for Service Bus, or deploy RabbitMQ
> in-cluster via the infrastructure Helm chart.

### Key outputs

```bash
terraform output kubeconfig_command
terraform output postgres_fqdn
terraform output redis_hostname
terraform output servicebus_endpoint
terraform output storage_account_name
terraform output acr_login_server
```

---

## Configure kubectl

```bash
az aks get-credentials \
  --resource-group $(terraform output -raw resource_group_name) \
  --name $(terraform output -raw aks_cluster_name)
```

---

## Deploy with Helm

Follow the [Kubernetes guide](./kubernetes.md) starting at step 1.

Override endpoints for managed services:

```yaml
config:
  DB_HOST: quantify-prod-pg.postgres.database.azure.com
  REDIS_CLUSTER_ADDRS: quantify-prod-redis.redis.cache.windows.net:6380
```

Push images to ACR:

```bash
az acr login --name $(terraform output -raw acr_login_server | cut -d/ -f1)
docker tag quantify-backend:latest \
  $(terraform output -raw acr_login_server)/backend:<tag>
docker push $(terraform output -raw acr_login_server)/backend:<tag>
```

---

## Workload Identity

The Terraform `iam` module creates user-assigned managed identities with
federated credentials for AKS Workload Identity:

| K8s ServiceAccount | Managed Identity | Permissions |
| :--- | :--- | :--- |
| `quantify-backend` | `quantify-prod-backend` | Key Vault, Storage, Service Bus |
| `quantify-ai-service` | `quantify-prod-ai` | Storage read |

Annotate the Helm chart service accounts:

```yaml
serviceAccount:
  annotations:
    azure.workload.identity/client-id: <backend-client-id>
```

Label pods:

```yaml
podLabels:
  azure.workload.identity/use: "true"
```

---

## DNS and TLS (Azure DNS + App Gateway / cert-manager)

### Option A: cert-manager (recommended)

```bash
helm repo add jetstack https://charts.jetstack.io
helm upgrade --install cert-manager jetstack/cert-manager \
  --namespace cert-manager --create-namespace \
  --set installCRDs=true
kubectl apply -f deploy/k8s/cert-manager/cluster-issuers.yaml
```

Create an Azure DNS A record pointing at the ingress load balancer IP.

### Option B: Application Gateway Ingress Controller

Use AGIC with an Azure Application Gateway for L7 routing, WAF, and
managed TLS certificates.

---

## Cost optimization

| Strategy | How |
| :--- | :--- |
| Spot VMs | Add `spot_max_price` to non-critical node pools |
| Disable HA in dev | `db_high_availability = false` |
| Disable replicas | `db_create_read_replica = false` in dev |
| Smaller SKUs | `db_sku_name = "B_Standard_B1ms"` for dev |
| Disable Service Bus | `enable_servicebus = false` if using in-cluster RabbitMQ |
| ACR geo-replication | Only enable for multi-region prod |
| Log Analytics retention | `retention_in_days = 30` (default), lower for dev |

---

## Terraform CI/CD

The `terraform.yml` workflow handles Azure changes under
`infrastructure/azure/**`. Required secrets: `ARM_CLIENT_ID`,
`ARM_CLIENT_SECRET`, `ARM_SUBSCRIPTION_ID`, `ARM_TENANT_ID`.

---

## Related

- [Kubernetes guide](./kubernetes.md)
- [CI/CD](./ci-cd.md)
- [Monitoring](./monitoring.md)
