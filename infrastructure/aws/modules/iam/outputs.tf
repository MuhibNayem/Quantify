output "backend_role_arn" {
  description = "IRSA role ARN for the backend service account."
  value       = aws_iam_role.backend.arn
}

output "ai_service_role_arn" {
  description = "IRSA role ARN for the AI service account."
  value       = aws_iam_role.ai_service.arn
}

output "s3_access_policy_arn" {
  description = "ARN of the shared S3 access policy."
  value       = aws_iam_policy.s3_access.arn
}

output "cloudwatch_logs_policy_arn" {
  description = "ARN of the shared CloudWatch Logs policy."
  value       = aws_iam_policy.cloudwatch_logs.arn
}
