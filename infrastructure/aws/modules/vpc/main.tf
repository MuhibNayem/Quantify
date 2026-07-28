locals {
  public_subnet_cidrs  = { for i, az in var.azs : az => cidrsubnet(var.cidr, 4, i) }
  private_subnet_cidrs = { for i, az in var.azs : az => cidrsubnet(var.cidr, 4, i + 8) }

  nat_azs           = var.single_nat_gateway ? [var.azs[0]] : var.azs
  private_route_nat = { for az in var.azs : az => var.single_nat_gateway ? var.azs[0] : az }
}

resource "aws_vpc" "this" {
  cidr_block           = var.cidr
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = merge(var.tags, { Name = var.name })
}

resource "aws_internet_gateway" "this" {
  vpc_id = aws_vpc.this.id

  tags = merge(var.tags, { Name = var.name })
}

resource "aws_subnet" "public" {
  for_each = toset(var.azs)

  vpc_id                  = aws_vpc.this.id
  cidr_block              = local.public_subnet_cidrs[each.value]
  availability_zone       = each.value
  map_public_ip_on_launch = true

  tags = merge(var.tags, {
    Name                     = "${var.name}-public-${each.value}"
    Tier                     = "public"
    "kubernetes.io/role/elb" = "1"
  })
}

resource "aws_subnet" "private" {
  for_each = toset(var.azs)

  vpc_id            = aws_vpc.this.id
  cidr_block        = local.private_subnet_cidrs[each.value]
  availability_zone = each.value

  tags = merge(var.tags, {
    Name                              = "${var.name}-private-${each.value}"
    Tier                              = "private"
    "kubernetes.io/role/internal-elb" = "1"
  })
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.this.id

  tags = merge(var.tags, { Name = "${var.name}-public" })
}

resource "aws_route" "public_internet" {
  route_table_id         = aws_route_table.public.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.this.id
}

resource "aws_route_table" "private" {
  for_each = toset(var.azs)

  vpc_id = aws_vpc.this.id

  tags = merge(var.tags, { Name = "${var.name}-private-${each.value}" })
}

resource "aws_eip" "nat" {
  for_each = var.enable_nat_gateway ? toset(local.nat_azs) : toset([])

  domain = "vpc"

  tags = merge(var.tags, { Name = "${var.name}-nat-${each.value}" })

  depends_on = [aws_internet_gateway.this]
}

resource "aws_nat_gateway" "this" {
  for_each = var.enable_nat_gateway ? toset(local.nat_azs) : toset([])

  allocation_id = aws_eip.nat[each.value].id
  subnet_id     = aws_subnet.public[each.value].id

  tags = merge(var.tags, { Name = "${var.name}-nat-${each.value}" })

  depends_on = [aws_internet_gateway.this]
}

resource "aws_route" "private_nat" {
  for_each = var.enable_nat_gateway ? toset(var.azs) : toset([])

  route_table_id         = aws_route_table.private[each.value].id
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = aws_nat_gateway.this[local.private_route_nat[each.value]].id
}

resource "aws_route_table_association" "public" {
  for_each = toset(var.azs)

  subnet_id      = aws_subnet.public[each.value].id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table_association" "private" {
  for_each = toset(var.azs)

  subnet_id      = aws_subnet.private[each.value].id
  route_table_id = aws_route_table.private[each.value].id
}

resource "aws_cloudwatch_log_group" "flow_logs" {
  count = var.enable_flow_logs ? 1 : 0

  name              = "/aws/vpc/flow-logs/${var.name}"
  retention_in_days = 30

  tags = var.tags
}

data "aws_iam_policy_document" "flow_logs_assume" {
  count = var.enable_flow_logs ? 1 : 0

  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["vpc-flow-logs.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "flow_logs" {
  count = var.enable_flow_logs ? 1 : 0

  name               = "${var.name}-vpc-flow-logs"
  assume_role_policy = data.aws_iam_policy_document.flow_logs_assume[0].json

  tags = var.tags
}

data "aws_iam_policy_document" "flow_logs_write" {
  count = var.enable_flow_logs ? 1 : 0

  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:DescribeLogGroups",
      "logs:DescribeLogStreams",
    ]
    resources = ["${aws_cloudwatch_log_group.flow_logs[0].arn}:*"]
  }
}

resource "aws_iam_role_policy" "flow_logs" {
  count = var.enable_flow_logs ? 1 : 0

  name   = "${var.name}-vpc-flow-logs"
  role   = aws_iam_role.flow_logs[0].id
  policy = data.aws_iam_policy_document.flow_logs_write[0].json
}

resource "aws_flow_log" "this" {
  count = var.enable_flow_logs ? 1 : 0

  vpc_id          = aws_vpc.this.id
  traffic_type    = "ALL"
  iam_role_arn    = aws_iam_role.flow_logs[0].arn
  log_destination = aws_cloudwatch_log_group.flow_logs[0].arn

  tags = var.tags
}
