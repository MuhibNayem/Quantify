output "replication_group_id" {
  description = "ID of the replication group."
  value       = aws_elasticache_replication_group.this.id
}

output "primary_endpoint" {
  description = "Primary endpoint address."
  value       = aws_elasticache_replication_group.this.primary_endpoint_address
}

output "reader_endpoint" {
  description = "Reader endpoint address."
  value       = aws_elasticache_replication_group.this.reader_endpoint_address
}

output "configuration_endpoint" {
  description = "Configuration endpoint for cluster-mode clients."
  value       = aws_elasticache_replication_group.this.configuration_endpoint_address
}

output "port" {
  description = "Redis port."
  value       = var.port
}

output "auth_token" {
  description = "Auth token for transit-encrypted connections."
  value       = random_password.auth_token.result
  sensitive   = true
}

output "security_group_id" {
  description = "Security group ID attached to the cluster."
  value       = aws_security_group.this.id
}
