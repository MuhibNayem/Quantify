variable "name" {
  description = "Name of the PostgreSQL server."
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

variable "sku_name" {
  description = "Server SKU."
  type        = string
  default     = "GP_Standard_D2ds_v4"
}

variable "storage_size_gb" {
  description = "Storage size in GiB."
  type        = number
  default     = 128
}

variable "high_availability" {
  description = "Enable zone-redundant high availability."
  type        = bool
  default     = true
}

variable "create_read_replica" {
  description = "Create a read replica."
  type        = bool
  default     = true
}

variable "administrator_login" {
  description = "Administrator username."
  type        = string
}

variable "database_name" {
  description = "Initial database name."
  type        = string
}

variable "subnet_id" {
  description = "Delegated subnet ID."
  type        = string
}

variable "vnet_id" {
  description = "Virtual network ID for the private DNS link."
  type        = string
}

variable "tags" {
  description = "Tags applied to all resources."
  type        = map(string)
  default     = {}
}
