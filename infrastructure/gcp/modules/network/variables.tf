variable "name" {
  description = "Name prefix for network resources."
  type        = string
}

variable "region" {
  description = "Region for the subnet and router."
  type        = string
}

variable "cidr" {
  description = "Overall VPC CIDR (informational)."
  type        = string
  default     = "10.0.0.0/16"
}

variable "subnet_cidr" {
  description = "Primary CIDR range for the subnet."
  type        = string
  default     = "10.0.0.0/20"
}

variable "pods_cidr" {
  description = "Secondary range for GKE pods."
  type        = string
  default     = "10.4.0.0/14"
}

variable "services_cidr" {
  description = "Secondary range for GKE services."
  type        = string
  default     = "10.8.0.0/20"
}

variable "pods_secondary_range_name" {
  description = "Name of the pods secondary range."
  type        = string
  default     = "pods"
}

variable "services_secondary_range_name" {
  description = "Name of the services secondary range."
  type        = string
  default     = "services"
}

variable "labels" {
  description = "Labels applied to resources."
  type        = map(string)
  default     = {}
}
