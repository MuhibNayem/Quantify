output "sns_topic_arn" {
  description = "ARN of the SNS alert topic."
  value       = aws_sns_topic.alerts.arn
}

output "application_log_group_name" {
  description = "Name of the application CloudWatch log group."
  value       = aws_cloudwatch_log_group.application.name
}

output "prometheus_workspace_endpoint" {
  description = "Endpoint of the Managed Prometheus workspace."
  value       = var.enable_managed_prometheus ? aws_prometheus_workspace.this[0].prometheus_endpoint : null
}

output "grafana_endpoint" {
  description = "Endpoint of the Managed Grafana workspace."
  value       = var.enable_managed_prometheus ? aws_grafana_workspace.this[0].endpoint : null
}
