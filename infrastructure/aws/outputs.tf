output "vpc_id" {
  description = "ID of the VPC."
  value       = module.vpc.vpc_id
}

output "eks_cluster_name" {
  description = "Name of the EKS cluster."
  value       = module.eks.cluster_name
}

output "eks_cluster_endpoint" {
  description = "Endpoint of the EKS API server."
  value       = module.eks.cluster_endpoint
}

output "eks_cluster_certificate_authority" {
  description = "Base64 encoded certificate data for the cluster."
  value       = module.eks.cluster_certificate_authority_data
  sensitive   = true
}

output "eks_oidc_provider_arn" {
  description = "ARN of the IAM OIDC provider for IRSA."
  value       = module.eks.oidc_provider_arn
}

output "kubeconfig_command" {
  description = "Command to configure kubectl for the cluster."
  value       = "aws eks update-kubeconfig --name ${module.eks.cluster_name} --region ${var.region}"
}

output "rds_endpoint" {
  description = "Writer endpoint of the PostgreSQL database."
  value       = var.enable_rds ? module.rds[0].endpoint : null
}

output "rds_reader_endpoint" {
  description = "Reader endpoint of the PostgreSQL database."
  value       = var.enable_rds ? module.rds[0].reader_endpoint : null
}

output "rds_secret_arn" {
  description = "Secrets Manager ARN holding database credentials."
  value       = var.enable_rds ? module.rds[0].secret_arn : null
}

output "redis_configuration_endpoint" {
  description = "Configuration endpoint of the Redis cluster."
  value       = var.enable_elasticache ? module.elasticache[0].configuration_endpoint : null
}

output "rabbitmq_endpoint" {
  description = "AMQP endpoint of the RabbitMQ broker."
  value       = var.enable_mq ? module.mq[0].endpoint : null
}

output "rabbitmq_management_endpoint" {
  description = "Management console URL of the RabbitMQ broker."
  value       = var.enable_mq ? module.mq[0].management_endpoint : null
}

output "s3_bucket_name" {
  description = "Name of the object storage bucket."
  value       = module.s3.bucket_id
}

output "s3_bucket_arn" {
  description = "ARN of the object storage bucket."
  value       = module.s3.bucket_arn
}

output "ecr_repository_urls" {
  description = "Repository URLs keyed by repository name."
  value       = module.ecr.repository_urls
}

output "backend_service_account_role_arn" {
  description = "IRSA role ARN for the backend service account."
  value       = module.iam.backend_role_arn
}

output "ai_service_account_role_arn" {
  description = "IRSA role ARN for the AI service account."
  value       = module.iam.ai_service_role_arn
}

output "sns_alert_topic_arn" {
  description = "ARN of the SNS alert topic."
  value       = var.enable_monitoring ? module.monitoring[0].sns_topic_arn : null
}
