variable "project_id" {
  description = "GCP project ID."
  type        = string
}

variable "cluster_name" {
  description = "Name of the GKE cluster (used for naming)."
  type        = string
}

variable "workload_identity_pool" {
  description = "Workload Identity pool (PROJECT.svc.id.goog)."
  type        = string
}

variable "namespace" {
  description = "Kubernetes namespace of the service accounts."
  type        = string
  default     = "quantify"
}

variable "backend_service_account" {
  description = "Kubernetes service account name for the backend."
  type        = string
  default     = "quantify-backend"
}

variable "ai_service_account" {
  description = "Kubernetes service account name for the AI service."
  type        = string
  default     = "quantify-ai-service"
}

variable "storage_bucket_name" {
  description = "Name of the object storage bucket."
  type        = string
  default     = null
}

variable "cloudsql_instance" {
  description = "Cloud SQL instance name (enables DB bindings when set)."
  type        = string
  default     = null
}

variable "labels" {
  description = "Labels applied to resources."
  type        = map(string)
  default     = {}
}
