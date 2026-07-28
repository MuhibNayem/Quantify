variable "name" {
  description = "Name of the GKE cluster."
  type        = string
}

variable "project_id" {
  description = "GCP project ID."
  type        = string
}

variable "region" {
  description = "Region for the cluster."
  type        = string
}

variable "network_id" {
  description = "Self link of the VPC network."
  type        = string
}

variable "subnet_name" {
  description = "Name of the subnet."
  type        = string
}

variable "pods_range" {
  description = "Name of the pods secondary range."
  type        = string
}

variable "services_range" {
  description = "Name of the services secondary range."
  type        = string
}

variable "release_channel" {
  description = "GKE release channel."
  type        = string
  default     = "REGULAR"
}

variable "master_ipv4_cidr_block" {
  description = "CIDR for the private control plane."
  type        = string
  default     = "172.16.0.0/28"
}

variable "deletion_protection" {
  description = "Enable cluster deletion protection."
  type        = bool
  default     = true
}

variable "node_pools" {
  description = "Node pool definitions."
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
  default = {}
}

variable "labels" {
  description = "Labels applied to resources."
  type        = map(string)
  default     = {}
}
