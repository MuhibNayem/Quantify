resource "random_password" "this" {
  length           = 32
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

resource "google_compute_global_address" "private_ip" {
  name          = "${var.name}-sql-private-ip"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = var.network_id
}

resource "google_service_networking_connection" "private_vpc" {
  network                 = var.network_id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip.name]
}

resource "google_sql_database_instance" "primary" {
  name                = "${var.name}-postgres"
  database_version    = "POSTGRES_15"
  region              = var.region
  deletion_protection = var.deletion_protection

  settings {
    tier              = var.tier
    availability_type = var.high_availability ? "REGIONAL" : "ZONAL"
    disk_size         = var.disk_size_gb
    disk_autoresize   = true
    disk_type         = "PD_SSD"

    backup_configuration {
      enabled                        = true
      point_in_time_recovery_enabled = true
      start_time                     = "03:00"
      transaction_log_retention_days = 7

      backup_retention_settings {
        retention_unit   = "COUNT"
        retained_backups = 7
      }
    }

    insights_config {
      query_insights_enabled  = true
      query_string_length     = 1024
      record_application_tags = true
    }

    database_flags {
      name  = "max_connections"
      value = var.max_connections
    }

    database_flags {
      name  = "log_min_duration_statement"
      value = "1000"
    }

    ip_configuration {
      ipv4_enabled    = false
      private_network = var.network_id
      ssl_mode        = "ENCRYPTED_ONLY"
    }

    maintenance_window {
      day  = 7
      hour = 4
    }

    user_labels = var.labels
  }

  depends_on = [google_service_networking_connection.private_vpc]
}

resource "google_sql_database" "this" {
  name     = var.db_name
  instance = google_sql_database_instance.primary.name
}

resource "google_sql_user" "this" {
  name     = var.username
  instance = google_sql_database_instance.primary.name
  password = random_password.this.result
}

resource "google_sql_database_instance" "replica" {
  count = var.create_read_replica ? 1 : 0

  name                 = "${var.name}-postgres-replica"
  master_instance_name = google_sql_database_instance.primary.name
  database_version     = "POSTGRES_15"
  region               = var.region
  deletion_protection  = var.deletion_protection

  replica_configuration {
    failover_target = false
  }

  settings {
    tier            = var.tier
    disk_size       = var.disk_size_gb
    disk_autoresize = true
    disk_type       = "PD_SSD"

    insights_config {
      query_insights_enabled = true
    }

    ip_configuration {
      ipv4_enabled    = false
      private_network = var.network_id
      ssl_mode        = "ENCRYPTED_ONLY"
    }

    user_labels = var.labels
  }
}
