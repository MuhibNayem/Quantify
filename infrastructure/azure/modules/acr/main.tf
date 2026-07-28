resource "azurerm_container_registry" "this" {
  name                = "${replace(var.name, "-", "")}acr"
  resource_group_name = var.resource_group_name
  location            = var.location
  sku                    = "Premium"
  admin_enabled          = false
  retention_policy_in_days = 30
  trust_policy_enabled   = true

  tags = var.tags
}
