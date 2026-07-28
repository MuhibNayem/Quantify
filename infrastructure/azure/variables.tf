variable "environment" {
  description = "Deployment environment."
  type        = string

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "environment must be one of: dev, staging, prod."
  }
}

variable "location" {
  description = "Azure region to deploy into."
  type        = string
  default     = "eastus"
}

variable "subscription_id" {
  description = "Azure subscription ID. Uses the provider default when null."
  type        = string
  default     = null
}

variable "vnet_cidr" {
  description = "Address space for the virtual network."
  type        = string
  default     = "10.0.0.0/16"
}

variable "aks_kubernetes_version" {
  description = "Kubernetes control plane version."
  type        = string
  default     = "1.30"
}

variable "aks_private_cluster" {
  description = "Enable a private AKS API server."
  type        = bool
  default     = true
}

variable "aks_node_pools" {
  description = "AKS user node pool definitions."
  type = map(object({
    vm_size         = string
    min_count       = number
    max_count       = number
    node_count      = number
    os_disk_size_gb = number
    labels          = map(string)
    taints          = list(string)
  }))
  default = {
    general = {
      vm_size         = "Standard_D2s_v3"
      min_count       = 2
      max_count       = 6
      node_count      = 3
      os_disk_size_gb = 128
      labels          = {}
      taints          = []
    }
    compute = {
      vm_size         = "Standard_F4s_v2"
      min_count       = 1
      max_count       = 8
      node_count      = 2
      os_disk_size_gb = 128
      labels = {
        workload = "backend"
      }
      taints = []
    }
  }
}

variable "db_sku_name" {
  description = "Azure Database for PostgreSQL SKU."
  type        = string
  default     = "GP_Standard_D2ds_v4"
}

variable "db_storage_size_gb" {
  description = "PostgreSQL storage size in GiB."
  type        = number
  default     = 128
}

variable "db_high_availability" {
  description = "Enable zone-redundant high availability."
  type        = bool
  default     = true
}

variable "db_create_read_replica" {
  description = "Create a PostgreSQL read replica."
  type        = bool
  default     = true
}

variable "redis_sku_capacity" {
  description = "Azure Cache for Redis capacity (Premium family P)."
  type        = number
  default     = 2
}

variable "redis_shard_count" {
  description = "Number of Redis cluster shards."
  type        = number
  default     = 3
}

variable "redis_replica_count" {
  description = "Replica count per shard."
  type        = number
  default     = 1
}

variable "alert_email_endpoints" {
  description = "Email addresses subscribed to operational alerts."
  type        = list(string)
  default     = []
}

variable "enable_postgres" {
  description = "Create the PostgreSQL module."
  type        = bool
  default     = true
}

variable "enable_redis" {
  description = "Create the Redis module."
  type        = bool
  default     = true
}

variable "enable_servicebus" {
  description = "Create the Service Bus module."
  type        = bool
  default     = true
}

variable "enable_monitoring" {
  description = "Create the monitoring module."
  type        = bool
  default     = true
}

variable "tags" {
  description = "Tags applied to all resources."
  type        = map(string)
  default     = {}
}
