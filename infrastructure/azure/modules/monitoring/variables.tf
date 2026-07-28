variable "name" {
  description = "Name prefix for monitoring resources."
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

variable "log_analytics_workspace_id" {
  description = "ID of the Log Analytics workspace."
  type        = string
}

variable "postgres_server_id" {
  description = "ID of the PostgreSQL server to alarm on."
  type        = string
  default     = null
}

variable "redis_id" {
  description = "ID of the Redis cache to alarm on."
  type        = string
  default     = null
}

variable "alert_email_endpoints" {
  description = "Email addresses subscribed to alerts."
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "Tags applied to all resources."
  type        = map(string)
  default     = {}
}
