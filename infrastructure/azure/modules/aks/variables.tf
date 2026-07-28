variable "name" {
  description = "Name of the AKS cluster."
  type        = string
}

variable "resource_group_name" {
  description = "Resource group name."
  type        = string
}

variable "location" {
  description = "Azure region."
  type        = string
}

variable "kubernetes_version" {
  description = "Kubernetes control plane version."
  type        = string
  default     = "1.30"
}

variable "private_cluster" {
  description = "Enable a private API server."
  type        = bool
  default     = true
}

variable "vnet_subnet_id" {
  description = "Subnet ID for the node pools."
  type        = string
}

variable "log_analytics_workspace_id" {
  description = "Log Analytics workspace ID for Container Insights."
  type        = string
}

variable "node_pools" {
  description = "User node pool definitions."
  type = map(object({
    vm_size         = string
    min_count       = number
    max_count       = number
    node_count      = number
    os_disk_size_gb = number
    labels          = map(string)
    taints          = list(string)
  }))
  default = {}
}

variable "tags" {
  description = "Tags applied to all resources."
  type        = map(string)
  default     = {}
}
