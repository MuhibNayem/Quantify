output "cluster_name" {
  description = "Name of the GKE cluster."
  value       = google_container_cluster.this.name
}

output "cluster_endpoint" {
  description = "Endpoint of the GKE cluster."
  value       = google_container_cluster.this.endpoint
  sensitive   = true
}

output "cluster_ca_certificate" {
  description = "Base64 encoded CA certificate of the cluster."
  value       = google_container_cluster.this.master_auth[0].cluster_ca_certificate
  sensitive   = true
}

output "workload_identity_pool" {
  description = "Workload Identity pool for the cluster."
  value       = "${var.project_id}.svc.id.goog"
}

output "node_service_account_email" {
  description = "Email of the node service account."
  value       = google_service_account.node.email
}
