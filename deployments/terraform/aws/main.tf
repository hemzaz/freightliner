# AWS Infrastructure for Freightliner Container Registry Replication
# This configuration creates AWS ECR repositories and associated resources

terraform {
  required_version = ">= 1.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
  
  # Configure remote state storage
  backend "s3" {
    # Configure these values according to your setup
    # bucket = "your-terraform-state-bucket"
    # key    = "freightliner/aws/terraform.tfstate"
    # region = "us-west-2"
    # dynamodb_table = "terraform-state-locks"
    # encrypt = true
  }
}

# Configure AWS Provider
provider "aws" {
  region = var.aws_region
  
  default_tags {
    tags = local.common_tags
  }
}

# Data sources
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# Local values
locals {
  common_tags = {
    Project     = "freightliner"
    Environment = var.environment
    ManagedBy   = "terraform"
    Component   = "container-registry"
    Team        = "platform"
    CostCenter  = var.cost_center
  }
  
  # Repository naming convention
  source_repositories = [
    for repo in var.source_repository_names : "${var.repository_prefix}${repo}"
  ]
  
  destination_repositories = [
    for repo in var.destination_repository_names : "${var.repository_prefix}${repo}"
  ]
}

# ECR Module
module "ecr" {
  source = "../modules/aws-ecr"
  
  name_prefix              = var.name_prefix
  environment             = var.environment
  source_repositories     = local.source_repositories
  destination_repositories = local.destination_repositories
  
  # Repository configuration
  image_tag_mutability    = var.image_tag_mutability
  scan_on_push           = var.scan_on_push
  force_delete_repository = var.force_delete_repository
  max_image_count        = var.max_image_count
  
  # Cross-account access
  enable_cross_account_access = var.enable_cross_account_access
  cross_account_arns         = var.cross_account_arns
  assume_role_arns          = var.assume_role_arns
  
  # Encryption
  enable_kms_encryption     = var.enable_kms_encryption
  kms_key_deletion_window   = var.kms_key_deletion_window
  
  # Monitoring
  enable_event_monitoring = var.enable_event_monitoring
  log_retention_days     = var.log_retention_days
  
  common_tags = local.common_tags
}

# Additional S3 bucket for storing replication checkpoints
resource "aws_s3_bucket" "replication_checkpoints" {
  bucket = "${var.name_prefix}-replication-checkpoints-${var.environment}-${random_id.bucket_suffix.hex}"
  
  tags = merge(local.common_tags, {
    Name        = "${var.name_prefix}-replication-checkpoints"
    Purpose     = "replication-state"
  })
}

# Random ID for unique bucket naming
resource "random_id" "bucket_suffix" {
  byte_length = 4
}

# S3 bucket versioning
resource "aws_s3_bucket_versioning" "checkpoint_versioning" {
  bucket = aws_s3_bucket.replication_checkpoints.id
  versioning_configuration {
    status = "Enabled"
  }
}

# S3 bucket encryption
resource "aws_s3_bucket_server_side_encryption_configuration" "checkpoint_encryption" {
  bucket = aws_s3_bucket.replication_checkpoints.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
    bucket_key_enabled = true
  }
}

# S3 bucket public access block
resource "aws_s3_bucket_public_access_block" "checkpoint_pab" {
  bucket = aws_s3_bucket.replication_checkpoints.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# S3 bucket lifecycle configuration
resource "aws_s3_bucket_lifecycle_configuration" "checkpoint_lifecycle" {
  bucket = aws_s3_bucket.replication_checkpoints.id

  rule {
    id     = "checkpoint_lifecycle"
    status = "Enabled"

    expiration {
      days = var.checkpoint_retention_days
    }

    noncurrent_version_expiration {
      noncurrent_days = 30
    }

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }
}

# IAM policy for S3 checkpoint access
resource "aws_iam_policy" "checkpoint_s3_access" {
  name        = "${var.name_prefix}-checkpoint-s3-access"
  description = "IAM policy for Freightliner S3 checkpoint access"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.replication_checkpoints.arn,
          "${aws_s3_bucket.replication_checkpoints.arn}/*"
        ]
      }
    ]
  })

  tags = local.common_tags
}

# Attach S3 policy to Freightliner role
resource "aws_iam_role_policy_attachment" "checkpoint_s3_attachment" {
  role       = module.ecr.freightliner_role_name
  policy_arn = aws_iam_policy.checkpoint_s3_access.arn
}

# CloudWatch dashboard for monitoring
resource "aws_cloudwatch_dashboard" "freightliner_dashboard" {
  count          = var.enable_monitoring_dashboard ? 1 : 0
  dashboard_name = "${var.name_prefix}-dashboard-${var.environment}"

  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        x      = 0
        y      = 0
        width  = 12
        height = 6

        properties = {
          metrics = [
            ["AWS/ECR", "RepositoryCount", "RepositoryName", "ALL_REPOSITORIES"],
            [".", "ImageCount", ".", "."]
          ]
          view    = "timeSeries"
          stacked = false
          region  = var.aws_region
          title   = "ECR Repository Metrics"
          period  = 300
        }
      },
      {
        type   = "log"
        x      = 0
        y      = 6
        width  = 12
        height = 6

        properties = {
          query   = "SOURCE '${module.ecr.cloudwatch_log_group_name}'\n| fields @timestamp, @message\n| sort @timestamp desc\n| limit 20"
          region  = var.aws_region
          title   = "ECR Events"
          view    = "table"
        }
      }
    ]
  })
}

# SNS topic for alerts
resource "aws_sns_topic" "freightliner_alerts" {
  count = var.enable_sns_alerts ? 1 : 0
  name  = "${var.name_prefix}-alerts-${var.environment}"
  
  tags = local.common_tags
}

# SNS topic subscription
resource "aws_sns_topic_subscription" "email_alerts" {
  count     = var.enable_sns_alerts ? length(var.alert_email_addresses) : 0
  topic_arn = aws_sns_topic.freightliner_alerts[0].arn
  protocol  = "email"
  endpoint  = var.alert_email_addresses[count.index]
}

# CloudWatch alarm for repository count
resource "aws_cloudwatch_metric_alarm" "repository_count_alarm" {
  count = var.enable_repository_alarms ? 1 : 0
  
  alarm_name          = "${var.name_prefix}-repository-count-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "RepositoryCount"
  namespace           = "AWS/ECR"
  period              = "300"
  statistic           = "Maximum"
  threshold           = var.repository_count_threshold
  alarm_description   = "This metric monitors ECR repository count"
  alarm_actions       = var.enable_sns_alerts ? [aws_sns_topic.freightliner_alerts[0].arn] : []

  tags = local.common_tags
}

# Secrets Manager secret for cross-cloud credentials
resource "aws_secretsmanager_secret" "gcp_credentials" {
  count       = var.create_gcp_secret ? 1 : 0
  name        = "${var.name_prefix}-gcp-credentials-${var.environment}"
  description = "GCP service account credentials for Freightliner replication"
  
  tags = local.common_tags
}

# Secrets Manager secret version (placeholder)
resource "aws_secretsmanager_secret_version" "gcp_credentials_version" {
  count         = var.create_gcp_secret ? 1 : 0
  secret_id     = aws_secretsmanager_secret.gcp_credentials[0].id
  secret_string = jsonencode({
    type = "service_account"
    # Add your GCP service account key here
    # This should be populated via terraform variables or external process
  })
  
  lifecycle {
    ignore_changes = [secret_string]
  }
}

# IAM policy for Secrets Manager access
resource "aws_iam_policy" "secrets_manager_access" {
  count       = var.create_gcp_secret ? 1 : 0
  name        = "${var.name_prefix}-secrets-manager-access"
  description = "IAM policy for Secrets Manager access"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret"
        ]
        Resource = aws_secretsmanager_secret.gcp_credentials[0].arn
      }
    ]
  })

  tags = local.common_tags
}

# Attach Secrets Manager policy to Freightliner role
resource "aws_iam_role_policy_attachment" "secrets_manager_attachment" {
  count      = var.create_gcp_secret ? 1 : 0
  role       = module.ecr.freightliner_role_name
  policy_arn = aws_iam_policy.secrets_manager_access[0].arn
}