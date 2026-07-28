variable "name" {
  description = "Name prefix for monitoring resources."
  type        = string
}

variable "eks_cluster_name" {
  description = "Name of the EKS cluster."
  type        = string
}

variable "eks_log_group_name" {
  description = "Name of the EKS control plane log group."
  type        = string
  default     = null
}

variable "log_retention_days" {
  description = "Retention for application log groups."
  type        = number
  default     = 30
}

variable "rds_identifier" {
  description = "RDS instance/cluster identifier to alarm on."
  type        = string
  default     = null
}

variable "elasticache_replication_group_id" {
  description = "ElastiCache replication group ID to alarm on."
  type        = string
  default     = null
}

variable "mq_broker_id" {
  description = "Amazon MQ broker ID to alarm on."
  type        = string
  default     = null
}

variable "alert_email_endpoints" {
  description = "Email addresses subscribed to alerts."
  type        = list(string)
  default     = []
}

variable "enable_managed_prometheus" {
  description = "Provision Amazon Managed Prometheus and Grafana."
  type        = bool
  default     = false
}

variable "tags" {
  description = "Tags applied to all resources."
  type        = map(string)
  default     = {}
}
