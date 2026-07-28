variable "eks_cluster_name" {
  description = "Name of the EKS cluster (used for role naming)."
  type        = string
}

variable "oidc_provider_arn" {
  description = "ARN of the cluster OIDC provider."
  type        = string
}

variable "oidc_provider_url" {
  description = "Issuer URL of the OIDC provider."
  type        = string
}

variable "region" {
  description = "AWS region (used for policy conditions)."
  type        = string
  default     = "us-east-1"
}

variable "s3_bucket_arn" {
  description = "ARN of the object storage bucket."
  type        = string
}

variable "rds_secret_arn" {
  description = "ARN of the Secrets Manager secret holding DB credentials."
  type        = string
  default     = null
}

variable "backend_namespace" {
  description = "Kubernetes namespace of the backend service account."
  type        = string
  default     = "quantify"
}

variable "backend_service_account" {
  description = "Kubernetes service account name for the backend."
  type        = string
  default     = "quantify-backend"
}

variable "ai_namespace" {
  description = "Kubernetes namespace of the AI service account."
  type        = string
  default     = "quantify"
}

variable "ai_service_account" {
  description = "Kubernetes service account name for the AI service."
  type        = string
  default     = "quantify-ai-service"
}

variable "log_group_prefix" {
  description = "CloudWatch log group prefix the roles may write to."
  type        = string
  default     = "quantify"
}

variable "tags" {
  description = "Tags applied to all resources."
  type        = map(string)
  default     = {}
}
