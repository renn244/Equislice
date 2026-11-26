terraform {
  required_version = ">= 1.7, < 2.0"

  # State lives in HCP Terraform. The workspace uses local execution, so plans and
  # applies still run on this machine with the signed-in Azure CLI session.
  cloud {
    organization = "Renn224"
    workspaces {
      name = "equislice-production"
    }
  }

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "4.81.0"
    }
    azapi = {
      source  = "Azure/azapi"
      version = "2.8.0"
    }
  }
}

provider "azurerm" {
  subscription_id                 = var.subscription_id
  resource_provider_registrations = "none"
  features {}
}

provider "azapi" {
  subscription_id = var.subscription_id
}
