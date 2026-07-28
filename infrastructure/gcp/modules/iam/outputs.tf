output "backend_email" {
  description = "Email of the backend service account."
  value       = google_service_account.backend.email
}

output "ai_email" {
  description = "Email of the AI service account."
  value       = google_service_account.ai.email
}
