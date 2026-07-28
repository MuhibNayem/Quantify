variable "bucket_name" {
  description = "Globally unique bucket name."
  type        = string
}

variable "versioning_enabled" {
  description = "Enable object versioning."
  type        = bool
  default     = true
}

variable "expiration_days" {
  description = "Days after which current objects expire."
  type        = number
  default     = 365
}

variable "noncurrent_expiration_days" {
  description = "Days after which noncurrent object versions expire."
  type        = number
  default     = 90
}

variable "cors_allowed_origins" {
  description = "Origins allowed by the CORS configuration."
  type        = list(string)
  default     = ["*"]
}

variable "tags" {
  description = "Tags applied to all resources."
  type        = map(string)
  default     = {}
}
