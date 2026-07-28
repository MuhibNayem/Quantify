output "instance_name" {
  description = "Name of the primary instance."
  value       = google_sql_database_instance.primary.name
}

output "connection_name" {
  description = "Connection name of the primary instance."
  value       = google_sql_database_instance.primary.connection_name
}

output "private_ip_address" {
  description = "Private IP address of the primary instance."
  value       = google_sql_database_instance.primary.private_ip_address
}

output "replica_private_ip_address" {
  description = "Private IP address of the read replica."
  value       = var.create_read_replica ? google_sql_database_instance.replica[0].private_ip_address : null
}

output "port" {
  description = "Database port."
  value       = 5432
}

output "db_name" {
  description = "Initial database name."
  value       = var.db_name
}

output "username" {
  description = "Master username."
  value       = var.username
}

output "password" {
  description = "Master password."
  value       = random_password.this.result
  sensitive   = true
}
