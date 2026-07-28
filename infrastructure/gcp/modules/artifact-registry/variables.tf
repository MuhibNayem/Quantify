variable "name" {
  description = "Prefix applied to repository IDs."
  type        = string
}

variable "region" {
  description = "Region for the repositories."
  type        = string
}

variable "repositories" {
  description = "Repository names to create."
  type        = list(string)
}

variable "keep_count" {
  description = "Number of most recent image versions to retain."
  type        = number
  default     = 30
}

variable "labels" {
  description = "Labels applied to resources."
  type        = map(string)
  default     = {}
}
