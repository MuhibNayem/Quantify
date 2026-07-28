output "broker_id" {
  description = "ID of the MQ broker."
  value       = aws_mq_broker.this.id
}

output "endpoint" {
  description = "AMQP endpoint of the broker."
  value       = aws_mq_broker.this.instances[0].endpoints[0]
}

output "management_endpoint" {
  description = "Management console URL of the broker."
  value       = aws_mq_broker.this.instances[0].console_url
}

output "username" {
  description = "Admin username."
  value       = var.username
}

output "password" {
  description = "Admin password."
  value       = local.admin_password
  sensitive   = true
}

output "security_group_id" {
  description = "Security group ID attached to the broker."
  value       = aws_security_group.this.id
}
