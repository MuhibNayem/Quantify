variable "name" {
  description = "Name prefix for monitoring resources."
  type        = string
}

variable "project_id" {
  description = "GCP project ID."
  type        = string
}

variable "cloudsql_instance" {
  description = "Cloud SQL instance name to alarm on."
  type        = string
  default     = null
}

variable "redis_cluster" {
  description = "Redis cluster name to alarm on."
  type        = string
  default     = null
}

variable "gke_cluster_name" {
  description = "GKE cluster name for node alarms."
  type        = string
  default     = ""
}

variable "alert_email_endpoints" {
  description = "Email addresses subscribed to alerts."
  type        = list(string)
  default     = []
}

variable "labels" {
  description = "Labels applied to resources."
  type        = map(string)
  default     = {}
}
