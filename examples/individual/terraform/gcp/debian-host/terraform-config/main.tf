terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }
}

provider "google" {
  project = var.gcp_project_id
  region  = var.gcp_region
  # Credentials handled by GOOGLE_CREDENTIALS env var set in entrypoint.sh
}

# Define common labels for resources
locals {
  common_labels = {
    lemc_scope = var.LEMC_SCOPE
    lemc_user  = var.lemc_username
    lemc_uuid  = var.lemc_uuid
  }
}

resource "tls_private_key" "ssh" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "google_compute_network" "vpc" {
  name                    = "${var.instance_name_prefix}-${var.LEMC_SCOPE}-${var.lemc_username}-vpc-${substr(var.lemc_uuid, 0, 8)}"
  auto_create_subnetworks = false
  project                 = var.gcp_project_id
}

resource "google_compute_subnetwork" "subnet" {
  name          = "${var.instance_name_prefix}-${var.LEMC_SCOPE}-${var.lemc_username}-subnet-${substr(var.lemc_uuid, 0, 8)}"
  ip_cidr_range = "10.0.1.0/24"
  region        = var.gcp_region
  network       = google_compute_network.vpc.id
  project       = var.gcp_project_id
}

resource "google_compute_firewall" "allow_ssh" {
  name    = "${var.instance_name_prefix}-${var.LEMC_SCOPE}-${var.lemc_username}-allow-ssh-${substr(var.lemc_uuid, 0, 8)}"
  network = google_compute_network.vpc.id
  project = var.gcp_project_id

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["0.0.0.0/0"] # Allow SSH from anywhere (adjust if needed)
  target_tags   = ["${var.instance_name_prefix}-${var.lemc_username}-ssh"]
}

resource "google_compute_address" "public_ip" {
  name    = "${var.instance_name_prefix}-${var.LEMC_SCOPE}-${var.lemc_username}-public-ip-${substr(var.lemc_uuid, 0, 8)}"
  project = var.gcp_project_id
  region  = var.gcp_region
}

resource "google_compute_instance" "vm" {
  name         = "${var.instance_name_prefix}-${var.LEMC_SCOPE}-${var.lemc_username}-${substr(var.lemc_uuid, 0, 8)}"
  machine_type = "e2-small" # Or another desired machine type
  zone         = var.gcp_zone
  project      = var.gcp_project_id

  tags = ["${var.instance_name_prefix}-${var.lemc_username}-ssh", "lemc-managed"]

  labels = local.common_labels

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-11" # Debian 11 image
    }
  }

  network_interface {
    subnetwork = google_compute_subnetwork.subnet.id
    access_config {
      nat_ip = google_compute_address.public_ip.address
    }
  }

  metadata = {
    ssh-keys = "${var.lemc_username}:${tls_private_key.ssh.public_key_openssh}" # Format: username:key
  }

  service_account {
    # Uses default compute service account. Specify scopes if needed.
    scopes = ["cloud-platform"]
  }

  allow_stopping_for_update = true
} 