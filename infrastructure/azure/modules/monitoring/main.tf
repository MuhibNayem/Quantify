resource "azurerm_monitor_action_group" "this" {
  name                = "${var.name}-alerts"
  resource_group_name = var.resource_group_name
  short_name          = "quantify"

  dynamic "email_receiver" {
    for_each = toset(var.alert_email_endpoints)
    content {
      name          = replace(email_receiver.value, "@", "-at-")
      email_address = email_receiver.value
    }
  }

  tags = var.tags
}

resource "azurerm_monitor_metric_alert" "postgres_cpu" {
  count = var.postgres_server_id != null ? 1 : 0

  name                 = "${var.name}-postgres-cpu"
  resource_group_name  = var.resource_group_name
  scopes               = [var.postgres_server_id]
  frequency            = "PT5M"
  window_size          = "PT5M"
  severity             = 2

  criteria {
    metric_namespace = "Microsoft.DBforPostgreSQL/flexibleServers"
    metric_name      = "cpu_percent"
    aggregation      = "Average"
    operator         = "GreaterThan"
    threshold        = 80
  }

  action {
    action_group_id = azurerm_monitor_action_group.this.id
  }

  tags = var.tags
}

resource "azurerm_monitor_metric_alert" "postgres_storage" {
  count = var.postgres_server_id != null ? 1 : 0

  name                 = "${var.name}-postgres-storage"
  resource_group_name  = var.resource_group_name
  scopes               = [var.postgres_server_id]
  frequency            = "PT5M"
  window_size          = "PT5M"
  severity             = 1

  criteria {
    metric_namespace = "Microsoft.DBforPostgreSQL/flexibleServers"
    metric_name      = "storage_percent"
    aggregation      = "Average"
    operator         = "GreaterThan"
    threshold        = 85
  }

  action {
    action_group_id = azurerm_monitor_action_group.this.id
  }

  tags = var.tags
}

resource "azurerm_monitor_metric_alert" "redis_cpu" {
  count = var.redis_id != null ? 1 : 0

  name                 = "${var.name}-redis-cpu"
  resource_group_name  = var.resource_group_name
  scopes               = [var.redis_id]
  frequency            = "PT5M"
  window_size          = "PT5M"
  severity             = 2

  criteria {
    metric_namespace = "Microsoft.Cache/redis"
    metric_name      = "percentProcessorTime"
    aggregation      = "Average"
    operator         = "GreaterThan"
    threshold        = 75
  }

  action {
    action_group_id = azurerm_monitor_action_group.this.id
  }

  tags = var.tags
}

resource "azurerm_monitor_metric_alert" "redis_memory" {
  count = var.redis_id != null ? 1 : 0

  name                 = "${var.name}-redis-memory"
  resource_group_name  = var.resource_group_name
  scopes               = [var.redis_id]
  frequency            = "PT5M"
  window_size          = "PT5M"
  severity             = 2

  criteria {
    metric_namespace = "Microsoft.Cache/redis"
    metric_name      = "usedmemorypercentage"
    aggregation      = "Average"
    operator         = "GreaterThan"
    threshold        = 80
  }

  action {
    action_group_id = azurerm_monitor_action_group.this.id
  }

  tags = var.tags
}
