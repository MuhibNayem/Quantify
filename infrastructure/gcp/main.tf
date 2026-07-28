locals {
  name = "${var.project_id}-${var.environment}"

  common_labels = merge(
    {
      project     = "quantify"
      environment = var.environment
      managed_by  = "terraform"
    },
    var.labels
  )
}

module "network" {
  source = "./modules/network"

  name   = "quantify-${var.environment}"
  region = var.region
  cidr   = var.network_cidr

  pods_secondary_range_name     = var.pods_secondary_range_name
  services_secondary_range_name = var.services_secondary_range_name

  labels = local.common_labels
}

module "gke" {
  source = "./modules/gke"

  name       = "quantify-${var.environment}"
  project_id = var.project_id
  region     = var.region

  network_id     = module.network.network_id
  subnet_name    = module.network.subnet_name
  pods_range     = module.network.pods_secondary_range_name
  services_range = module.network.services_secondary_range_name

  release_channel = var.gke_release_channel
  node_pools      = var.gke_node_pools

  labels = local.common_labels
}

module "cloudsql" {
  source = "./modules/cloudsql"
  count  = var.enable_cloudsql ? 1 : 0

  name   = "quantify-${var.environment}"
  region = var.region

  tier                = var.db_tier
  high_availability   = var.db_ha
  disk_size_gb        = var.db_disk_size_gb
  create_read_replica = var.db_create_read_replica

  db_name  = "inventory_db"
  username = "quantify"

  network_id = module.network.network_id

  labels = local.common_labels
}

module "memorystore" {
  source = "./modules/memorystore"
  count  = var.enable_redis ? 1 : 0

  name   = "quantify-${var.environment}"
  region = var.region

  shard_count   = var.redis_shard_count
  replica_count = var.redis_replica_count
  network_id    = module.network.network_id
}

module "pubsub" {
  source = "./modules/pubsub"
  count  = var.enable_pubsub ? 1 : 0

  name = "quantify-${var.environment}"

  labels = local.common_labels
}

module "storage" {
  source = "./modules/storage"

  name     = "quantify-${var.environment}-storage"
  location = var.region

  labels = local.common_labels
}

module "artifact_registry" {
  source = "./modules/artifact-registry"

  name         = "quantify"
  region       = var.region
  repositories = ["backend", "ai-service"]

  labels = local.common_labels
}

module "iam" {
  source = "./modules/iam"

  project_id              = var.project_id
  cluster_name            = module.gke.cluster_name
  workload_identity_pool  = module.gke.workload_identity_pool
  namespace               = "quantify"
  backend_service_account = "quantify-backend"
  ai_service_account      = "quantify-ai-service"

  storage_bucket_name = module.storage.bucket_name
  cloudsql_instance   = var.enable_cloudsql ? module.cloudsql[0].instance_name : null

  labels = local.common_labels
}

module "monitoring" {
  source = "./modules/monitoring"
  count  = var.enable_monitoring ? 1 : 0

  name       = "quantify-${var.environment}"
  project_id = var.project_id

  cloudsql_instance = var.enable_cloudsql ? module.cloudsql[0].instance_name : null
  redis_cluster     = var.enable_redis ? module.memorystore[0].cluster_name : null
  gke_cluster_name  = module.gke.cluster_name

  alert_email_endpoints = var.alert_email_endpoints

  labels = local.common_labels
}
