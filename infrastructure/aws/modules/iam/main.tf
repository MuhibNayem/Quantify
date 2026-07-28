locals {
  oidc_issuer = replace(var.oidc_provider_url, "https://", "")

  backend_subject = "system:serviceaccount:${var.backend_namespace}:${var.backend_service_account}"
  ai_subject      = "system:serviceaccount:${var.ai_namespace}:${var.ai_service_account}"
}

# ---------------------------------------------------------------------------
# Shared policies
# ---------------------------------------------------------------------------
data "aws_iam_policy_document" "s3_access" {
  statement {
    sid = "BucketLevel"

    actions = [
      "s3:ListBucket",
      "s3:GetBucketLocation",
    ]

    resources = [var.s3_bucket_arn]
  }

  statement {
    sid = "ObjectLevel"

    actions = [
      "s3:GetObject",
      "s3:PutObject",
      "s3:DeleteObject",
      "s3:AbortMultipartUpload",
    ]

    resources = ["${var.s3_bucket_arn}/*"]
  }
}

resource "aws_iam_policy" "s3_access" {
  name   = "${var.eks_cluster_name}-s3-access"
  policy = data.aws_iam_policy_document.s3_access.json

  tags = var.tags
}

data "aws_iam_policy_document" "rds_secret_access" {
  count = var.rds_secret_arn != null ? 1 : 0

  statement {
    sid = "ReadSecret"

    actions = [
      "secretsmanager:GetSecretValue",
      "secretsmanager:DescribeSecret",
    ]

    resources = [var.rds_secret_arn]
  }

  statement {
    sid = "DecryptSecret"

    actions   = ["kms:Decrypt"]
    resources = ["*"]

    condition {
      test     = "StringLike"
      variable = "kms:ViaService"
      values   = ["secretsmanager.${var.region}.amazonaws.com"]
    }
  }
}

resource "aws_iam_policy" "rds_secret_access" {
  count = var.rds_secret_arn != null ? 1 : 0

  name   = "${var.eks_cluster_name}-rds-secret-access"
  policy = data.aws_iam_policy_document.rds_secret_access[0].json

  tags = var.tags
}

data "aws_iam_policy_document" "cloudwatch_logs" {
  statement {
    sid = "WriteLogs"

    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:DescribeLogStreams",
    ]

    resources = ["arn:aws:logs:${var.region}:*:log-group:/${var.log_group_prefix}/*"]
  }
}

resource "aws_iam_policy" "cloudwatch_logs" {
  name   = "${var.eks_cluster_name}-cloudwatch-logs"
  policy = data.aws_iam_policy_document.cloudwatch_logs.json

  tags = var.tags
}

# ---------------------------------------------------------------------------
# IRSA role: backend
# ---------------------------------------------------------------------------
data "aws_iam_policy_document" "backend_assume" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [var.oidc_provider_arn]
    }

    condition {
      test     = "StringEquals"
      variable = "${local.oidc_issuer}:aud"
      values   = ["sts.amazonaws.com"]
    }

    condition {
      test     = "StringEquals"
      variable = "${local.oidc_issuer}:sub"
      values   = [local.backend_subject]
    }
  }
}

resource "aws_iam_role" "backend" {
  name               = "${var.eks_cluster_name}-backend"
  assume_role_policy = data.aws_iam_policy_document.backend_assume.json

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "backend_s3" {
  role       = aws_iam_role.backend.name
  policy_arn = aws_iam_policy.s3_access.arn
}

resource "aws_iam_role_policy_attachment" "backend_rds_secret" {
  count = var.rds_secret_arn != null ? 1 : 0

  role       = aws_iam_role.backend.name
  policy_arn = aws_iam_policy.rds_secret_access[0].arn
}

resource "aws_iam_role_policy_attachment" "backend_logs" {
  role       = aws_iam_role.backend.name
  policy_arn = aws_iam_policy.cloudwatch_logs.arn
}

# ---------------------------------------------------------------------------
# IRSA role: ai-service
# ---------------------------------------------------------------------------
data "aws_iam_policy_document" "ai_assume" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [var.oidc_provider_arn]
    }

    condition {
      test     = "StringEquals"
      variable = "${local.oidc_issuer}:aud"
      values   = ["sts.amazonaws.com"]
    }

    condition {
      test     = "StringEquals"
      variable = "${local.oidc_issuer}:sub"
      values   = [local.ai_subject]
    }
  }
}

resource "aws_iam_role" "ai_service" {
  name               = "${var.eks_cluster_name}-ai-service"
  assume_role_policy = data.aws_iam_policy_document.ai_assume.json

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "ai_s3" {
  role       = aws_iam_role.ai_service.name
  policy_arn = aws_iam_policy.s3_access.arn
}

resource "aws_iam_role_policy_attachment" "ai_logs" {
  role       = aws_iam_role.ai_service.name
  policy_arn = aws_iam_policy.cloudwatch_logs.arn
}
