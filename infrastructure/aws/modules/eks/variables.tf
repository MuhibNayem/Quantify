variable "cluster_name" {
  description = "Name of the EKS cluster."
  type        = string
}

variable "cluster_version" {
  description = "Kubernetes control plane version."
  type        = string
  default     = "1.30"
}

variable "vpc_id" {
  description = "VPC ID hosting the cluster."
  type        = string
}

variable "subnet_ids" {
  description = "Subnet IDs for the control plane and node groups."
  type        = list(string)
}

variable "endpoint_private_access" {
  description = "Enable private API server endpoint."
  type        = bool
  default     = true
}

variable "endpoint_public_access" {
  description = "Enable public API server endpoint."
  type        = bool
  default     = false
}

variable "cluster_log_types" {
  description = "Control plane log types to enable."
  type        = list(string)
  default     = ["api", "audit", "authenticator", "controllerManager", "scheduler"]
}

variable "control_plane_log_retention" {
  description = "Retention in days for control plane logs."
  type        = number
  default     = 30
}

variable "node_groups" {
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
  default = {}
}

variable "install_cluster_autoscaler" {
  description = "Install the cluster-autoscaler Helm release."
  type        = bool
  default     = true
}

variable "tags" {
  description = "Tags applied to all resources."
  type        = map(string)
  default     = {}
}
