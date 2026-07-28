output "vnet_id" {
  description = "ID of the virtual network."
  value       = azurerm_virtual_network.this.id
}

output "vnet_name" {
  description = "Name of the virtual network."
  value       = azurerm_virtual_network.this.name
}

output "aks_subnet_id" {
  description = "ID of the AKS subnet."
  value       = azurerm_subnet.aks.id
}

output "postgres_subnet_id" {
  description = "ID of the PostgreSQL subnet."
  value       = azurerm_subnet.postgres.id
}

output "redis_subnet_id" {
  description = "ID of the Redis subnet."
  value       = azurerm_subnet.redis.id
}

output "private_endpoint_subnet_id" {
  description = "ID of the private endpoints subnet."
  value       = azurerm_subnet.private_endpoints.id
}
