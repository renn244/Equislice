# Secrets live only in Key Vault. Terraform manages the vault, access, and the
# app setting references, but never the secret values, so they stay out of state.
data "azurerm_client_config" "current" {}

locals {
  key_vault_name = "equislice-kv-renn244"
  # Values are added once by hand (see README); these are only their names.
  key_vault_secrets = {
    backend_sentry_dsn   = "backend-sentry-dsn"
    function_sentry_dsn  = "function-sentry-dsn"
    grafana_otlp_headers = "grafana-otlp-headers"
  }
  # Versionless references follow secret rotation automatically.
  key_vault_refs = {
    for key, name in local.key_vault_secrets :
    key => "@Microsoft.KeyVault(VaultName=${local.key_vault_name};SecretName=${name})"
  }
}

resource "azurerm_key_vault" "equislice" {
  name                       = local.key_vault_name
  resource_group_name        = data.azurerm_resource_group.existing.name
  location                   = local.storage_location
  tenant_id                  = data.azurerm_client_config.current.tenant_id
  sku_name                   = "standard"
  rbac_authorization_enabled = true
  soft_delete_retention_days = 7
  purge_protection_enabled   = false

  lifecycle {
    prevent_destroy = true
  }
}

# One identity shared by both apps, created separately from the apps so the vault
# and its access can exist before any app setting points at it.
resource "azurerm_user_assigned_identity" "apps" {
  name                = "equislice-apps"
  resource_group_name = data.azurerm_resource_group.existing.name
  location            = local.storage_location
}

resource "azurerm_role_assignment" "apps_kv_secrets_user" {
  scope                = azurerm_key_vault.equislice.id
  role_definition_name = "Key Vault Secrets User"
  principal_id         = azurerm_user_assigned_identity.apps.principal_id
  principal_type       = "ServicePrincipal"
}

# Lets the person in var.secrets_admin_object_id add and rotate secret values.
resource "azurerm_role_assignment" "admin_kv_secrets_officer" {
  scope                = azurerm_key_vault.equislice.id
  role_definition_name = "Key Vault Secrets Officer"
  principal_id         = var.secrets_admin_object_id
  principal_type       = "User"
}
