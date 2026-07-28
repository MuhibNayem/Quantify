variable "identifier" {
  description = "Identifier prefix for RDS resources."
  type        = string
}

variable "engine_version" {
  description = "PostgreSQL engine version."
  type        = string
  default     = "15"
}

variable "instance_class" {
  description = "Instance class for the primary and replica."
  type        = string
  default     = "db.m6g.large"
}

variable "allocated_storage" {
  description = "Allocated storage in GiB."
  type        = number
  default     = 100
}

variable "max_allocated_storage" {
  description = "Upper bound for storage autoscaling in GiB (0 disables)."
  type        = number
  default     = 500
}

variable "port" {
  description = "Database port."
  type        = number
  default     = 5432
}

variable "db_name" {
  description = "Initial database name."
  type        = string
}

variable "username" {
  description = "Master username."
  type        = string
}

variable "password" {
  description = "Master password. Auto-generated when null."
  type        = string
  default     = null
  sensitive   = true
}

variable "multi_az" {
  description = "Enable Multi-AZ deployment."
  type        = bool
  default     = true
}

variable "create_read_replica" {
  description = "Create a read replica (ignored when use_aurora is true)."
  type        = bool
  default     = true
}

variable "backup_retention_period" {
  description = "Automated backup retention in days."
  type        = number
  default     = 7
}

variable "max_connections" {
  description = "max_connections parameter value."
  type        = string
  default     = "300"
}

variable "deletion_protection" {
  description = "Enable deletion protection."
  type        = bool
  default     = true
}

variable "skip_final_snapshot" {
  description = "Skip the final snapshot on destroy."
  type        = bool
  default     = false
}

variable "use_aurora" {
  description = "Use Aurora PostgreSQL Serverless v2 instead of standard RDS."
  type        = bool
  default     = false
}

variable "aurora_engine_version" {
  description = "Aurora PostgreSQL engine version."
  type        = string
  default     = "15.5"
}

variable "aurora_min_capacity" {
  description = "Aurora Serverless v2 minimum capacity (ACU)."
  type        = number
  default     = 0.5
}

variable "aurora_max_capacity" {
  description = "Aurora Serverless v2 maximum capacity (ACU)."
  type        = number
  default     = 16
}

variable "aurora_instance_count" {
  description = "Number of Aurora cluster instances."
  type        = number
  default     = 2
}

variable "subnet_ids" {
  description = "Subnet IDs for the DB subnet group."
  type        = list(string)
}

variable "vpc_id" {
  description = "VPC ID for the security group."
  type        = string
}

variable "allowed_security_group_ids" {
  description = "Security group IDs allowed to connect (typically EKS nodes)."
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "Tags applied to all resources."
  type        = map(string)
  default     = {}
}
