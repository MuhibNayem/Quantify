output "cluster_name" {
  description = "Name of the Redis cluster."
  value       = google_redis_cluster.this.name
}

output "discovery_endpoint" {
  description = "Discovery endpoint address for cluster clients."
  value       = google_redis_cluster.this.discovery_endpoints[0].address
}

output "port" {
  description = "Discovery endpoint port."
  value       = google_redis_cluster.this.discovery_endpoints[0].port
}
