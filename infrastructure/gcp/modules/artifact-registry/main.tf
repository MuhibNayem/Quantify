resource "google_artifact_registry_repository" "this" {
  for_each = toset(var.repositories)

  repository_id = "${var.name}-${each.value}"
  location      = var.region
  format        = "DOCKER"
  mode          = "STANDARD_REPOSITORY"

  cleanup_policies {
    id = "keep-last-images"

    most_recent_versions {
      keep_count = var.keep_count
    }
  }

  labels = var.labels
}
