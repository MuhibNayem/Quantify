terraform {
  required_version = ">= 1.5.0"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }

  # Remote state: configure via `terraform init -backend-config=backend.hcl`.
  # backend.hcl:
  #   resource_group_name  = "quantify-terraform"
  #   storage_account_name = "quantifytfstate"
  #   container_name       = "tfstate"
  #   key                  = "azure/prod/terraform.tfstate"
  #   use_azuread_auth     = true
  backend "azurerm" {}
}
