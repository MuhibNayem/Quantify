variable "repositories" {
  description = "Repository names to create."
  type        = list(string)
}

variable "name_prefix" {
  description = "Prefix applied to repository names."
  type        = string
  default     = "quantify"
}

variable "max_image_count" {
  description = "Number of most recent images to retain."
  type        = number
  default     = 30
}

variable "cross_account_principals" {
  description = "AWS account IDs allowed to pull images."
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "Tags applied to all resources."
  type        = map(string)
  default     = {}
}
