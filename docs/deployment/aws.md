# AWS Deployment

Provision infrastructure with Terraform (`infrastructure/aws/`) and deploy
Quantify to EKS with Helm.

---

## Prerequisites

| Tool | Version | Purpose |
| :--- | :--- | :--- |
| AWS CLI | ≥ 2.x | Authentication |
| Terraform | ≥ 1.9 | Infrastructure provisioning |
| kubectl | ≥ 1.26 | Cluster access |
| helm | ≥ 3.12 | Application deployment |

AWS credentials must be configured (`aws sts get-caller-identity` succeeds).

---

## Terraform setup

### State backend

Configure a remote backend (S3 + DynamoDB) in `infrastructure/aws/`:

```hcl
terraform {
  backend "s3" {
    bucket         = "quantify-tfstate"
    key            = "aws/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "quantify-tf-locks"
    encrypt        = true
  }
}
```

### Variables

Create `terraform.tfvars`:

```hcl
project_name = "quantify"
environment  = "prod"
region       = "us-east-1"
owner        = "platform"

# EKS
eks_cluster_version = "1.30"
eks_public_endpoint = false

# RDS
db_instance_class     = "db.m6g.large"
db_multi_az           = true
db_create_read_replica = true

# ElastiCache
redis_node_type              = "cache.m6g.large"
redis_num_node_groups        = 3
redis_replicas_per_node_group = 1

# Amazon MQ
mq_instance_type = "mq.m5.large"

# Monitoring
alert_email_endpoints     = ["oncall@quantify.com"]
enable_managed_prometheus = false
```

All variables and defaults: `infrastructure/aws/variables.tf`.

---

## Provisioning infrastructure

```bash
cd infrastructure/aws

terraform init
terraform plan  -var-file=terraform.tfvars -out=tfplan
terraform apply tfplan
```

### Resources created

| Module | Resources |
| :--- | :--- |
| `vpc` | VPC, public/private subnets (3 AZs), NAT gateways, flow logs |
| `eks` | EKS cluster, managed node groups (general + compute), OIDC provider |
| `rds` | PostgreSQL 15 (Multi-AZ), read replica, Secrets Manager credential |
| `elasticache` | Redis cluster (3 shards, 1 replica each) |
| `mq` | Amazon MQ RabbitMQ broker |
| `s3` | Object storage bucket (versioned) |
| `ecr` | Container registries (`backend`, `ai-service`) |
| `iam` | IRSA roles for backend and AI service |
| `monitoring` | CloudWatch alarms, SNS topic, optional Managed Prometheus |

### Key outputs

```bash
terraform output kubeconfig_command
terraform output rds_endpoint
terraform output redis_configuration_endpoint
terraform output rabbitmq_endpoint
terraform output s3_bucket_name
terraform output ecr_repository_urls
```

---

## Configure kubectl

```bash
aws eks update-kubeconfig \
  --name $(terraform output -raw eks_cluster_name) \
  --region us-east-1
```

---

## Deploy with Helm

Follow the [Kubernetes guide](./kubernetes.md) starting at step 1 (namespaces).

Override infrastructure endpoints in the Helm values to point at the
Terraform-provisioned managed services instead of in-cluster Bitnami charts:

```yaml
config:
  DB_HOST: quantify-prod-postgres.xxxx.us-east-1.rds.amazonaws.com
  REDIS_CLUSTER_ADDRS: quantify-prod-redis.xxxx.cache.amazonaws.com:6379
  MINIO_ENDPOINT: s3.us-east-1.amazonaws.com
```

---

## External Secrets Operator

Use AWS Secrets Manager as the secret backend:

```bash
helm repo add external-secrets https://charts.external-secrets.io
helm upgrade --install external-secrets external-secrets/external-secrets \
  --namespace external-secrets --create-namespace \
  --set installCRDs=true
```

Create a ClusterSecretStore pointing at Secrets Manager, then apply the
ExternalSecret manifest:

```bash
kubectl apply -f deploy/k8s/external-secrets/backend-external-secret.yaml
```

Enable in the backend chart:

```yaml
externalSecrets:
  enabled: true
  secretStoreName: vault-backend
  secretStoreKind: ClusterSecretStore
  remoteRef: quantify/backend
```

The Terraform `iam` module creates IRSA roles with Secrets Manager read access.

---

## DNS and TLS (Route53 + ACM)

1. Create a hosted zone in Route53 for `quantify.com`.
2. Request an ACM certificate for `api.quantify.com` (DNS validation).
3. Annotate the ingress-nginx service with the NLB/ALB and ACM ARN:

```yaml
controller:
  service:
    annotations:
      service.beta.kubernetes.io/aws-load-balancer-type: nlb
      service.beta.kubernetes.io/aws-load-balancer-ssl-cert: arn:aws:acm:...
```

4. Create a Route53 A record (alias) pointing at the load balancer.

Alternatively, use cert-manager with the `letsencrypt-prod` ClusterIssuer
(see [kubernetes.md §5](./kubernetes.md)).

---

## Cost optimization

| Strategy | How |
| :--- | :--- |
| Single NAT gateway | `single_nat_gateway = true` (dev/staging) |
| Spot instances | Set `capacity_type = "SPOT"` on non-critical node groups |
| Aurora Serverless v2 | `db_use_aurora = true` for variable workloads |
| Right-size instances | Start with `db.m6g.large` / `cache.m6g.large`, monitor, adjust |
| Disable in dev | `enable_rds = false`, `enable_elasticache = false`, `enable_mq = false` |
| ECR lifecycle | `ecr_max_image_count = 30` prunes old images automatically |
| S3 lifecycle | Add transition rules to Glacier for old report artifacts |
| VPC flow logs | `enable_vpc_flow_logs = false` in dev |

---

## Terraform CI/CD

The `terraform.yml` GitHub Actions workflow runs on changes to
`infrastructure/aws/**`:

- **PR**: `fmt` → `validate` → `plan` (commented on the PR)
- **Push to main**: `fmt` → `validate` → `apply` (with environment approval)
- **Manual dispatch**: choose cloud + action

Required repository secrets: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`.

---

## Related

- [Kubernetes guide](./kubernetes.md)
- [CI/CD](./ci-cd.md)
- [Monitoring](./monitoring.md)
