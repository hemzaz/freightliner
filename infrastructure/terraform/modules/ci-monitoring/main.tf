# CI/CD Monitoring and Alerting System
# Comprehensive monitoring for GitHub Actions and infrastructure

terraform {
  required_version = ">= 1.5"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    github = {
      source  = "integrations/github"
      version = "~> 5.0"
    }
  }
}

locals {
  name_prefix = "${var.project_name}-${var.environment}"

  # Common tags for all resources
  common_tags = merge(var.tags, {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "terraform"
    Module      = "ci-monitoring"
    CreatedAt   = timestamp()
  })

  # Pipeline metrics configuration
  pipeline_metrics = {
    success_rate_threshold = 0.95
    duration_threshold_minutes = 30
    failure_rate_alert_threshold = 0.1
    cost_alert_threshold_usd = 100
  }

  # Dashboard configuration
  dashboard_config = {
    refresh_interval = "30s"
    time_range = "24h"
    data_retention_days = 90
  }
}

# CloudWatch Log Groups for CI/CD Metrics
resource "aws_cloudwatch_log_group" "ci_pipeline_metrics" {
  name              = "/aws/lambda/${local.name_prefix}-pipeline-metrics"
  retention_in_days = var.log_retention_days
  kms_key_id        = aws_kms_key.monitoring_encryption.arn

  tags = local.common_tags
}

resource "aws_cloudwatch_log_group" "ci_performance_metrics" {
  name              = "/aws/lambda/${local.name_prefix}-performance-metrics"
  retention_in_days = var.log_retention_days
  kms_key_id        = aws_kms_key.monitoring_encryption.arn

  tags = local.common_tags
}

# KMS Key for monitoring data encryption
resource "aws_kms_key" "monitoring_encryption" {
  description             = "KMS key for CI/CD monitoring encryption"
  deletion_window_in_days = var.kms_deletion_window

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "EnableIAMUserPermissions"
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "AllowCloudWatchAccess"
        Effect = "Allow"
        Principal = {
          Service = [
            "logs.amazonaws.com",
            "cloudwatch.amazonaws.com"
          ]
        }
        Action = [
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:ReEncrypt*",
          "kms:GenerateDataKey*",
          "kms:DescribeKey"
        ]
        Resource = "*"
      }
    ]
  })

  tags = local.common_tags
}

resource "aws_kms_alias" "monitoring_encryption" {
  name          = "alias/${local.name_prefix}-monitoring"
  target_key_id = aws_kms_key.monitoring_encryption.key_id
}

# S3 Bucket for monitoring data and dashboards
resource "aws_s3_bucket" "monitoring_data" {
  bucket = "${local.name_prefix}-ci-monitoring-data"

  tags = local.common_tags
}

resource "aws_s3_bucket_versioning" "monitoring_data" {
  bucket = aws_s3_bucket.monitoring_data.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_encryption" "monitoring_data" {
  bucket = aws_s3_bucket.monitoring_data.id

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        kms_master_key_id = aws_kms_key.monitoring_encryption.arn
        sse_algorithm     = "aws:kms"
      }
      bucket_key_enabled = true
    }
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "monitoring_data" {
  bucket = aws_s3_bucket.monitoring_data.id

  rule {
    id     = "monitoring_data_lifecycle"
    status = "Enabled"

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 90
      storage_class = "GLACIER"
    }

    expiration {
      days = var.data_retention_days
    }

    noncurrent_version_expiration {
      noncurrent_days = 30
    }
  }
}

# DynamoDB table for pipeline metadata and state
resource "aws_dynamodb_table" "pipeline_metadata" {
  name           = "${local.name_prefix}-pipeline-metadata"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "pipeline_id"
  range_key      = "execution_id"

  attribute {
    name = "pipeline_id"
    type = "S"
  }

  attribute {
    name = "execution_id"
    type = "S"
  }

  attribute {
    name = "status"
    type = "S"
  }

  attribute {
    name = "created_at"
    type = "S"
  }

  global_secondary_index {
    name     = "status-created_at-index"
    hash_key = "status"
    range_key = "created_at"
  }

  ttl {
    attribute_name = "ttl"
    enabled        = true
  }

  server_side_encryption {
    enabled     = true
    kms_key_arn = aws_kms_key.monitoring_encryption.arn
  }

  point_in_time_recovery {
    enabled = true
  }

  tags = local.common_tags
}

# SNS Topics for different types of alerts
resource "aws_sns_topic" "critical_alerts" {
  name              = "${local.name_prefix}-critical-alerts"
  kms_master_key_id = aws_kms_key.monitoring_encryption.id

  tags = local.common_tags
}

resource "aws_sns_topic" "warning_alerts" {
  name              = "${local.name_prefix}-warning-alerts"
  kms_master_key_id = aws_kms_key.monitoring_encryption.id

  tags = local.common_tags
}

resource "aws_sns_topic" "cost_alerts" {
  name              = "${local.name_prefix}-cost-alerts"
  kms_master_key_id = aws_kms_key.monitoring_encryption.id

  tags = local.common_tags
}

# Lambda function for pipeline metrics collection
resource "aws_lambda_function" "pipeline_metrics_collector" {
  filename         = "pipeline_metrics_collector.zip"
  function_name    = "${local.name_prefix}-pipeline-metrics-collector"
  role            = aws_iam_role.lambda_execution_role.arn
  handler         = "index.handler"
  source_code_hash = data.archive_file.pipeline_metrics_collector.output_base64sha256
  runtime         = "python3.11"
  timeout         = 300
  memory_size     = 256

  environment {
    variables = {
      GITHUB_TOKEN                    = var.github_token
      GITHUB_OWNER                    = var.github_owner
      GITHUB_REPO                     = var.github_repo
      DYNAMODB_TABLE                  = aws_dynamodb_table.pipeline_metadata.name
      DYNAMODB_TABLE_CIRCUIT_BREAKER  = aws_dynamodb_table.circuit_breaker_state.name
      S3_BUCKET                       = aws_s3_bucket.monitoring_data.id
      SNS_CRITICAL_TOPIC              = aws_sns_topic.critical_alerts.arn
      SNS_WARNING_TOPIC               = aws_sns_topic.warning_alerts.arn
      CIRCUIT_BREAKER_ENABLED         = var.enable_circuit_breaker
      CIRCUIT_BREAKER_FAILURE_THRESHOLD = var.circuit_breaker_failure_threshold
      CIRCUIT_BREAKER_TIMEOUT         = var.circuit_breaker_timeout_seconds
      LOG_LEVEL                       = var.log_level
    }
  }

  vpc_config {
    subnet_ids         = var.subnet_ids
    security_group_ids = [aws_security_group.lambda_sg.id]
  }

  tags = local.common_tags

  depends_on = [
    aws_iam_role_policy_attachment.lambda_vpc_execution,
    aws_cloudwatch_log_group.ci_pipeline_metrics,
    aws_dynamodb_table.circuit_breaker_state,
  ]
}

# Data source for Lambda deployment package
data "archive_file" "pipeline_metrics_collector" {
  type        = "zip"
  output_path = "${path.module}/pipeline_metrics_collector.zip"
  source {
    content = templatefile("${path.module}/lambda/pipeline_metrics_collector.py", {
      success_rate_threshold = local.pipeline_metrics.success_rate_threshold
      duration_threshold = local.pipeline_metrics.duration_threshold_minutes
      failure_rate_threshold = local.pipeline_metrics.failure_rate_alert_threshold
    })
    filename = "index.py"
  }
}

# Lambda function for performance monitoring
resource "aws_lambda_function" "performance_monitor" {
  filename         = "performance_monitor.zip"
  function_name    = "${local.name_prefix}-performance-monitor"
  role            = aws_iam_role.lambda_execution_role.arn
  handler         = "index.handler"
  source_code_hash = data.archive_file.performance_monitor.output_base64sha256
  runtime         = "python3.11"
  timeout         = 300
  memory_size     = 512

  environment {
    variables = {
      GITHUB_TOKEN        = var.github_token
      GITHUB_OWNER        = var.github_owner
      GITHUB_REPO         = var.github_repo
      S3_BUCKET          = aws_s3_bucket.monitoring_data.id
      CLOUDWATCH_NAMESPACE = "CI-CD/Performance"
      LOG_LEVEL          = var.log_level
    }
  }

  vpc_config {
    subnet_ids         = var.subnet_ids
    security_group_ids = [aws_security_group.lambda_sg.id]
  }

  tags = local.common_tags

  depends_on = [
    aws_iam_role_policy_attachment.lambda_vpc_execution,
    aws_cloudwatch_log_group.ci_performance_metrics,
  ]
}

data "archive_file" "performance_monitor" {
  type        = "zip"
  output_path = "${path.module}/performance_monitor.zip"
  source {
    content  = file("${path.module}/lambda/performance_monitor.py")
    filename = "index.py"
  }
}

# EventBridge rules for scheduled monitoring
resource "aws_cloudwatch_event_rule" "pipeline_metrics_schedule" {
  name                = "${local.name_prefix}-pipeline-metrics-schedule"
  description         = "Trigger pipeline metrics collection"
  schedule_expression = "rate(5 minutes)"

  tags = local.common_tags
}

resource "aws_cloudwatch_event_target" "pipeline_metrics_target" {
  rule      = aws_cloudwatch_event_rule.pipeline_metrics_schedule.name
  target_id = "PipelineMetricsCollectorTarget"
  arn       = aws_lambda_function.pipeline_metrics_collector.arn
}

resource "aws_lambda_permission" "allow_eventbridge_pipeline_metrics" {
  statement_id  = "AllowExecutionFromEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.pipeline_metrics_collector.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.pipeline_metrics_schedule.arn
}

resource "aws_cloudwatch_event_rule" "performance_monitor_schedule" {
  name                = "${local.name_prefix}-performance-monitor-schedule"
  description         = "Trigger performance monitoring"
  schedule_expression = "rate(10 minutes)"

  tags = local.common_tags
}

resource "aws_cloudwatch_event_target" "performance_monitor_target" {
  rule      = aws_cloudwatch_event_rule.performance_monitor_schedule.name
  target_id = "PerformanceMonitorTarget"
  arn       = aws_lambda_function.performance_monitor.arn
}

resource "aws_lambda_permission" "allow_eventbridge_performance_monitor" {
  statement_id  = "AllowExecutionFromEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.performance_monitor.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.performance_monitor_schedule.arn
}

# Security Group for Lambda functions
resource "aws_security_group" "lambda_sg" {
  name_prefix = "${local.name_prefix}-lambda-"
  vpc_id      = var.vpc_id
  description = "Security group for CI/CD monitoring Lambda functions"

  egress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTPS outbound for GitHub API and AWS services"
  }

  egress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTP outbound for package downloads"
  }

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-lambda-sg"
  })
}

# Data sources
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
