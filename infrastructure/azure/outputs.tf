output "resource_group_name" {
  description = "Name of the resource group."
  value       = azurerm_resource_group.this.name
}

output "aks_cluster_name" {
  description = "Name of the AKS cluster."
  value       = module.aks.cluster_name
}

output "aks_oidc_issuer_url" {
  description = "OIDC issuer URL for workload identity."
  value       = module.aks.oidc_issuer_url
}

output "kubeconfig_command" {
  description = "Command to configure kubectl for the cluster."
  value       = "az aks get-credentials --resource-group ${azurerm_resource_group.this.name} --name ${module.aks.cluster_name}"
}

output "postgres_fqdn" {
  description = "FQDN of the PostgreSQL server."
  value       = var.enable_postgres ? module.postgres[0].fqdn : null
}

output "postgres_replica_fqdn" {
  description = "FQDN of the PostgreSQL read replica."
  value       = var.enable_postgres ? module.postgres[0].replica_fqdn : null
}

output "redis_hostname" {
  description = "Hostname of the Redis cache."
  value       = var.enable_redis ? module.redis[0].hostname : null
}

output "servicebus_endpoint" {
  description = "Endpoint of the Service Bus namespace."
  value       = var.enable_servicebus ? module.servicebus[0].namespace_endpoint : null
}

output "storage_account_name" {
  description = "Name of the storage account."
  value       = module.storage.storage_account_name
}

output "storage_primary_blob_endpoint" {
  description = "Primary blob endpoint of the storage account."
  value       = module.storage.primary_blob_endpoint
}

output "acr_login_server" {
  description = "Login server of the container registry."
  value       = module.acr.login_server
}

output "backend_workload_identity_client_id" {
  description = "Client ID of the backend workload identity."
  value       = module.iam.backend_client_id
}

output "ai_workload_identity_client_id" {
  description = "Client ID of the AI service workload identity."
  value       = module.iam.ai_client_id
}
