variable "gcp_project_id" {
  description = "The GCP project ID where resources will be created."
  type        = string
  # No default, should be provided via env var or form
}

variable "gcp_region" {
  description = "The GCP region for the resources."
  type        = string
  default     = "us-central1"
}

variable "gcp_zone" {
  description = "The GCP zone for the VM instance."
  type        = string
  default     = "us-central1-a"
}

variable "instance_name_prefix" {
  description = "Prefix for the VM instance name. Set via LEMC form/env or defaults in entrypoint."
  type        = string
  default     = "lemc-tf"
}

variable "lemc_uuid" {
  description = "LEMC-provided UUID for uniqueness (passed as env var)."
  type        = string
  default     = "" # Will be overridden by env var
}

variable "lemc_username" {
  description = "LEMC-provided username (passed as env var)."
  type        = string
  default     = "" # Will be overridden by env var
}

variable "LEMC_SCOPE" {
  description = "Scope of the LEMC job (e.g., 'individual' or 'shared'). Determines resource linkage."
  type        = string
  default     = "unknown" # Default value if not provided
}

variable "domain_name" {
  description = "The domain name for the SSL certificate and DNS record (e.g., myapp.example.com)."
  type        = string
  # No default, must be provided via env var (TF_VAR_domain_name)
}

variable "dns_zone_name" {
  description = "The name of the managed DNS zone in GCP where the DNS record will be created."
  type        = string
  # No default, must be provided via env var (TF_VAR_dns_zone_name)
} 