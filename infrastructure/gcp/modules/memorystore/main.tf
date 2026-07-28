resource "google_redis_cluster" "this" {
  name           = var.name
  region         = var.region
  shard_count    = var.shard_count
  replica_count  = var.replica_count

  transit_encryption_mode = "TRANSIT_ENCRYPTION_MODE_SERVER_AUTHENTICATION"
  authorization_mode      = "AUTH_MODE_DISABLED"

  psc_configs {
    network = replace(var.network_id, "https://www.googleapis.com/compute/v1/", "")
  }
}
