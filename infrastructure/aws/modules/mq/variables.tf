variable "name" {
  description = "Name prefix for the broker."
  type        = string
}

variable "engine_version" {
  description = "RabbitMQ engine version."
  type        = string
  default     = "3.13"
}

variable "instance_type" {
  description = "Broker instance type."
  type        = string
  default     = "mq.m5.large"
}

variable "username" {
  description = "Admin username."
  type        = string
}

variable "password" {
  description = "Admin password. Auto-generated when null."
  type        = string
  default     = null
  sensitive   = true
}

variable "subnet_ids" {
  description = "Subnet IDs (3 required for CLUSTER_MULTI_AZ)."
  type        = list(string)
}

variable "vpc_id" {
  description = "VPC ID for the security group."
  type        = string
}

variable "allowed_security_group_ids" {
  description = "Security group IDs allowed to connect."
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "Tags applied to all resources."
  type        = map(string)
  default     = {}
}
