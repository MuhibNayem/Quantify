locals {
  name = "${var.project_name}-${var.environment}"

  common_tags = merge(
    {
      Project     = var.project_name
      Environment = var.environment
      ManagedBy   = "terraform"
      Owner       = var.owner
    },
    var.tags
  )
}

data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_caller_identity" "current" {}

locals {
  azs = length(var.availability_zones) > 0 ? var.availability_zones : slice(data.aws_availability_zones.available.names, 0, 3)
}

module "vpc" {
  source = "./modules/vpc"

  name = local.name
  cidr = var.vpc_cidr
  azs  = local.azs

  enable_nat_gateway = var.enable_nat_gateway
  single_nat_gateway = var.single_nat_gateway
  enable_flow_logs   = var.enable_vpc_flow_logs

  tags = local.common_tags
}

module "eks" {
  source = "./modules/eks"

  cluster_name    = "${local.name}-eks"
  cluster_version = var.eks_cluster_version

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnet_ids

  endpoint_public_access  = var.eks_public_endpoint
  endpoint_private_access = true

  node_groups                = var.eks_node_groups
  install_cluster_autoscaler = var.install_cluster_autoscaler

  tags = local.common_tags
}

module "rds" {
  source = "./modules/rds"
  count  = var.enable_rds ? 1 : 0

  identifier     = "${local.name}-postgres"
  engine_version = "15"
  instance_class = var.db_instance_class

  allocated_storage   = var.db_allocated_storage
  multi_az            = var.db_multi_az
  create_read_replica = var.db_create_read_replica
  use_aurora          = var.db_use_aurora

  db_name  = "inventory_db"
  username = var.db_username
  password = var.db_password

  backup_retention_period = var.db_backup_retention_period

  subnet_ids                 = module.vpc.private_subnet_ids
  vpc_id                     = module.vpc.vpc_id
  allowed_security_group_ids = [module.eks.node_security_group_id]

  tags = local.common_tags
}

module "elasticache" {
  source = "./modules/elasticache"
  count  = var.enable_elasticache ? 1 : 0

  name                    = "${local.name}-redis"
  node_type               = var.redis_node_type
  num_node_groups         = var.redis_num_node_groups
  replicas_per_node_group = var.redis_replicas_per_node_group

  subnet_ids                 = module.vpc.private_subnet_ids
  vpc_id                     = module.vpc.vpc_id
  allowed_security_group_ids = [module.eks.node_security_group_id]

  tags = local.common_tags
}

module "mq" {
  source = "./modules/mq"
  count  = var.enable_mq ? 1 : 0

  name          = "${local.name}-rabbitmq"
  instance_type = var.mq_instance_type
  username      = var.mq_username
  password      = var.mq_password

  subnet_ids                 = module.vpc.private_subnet_ids
  vpc_id                     = module.vpc.vpc_id
  allowed_security_group_ids = [module.eks.node_security_group_id]

  tags = local.common_tags
}

module "s3" {
  source = "./modules/s3"

  bucket_name        = "${local.name}-${var.s3_bucket_suffix}-${data.aws_caller_identity.current.account_id}"
  versioning_enabled = var.s3_versioning_enabled

  tags = local.common_tags
}

module "ecr" {
  source = "./modules/ecr"

  repositories             = var.ecr_repositories
  max_image_count          = var.ecr_max_image_count
  cross_account_principals = var.ecr_cross_account_principals

  tags = local.common_tags
}

module "iam" {
  source = "./modules/iam"

  eks_cluster_name  = module.eks.cluster_name
  oidc_provider_arn = module.eks.oidc_provider_arn
  oidc_provider_url = module.eks.oidc_provider_url
  region            = var.region

  s3_bucket_arn  = module.s3.bucket_arn
  rds_secret_arn = var.enable_rds ? module.rds[0].secret_arn : null

  tags = local.common_tags
}

module "monitoring" {
  source = "./modules/monitoring"
  count  = var.enable_monitoring ? 1 : 0

  name = local.name

  eks_cluster_name                 = module.eks.cluster_name
  eks_log_group_name               = module.eks.cloudwatch_log_group_name
  rds_identifier                   = var.enable_rds ? module.rds[0].identifier : null
  elasticache_replication_group_id = var.enable_elasticache ? module.elasticache[0].replication_group_id : null
  mq_broker_id                     = var.enable_mq ? module.mq[0].broker_id : null

  alert_email_endpoints     = var.alert_email_endpoints
  enable_managed_prometheus = var.enable_managed_prometheus

  tags = local.common_tags
}
