output "server_name" {
  description = "Name of the primary server."
  value       = azurerm_postgresql_flexible_server.primary.name
}

output "server_id" {
  description = "ID of the primary server."
  value       = azurerm_postgresql_flexible_server.primary.id
}

output "fqdn" {
  description = "FQDN of the primary server."
  value       = azurerm_postgresql_flexible_server.primary.fqdn
}

output "replica_fqdn" {
  description = "FQDN of the read replica."
  value       = var.create_read_replica ? azurerm_postgresql_flexible_server.replica[0].fqdn : null
}

output "port" {
  description = "Database port."
  value       = 5432
}

output "database_name" {
  description = "Initial database name."
  value       = var.database_name
}

output "administrator_login" {
  description = "Administrator username."
  value       = var.administrator_login
}

output "administrator_password" {
  description = "Administrator password."
  value       = random_password.this.result
  sensitive   = true
}
