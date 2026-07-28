output "cluster_name" {
  description = "Name of the EKS cluster."
  value       = aws_eks_cluster.this.name
}

output "cluster_endpoint" {
  description = "Endpoint of the EKS API server."
  value       = aws_eks_cluster.this.endpoint
}

output "cluster_certificate_authority_data" {
  description = "Base64 encoded certificate data for the cluster."
  value       = aws_eks_cluster.this.certificate_authority[0].data
}

output "cluster_security_group_id" {
  description = "Security group ID attached to the cluster control plane."
  value       = aws_security_group.cluster.id
}

output "node_security_group_id" {
  description = "Security group ID attached to the node groups."
  value       = aws_security_group.node.id
}

output "cluster_iam_role_arn" {
  description = "IAM role ARN of the cluster control plane."
  value       = aws_iam_role.cluster.arn
}

output "node_iam_role_arn" {
  description = "IAM role ARN of the node groups."
  value       = aws_iam_role.node.arn
}

output "oidc_provider_arn" {
  description = "ARN of the IAM OIDC provider for IRSA."
  value       = aws_iam_openid_connect_provider.this.arn
}

output "oidc_provider_url" {
  description = "Issuer URL of the OIDC provider (without scheme)."
  value       = local.oidc_issuer
}

output "cloudwatch_log_group_name" {
  description = "Name of the control plane CloudWatch log group."
  value       = aws_cloudwatch_log_group.control_plane.name
}
