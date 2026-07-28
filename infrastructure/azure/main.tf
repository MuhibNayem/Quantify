locals {
  name = "quantify-${var.environment}"

  common_tags = merge(
    {
      Project     = "quantify"
      Environment = var.environment
      ManagedBy   = "terraform"
    },
    var.tags
  )
}

resource "azurerm_resource_group" "this" {
  name     = local.name
  location = var.location

  tags = local.common_tags
}

resource "azurerm_log_analytics_workspace" "this" {
  name                = "${local.name}-logs"
  resource_group_name = azurerm_resource_group.this.name
  location            = var.location
  sku                 = "PerGB2018"
  retention_in_days   = 30

  tags = local.common_tags
}

module "network" {
  source = "./modules/network"

  name                = local.name
  resource_group_name = azurerm_resource_group.this.name
  location            = var.location
  address_space       = var.vnet_cidr

  tags = local.common_tags
}

module "aks" {
  source = "./modules/aks"

  name                = "${local.name}-aks"
  resource_group_name = azurerm_resource_group.this.name
  location            = var.location

  kubernetes_version = var.aks_kubernetes_version
  private_cluster    = var.aks_private_cluster
  vnet_subnet_id     = module.network.aks_subnet_id

  node_pools = var.aks_node_pools

  log_analytics_workspace_id = azurerm_log_analytics_workspace.this.id

  tags = local.common_tags
}

module "postgres" {
  source = "./modules/postgres"
  count  = var.enable_postgres ? 1 : 0

  name                = "${local.name}-pg"
  resource_group_name = azurerm_resource_group.this.name
  location            = var.location

  sku_name            = var.db_sku_name
  storage_size_gb     = var.db_storage_size_gb
  high_availability   = var.db_high_availability
  create_read_replica = var.db_create_read_replica

  administrator_login = "quantify"
  database_name       = "inventory_db"

  subnet_id = module.network.postgres_subnet_id
  vnet_id   = module.network.vnet_id

  tags = local.common_tags
}

module "redis" {
  source = "./modules/redis"
  count  = var.enable_redis ? 1 : 0

  name                = "${local.name}-redis"
  resource_group_name = azurerm_resource_group.this.name
  location            = var.location

  capacity      = var.redis_sku_capacity
  shard_count   = var.redis_shard_count
  replica_count = var.redis_replica_count

  subnet_id = module.network.redis_subnet_id

  tags = local.common_tags
}

module "servicebus" {
  source = "./modules/servicebus"
  count  = var.enable_servicebus ? 1 : 0

  name                = local.name
  resource_group_name = azurerm_resource_group.this.name
  location            = var.location

  tags = local.common_tags
}

module "storage" {
  source = "./modules/storage"

  name                = local.name
  resource_group_name = azurerm_resource_group.this.name
  location            = var.location

  tags = local.common_tags
}

module "acr" {
  source = "./modules/acr"

  name                = local.name
  resource_group_name = azurerm_resource_group.this.name
  location            = var.location

  tags = local.common_tags
}

module "iam" {
  source = "./modules/iam"

  name                = local.name
  resource_group_name = azurerm_resource_group.this.name
  location            = var.location

  aks_oidc_issuer_url      = module.aks.oidc_issuer_url
  aks_kubelet_principal_id = module.aks.kubelet_identity_object_id
  acr_id                   = module.acr.acr_id
  storage_account_id       = module.storage.storage_account_id

  namespace               = "quantify"
  backend_service_account = "quantify-backend"
  ai_service_account      = "quantify-ai-service"

  tags = local.common_tags
}

module "monitoring" {
  source = "./modules/monitoring"
  count  = var.enable_monitoring ? 1 : 0

  name                = local.name
  resource_group_name = azurerm_resource_group.this.name
  location            = var.location

  log_analytics_workspace_id = azurerm_log_analytics_workspace.this.id
  postgres_server_id         = var.enable_postgres ? module.postgres[0].server_id : null
  redis_id                   = var.enable_redis ? module.redis[0].id : null

  alert_email_endpoints = var.alert_email_endpoints

  tags = local.common_tags
}
