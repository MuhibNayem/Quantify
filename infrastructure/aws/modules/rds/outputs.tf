output "identifier" {
  description = "Identifier of the primary database or cluster."
  value       = var.identifier
}

output "endpoint" {
  description = "Writer endpoint."
  value       = local.endpoint
}

output "reader_endpoint" {
  description = "Reader endpoint."
  value       = local.reader_endpoint
}

output "port" {
  description = "Database port."
  value       = var.port
}

output "db_name" {
  description = "Initial database name."
  value       = var.db_name
}

output "username" {
  description = "Master username."
  value       = var.username
}

output "secret_arn" {
  description = "ARN of the Secrets Manager secret holding credentials."
  value       = aws_secretsmanager_secret.credentials.arn
}

output "security_group_id" {
  description = "Security group ID attached to the database."
  value       = aws_security_group.this.id
}
