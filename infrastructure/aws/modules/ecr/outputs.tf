output "repository_urls" {
  description = "Repository URLs keyed by repository name."
  value       = { for name, repo in aws_ecr_repository.this : name => repo.repository_url }
}

output "repository_arns" {
  description = "Repository ARNs keyed by repository name."
  value       = { for name, repo in aws_ecr_repository.this : name => repo.arn }
}

output "registry_id" {
  description = "Registry ID owning the repositories."
  value       = values(aws_ecr_repository.this)[0].registry_id
}
