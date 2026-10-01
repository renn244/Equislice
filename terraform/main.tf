# This group also hosts unrelated projects. Only EquiSlice resources are managed.
data "azurerm_resource_group" "existing" {
  name = var.resource_group_name
}

locals {
  storage_name              = "equislice"
  storage_location          = "koreacentral"
  backend_location          = "japaneast"
  backend_name              = "equislice-api-renn244"
  function_name             = "equislice-slicer-renn244"
  deployment_container_name = "app-package-equislice-slicer-renn244-5de8687"
  resource_group_id         = "/subscriptions/${var.subscription_id}/resourceGroups/${var.resource_group_name}"
  storage_id                = "${local.resource_group_id}/providers/Microsoft.Storage/storageAccounts/${local.storage_name}"
  backend_id                = "${local.resource_group_id}/providers/Microsoft.Web/sites/${local.backend_name}"
  function_id               = "${local.resource_group_id}/providers/Microsoft.Web/sites/${local.function_name}"
}

resource "azurerm_storage_account" "equislice-sa" {
  name                            = local.storage_name
  resource_group_name             = data.azurerm_resource_group.existing.name
  location                        = local.storage_location
  account_kind                    = "StorageV2"
  account_tier                    = "Standard"
  account_replication_type        = "LRS"
  access_tier                     = "Hot"
  min_tls_version                 = "TLS1_2"
  https_traffic_only_enabled      = true
  allow_nested_items_to_be_public = false
  shared_access_key_enabled       = true
  public_network_access_enabled   = true

  blob_properties {
    delete_retention_policy {
      days = 7
    }
    container_delete_retention_policy {
      days = 7
    }
    dynamic "cors_rule" {
      for_each = ["http://localhost:5173", var.frontend_url]
      content {
        allowed_origins    = [cors_rule.value]
        allowed_methods    = ["PUT", "OPTIONS", "GET", "HEAD"]
        allowed_headers    = ["*"]
        exposed_headers    = ["*"]
        max_age_in_seconds = 86400
      }
    }
  }

  lifecycle {
    prevent_destroy = true
  }
}

resource "azurerm_storage_container" "equislice-sa-containers" {
  for_each              = toset(["equirectangular", "equirectangular-slice"])
  name                  = each.value
  storage_account_id    = azurerm_storage_account.equislice-sa.id
  container_access_type = "private"
}

resource "azurerm_storage_container" "deployment" {
  name                  = local.deployment_container_name
  storage_account_id    = azurerm_storage_account.equislice-sa.id
  container_access_type = "private"
}

# azure-webjobs-hosts and azure-webjobs-secrets remain owned by the Functions host.
resource "azurerm_storage_queue" "equislice-sa-queue" {
  name               = "panorama-slice"
  storage_account_id = azurerm_storage_account.equislice-sa.id
}

resource "azurerm_storage_queue" "poison" {
  name               = "panorama-slice-poison"
  storage_account_id = azurerm_storage_account.equislice-sa.id
}

resource "azurerm_storage_table" "equislice-sa-table" {
  name                 = "Panorama"
  storage_account_name = azurerm_storage_account.equislice-sa.name
}

resource "azurerm_service_plan" "backend-sp" {
  name                = "equislice-api-japaneast-plan"
  resource_group_name = data.azurerm_resource_group.existing.name
  location            = local.backend_location
  os_type             = "Linux"
  sku_name            = "F1"
}

resource "azurerm_linux_web_app" "backend" {
  name                                           = local.backend_name
  resource_group_name                            = data.azurerm_resource_group.existing.name
  location                                       = local.backend_location
  service_plan_id                                = azurerm_service_plan.backend-sp.id
  https_only                                     = true
  ftp_publish_basic_authentication_enabled       = false
  webdeploy_publish_basic_authentication_enabled = false

  logs {
    http_logs {
      file_system {
        retention_in_days = 3
        retention_in_mb   = 100
      }
    }
  }

  app_settings = {
    AZURE_CONNECTION_STRING             = azurerm_storage_account.equislice-sa.primary_connection_string
    FRONTEND_URL                        = var.frontend_url
    SENTRY_DSN                          = var.backend_sentry_dsn
    SENTRY_ENVIRONMENT                  = var.backend_sentry_environment
    WEBSITE_PORT                        = "3000"
    WEBSITES_ENABLE_APP_SERVICE_STORAGE = "false"
    OTEL_EXPORTER_OTLP_ENDPOINT         = var.grafana_otlp_endpoint
    OTEL_EXPORTER_OTLP_HEADERS          = var.grafana_otlp_headers
    OTEL_SERVICE_NAME                   = "equislice-api"
    OTEL_RESOURCE_ATTRIBUTES            = "deployment.environment=production"
  }

  site_config {
    always_on               = false
    ftps_state              = "FtpsOnly"
    http2_enabled           = false
    minimum_tls_version     = "1.2"
    scm_minimum_tls_version = "1.2"
    use_32_bit_worker       = true
    application_stack {
      docker_image_name   = var.backend_image
      docker_registry_url = "https://ghcr.io"
    }
  }

  lifecycle {
    # GitHub Actions owns releases, so an infra apply must not roll the image back.
    ignore_changes = [site_config[0].application_stack[0].docker_image_name]
  }
}

resource "azurerm_service_plan" "function" {
  name                = "ASP-rgrenatodsantosjr91317-af55"
  resource_group_name = data.azurerm_resource_group.existing.name
  location            = local.storage_location
  os_type             = "Linux"
  sku_name            = "FC1"
}

# The shared default workspace is not owned by EquiSlice.
data "azurerm_log_analytics_workspace" "existing" {
  name                = "DefaultWorkspace-${var.subscription_id}-SE"
  resource_group_name = "DefaultResourceGroup-SE"
}

resource "azurerm_application_insights" "function" {
  name                = local.function_name
  resource_group_name = data.azurerm_resource_group.existing.name
  location            = local.storage_location
  application_type    = "web"
  retention_in_days   = 90
  workspace_id        = data.azurerm_log_analytics_workspace.existing.id
}

# AzureRM 4.81.0 rejects runtime_name = "go". The ARM API supports this live app.
resource "azapi_resource" "function" {
  type      = "Microsoft.Web/sites@2024-04-01"
  name      = local.function_name
  parent_id = data.azurerm_resource_group.existing.id
  location  = local.storage_location
  tags = {
    "hidden-link: /app-insights-resource-id" = azurerm_application_insights.function.id
  }

  body = {
    kind = "functionapp,linux"
    properties = {
      serverFarmId        = azurerm_service_plan.function.id
      httpsOnly           = true
      publicNetworkAccess = "Enabled"
      functionAppConfig = {
        runtime = {
          name    = "go"
          version = "1.0"
        }
        deployment = {
          storage = {
            type  = "blobContainer"
            value = "${azurerm_storage_account.equislice-sa.primary_blob_endpoint}${azurerm_storage_container.deployment.name}"
            authentication = {
              type                               = "StorageAccountConnectionString"
              storageAccountConnectionStringName = "DEPLOYMENT_STORAGE_CONNECTION_STRING"
            }
          }
        }
        scaleAndConcurrency = {
          instanceMemoryMB     = 2048
          maximumInstanceCount = 100
        }
      }
      siteConfig = {
        http20Enabled    = false
        ftpsState        = "FtpsOnly"
        minTlsVersion    = "1.2"
        scmMinTlsVersion = "1.2"
        cors = {
          allowedOrigins     = ["https://portal.azure.com"]
          supportCredentials = false
        }
        appSettings = [for name, value in {
          APPLICATIONINSIGHTS_CONNECTION_STRING = azurerm_application_insights.function.connection_string
          AzureWebJobsStorage                   = azurerm_storage_account.equislice-sa.primary_connection_string
          DEPLOYMENT_STORAGE_CONNECTION_STRING  = azurerm_storage_account.equislice-sa.primary_connection_string
          EQUISLICE_STORAGE_CONNECTION_STRING   = azurerm_storage_account.equislice-sa.primary_connection_string
          SENTRY_DSN                            = var.function_sentry_dsn
          SENTRY_ENVIRONMENT                    = var.function_sentry_environment
        } : { name = name, value = value }]
      }
    }
  }
  response_export_values = []
}

# Reuse the existing GitHub OIDC identity; no new GitHub credentials are needed.
resource "azurerm_role_assignment" "github_backend" {
  name                 = "98989f27-3011-4daa-a8ab-ece1bb3c97aa"
  scope                = azurerm_linux_web_app.backend.id
  role_definition_name = "Website Contributor"
  principal_id         = var.github_principal_id
}

resource "azurerm_role_assignment" "github_function" {
  name                 = "80df005d-1dfc-4945-8054-95ce694a1b3d"
  scope                = azapi_resource.function.id
  role_definition_name = "Website Contributor"
  principal_id         = var.github_principal_id
}
