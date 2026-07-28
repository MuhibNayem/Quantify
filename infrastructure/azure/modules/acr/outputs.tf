output "login_server" {
  description = "Login server of the registry."
  value       = azurerm_container_registry.this.login_server
}

output "acr_id" {
  description = "ID of the registry."
  value       = azurerm_container_registry.this.id
}
