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