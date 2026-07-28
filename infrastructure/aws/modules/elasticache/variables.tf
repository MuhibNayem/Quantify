variable "name" {
  description = "Name prefix for ElastiCache resources."
  type        = string
}

variable "node_type" {
  description = "Cache node type."
  type        = string
  default     = "cache.m6g.large"
}

variable "engine_version" {
  description = "Redis engine version."
  type        = string
  default     = "7.1"
}

variable "port" {
  description = "Redis port."
  type        = number
  default     = 6379
}

variable "num_node_groups" {
  description = "Number of shards (cluster mode enabled)."
  type        = number
  default     = 3
}

variable "replicas_per_node_group" {
  description = "Replica count per shard."
  type        = number
  default     = 1
}

variable "snapshot_retention_limit" {
  description = "Days to retain automatic snapshots."
  type        = number
  default     = 7
}

variable "subnet_ids" {
  description = "Subnet IDs for the subnet group."
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
