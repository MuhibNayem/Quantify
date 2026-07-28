variable "name" {
  description = "Name prefix for Cloud SQL resources."
  type        = string
}

variable "region" {
  description = "Region for the instance."
  type        = string
}

variable "tier" {
  description = "Machine tier."
  type        = string
  default     = "db-custom-2-8192"
}

variable "high_availability" {
  description = "Enable regional high availability."
  type        = bool
  default     = true
}

variable "disk_size_gb" {
  description = "Disk size in GiB."
  type        = number
  default     = 100
}

variable "max_connections" {
  description = "max_connections database flag."
  type        = string
  default     = "300"
}

variable "db_name" {
  description = "Initial database name."
  type        = string
}

variable "username" {
  description = "Master username."
  type        = string
}

variable "create_read_replica" {
  description = "Create a read replica."
  type        = bool
  default     = true
}

variable "deletion_protection" {
  description = "Enable deletion protection."
  type        = bool
  default     = true
}

variable "network_id" {
  description = "Self link of the VPC network for private IP."
  type        = string
}

variable "labels" {
  description = "Labels applied to resources."
  type        = map(string)
  default     = {}
}
