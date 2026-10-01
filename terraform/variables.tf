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

variable "backend_image" {
  type        = string
  description = "Bootstrap image. After creation GitHub Actions owns image releases."
  default     = "renn244/equislice-backend:661783148abcf5a1b5727117b498e1bff663db28"
}

variable "frontend_url" {
  type    = string
  default = "https://equislice.vercel.app"
}

variable "backend_sentry_dsn" {
  type        = string
  sensitive   = true
  description = "Existing backend Sentry DSN, loaded by Export-LiveSettings.ps1."
}

variable "backend_sentry_environment" {
  type        = string
  description = "Existing backend Sentry environment."
}

variable "function_sentry_dsn" {
  type        = string
  sensitive   = true
  description = "Existing Function Sentry DSN, loaded by Export-LiveSettings.ps1."
}

variable "function_sentry_environment" {
  type        = string
  description = "Existing Function Sentry environment."
}

variable "grafana_otlp_endpoint" {
  type        = string
  description = "Grafana Cloud OTLP HTTP endpoint."
}

variable "grafana_otlp_headers" {
  type        = string
  sensitive   = true
  description = "Grafana Cloud OTLP authentication header."
}
