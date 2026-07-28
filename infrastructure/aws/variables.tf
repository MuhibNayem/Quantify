variable "project_name" {
  description = "Project name used as a prefix for all resources."
  type        = string
  default     = "quantify"

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{1,30}$", var.project_name))
    error_message = "project_name must be lowercase alphanumeric with hyphens, 2-31 chars."
  }
}

variable "environment" {
  description = "Deployment environment."
  type        = string

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "environment must be one of: dev, staging, prod."
  }
}

variable "region" {
  description = "AWS region to deploy into."
  type        = string
  default     = "us-east-1"
}

variable "owner" {
  description = "Owner tag applied to all resources."
  type        = string
  default     = "platform"
}

variable "tags" {
  description = "Additional tags merged into every resource."
  type        = map(string)
  default     = {}
}

# ---------------------------------------------------------------------------
# VPC
# ---------------------------------------------------------------------------
variable "vpc_cidr" {
  description = "CIDR block for the VPC."
  type        = string
  default     = "10.0.0.0/16"

  validation {
    condition     = can(cidrhost(var.vpc_cidr, 0))
    error_message = "vpc_cidr must be a valid IPv4 CIDR block."
  }
}

variable "availability_zones" {
  description = "Availability zones. Defaults to the first 3 available in the region."
  type        = list(string)
  default     = []
}

variable "enable_nat_gateway" {
  description = "Provision NAT gateway(s) for private subnets."
  type        = bool
  default     = true
}

variable "single_nat_gateway" {
  description = "Use a single shared NAT gateway (cost) instead of one per AZ (HA)."
  type        = bool
  default     = false
}

variable "enable_vpc_flow_logs" {
  description = "Enable VPC flow logs to CloudWatch."
  type        = bool
  default     = true
}

# ---------------------------------------------------------------------------
# EKS
# ---------------------------------------------------------------------------
variable "eks_cluster_version" {
  description = "Kubernetes control plane version."
  type        = string
  default     = "1.30"
}

variable "eks_public_endpoint" {
  description = "Expose the EKS API server endpoint publicly (private access is always on)."
  type        = bool
  default     = false
}

variable "install_cluster_autoscaler" {
  description = "Install the cluster-autoscaler Helm release on the cluster."
  type        = bool
  default     = true
}

variable "eks_node_groups" {
  description = "Managed node group definitions."
  type = map(object({
    instance_types = list(string)
    capacity_type  = string
    min_size       = number
    max_size       = number
    desired_size   = number
    disk_size      = number
    labels         = map(string)
    taints = map(object({
      key    = string
      value  = string
      effect = string
    }))
  }))
  default = {
    general = {
      instance_types = ["m5.large"]
      capacity_type  = "ON_DEMAND"
      min_size       = 2
      max_size       = 6
      desired_size   = 3
      disk_size      = 50
      labels         = {}
      taints         = {}
    }
    compute = {
      instance_types = ["c5.xlarge"]
      capacity_type  = "ON_DEMAND"
      min_size       = 1
      max_size       = 8
      desired_size   = 2
      disk_size      = 100
      labels = {
        workload = "backend"
      }
      taints = {}
    }
  }
}

# ---------------------------------------------------------------------------
# RDS
# ---------------------------------------------------------------------------
variable "db_instance_class" {
  description = "RDS instance class."
  type        = string
  default     = "db.m6g.large"
}

variable "db_allocated_storage" {
  description = "Allocated storage in GiB for the primary instance."
  type        = number
  default     = 100
}

variable "db_multi_az" {
  description = "Enable Multi-AZ for the primary database."
  type        = bool
  default     = true
}

variable "db_create_read_replica" {
  description = "Create a read replica for the primary database."
  type        = bool
  default     = true
}

variable "db_backup_retention_period" {
  description = "Automated backup retention in days."
  type        = number
  default     = 7
}

variable "db_use_aurora" {
  description = "Use Aurora PostgreSQL Serverless v2 instead of standard RDS."
  type        = bool
  default     = false
}

variable "db_username" {
  description = "Master username. A random password is generated when db_password is null."
  type        = string
  default     = "quantify"
}

variable "db_password" {
  description = "Master password. Leave null to auto-generate and store in Secrets Manager."
  type        = string
  default     = null
  sensitive   = true
}

# ---------------------------------------------------------------------------
# ElastiCache
# ---------------------------------------------------------------------------
variable "redis_node_type" {
  description = "ElastiCache node type."
  type        = string
  default     = "cache.m6g.large"
}

variable "redis_num_node_groups" {
  description = "Number of shards (cluster mode enabled)."
  type        = number
  default     = 3
}

variable "redis_replicas_per_node_group" {
  description = "Replica count per shard."
  type        = number
  default     = 1
}

# ---------------------------------------------------------------------------
# Amazon MQ
# ---------------------------------------------------------------------------
variable "mq_instance_type" {
  description = "Amazon MQ broker instance type."
  type        = string
  default     = "mq.m5.large"
}

variable "mq_username" {
  description = "RabbitMQ admin username."
  type        = string
  default     = "quantify"
}

variable "mq_password" {
  description = "RabbitMQ admin password. Leave null to auto-generate."
  type        = string
  default     = null
  sensitive   = true
}

# ---------------------------------------------------------------------------
# S3
# ---------------------------------------------------------------------------
variable "s3_bucket_suffix" {
  description = "Suffix appended to the object storage bucket name."
  type        = string
  default     = "storage"
}

variable "s3_versioning_enabled" {
  description = "Enable object versioning on the storage bucket."
  type        = bool
  default     = true
}

# ---------------------------------------------------------------------------
# ECR
# ---------------------------------------------------------------------------
variable "ecr_repositories" {
  description = "ECR repositories to create."
  type        = list(string)
  default     = ["backend", "ai-service"]
}

variable "ecr_max_image_count" {
  description = "Number of most recent images to retain per repository."
  type        = number
  default     = 30
}

variable "ecr_cross_account_principals" {
  description = "Optional AWS account IDs allowed to pull images."
  type        = list(string)
  default     = []
}

# ---------------------------------------------------------------------------
# Monitoring
# ---------------------------------------------------------------------------
variable "alert_email_endpoints" {
  description = "Email addresses subscribed to operational alerts."
  type        = list(string)
  default     = []
}

variable "enable_managed_prometheus" {
  description = "Provision Amazon Managed Prometheus and Grafana."
  type        = bool
  default     = false
}

# ---------------------------------------------------------------------------
# Feature toggles
# ---------------------------------------------------------------------------
variable "enable_rds" {
  description = "Create the RDS module."
  type        = bool
  default     = true
}

variable "enable_elasticache" {
  description = "Create the ElastiCache module."
  type        = bool
  default     = true
}

variable "enable_mq" {
  description = "Create the Amazon MQ module."
  type        = bool
  default     = true
}

variable "enable_monitoring" {
  description = "Create the monitoring module."
  type        = bool
  default     = true
}
