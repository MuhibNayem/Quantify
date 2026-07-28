output "name" {
  description = "Name of the Redis cache."
  value       = azurerm_redis_cache.this.name
}

output "id" {
  description = "ID of the Redis cache."
  value       = azurerm_redis_cache.this.id
}

output "hostname" {
  description = "Hostname of the Redis cache."
  value       = azurerm_redis_cache.this.hostname
}

output "ssl_port" {
  description = "TLS port of the Redis cache."
  value       = azurerm_redis_cache.this.ssl_port
}

output "primary_access_key" {
  description = "Primary access key."
  value       = azurerm_redis_cache.this.primary_access_key
  sensitive   = true
}
