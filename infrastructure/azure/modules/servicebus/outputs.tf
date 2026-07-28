output "namespace_name" {
  description = "Name of the Service Bus namespace."
  value       = azurerm_servicebus_namespace.this.name
}

output "namespace_endpoint" {
  description = "Endpoint of the Service Bus namespace."
  value       = azurerm_servicebus_namespace.this.endpoint
}

output "queue_name" {
  description = "Name of the tasks queue."
  value       = azurerm_servicebus_queue.tasks.name
}
