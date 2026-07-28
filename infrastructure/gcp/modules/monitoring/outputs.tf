output "notification_channel_ids" {
  description = "IDs of the email notification channels."
  value       = [for ch in google_monitoring_notification_channel.email : ch.id]
}
