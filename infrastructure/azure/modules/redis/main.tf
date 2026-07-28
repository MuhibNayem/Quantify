resource "azurerm_redis_cache" "this" {
  name                = var.name
  resource_group_name = var.resource_group_name
  location            = var.location

  capacity                      = var.capacity
  family                        = "P"
  sku_name                      = "Premium"
  subnet_id                     = var.subnet_id
  redis_version                 = 6
  minimum_tls_version           = "1.2"
  non_ssl_port_enabled          = false
  public_network_access_enabled = false

  shard_count          = var.shard_count
  replicas_per_primary = var.replica_count

  redis_configuration {
    maxmemory_policy   = "volatile-lru"
    aof_backup_enabled = true
  }

  patch_schedule {
    day_of_week    = "Sunday"
    start_hour_utc = 4
  }

  tags = var.tags
}
