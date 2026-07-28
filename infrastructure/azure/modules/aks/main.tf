resource "azurerm_user_assigned_identity" "cluster" {
  name                = "${var.name}-identity"
  resource_group_name = var.resource_group_name
  location            = var.location

  tags = var.tags
}

resource "azurerm_role_assignment" "cluster_network" {
  scope                = var.vnet_subnet_id
  role_definition_name = "Network Contributor"
  principal_id         = azurerm_user_assigned_identity.cluster.principal_id
}

resource "azurerm_kubernetes_cluster" "this" {
  name                      = var.name
  resource_group_name       = var.resource_group_name
  location                  = var.location
  dns_prefix                = replace(var.name, "-", "")
  kubernetes_version        = var.kubernetes_version
  private_cluster_enabled   = var.private_cluster
  oidc_issuer_enabled       = true
  workload_identity_enabled = true
  automatic_upgrade_channel = "stable"
  sku_tier                  = "Standard"

  default_node_pool {
    name                 = "system"
    vm_size              = "Standard_D2s_v3"
    node_count           = 1
    vnet_subnet_id       = var.vnet_subnet_id
    os_disk_size_gb      = 128
    zones                = [1, 2, 3]
    auto_scaling_enabled = false
  }

  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.cluster.id]
  }

  network_profile {
    network_plugin    = "azure"
    network_policy    = "azure"
    outbound_type     = "userAssignedNATGateway"
    load_balancer_sku = "standard"
  }

  oms_agent {
    log_analytics_workspace_id = var.log_analytics_workspace_id
  }

  tags = var.tags
}

resource "azurerm_kubernetes_cluster_node_pool" "this" {
  for_each = var.node_pools

  name                  = each.key
  kubernetes_cluster_id = azurerm_kubernetes_cluster.this.id
  vm_size               = each.value.vm_size
  node_count            = each.value.node_count
  auto_scaling_enabled  = true
  min_count             = each.value.min_count
  max_count             = each.value.max_count
  os_disk_size_gb       = each.value.os_disk_size_gb
  vnet_subnet_id        = var.vnet_subnet_id
  zones                 = [1, 2, 3]
  node_labels           = each.value.labels
  node_taints           = each.value.taints

  tags = var.tags
}
