resource "azurerm_user_assigned_identity" "backend" {
  name                = "${var.name}-backend"
  resource_group_name = var.resource_group_name
  location            = var.location

  tags = var.tags
}

resource "azurerm_user_assigned_identity" "ai" {
  name                = "${var.name}-ai"
  resource_group_name = var.resource_group_name
  location            = var.location

  tags = var.tags
}

resource "azurerm_federated_identity_credential" "backend" {
  name                = "${var.name}-backend"
  resource_group_name = var.resource_group_name
  parent_id           = azurerm_user_assigned_identity.backend.id
  audience            = ["api://AzureADTokenExchange"]
  issuer              = var.aks_oidc_issuer_url
  subject             = "system:serviceaccount:${var.namespace}:${var.backend_service_account}"
}

resource "azurerm_federated_identity_credential" "ai" {
  name                = "${var.name}-ai"
  resource_group_name = var.resource_group_name
  parent_id           = azurerm_user_assigned_identity.ai.id
  audience            = ["api://AzureADTokenExchange"]
  issuer              = var.aks_oidc_issuer_url
  subject             = "system:serviceaccount:${var.namespace}:${var.ai_service_account}"
}

resource "azurerm_role_assignment" "kubelet_acr_pull" {
  scope                = var.acr_id
  role_definition_name = "AcrPull"
  principal_id         = var.aks_kubelet_principal_id
}

resource "azurerm_role_assignment" "backend_acr_pull" {
  scope                = var.acr_id
  role_definition_name = "AcrPull"
  principal_id         = azurerm_user_assigned_identity.backend.principal_id
}

resource "azurerm_role_assignment" "ai_acr_pull" {
  scope                = var.acr_id
  role_definition_name = "AcrPull"
  principal_id         = azurerm_user_assigned_identity.ai.principal_id
}

resource "azurerm_role_assignment" "backend_storage" {
  scope                = var.storage_account_id
  role_definition_name = "Storage Blob Data Contributor"
  principal_id         = azurerm_user_assigned_identity.backend.principal_id
}

resource "azurerm_role_assignment" "ai_storage" {
  scope                = var.storage_account_id
  role_definition_name = "Storage Blob Data Contributor"
  principal_id         = azurerm_user_assigned_identity.ai.principal_id
}

resource "azurerm_role_assignment" "backend_metrics" {
  scope                = var.storage_account_id
  role_definition_name = "Monitoring Metrics Publisher"
  principal_id         = azurerm_user_assigned_identity.backend.principal_id
}
