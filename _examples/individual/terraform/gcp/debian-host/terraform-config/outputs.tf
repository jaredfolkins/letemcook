output "private_ssh_key" {
  description = "The generated private SSH key (PEM format). Save this securely!"
  value       = tls_private_key.ssh.private_key_pem
  sensitive   = true
}

output "public_ssh_key" {
  description = "The generated public SSH key (OpenSSH format)."
  value       = tls_private_key.ssh.public_key_openssh
}

output "public_ip" {
  description = "The public IP address of the created VM instance."
  value       = google_compute_address.public_ip.address
}

output "instance_name" {
  description = "The name of the created VM instance."
  value       = google_compute_instance.vm.name
}

output "vpc_name" {
  description = "The name of the created VPC network."
  value       = google_compute_network.vpc.name
}

output "subnet_name" {
  description = "The name of the created subnet."
  value       = google_compute_subnetwork.subnet.name
} 