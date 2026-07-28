variable "name" {
  description = "Name prefix for the storage account."
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

variable "cors_origins" {
  description = "Origins allowed by the blob CORS configuration."
  type        = list(string)
  default     = ["*"]
}

variable "tags" {
  description = "Tags applied to all resources."
  type        = map(string)
  default     = {}
}
