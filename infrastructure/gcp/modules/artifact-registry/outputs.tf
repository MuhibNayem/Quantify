output "repository_urls" {
  description = "Docker repository paths keyed by name."
  value = {
    for name, repo in google_artifact_registry_repository.this :
    name => "${repo.location}-docker.pkg.dev/${repo.project}/${repo.repository_id}"
  }
}
