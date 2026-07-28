resource "google_pubsub_topic" "this" {
  name                       = "${var.name}-tasks"
  labels                     = var.labels
  message_retention_duration = "604800s"
}

resource "google_pubsub_subscription" "this" {
  name  = "${var.name}-tasks-sub"
  topic = google_pubsub_topic.this.id

  ack_deadline_seconds = 60

  expiration_policy {
    ttl = ""
  }

  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }

  labels = var.labels
}
