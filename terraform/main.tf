terraform {
    required_providers {
        azurerm = {
            source = "hashicorp/azurerm"
            version = "4.81.0"
        }
    }
}

provider "azurerm" {
    resource_provider_registrations = "none"
    features {}
}

resource "azurerm_resource_group" "equislice-rg" {
    name = "equiclice-rg"
    location = "Korea Central"
}

resource "azurerm_storage_account" "equislice-sa" {
    name = "equislice"
    location = azurerm_resource_group.equislice-rg.location
    resource_group_name = azurerm_resource_group.equislice-rg.name
    account_tier = "Standard"
    account_replication_type = "LRS"

    tags = {
        environment = "production"
        provider = "terraform"
    }
}

resource "azurerm_storage_container" "equislice-sa-containers" {
    for_each = toset(["equirectangular", "equirectangular-slice"])
    name = each.value
    storage_account_id = azurerm_storage_account.equislice-sa.id
    container_access_type = "private"
}

resource "azurerm_storage_queue" "equislice-sa-queue" {
    name = "panorama-slice"
    storage_account_id = azurerm_storage_account.equislice-sa.id
}

resource "azurerm_storage_table" "equislice-sa-table" {
    name = "Panorama"
    storage_account_name = azurerm_storage_account.equislice-sa.name
}

resource "azurerm_service_plan" "backend-sp" {
    name                = "backend-sp"
    location            = "Korea Central"

    os_type             = "Linux"
    resource_group_name = azurerm_resource_group.equislice-rg.name
    sku_name            = "F1"
}

resource "azurerm_linux_web_app" "backend" {
    name                = "equislicebackend"
    location            = "Korea Central"
    resource_group_name = azurerm_resource_group.equislice-rg.name
    service_plan_id     = azurerm_service_plan.backend-sp.id

    app_settings = {
        AZURE_CONNECTION_STRING=""
        FRONTEND_URL=""
    }

    site_config {
        application_stack {
            go_version = "1.24"
            docker_image_name = ""
            docker_registry_url = ""
            docker_registry_username = ""
            docker_registry_password = ""
        }
    }
}

# WHY COMMENTED? BECAUSE TERRAFORM  DOES NOT YET SUPPORT GO RUNTIME IN FUNCTION APP IN TERRAFORM
#
# resource "azurerm_service_plan" "equislice-plan" {
#     name                = "flex-plan"
#     location            = "Korea Central"
#     resource_group_name = azurerm_resource_group.equislice-rg.name
#     os_type             = "Linux"
#     sku_name            = "FC1"
# }

# resource "azurerm_storage_container" "deployment-package" {
#     name = "deployment-package"
#     storage_account_id = azurerm_storage_account.equislice-sa.id
#     container_access_type = "private"
# }

# resource "azurerm_function_app_flex_consumption" "equislice-slicer-app" {
#     name                       = "equislice-slicer"
#     location                    = "Korea Central"
#     resource_group_name         = azurerm_resource_group.equislice-rg.name
#     runtime_name                = "go"
#     runtime_version             = "1.0"
#     service_plan_id             = azurerm_service_plan.equislice-plan.id
#
#     storage_authentication_type = "StorageAccountConnectionString"
#     storage_container_endpoint  = "${azurerm_storage_account.equislice-sa.primary_blob_endpoint}${azurerm_storage_container.deployment-package.name}"
#     storage_container_type      = "blobContainer"
#     storage_access_key= ""
#
#     instance_memory_in_mb = 2048
#     maximum_instance_count = 3
#
#     app_settings = {
#         EQUISLICE_STORAGE_CONNECTION_STRING = ""
#         AzureWebJobsStorage = ""
#     }
#
#     site_config {}
# }