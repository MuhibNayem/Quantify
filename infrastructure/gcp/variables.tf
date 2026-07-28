variable "project_id" {
  description = "GCP project ID to deploy into."
  type        = string
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
  description = "GCP region to deploy into."
  type        = string
  default     = "us-central1"
}

variable "network_cidr" {
  description = "CIDR block for the VPC."
  type        = string
  default     = "10.0.0.0/16"
}

variable "pods_secondary_range_name" {
  description = "Name of the pods secondary range on the GKE subnet."
  type        = string
  default     = "pods"
}

variable "services_secondary_range_name" {
  description = "Name of the services secondary range on the GKE subnet."
  type        = string
  default     = "services"
}

variable "gke_release_channel" {
  description = "GKE release channel."
  type        = string
  default     = "REGULAR"
}

variable "gke_node_pools" {
  description = "GKE node pool definitions."
  type = map(object({
    machine_type  = string
    min_count     = number
    max_count     = number
    initial_count = number
    disk_size_gb  = number
    labels        = map(string)
    taints = list(object({
      key    = string
      value  = string
      effect = string
    }))
  }))
  default = {
    general = {
      machine_type  = "e2-standard-2"
      min_count     = 2
      max_count     = 6
      initial_count = 3
      disk_size_gb  = 100
      labels        = {}
      taints        = []
    }
    compute = {
      machine_type  = "c2-standard-4"
      min_count     = 1
      max_count     = 8
      initial_count = 2
      disk_size_gb  = 100
      labels = {
        workload = "backend"
      }
      taints = []
    }
  }
}

variable "db_tier" {
  description = "Cloud SQL machine tier."
  type        = string
  default     = "db-custom-2-8192"
}

variable "db_ha" {
  description = "Enable regional high availability for Cloud SQL."
  type        = bool
  default     = true
}

variable "db_create_read_replica" {
  description = "Create a Cloud SQL read replica."
  type        = bool
  default     = true
}

variable "db_disk_size_gb" {
  description = "Cloud SQL disk size in GiB."
  type        = number
  default     = 100
}

variable "redis_shard_count" {
  description = "Number of Redis cluster shards."
  type        = number
  default     = 3
}

variable "redis_replica_count" {
  description = "Replica count per Redis shard."
  type        = number
  default     = 1
}

variable "alert_email_endpoints" {
  description = "Email addresses subscribed to operational alerts."
  type        = list(string)
  default     = []
}

variable "enable_cloudsql" {
  description = "Create the Cloud SQL module."
  type        = bool
  default     = true
}

variable "enable_redis" {
  description = "Create the Memorystore module."
  type        = bool
  default     = true
}

variable "enable_pubsub" {
  description = "Create the Pub/Sub module."
  type        = bool
  default     = true
}

variable "enable_monitoring" {
  description = "Create the monitoring module."
  type        = bool
  default     = true
}

variable "labels" {
  description = "Labels applied to all resources."
  type        = map(string)
  default     = {}
}
