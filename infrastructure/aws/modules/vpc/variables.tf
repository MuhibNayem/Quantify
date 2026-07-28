variable "name" {
  description = "Name prefix for VPC resources."
  type        = string
}

variable "cidr" {
  description = "VPC CIDR block."
  type        = string
}

variable "azs" {
  description = "Availability zones to span."
  type        = list(string)
}

variable "enable_nat_gateway" {
  description = "Provision NAT gateway(s)."
  type        = bool
  default     = true
}

variable "single_nat_gateway" {
  description = "Use a single shared NAT gateway instead of one per AZ."
  type        = bool
  default     = false
}

variable "enable_flow_logs" {
  description = "Enable VPC flow logs to CloudWatch."
  type        = bool
  default     = true
}

variable "tags" {
  description = "Tags applied to all resources."
  type        = map(string)
  default     = {}
}
