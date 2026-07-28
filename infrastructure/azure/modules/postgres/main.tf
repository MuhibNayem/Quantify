resource "random_password" "this" {
  length           = 32
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

resource "azurerm_private_dns_zone" "this" {
  name                = "${var.name}.private.postgres.database.azure.com"
  resource_group_name = var.resource_group_name

  tags = var.tags
}

resource "azurerm_private_dns_zone_virtual_network_link" "this" {
  name                  = "${var.name}-dns-link"
  private_dns_zone_name = azurerm_private_dns_zone.this.name
  virtual_network_id    = var.vnet_id
  resource_group_name   = var.resource_group_name

  tags = var.tags
}

resource "azurerm_postgresql_flexible_server" "primary" {
  name                = var.name
  resource_group_name = var.resource_group_name
  location            = var.location

  version                = "15"
  administrator_login    = var.administrator_login
  administrator_password = random_password.this.result

  sku_name            = var.sku_name
  storage_mb          = var.storage_size_gb * 1024
  auto_grow_enabled   = true
  backup_retention_days = 7

  delegated_subnet_id = var.subnet_id
  private_dns_zone_id = azurerm_private_dns_zone.this.id

  dynamic "high_availability" {
    for_each = var.high_availability ? [1] : []
    content {
      mode                      = "ZoneRedundant"
      standby_availability_zone = "2"
    }
  }

  tags = var.tags
}

resource "azurerm_postgresql_flexible_server_database" "this" {
  name      = var.database_name
  server_id = azurerm_postgresql_flexible_server.primary.id
  charset   = "UTF8"
  collation = "en_US.utf8"
}

resource "azurerm_postgresql_flexible_server_configuration" "max_connections" {
  name      = "max_connections"
  server_id = azurerm_postgresql_flexible_server.primary.id
  value     = "300"
}

resource "azurerm_postgresql_flexible_server_configuration" "log_min_duration" {
  name      = "log_min_duration_statement"
  server_id = azurerm_postgresql_flexible_server.primary.id
  value     = "1000"
}

resource "azurerm_postgresql_flexible_server" "replica" {
  count = var.create_read_replica ? 1 : 0

  name                = "${var.name}-replica"
  resource_group_name = var.resource_group_name
  location            = var.location

  create_mode      = "Replica"
  source_server_id = azurerm_postgresql_flexible_server.primary.id

  sku_name          = var.sku_name
  auto_grow_enabled = true

  delegated_subnet_id = var.subnet_id
  private_dns_zone_id = azurerm_private_dns_zone.this.id

  tags = var.tags
}
