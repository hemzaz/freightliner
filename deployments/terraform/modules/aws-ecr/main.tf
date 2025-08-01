# AWS ECR Repository Module
# Creates and configures ECR repositories for container registry replication

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
  required_version = ">= 1.0"
}

# ECR Repository for source registry
resource "aws_ecr_repository" "source_repositories" {
  for_each = toset(var.source_repositories)
  
  name                 = each.value
  image_tag_mutability = var.image_tag_mutability
  
  force_delete = var.force_delete_repository

  encryption_configuration {
    encryption_type = "AES256"
  }

  image_scanning_configuration {
    scan_on_push = var.scan_on_push
  }

  tags = merge(var.common_tags, {
    Name        = each.value
    Type        = "source"
    Environment = var.environment
  })
}

# ECR Repository for destination registry
resource "aws_ecr_repository" "destination_repositories" {
  for_each = toset(var.destination_repositories)
  
  name                 = each.value
  image_tag_mutability = var.image_tag_mutability
  
  force_delete = var.force_delete_repository

  encryption_configuration {
    encryption_type = "AES256"
  }

  image_scanning_configuration {
    scan_on_push = var.scan_on_push
  }

  tags = merge(var.common_tags, {
    Name        = each.value
    Type        = "destination"
    Environment = var.environment
  })
}

# Lifecycle policy for all repositories
resource "aws_ecr_lifecycle_policy" "repository_lifecycle" {
  for_each = merge(
    { for repo in var.source_repositories : repo => aws_ecr_repository.source_repositories[repo] },
    { for repo in var.destination_repositories : repo => aws_ecr_repository.destination_repositories[repo] }
  )
  
  repository = each.value.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last ${var.max_image_count} images"
        selection = {
          tagStatus     = "tagged"
          countType     = "imageCountMoreThan"
          countNumber   = var.max_image_count
        }
        action = {
          type = "expire"
        }
      },
      {
        rulePriority = 2
        description  = "Delete untagged images older than 1 day"
        selection = {
          tagStatus   = "untagged"
          countType   = "sinceImagePushed"
          countUnit   = "days"
          countNumber = 1
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

# Repository policy for cross-account access
resource "aws_ecr_repository_policy" "repository_policy" {
  for_each = var.enable_cross_account_access ? merge(
    { for repo in var.source_repositories : repo => aws_ecr_repository.source_repositories[repo] },
    { for repo in var.destination_repositories : repo => aws_ecr_repository.destination_repositories[repo] }
  ) : {}
  
  repository = each.value.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowCrossAccountAccess"
        Effect = "Allow"
        Principal = {
          AWS = var.cross_account_arns
        }
        Action = [
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage",
          "ecr:BatchCheckLayerAvailability",
          "ecr:PutImage",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload",
          "ecr:BatchDeleteImage"
        ]
      }
    ]
  })
}

# IAM role for Freightliner application
resource "aws_iam_role" "freightliner_role" {
  name = "${var.name_prefix}-freightliner-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          AWS = var.assume_role_arns
        }
      }
    ]
  })

  tags = merge(var.common_tags, {
    Name        = "${var.name_prefix}-freightliner-role"
    Environment = var.environment
  })
}

# IAM policy for ECR access
resource "aws_iam_policy" "ecr_access_policy" {
  name        = "${var.name_prefix}-ecr-access-policy"
  description = "Policy for Freightliner ECR access"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage",
          "ecr:BatchCheckLayerAvailability",
          "ecr:PutImage",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload",
          "ecr:BatchDeleteImage",
          "ecr:ListImages",
          "ecr:DescribeImages",
          "ecr:DescribeRepositories"
        ]
        Resource = concat(
          [for repo in aws_ecr_repository.source_repositories : repo.arn],
          [for repo in aws_ecr_repository.destination_repositories : repo.arn]
        )
      }
    ]
  })

  tags = merge(var.common_tags, {
    Name        = "${var.name_prefix}-ecr-access-policy"
    Environment = var.environment
  })
}

# Attach ECR policy to role
resource "aws_iam_role_policy_attachment" "ecr_access_attachment" {
  role       = aws_iam_role.freightliner_role.name
  policy_arn = aws_iam_policy.ecr_access_policy.arn
}

# KMS key for ECR encryption (optional)
resource "aws_kms_key" "ecr_kms_key" {
  count = var.enable_kms_encryption ? 1 : 0
  
  description             = "KMS key for ECR encryption"
  deletion_window_in_days = var.kms_key_deletion_window
  
  tags = merge(var.common_tags, {
    Name        = "${var.name_prefix}-ecr-kms-key"
    Environment = var.environment
  })
}

# KMS key alias
resource "aws_kms_alias" "ecr_kms_key_alias" {
  count = var.enable_kms_encryption ? 1 : 0
  
  name          = "alias/${var.name_prefix}-ecr-key"
  target_key_id = aws_kms_key.ecr_kms_key[0].key_id
}

# CloudWatch Log Group for ECR events
resource "aws_cloudwatch_log_group" "ecr_events" {
  name              = "/aws/ecr/${var.name_prefix}"
  retention_in_days = var.log_retention_days
  
  tags = merge(var.common_tags, {
    Name        = "${var.name_prefix}-ecr-events"
    Environment = var.environment
  })
}

# EventBridge rule for ECR image pushes
resource "aws_cloudwatch_event_rule" "ecr_image_push" {
  count = var.enable_event_monitoring ? 1 : 0
  
  name        = "${var.name_prefix}-ecr-image-push"
  description = "Capture ECR image push events"

  event_pattern = jsonencode({
    source      = ["aws.ecr"]
    detail-type = ["ECR Image Action"]
    detail = {
      action-type = ["PUSH"]
      result      = ["SUCCESS"]
    }
  })

  tags = merge(var.common_tags, {
    Name        = "${var.name_prefix}-ecr-image-push"
    Environment = var.environment
  })
}

# CloudWatch Log Stream for EventBridge
resource "aws_cloudwatch_log_stream" "ecr_events_stream" {
  count          = var.enable_event_monitoring ? 1 : 0
  name           = "ecr-image-events"
  log_group_name = aws_cloudwatch_log_group.ecr_events.name
}

# EventBridge target to send events to CloudWatch Logs
resource "aws_cloudwatch_event_target" "ecr_events_target" {
  count = var.enable_event_monitoring ? 1 : 0
  
  rule      = aws_cloudwatch_event_rule.ecr_image_push[0].name
  target_id = "ECREventsLogTarget"
  arn       = aws_cloudwatch_log_group.ecr_events.arn
}