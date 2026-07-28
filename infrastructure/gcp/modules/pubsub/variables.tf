variable "name" {
  description = "Name prefix for Pub/Sub resources."
  type        = string
}

variable "labels" {
  description = "Labels applied to resources."
  type        = map(string)
  default     = {}
}
