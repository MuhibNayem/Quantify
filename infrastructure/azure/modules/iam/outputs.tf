output "backend_identity_id" {
  description = "ID of the backend user-assigned identity."
  value       = azurerm_user_assigned_identity.backend.id
}

output "backend_client_id" {
  description = "Client ID of the backend user-assigned identity."
  value       = azurerm_user_assigned_identity.backend.client_id
}

output "ai_identity_id" {
  description = "ID of the AI service user-assigned identity."
  value       = azurerm_user_assigned_identity.ai.id
}

output "ai_client_id" {
  description = "Client ID of the AI service user-assigned identity."
  value       = azurerm_user_assigned_identity.ai.client_id
}
