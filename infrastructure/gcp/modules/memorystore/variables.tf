variable "name" {
  description = "Name of the Redis cluster."
  type        = string
}

variable "region" {
  description = "Region for the cluster."
  type        = string
}

variable "shard_count" {
  description = "Number of shards."
  type        = number
  default     = 3
}

variable "replica_count" {
  description = "Replica count per shard."
  type        = number
  default     = 1
}

variable "network_id" {
  description = "Self link of the VPC network."
  type        = string
}
