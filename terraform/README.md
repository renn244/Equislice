# Adopt the live EquiSlice deployment

This configuration describes the Azure deployment inspected on September 15, 2026.
It is an adoption configuration for this environment, not a new-environment template.

| Resource | Existing setup |
| --- | --- |
| Shared resource group | `rg-renatodsantosjr9-1317` (referenced, not owned) |
| Storage | `equislice`, Korea Central, Standard LRS |
| Backend plan | `equislice-api-japaneast-plan`, Japan East, Linux F1 |
| Backend | `equislice-api-renn244`, public GHCR container |
| Function plan | `ASP-rgrenatodsantosjr91317-af55`, Korea Central, FC1 |
| Function | `equislice-slicer-renn244`, Go 1.0, 2048 MB, maximum 100 instances |
| Monitoring | Existing Function Application Insights, referencing the shared default workspace; both apps export OpenTelemetry to Grafana Cloud |
| Secrets | `equislice-kv-renn244` Key Vault (RBAC), read by the `equislice-apps` user-assigned identity |
| Deployment identity | Existing `equislice-github-actions`, Website Contributor on each app |

Terraform owns storage configuration, application environment settings, hosting,
and the existing app-scoped role assignments. GitHub Actions owns application code
releases. The backend image has a narrow `ignore_changes` rule so applying
infrastructure does not roll back a GitHub deployment.

The shared resource group, shared Log Analytics workspace, existing Entra application
and its GitHub OIDC trust are prerequisites. This configuration does not recreate or
broaden that identity. Azure Functions owns its `azure-webjobs-hosts` and
`azure-webjobs-secrets` containers and their contents.

## Prepare and review

Run from this directory in PowerShell with Terraform 1.7+ and Azure CLI installed:

```powershell
az login        # Only if the existing session has expired.
terraform login # Once per machine, for HCP Terraform state.
terraform init
terraform fmt -check
terraform validate
terraform plan '-out=adopt.tfplan'
```

`imports.tf` tells Terraform to adopt the existing resource IDs. Planning reads Azure
but does not change Azure resources or persist those imports into state. Do not apply
a plan that proposes unintended replacements, deletions, or configuration changes.
Once the reviewed plan matches the intended adoption:

```powershell
terraform apply adopt.tfplan
terraform plan
```

Import blocks can remain after adoption; Terraform skips resources already in state.
Do not use these fixed import IDs for a different environment.

## How credentials are connected

- Backend `AZURE_CONNECTION_STRING` comes from the managed storage account.
- Function `AzureWebJobsStorage`, `EQUISLICE_STORAGE_CONNECTION_STRING`, and
  `DEPLOYMENT_STORAGE_CONNECTION_STRING` come from that same account.
- Application Insights connection configuration comes from its Terraform resource.
- External secrets (Sentry DSNs and the Grafana OTLP auth header) live only in the
  `equislice-kv-renn244` Key Vault. App settings hold Key Vault references such as
  `@Microsoft.KeyVault(VaultName=equislice-kv-renn244;SecretName=grafana-otlp-headers)`,
  which the apps resolve at runtime through the shared `equislice-apps` managed
  identity. Terraform never reads or writes the secret values, so they are not in
  state and Terraform needs no variable files.
- The existing GitHub Azure login secrets stay as they are. No storage connection
  string needs to be copied into GitHub.

## Key Vault secrets

| Secret | Used by |
| --- | --- |
| `backend-sentry-dsn` | Backend `SENTRY_DSN` |
| `function-sentry-dsn` | Function `SENTRY_DSN` |
| `grafana-otlp-headers` | Both apps' `OTEL_EXPORTER_OTLP_HEADERS`, e.g. `Authorization=Basic <base64 of instanceId:token>` |

`var.secrets_admin_object_id` gets Key Vault Secrets Officer so that person can set
values. Set or rotate one in the portal, or without leaving it in shell history:

```powershell
$value = Read-Host -AsSecureString 'Secret value'
az keyvault secret set --vault-name equislice-kv-renn244 --name grafana-otlp-headers `
  --value ([System.Net.NetworkCredential]::new('', $value).Password) --output none
```

References are versionless, so apps pick up a new value on their next restart (or
within 24 hours). Restart an app to apply a rotation immediately.

### First rollout order

The Function panics on an invalid Sentry DSN, so secrets must exist before any app
setting points at the vault:

1. Create only the vault, identity, and access:
   `terraform apply '-target=azurerm_key_vault.equislice' '-target=azurerm_user_assigned_identity.apps' '-target=azurerm_role_assignment.apps_kv_secrets_user' '-target=azurerm_role_assignment.admin_kv_secrets_officer'`
2. Add the three secrets above. Role assignments can take a few minutes to apply.
3. Run a full `terraform plan` and `terraform apply` to switch the apps to references.
4. In each app's Environment variables blade, confirm the references show as resolved.

## State

State lives in HCP Terraform: organization `Renn224`, workspace
`equislice-production`. The workspace uses **local execution**, so `plan` and
`apply` run on your machine with your Azure CLI session; HCP Terraform stores and
locks the state and keeps its version history. Nothing is stored locally besides
the `.terraform` cache.

The state no longer holds Sentry or Grafana secrets, but it still contains the
storage account keys and connection strings plus the Application Insights
connection string, so keep workspace access restricted. Saved plan files contain
the same values and are gitignored. Do not start another independent state for
these same resources.

Remote execution (runs on HCP Terraform, or triggered from GitHub) would need
Azure credentials configured in the workspace, ideally HCP Terraform dynamic
provider credentials via OIDC, plus an identity allowed to manage these resources
and their role assignments. That is not set up.

## Existing GitHub workflows

The backend workflow already deploys to `equislice-api-renn244`; the Function workflow
uses the existing `AZURE_FUNCTIONAPP_NAME` repository variable. Those targets remain
unchanged. Outputs expose names and API URL for future workflow integration:

```powershell
terraform output -raw backend_app_name
terraform output -raw function_app_name
terraform output -raw backend_api_url
```

GitHub's current Website Contributor assignments allow code deployments to these
apps. They are insufficient to run this entire infrastructure configuration. Run
the initial adoption with your existing authorized Azure CLI session. A future CI
infrastructure workflow needs separately scoped provisioning and role-assignment
permissions plus remote state; this change does not grant those permissions.

## Preserved live settings

The F1 plan, disabled Always On, and the existing `WEBSITE_PORT=3000` spelling are
preserved for parity. The documented legacy App Service custom-container setting is
`WEBSITES_PORT`; changing it should be a separate reviewed runtime fix. Blob CORS
allows both localhost and the deployed Vercel frontend. Storage retains deleted
blobs and containers for seven days.

AzureRM 4.81.0 does not accept Go for its Flex Consumption resource, so the Function
uses AzAPI against the ARM API. Terraform provisions/configures the Function;
the existing GitHub workflow continues to publish its deployment package.
