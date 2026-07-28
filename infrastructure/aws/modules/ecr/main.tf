resource "aws_ecr_repository" "this" {
  for_each = toset(var.repositories)

  name                 = "${var.name_prefix}/${each.value}"
  image_tag_mutability = "IMMUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "KMS"
  }

  tags = var.tags
}

resource "aws_ecr_lifecycle_policy" "this" {
  for_each = aws_ecr_repository.this

  repository = each.value.name

  policy = jsonencode({
    rules = [{
      rulePriority = 1
      description  = "Keep last ${var.max_image_count} images"
      selection = {
        tagStatus   = "any"
        countType   = "imageCountMoreThan"
        countNumber = var.max_image_count
      }
      action = {
        type = "expire"
      }
    }]
  })
}

data "aws_iam_policy_document" "registry" {
  count = length(var.cross_account_principals) > 0 ? 1 : 0

  statement {
    sid    = "CrossAccountPull"
    effect = "Allow"

    actions = [
      "ecr:GetAuthorizationToken",
      "ecr:BatchCheckLayerAvailability",
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage",
    ]

    principals {
      type        = "AWS"
      identifiers = [for account in var.cross_account_principals : "arn:aws:iam::${account}:root"]
    }
  }
}

resource "aws_ecr_registry_policy" "this" {
  count = length(var.cross_account_principals) > 0 ? 1 : 0

  policy = data.aws_iam_policy_document.registry[0].json
}
