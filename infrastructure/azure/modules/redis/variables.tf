variable "name" {
  description = "Name of the Redis cache."
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

variable "capacity" {
  description = "Cache capacity (Premium family P)."
  type        = number
  default     = 2
}

variable "shard_count" {
  description = "Number of cluster shards."
  type        = number
  default     = 3
}

variable "replica_count" {
  description = "Replica count per primary shard."
  type        = number
  default     = 1
}

variable "subnet_id" {
  description = "Subnet ID for VNet injection."
  type        = string
}

variable "tags" {
  description = "Tags applied to all resources."
  type        = map(string)
  default     = {}
}
