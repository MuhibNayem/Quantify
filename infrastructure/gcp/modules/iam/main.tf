resource "google_service_account" "backend" {
  account_id   = "${var.cluster_name}-backend"
  display_name = "Quantify backend workload identity"
}

resource "google_service_account" "ai" {
  account_id   = "${var.cluster_name}-ai"
  display_name = "Quantify AI service workload identity"
}

resource "google_service_account_iam_member" "backend_workload_identity" {
  service_account_id = google_service_account.backend.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${var.workload_identity_pool}[${var.namespace}/${var.backend_service_account}]"
}

resource "google_service_account_iam_member" "ai_workload_identity" {
  service_account_id = google_service_account.ai.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${var.workload_identity_pool}[${var.namespace}/${var.ai_service_account}]"
}

resource "google_project_iam_member" "backend_storage" {
  project = var.project_id
  role    = "roles/storage.objectUser"
  member  = "serviceAccount:${google_service_account.backend.email}"
}

resource "google_project_iam_member" "ai_storage" {
  project = var.project_id
  role    = "roles/storage.objectUser"
  member  = "serviceAccount:${google_service_account.ai.email}"
}

resource "google_project_iam_member" "backend_cloudsql_client" {
  count = var.cloudsql_instance != null ? 1 : 0

  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.backend.email}"
}

resource "google_project_iam_member" "backend_secret_accessor" {
  count = var.cloudsql_instance != null ? 1 : 0

  project = var.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${google_service_account.backend.email}"
}

resource "google_project_iam_member" "backend_metric_writer" {
  project = var.project_id
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${google_service_account.backend.email}"
}

resource "google_project_iam_member" "ai_metric_writer" {
  project = var.project_id
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${google_service_account.ai.email}"
}
