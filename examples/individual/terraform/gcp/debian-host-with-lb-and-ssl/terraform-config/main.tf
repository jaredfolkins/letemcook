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
  # Define a tag for backend instances
  lb_backend_tag = "${var.instance_name_prefix}-${var.lemc_username}-backend"
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

  # Add the lb_backend_tag to the existing tags
  tags = ["${var.instance_name_prefix}-${var.lemc_username}-ssh", local.lb_backend_tag, "lemc-managed"]

  labels = local.common_labels

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-11" # Debian 11 image
    }
  }

  network_interface {
    subnetwork = google_compute_subnetwork.subnet.id
    # Remove direct public IP assignment, LB will handle external access
    # Add an empty access_config block to ensure no ephemeral public IP is assigned
    access_config {}
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

# Firewall rule to allow traffic from Health Checker and LB to backend port (8080)
resource "google_compute_firewall" "allow_lb_healthcheck" {
  name    = "${var.instance_name_prefix}-${var.LEMC_SCOPE}-${var.lemc_username}-allow-lb-hc-${substr(var.lemc_uuid, 0, 8)}"
  network = google_compute_network.vpc.id
  project = var.gcp_project_id

  allow {
    protocol = "tcp"
    ports    = ["8080"] # Backend port
  }

  # GCP Health Checker IP ranges
  source_ranges = ["130.211.0.0/22", "35.191.0.0/16"]
  target_tags   = [local.lb_backend_tag]
}

# Reserve a global static IP for the Load Balancer
resource "google_compute_global_address" "lb_ip" {
  name    = "${var.instance_name_prefix}-${var.LEMC_SCOPE}-${var.lemc_username}-lb-ip-${substr(var.lemc_uuid, 0, 8)}"
  project = var.gcp_project_id
}

# Unmanaged Instance Group for the VM
# Note: For simplicity, using unmanaged. For scaling, consider managed instance groups.
resource "google_compute_instance_group" "instance_group" {
  name        = "${var.instance_name_prefix}-${var.LEMC_SCOPE}-${var.lemc_username}-ig-${substr(var.lemc_uuid, 0, 8)}"
  description = "Instance group for LEMC demo VM"
  zone        = var.gcp_zone
  project     = var.gcp_project_id

  instances = [
    google_compute_instance.vm.id,
  ]

  named_port {
    name = "http"
    port = 8080 # Backend port
  }
}

# Health Check for the Load Balancer Backend
resource "google_compute_health_check" "lb_health_check" {
  name    = "${var.instance_name_prefix}-${var.LEMC_SCOPE}-${var.lemc_username}-hc-${substr(var.lemc_uuid, 0, 8)}"
  project = var.gcp_project_id

  timeout_sec        = 5
  check_interval_sec = 10

  tcp_health_check {
    port = 8080 # Backend port
  }
}

# Backend Service
resource "google_compute_backend_service" "backend_service" {
  name      = "${var.instance_name_prefix}-${var.LEMC_SCOPE}-${var.lemc_username}-bes-${substr(var.lemc_uuid, 0, 8)}"
  port_name = "http" # Matches named_port in instance_group
  protocol  = "HTTP" # Protocol between LB and backend
  project   = var.gcp_project_id

  load_balancing_scheme = "EXTERNAL"
  timeout_sec           = 30

  backend {
    group = google_compute_instance_group.instance_group.id
  }

  health_checks = [
    google_compute_health_check.lb_health_check.id,
  ]
}

# URL Map (Basic: send all traffic to the backend service)
resource "google_compute_url_map" "url_map" {
  name            = "${var.instance_name_prefix}-${var.LEMC_SCOPE}-${var.lemc_username}-urlmap-${substr(var.lemc_uuid, 0, 8)}"
  default_service = google_compute_backend_service.backend_service.id
  project         = var.gcp_project_id
}

# Managed SSL Certificate
resource "google_compute_managed_ssl_certificate" "ssl_certificate" {
  name    = "${var.instance_name_prefix}-${var.LEMC_SCOPE}-${var.lemc_username}-ssl-cert-${substr(var.lemc_uuid, 0, 8)}"
  project = var.gcp_project_id
  managed {
    domains = [var.domain_name]
  }
}

# Target HTTPS Proxy
resource "google_compute_target_https_proxy" "https_proxy" {
  name             = "${var.instance_name_prefix}-${var.LEMC_SCOPE}-${var.lemc_username}-https-proxy-${substr(var.lemc_uuid, 0, 8)}"
  url_map          = google_compute_url_map.url_map.id
  ssl_certificates = [google_compute_managed_ssl_certificate.ssl_certificate.id]
  project          = var.gcp_project_id
}

# Global Forwarding Rule for HTTPS (Port 443)
resource "google_compute_global_forwarding_rule" "https_forwarding_rule" {
  name                  = "${var.instance_name_prefix}-${var.LEMC_SCOPE}-${var.lemc_username}-https-fwd-rule-${substr(var.lemc_uuid, 0, 8)}"
  target                = google_compute_target_https_proxy.https_proxy.id
  ip_address            = google_compute_global_address.lb_ip.address
  port_range            = "443"
  load_balancing_scheme = "EXTERNAL"
  project               = var.gcp_project_id
}

# Optional: Redirect HTTP to HTTPS
# Target HTTP Proxy
resource "google_compute_target_http_proxy" "http_proxy" {
  name    = "${var.instance_name_prefix}-${var.LEMC_SCOPE}-${var.lemc_username}-http-proxy-${substr(var.lemc_uuid, 0, 8)}"
  url_map = google_compute_url_map.url_map.id # Can reuse the same URL map
  project = var.gcp_project_id
}

# Global Forwarding Rule for HTTP (Port 80)
resource "google_compute_global_forwarding_rule" "http_forwarding_rule" {
  name                  = "${var.instance_name_prefix}-${var.LEMC_SCOPE}-${var.lemc_username}-http-fwd-rule-${substr(var.lemc_uuid, 0, 8)}"
  target                = google_compute_target_http_proxy.http_proxy.id
  ip_address            = google_compute_global_address.lb_ip.address
  port_range            = "80"
  load_balancing_scheme = "EXTERNAL"
  project               = var.gcp_project_id
}


# --- DNS Record ---

# Get details of the managed DNS zone
data "google_dns_managed_zone" "zone" {
  name    = var.dns_zone_name
  project = var.gcp_project_id # Assuming DNS zone is in the same project
}

# Create DNS A record for the domain pointing to the LB IP
resource "google_dns_record_set" "dns_record" {
  name    = "${var.domain_name}." # Ensure trailing dot
  type    = "A"
  ttl     = 300
  project = data.google_dns_managed_zone.zone.project

  managed_zone = data.google_dns_managed_zone.zone.name

  rrdatas = [google_compute_global_address.lb_ip.address]
}