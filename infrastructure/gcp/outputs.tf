output "network_id" {
  description = "Self link of the VPC network."
  value       = module.network.network_id
}

output "gke_cluster_name" {
  description = "Name of the GKE cluster."
  value       = module.gke.cluster_name
}

output "gke_cluster_endpoint" {
  description = "Endpoint of the GKE cluster."
  value       = module.gke.cluster_endpoint
  sensitive   = true
}

output "gke_workload_identity_pool" {
  description = "Workload Identity pool for the cluster."
  value       = module.gke.workload_identity_pool
}

output "kubeconfig_command" {
  description = "Command to configure kubectl for the cluster."
  value       = "gcloud container clusters get-credentials ${module.gke.cluster_name} --region ${var.region} --project ${var.project_id}"
}

output "cloudsql_connection_name" {
  description = "Connection name of the Cloud SQL instance."
  value       = var.enable_cloudsql ? module.cloudsql[0].connection_name : null
}

output "cloudsql_private_ip" {
  description = "Private IP of the Cloud SQL instance."
  value       = var.enable_cloudsql ? module.cloudsql[0].private_ip_address : null
}

output "redis_discovery_endpoint" {
  description = "Discovery endpoint of the Redis cluster."
  value       = var.enable_redis ? module.memorystore[0].discovery_endpoint : null
}

output "pubsub_topic" {
  description = "Pub/Sub topic used for async messaging."
  value       = var.enable_pubsub ? module.pubsub[0].topic_id : null
}

output "storage_bucket_name" {
  description = "Name of the object storage bucket."
  value       = module.storage.bucket_name
}

output "artifact_registry_urls" {
  description = "Docker repository paths keyed by name."
  value       = module.artifact_registry.repository_urls
}

output "backend_service_account_email" {
  description = "Workload Identity service account for the backend."
  value       = module.iam.backend_email
}

output "ai_service_account_email" {
  description = "Workload Identity service account for the AI service."
  value       = module.iam.ai_email
}
