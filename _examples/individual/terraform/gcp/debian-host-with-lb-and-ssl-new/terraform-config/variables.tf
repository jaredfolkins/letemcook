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

variable "resource_prefix" {
  description = "Dynamically generated prefix for resources (e.g., uuid-user-userid-scope)."
  type        = string
  # No default, generated by setup_tf.go
}

variable "lemc_uuid" {
  description = "LEMC-provided UUID for uniqueness (passed as env var)."
  type        = string
  default     = "" # Will be overridden by env var
}

variable "lemc_user_id" {
  description = "LEMC-provided user ID (passed as env var)."
  type        = string
  default     = "" # Will be overridden by env var
}

variable "lemc_username" {
  description = "LEMC-provided username (passed as env var)."
  type        = string
  default     = "" # Will be overridden by env var
}

variable "lemc_scope" {
  description = "Scope of the LEMC job (e.g., 'individual' or 'shared'). Determines resource linkage."
  type        = string
  default     = "" # Default value if not provided
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

variable "machine_type" {
  description = "The machine type for the VM instance (e.g., e2-small, n1-standard-1). Set via FORM_MACHINE_TYPE in .env."
  type        = string
  # Default can be set here if FORM_MACHINE_TYPE is not always provided
  # default = "e2-small"
}

variable "image" {
  description = "The OS image for the VM instance (e.g., debian-cloud/debian-11). Set via FORM_IMAGE in .env."
  type        = string
  # default = "debian-cloud/debian-11" # Default can be set here
} 