resource "random_string" "suffix" {
  length  = 6
  special = false
  upper   = false
}

locals {
  account_name = "${substr(replace(lower(var.name), "-", ""), 0, 18)}${random_string.suffix.result}"
}

resource "azurerm_storage_account" "this" {
  name                = local.account_name
  resource_group_name = var.resource_group_name
  location            = var.location

  account_tier             = "Standard"
  account_replication_type = "GRS"
  account_kind             = "StorageV2"

  min_tls_version                 = "TLS1_2"
  public_network_access_enabled   = false
  allow_nested_items_to_be_public = false

  blob_properties {
    versioning_enabled = true

    cors_rule {
      allowed_headers    = ["*"]
      allowed_methods    = ["GET", "PUT", "POST", "HEAD"]
      allowed_origins    = var.cors_origins
      exposed_headers    = ["ETag"]
      max_age_in_seconds = 3600
    }

    delete_retention_policy {
      days = 30
    }

    container_delete_retention_policy {
      days = 30
    }
  }

  tags = var.tags
}

resource "azurerm_storage_container" "this" {
  name                  = "reports"
  storage_account_id    = azurerm_storage_account.this.id
  container_access_type = "private"
}
