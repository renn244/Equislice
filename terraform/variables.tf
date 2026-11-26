variable "subscription_id" {
  type        = string
  description = "Azure subscription containing the existing EquiSlice deployment."
  default     = "733d81d4-7950-404d-a6de-66c1537e7560"
}

variable "resource_group_name" {
  type        = string
  description = "Existing shared resource group; Terraform does not own this group."
  default     = "rg-renatodsantosjr9-1317"
}

variable "github_principal_id" {
  type        = string
  description = "Object ID (not client ID) of the existing equislice-github-actions service principal."
  default     = "31930435-3956-4545-ba00-a147db4ffe4b"
}

variable "secrets_admin_object_id" {
  type        = string
  description = "Entra object ID of the user who manages Key Vault secret values."
  default     = "988c6f0f-7a7e-44a1-a16f-81040d367bae"
}

variable "backend_image" {
  type        = string
  description = "Bootstrap image. After creation GitHub Actions owns image releases."
  default     = "renn244/equislice-backend:661783148abcf5a1b5727117b498e1bff663db28"
}

variable "frontend_url" {
  type    = string
  default = "https://equislice.vercel.app"
}

variable "sentry_environment" {
  type        = string
  description = "Sentry environment for both apps."
  default     = "production"
}

variable "grafana_otlp_endpoint" {
  type        = string
  description = "Grafana Cloud OTLP HTTP endpoint (not secret; the auth header is in Key Vault)."
  default     = "https://otlp-gateway-prod-ap-southeast-1.grafana.net/otlp"
}
