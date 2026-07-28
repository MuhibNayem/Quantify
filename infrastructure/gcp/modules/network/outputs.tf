output "network_id" {
  description = "Self link of the VPC network."
  value       = google_compute_network.this.self_link
}

output "network_name" {
  description = "Name of the VPC network."
  value       = google_compute_network.this.name
}

output "subnet_name" {
  description = "Name of the subnet."
  value       = google_compute_subnetwork.this.name
}

output "subnet_id" {
  description = "Self link of the subnet."
  value       = google_compute_subnetwork.this.self_link
}

output "pods_secondary_range_name" {
  description = "Name of the pods secondary range."
  value       = var.pods_secondary_range_name
}

output "services_secondary_range_name" {
  description = "Name of the services secondary range."
  value       = var.services_secondary_range_name
}
