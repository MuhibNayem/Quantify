resource "random_password" "this" {
  count = var.password == null ? 1 : 0

  length           = 24
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

locals {
  admin_password = var.password != null ? var.password : random_password.this[0].result
}

resource "aws_kms_key" "this" {
  description             = "Amazon MQ encryption for ${var.name}"
  enable_key_rotation     = true
  deletion_window_in_days = 7

  tags = var.tags
}

resource "aws_security_group" "this" {
  name_prefix = "${var.name}-"
  description = "RabbitMQ access for ${var.name}"
  vpc_id      = var.vpc_id

  tags = var.tags

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_security_group_rule" "ingress" {
  for_each = toset(var.allowed_security_group_ids)

  type                     = "ingress"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
  source_security_group_id = each.value
  security_group_id        = aws_security_group.this.id
}

resource "aws_security_group_rule" "egress" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.this.id
}

resource "aws_mq_configuration" "this" {
  name           = "${var.name}-config"
  engine_type    = "RabbitMQ"
  engine_version = var.engine_version

  data = jsonencode({
    "rabbitmq" = {
      "consumer_timeout" = 1800000
    }
  })

  tags = var.tags
}

resource "aws_mq_broker" "this" {
  broker_name = var.name

  engine_type        = "RabbitMQ"
  engine_version     = var.engine_version
  host_instance_type = var.instance_type

  deployment_mode = "CLUSTER_MULTI_AZ"
  subnet_ids      = var.subnet_ids
  security_groups = [aws_security_group.this.id]

  publicly_accessible = false

  encryption_options {
    use_aws_owned_key = false
    kms_key_id        = aws_kms_key.this.arn
  }

  configuration {
    id       = aws_mq_configuration.this.id
    revision = aws_mq_configuration.this.latest_revision
  }

  user {
    username = var.username
    password = local.admin_password
  }

  logs {
    general = true
  }

  maintenance_window_start_time {
    day_of_week = "SUNDAY"
    time_of_day = "04:00"
    time_zone   = "UTC"
  }

  auto_minor_version_upgrade = true

  tags = var.tags
}
