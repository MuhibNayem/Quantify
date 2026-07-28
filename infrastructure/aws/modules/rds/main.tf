resource "random_password" "this" {
  count = var.password == null ? 1 : 0

  length           = 32
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

locals {
  master_password = var.password != null ? var.password : random_password.this[0].result

  endpoint = var.use_aurora ? aws_rds_cluster.this[0].endpoint : aws_db_instance.primary[0].address
  reader_endpoint = var.use_aurora ? aws_rds_cluster.this[0].reader_endpoint : (
    var.create_read_replica ? aws_db_instance.replica[0].address : null
  )
}

resource "aws_kms_key" "this" {
  description             = "RDS encryption for ${var.identifier}"
  enable_key_rotation     = true
  deletion_window_in_days = 7

  tags = var.tags
}

resource "aws_db_subnet_group" "this" {
  name       = var.identifier
  subnet_ids = var.subnet_ids

  tags = var.tags
}

resource "aws_security_group" "this" {
  name_prefix = "${var.identifier}-"
  description = "PostgreSQL access for ${var.identifier}"
  vpc_id      = var.vpc_id

  tags = var.tags

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_security_group_rule" "ingress" {
  for_each = toset(var.allowed_security_group_ids)

  type                     = "ingress"
  from_port                = var.port
  to_port                  = var.port
  protocol                 = "tcp"
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

data "aws_iam_policy_document" "monitoring_assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["monitoring.rds.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "enhanced_monitoring" {
  name               = "${var.identifier}-rds-monitoring"
  assume_role_policy = data.aws_iam_policy_document.monitoring_assume.json

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "enhanced_monitoring" {
  role       = aws_iam_role.enhanced_monitoring.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
}

# ---------------------------------------------------------------------------
# Standard RDS PostgreSQL (Multi-AZ + read replica)
# ---------------------------------------------------------------------------
resource "aws_db_parameter_group" "this" {
  count = var.use_aurora ? 0 : 1

  name   = "${var.identifier}-pg15"
  family = "postgres15"

  parameter {
    name  = "max_connections"
    value = var.max_connections
  }

  parameter {
    name  = "log_min_duration_statement"
    value = "1000"
  }

  parameter {
    name  = "idle_in_transaction_session_timeout"
    value = "60000"
  }

  parameter {
    name         = "shared_buffers"
    value        = "{DBInstanceClassMemory/32768}"
    apply_method = "pending-reboot"
  }

  tags = var.tags
}

resource "aws_db_instance" "primary" {
  count = var.use_aurora ? 0 : 1

  identifier     = var.identifier
  engine         = "postgres"
  engine_version = var.engine_version
  instance_class = var.instance_class
  port           = var.port

  allocated_storage     = var.allocated_storage
  max_allocated_storage = var.max_allocated_storage
  storage_type          = "gp3"
  storage_encrypted     = true
  kms_key_id            = aws_kms_key.this.arn

  db_name  = var.db_name
  username = var.username
  password = local.master_password

  multi_az               = var.multi_az
  db_subnet_group_name   = aws_db_subnet_group.this.name
  vpc_security_group_ids = [aws_security_group.this.id]
  parameter_group_name   = aws_db_parameter_group.this[0].name

  backup_retention_period = var.backup_retention_period
  backup_window           = "03:00-04:00"
  maintenance_window      = "sun:04:00-sun:05:00"

  performance_insights_enabled    = true
  performance_insights_kms_key_id = aws_kms_key.this.arn
  monitoring_interval             = 60
  monitoring_role_arn             = aws_iam_role.enhanced_monitoring.arn

  auto_minor_version_upgrade = true
  deletion_protection        = var.deletion_protection
  skip_final_snapshot        = var.skip_final_snapshot
  final_snapshot_identifier  = var.skip_final_snapshot ? null : "${var.identifier}-final"

  tags = var.tags
}

resource "aws_db_instance" "replica" {
  count = var.use_aurora ? 0 : (var.create_read_replica ? 1 : 0)

  identifier          = "${var.identifier}-replica"
  replicate_source_db = aws_db_instance.primary[0].identifier
  replica_mode        = "open-read-only"
  instance_class      = var.instance_class
  port                = var.port

  storage_encrypted = true
  kms_key_id        = aws_kms_key.this.arn

  db_subnet_group_name   = aws_db_subnet_group.this.name
  vpc_security_group_ids = [aws_security_group.this.id]
  parameter_group_name   = aws_db_parameter_group.this[0].name

  performance_insights_enabled    = true
  performance_insights_kms_key_id = aws_kms_key.this.arn
  monitoring_interval             = 60
  monitoring_role_arn             = aws_iam_role.enhanced_monitoring.arn

  auto_minor_version_upgrade = true
  deletion_protection        = var.deletion_protection
  skip_final_snapshot        = true

  tags = var.tags
}

# ---------------------------------------------------------------------------
# Aurora PostgreSQL Serverless v2 (optional)
# ---------------------------------------------------------------------------
resource "aws_rds_cluster_parameter_group" "aurora" {
  count = var.use_aurora ? 1 : 0

  name   = "${var.identifier}-aurora-pg15"
  family = "aurora-postgresql15"

  parameter {
    name  = "log_min_duration_statement"
    value = "1000"
  }

  tags = var.tags
}

resource "aws_rds_cluster" "this" {
  count = var.use_aurora ? 1 : 0

  cluster_identifier = var.identifier
  engine             = "aurora-postgresql"
  engine_version     = var.aurora_engine_version
  port               = var.port

  database_name   = var.db_name
  master_username = var.username
  master_password = local.master_password

  db_subnet_group_name            = aws_db_subnet_group.this.name
  vpc_security_group_ids          = [aws_security_group.this.id]
  db_cluster_parameter_group_name = aws_rds_cluster_parameter_group.aurora[0].name

  storage_encrypted = true
  kms_key_id        = aws_kms_key.this.arn

  backup_retention_period = var.backup_retention_period

  serverlessv2_scaling_configuration {
    min_capacity = var.aurora_min_capacity
    max_capacity = var.aurora_max_capacity
  }

  deletion_protection       = var.deletion_protection
  skip_final_snapshot       = var.skip_final_snapshot
  final_snapshot_identifier = var.skip_final_snapshot ? null : "${var.identifier}-final"

  tags = var.tags
}

resource "aws_rds_cluster_instance" "this" {
  count = var.use_aurora ? var.aurora_instance_count : 0

  identifier         = "${var.identifier}-${count.index}"
  cluster_identifier = aws_rds_cluster.this[0].id
  instance_class     = "db.serverless"
  engine             = aws_rds_cluster.this[0].engine
  engine_version     = aws_rds_cluster.this[0].engine_version

  performance_insights_enabled = true
  monitoring_interval          = 60
  monitoring_role_arn          = aws_iam_role.enhanced_monitoring.arn

  tags = var.tags
}

# ---------------------------------------------------------------------------
# Credentials in Secrets Manager
# ---------------------------------------------------------------------------
resource "aws_secretsmanager_secret" "credentials" {
  name                    = "${var.identifier}-credentials"
  recovery_window_in_days = 7

  tags = var.tags
}

resource "aws_secretsmanager_secret_version" "credentials" {
  secret_id = aws_secretsmanager_secret.credentials.id

  secret_string = jsonencode({
    engine   = "postgres"
    host     = local.endpoint
    port     = var.port
    username = var.username
    password = local.master_password
    dbname   = var.db_name
  })
}
