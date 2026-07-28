variable "name" {
  description = "Name prefix for identities."
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

variable "aks_oidc_issuer_url" {
  description = "OIDC issuer URL of the AKS cluster."
  type        = string
}

variable "aks_kubelet_principal_id" {
  description = "Object ID of the AKS kubelet identity."
  type        = string
}

variable "acr_id" {
  description = "ID of the container registry."
  type        = string
}

variable "storage_account_id" {
  description = "ID of the storage account."
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

variable "tags" {
  description = "Tags applied to all resources."
  type        = map(string)
  default     = {}
}
