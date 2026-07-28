resource "aws_cloudwatch_log_group" "application" {
  name              = "/${var.name}/eks/${var.eks_cluster_name}/application"
  retention_in_days = var.log_retention_days

  tags = var.tags
}

# ---------------------------------------------------------------------------
# SNS alerting
# ---------------------------------------------------------------------------
resource "aws_sns_topic" "alerts" {
  name              = "${var.name}-alerts"
  kms_master_key_id = "alias/aws/sns"

  tags = var.tags
}

resource "aws_sns_topic_subscription" "email" {
  for_each = toset(var.alert_email_endpoints)

  topic_arn = aws_sns_topic.alerts.arn
  protocol  = "email"
  endpoint  = each.value
}

# ---------------------------------------------------------------------------
# RDS alarms
# ---------------------------------------------------------------------------
resource "aws_cloudwatch_metric_alarm" "rds_cpu" {
  count = var.rds_identifier != null ? 1 : 0

  alarm_name          = "${var.rds_identifier}-high-cpu"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "CPUUtilization"
  namespace           = "AWS/RDS"
  period              = 300
  statistic           = "Average"
  threshold           = 80
  alarm_actions       = [aws_sns_topic.alerts.arn]
  ok_actions          = [aws_sns_topic.alerts.arn]

  dimensions = {
    DBInstanceIdentifier = var.rds_identifier
  }

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "rds_free_storage" {
  count = var.rds_identifier != null ? 1 : 0

  alarm_name          = "${var.rds_identifier}-low-storage"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = 1
  metric_name         = "FreeStorageSpace"
  namespace           = "AWS/RDS"
  period              = 300
  statistic           = "Average"
  threshold           = 10 * 1024 * 1024 * 1024
  alarm_actions       = [aws_sns_topic.alerts.arn]
  ok_actions          = [aws_sns_topic.alerts.arn]

  dimensions = {
    DBInstanceIdentifier = var.rds_identifier
  }

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "rds_connections" {
  count = var.rds_identifier != null ? 1 : 0

  alarm_name          = "${var.rds_identifier}-high-connections"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "DatabaseConnections"
  namespace           = "AWS/RDS"
  period              = 300
  statistic           = "Average"
  threshold           = 250
  alarm_actions       = [aws_sns_topic.alerts.arn]
  ok_actions          = [aws_sns_topic.alerts.arn]

  dimensions = {
    DBInstanceIdentifier = var.rds_identifier
  }

  tags = var.tags
}

# ---------------------------------------------------------------------------
# ElastiCache alarms
# ---------------------------------------------------------------------------
resource "aws_cloudwatch_metric_alarm" "redis_cpu" {
  count = var.elasticache_replication_group_id != null ? 1 : 0

  alarm_name          = "${var.elasticache_replication_group_id}-high-cpu"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "CPUUtilization"
  namespace           = "AWS/ElastiCache"
  period              = 300
  statistic           = "Average"
  threshold           = 75
  alarm_actions       = [aws_sns_topic.alerts.arn]
  ok_actions          = [aws_sns_topic.alerts.arn]

  dimensions = {
    ReplicationGroupId = var.elasticache_replication_group_id
  }

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "redis_memory" {
  count = var.elasticache_replication_group_id != null ? 1 : 0

  alarm_name          = "${var.elasticache_replication_group_id}-high-memory"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "DatabaseMemoryUsagePercentage"
  namespace           = "AWS/ElastiCache"
  period              = 300
  statistic           = "Average"
  threshold           = 80
  alarm_actions       = [aws_sns_topic.alerts.arn]
  ok_actions          = [aws_sns_topic.alerts.arn]

  dimensions = {
    ReplicationGroupId = var.elasticache_replication_group_id
  }

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "redis_evictions" {
  count = var.elasticache_replication_group_id != null ? 1 : 0

  alarm_name          = "${var.elasticache_replication_group_id}-evictions"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "Evictions"
  namespace           = "AWS/ElastiCache"
  period              = 300
  statistic           = "Sum"
  threshold           = 100
  alarm_actions       = [aws_sns_topic.alerts.arn]
  ok_actions          = [aws_sns_topic.alerts.arn]

  dimensions = {
    ReplicationGroupId = var.elasticache_replication_group_id
  }

  tags = var.tags
}

# ---------------------------------------------------------------------------
# Amazon MQ alarms
# ---------------------------------------------------------------------------
resource "aws_cloudwatch_metric_alarm" "mq_cpu" {
  count = var.mq_broker_id != null ? 1 : 0

  alarm_name          = "${var.name}-mq-high-cpu"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "CpuUtilization"
  namespace           = "AWS/AmazonMQ"
  period              = 300
  statistic           = "Average"
  threshold           = 80
  alarm_actions       = [aws_sns_topic.alerts.arn]
  ok_actions          = [aws_sns_topic.alerts.arn]

  dimensions = {
    Broker = var.mq_broker_id
  }

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "mq_memory" {
  count = var.mq_broker_id != null ? 1 : 0

  alarm_name          = "${var.name}-mq-high-memory"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "MemoryUsage"
  namespace           = "AWS/AmazonMQ"
  period              = 300
  statistic           = "Average"
  threshold           = 80
  alarm_actions       = [aws_sns_topic.alerts.arn]
  ok_actions          = [aws_sns_topic.alerts.arn]

  dimensions = {
    Broker = var.mq_broker_id
  }

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "mq_connections" {
  count = var.mq_broker_id != null ? 1 : 0

  alarm_name          = "${var.name}-mq-high-connections"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "RabbitMQConnections"
  namespace           = "AWS/AmazonMQ"
  period              = 300
  statistic           = "Average"
  threshold           = 1000
  alarm_actions       = [aws_sns_topic.alerts.arn]
  ok_actions          = [aws_sns_topic.alerts.arn]

  dimensions = {
    Broker = var.mq_broker_id
  }

  tags = var.tags
}

# ---------------------------------------------------------------------------
# Optional: Amazon Managed Prometheus + Grafana
# ---------------------------------------------------------------------------
resource "aws_prometheus_workspace" "this" {
  count = var.enable_managed_prometheus ? 1 : 0

  alias = "${var.name}-amp"

  tags = var.tags
}

data "aws_iam_policy_document" "grafana_assume" {
  count = var.enable_managed_prometheus ? 1 : 0

  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["grafana.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "grafana" {
  count = var.enable_managed_prometheus ? 1 : 0

  name               = "${var.name}-grafana"
  assume_role_policy = data.aws_iam_policy_document.grafana_assume[0].json

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "grafana_prometheus" {
  count = var.enable_managed_prometheus ? 1 : 0

  role       = aws_iam_role.grafana[0].name
  policy_arn = "arn:aws:iam::aws:policy/AmazonPrometheusQueryAccess"
}

resource "aws_grafana_workspace" "this" {
  count = var.enable_managed_prometheus ? 1 : 0

  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["AWS_SSO"]
  permission_type          = "SERVICE_MANAGED"
  role_arn                 = aws_iam_role.grafana[0].arn
  data_sources             = ["PROMETHEUS"]

  tags = var.tags
}
