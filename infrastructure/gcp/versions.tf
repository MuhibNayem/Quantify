terraform {
  required_version = ">= 1.5.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 6.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }

  # Remote state: configure via `terraform init -backend-config=backend.hcl`.
  # backend.hcl:
  #   bucket = "quantify-terraform-state"
  #   prefix = "gcp/prod/terraform.tfstate"
  backend "gcs" {}
}
