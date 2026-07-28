resource "random_string" "suffix" {
  length  = 6
  special = false
  upper   = false
}

resource "azurerm_servicebus_namespace" "this" {
  name                = "${var.name}-sb-${random_string.suffix.result}"
  resource_group_name = var.resource_group_name
  location            = var.location
  sku                 = "Premium"
  capacity            = 1

  tags = var.tags
}

resource "azurerm_servicebus_queue" "tasks" {
  name         = "tasks"
  namespace_id = azurerm_servicebus_namespace.this.id

  partitioning_enabled               = true
  max_delivery_count                   = 10
  lock_duration                        = "PT60S"
  default_message_ttl                  = "P14D"
  dead_lettering_on_message_expiration = true
}
