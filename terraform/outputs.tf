output "backend_app_name" {
  value = azurerm_linux_web_app.backend.name
}

output "backend_api_url" {
  value = "https://${azurerm_linux_web_app.backend.default_hostname}/api"
}

output "function_app_name" {
  value = azapi_resource.function.name
}

output "storage_account_name" {
  value = azurerm_storage_account.equislice-sa.name
}

output "resource_group_name" {
  value = data.azurerm_resource_group.existing.name
}
