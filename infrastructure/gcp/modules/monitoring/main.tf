resource "google_monitoring_notification_channel" "email" {
  for_each = toset(var.alert_email_endpoints)

  display_name = "${var.name}-email"
  type         = "email"

  labels = {
    email_address = each.value
  }
}

locals {
  notification_channels = [for ch in google_monitoring_notification_channel.email : ch.id]
}

resource "google_monitoring_alert_policy" "cloudsql_cpu" {
  count = var.cloudsql_instance != null ? 1 : 0

  display_name = "${var.name}-cloudsql-cpu"
  combiner     = "OR"

  conditions {
    display_name = "Cloud SQL CPU utilization"

    condition_threshold {
      filter          = "resource.type = \"cloudsql_database\" AND resource.labels.database_id = \"${var.project_id}:${var.cloudsql_instance}\" AND metric.type = \"cloudsql.googleapis.com/database/cpu/utilization\""
      duration        = "300s"
      comparison      = "COMPARISON_GT"
      threshold_value = 0.8

      aggregations {
        alignment_period     = "300s"
        per_series_aligner   = "ALIGN_MEAN"
        cross_series_reducer = "REDUCE_NONE"
      }
    }
  }

  notification_channels = local.notification_channels
}

resource "google_monitoring_alert_policy" "cloudsql_connections" {
  count = var.cloudsql_instance != null ? 1 : 0

  display_name = "${var.name}-cloudsql-connections"
  combiner     = "OR"

  conditions {
    display_name = "Cloud SQL connection count"

    condition_threshold {
      filter          = "resource.type = \"cloudsql_database\" AND resource.labels.database_id = \"${var.project_id}:${var.cloudsql_instance}\" AND metric.type = \"cloudsql.googleapis.com/database/postgresql/num_backends\""
      duration        = "300s"
      comparison      = "COMPARISON_GT"
      threshold_value = 250

      aggregations {
        alignment_period     = "300s"
        per_series_aligner   = "ALIGN_MEAN"
        cross_series_reducer = "REDUCE_NONE"
      }
    }
  }

  notification_channels = local.notification_channels
}

resource "google_monitoring_alert_policy" "redis_memory" {
  count = var.redis_cluster != null ? 1 : 0

  display_name = "${var.name}-redis-memory"
  combiner     = "OR"

  conditions {
    display_name = "Redis memory usage"

    condition_threshold {
      filter          = "resource.type = \"redis_instance\" AND metric.type = \"redis.googleapis.com/stats/memory/usage_ratio\""
      duration        = "300s"
      comparison      = "COMPARISON_GT"
      threshold_value = 0.8

      aggregations {
        alignment_period     = "300s"
        per_series_aligner   = "ALIGN_MEAN"
        cross_series_reducer = "REDUCE_NONE"
      }
    }
  }

  notification_channels = local.notification_channels
}

resource "google_monitoring_alert_policy" "gke_node_cpu" {
  display_name = "${var.name}-gke-node-cpu"
  combiner     = "OR"

  conditions {
    display_name = "GKE node CPU utilization"

    condition_threshold {
      filter          = "resource.type = \"k8s_node\" AND resource.labels.cluster_name = \"${var.gke_cluster_name}\" AND metric.type = \"kubernetes.io/node/cpu/allocatable_utilization\""
      duration        = "600s"
      comparison      = "COMPARISON_GT"
      threshold_value = 0.9

      aggregations {
        alignment_period     = "300s"
        per_series_aligner   = "ALIGN_MEAN"
        cross_series_reducer = "REDUCE_NONE"
      }
    }
  }

  notification_channels = local.notification_channels
}
