output "cluster_name" {
  description = "Name of the AKS cluster."
  value       = azurerm_kubernetes_cluster.this.name
}

output "cluster_id" {
  description = "ID of the AKS cluster."
  value       = azurerm_kubernetes_cluster.this.id
}

output "oidc_issuer_url" {
  description = "OIDC issuer URL for workload identity."
  value       = azurerm_kubernetes_cluster.this.oidc_issuer_url
}

output "kubelet_identity_object_id" {
  description = "Object ID of the kubelet identity."
  value       = azurerm_kubernetes_cluster.this.kubelet_identity[0].object_id
}

output "node_resource_group" {
  description = "Name of the node resource group."
  value       = azurerm_kubernetes_cluster.this.node_resource_group
}

output "kube_config_raw" {
  description = "Raw kubeconfig for the cluster."
  value       = azurerm_kubernetes_cluster.this.kube_config_raw
  sensitive   = true
}
