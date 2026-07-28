variable "name" {
  description = "Globally unique bucket name."
  type        = string
}

variable "location" {
  description = "Bucket location."
  type        = string
}

variable "cors_origins" {
  description = "Origins allowed by the CORS configuration."
  type        = list(string)
  default     = ["*"]
}

variable "labels" {
  description = "Labels applied to resources."
  type        = map(string)
  default     = {}
}
