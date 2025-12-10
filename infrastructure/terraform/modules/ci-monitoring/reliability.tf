# Reliability Engineering Components
# Circuit breakers, retry mechanisms, and automated recovery

# Lambda function for cost optimization
resource "aws_lambda_function" "cost_optimizer" {
  filename         = "cost_optimizer.zip"
  function_name    = "${local.name_prefix}-cost-optimizer"
  role            = aws_iam_role.cost_optimizer_role.arn
  handler         = "index.handler"
  source_code_hash = data.archive_file.cost_optimizer.output_base64sha256
  runtime         = "python3.11"
  timeout         = 600  # 10 minutes for cost analysis
  memory_size     = 512

  environment {
    variables = {
      S3_BUCKET               = aws_s3_bucket.monitoring_data.id
      SNS_COST_TOPIC         = aws_sns_topic.cost_alerts.arn
      COST_THRESHOLD_USD     = var.cost_alert_threshold_usd
      LOG_LEVEL              = var.log_level
      AUTO_SCALING_ENABLED   = var.auto_scaling_enabled
      SCALING_TARGET_UTIL    = var.scaling_target_utilization
    }
  }

  vpc_config {
    subnet_ids         = var.subnet_ids
    security_group_ids = [aws_security_group.lambda_sg.id]
  }

  tags = local.common_tags

  depends_on = [
    aws_iam_role_policy.cost_optimizer_policy,
    aws_cloudwatch_log_group.cost_optimizer_logs,
  ]
}

resource "aws_cloudwatch_log_group" "cost_optimizer_logs" {
  name              = "/aws/lambda/${local.name_prefix}-cost-optimizer"
  retention_in_days = var.log_retention_days
  kms_key_id        = aws_kms_key.monitoring_encryption.arn

  tags = local.common_tags
}

data "archive_file" "cost_optimizer" {
  type        = "zip"
  output_path = "${path.module}/cost_optimizer.zip"
  source {
    content  = file("${path.module}/lambda/cost_optimizer.py")
    filename = "index.py"
  }
}

# Lambda function for automated recovery management
resource "aws_lambda_function" "recovery_manager" {
  filename         = "recovery_manager.zip"
  function_name    = "${local.name_prefix}-recovery-manager"
  role            = aws_iam_role.lambda_execution_role.arn
  handler         = "index.handler"
  source_code_hash = data.archive_file.recovery_manager.output_base64sha256
  runtime         = "python3.11"
  timeout         = 300
  memory_size     = 256

  environment {
    variables = {
      GITHUB_TOKEN                    = var.github_token
      GITHUB_OWNER                    = var.github_owner
      GITHUB_REPO                     = var.github_repo
      DYNAMODB_TABLE                  = aws_dynamodb_table.pipeline_metadata.name
      S3_BUCKET                       = aws_s3_bucket.monitoring_data.id
      SNS_CRITICAL_TOPIC             = aws_sns_topic.critical_alerts.arn
      CIRCUIT_BREAKER_ENABLED        = var.enable_circuit_breaker
      CIRCUIT_BREAKER_FAILURE_THRESHOLD = var.circuit_breaker_failure_threshold
      CIRCUIT_BREAKER_TIMEOUT        = var.circuit_breaker_timeout_seconds
      LOG_LEVEL                      = var.log_level
    }
  }

  vpc_config {
    subnet_ids         = var.subnet_ids
    security_group_ids = [aws_security_group.lambda_sg.id]
  }

  tags = local.common_tags

  depends_on = [
    aws_iam_role_policy_attachment.lambda_vpc_execution,
    aws_cloudwatch_log_group.recovery_manager_logs,
  ]
}

resource "aws_cloudwatch_log_group" "recovery_manager_logs" {
  name              = "/aws/lambda/${local.name_prefix}-recovery-manager"
  retention_in_days = var.log_retention_days
  kms_key_id        = aws_kms_key.monitoring_encryption.arn

  tags = local.common_tags
}

data "archive_file" "recovery_manager" {
  type        = "zip"
  output_path = "${path.module}/recovery_manager.zip"
  source {
    content  = file("${path.module}/lambda/recovery_manager.py")
    filename = "index.py"
  }
}

# CloudWatch Alarms for automated triggering of recovery procedures
resource "aws_cloudwatch_metric_alarm" "pipeline_failure_rate_high" {
  alarm_name          = "${local.name_prefix}-pipeline-failure-rate-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "PipelineFailureRate"
  namespace           = "CI-CD/Pipeline"
  period              = "300"  # 5 minutes
  statistic           = "Average"
  threshold           = "0.2"  # 20% failure rate
  alarm_description   = "This metric monitors pipeline failure rate"
  alarm_actions       = [
    aws_sns_topic.critical_alerts.arn,
    aws_lambda_function.recovery_manager.arn
  ]

  dimensions = {
    Repository = "${var.github_owner}/${var.github_repo}"
  }

  tags = local.common_tags
}

resource "aws_cloudwatch_metric_alarm" "pipeline_duration_high" {
  alarm_name          = "${local.name_prefix}-pipeline-duration-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "3"
  metric_name         = "AveragePipelineDuration"
  namespace           = "CI-CD/Pipeline"
  period              = "600"  # 10 minutes
  statistic           = "Average"
  threshold           = tostring(var.pipeline_duration_threshold_minutes)
  alarm_description   = "This metric monitors average pipeline duration"
  alarm_actions       = [
    aws_sns_topic.warning_alerts.arn
  ]

  dimensions = {
    Repository = "${var.github_owner}/${var.github_repo}"
  }

  tags = local.common_tags
}

resource "aws_cloudwatch_metric_alarm" "performance_regression_critical" {
  alarm_name          = "${local.name_prefix}-performance-regression-critical"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "1"
  metric_name         = "total_duration_baseline_deviation"
  namespace           = "CI-CD/Performance"
  period              = "300"
  statistic           = "Maximum"
  threshold           = "50"  # 50% performance degradation
  alarm_description   = "Critical performance regression detected"
  alarm_actions       = [
    aws_sns_topic.critical_alerts.arn,
    aws_lambda_function.recovery_manager.arn
  ]

  dimensions = {
    Repository = "${var.github_owner}/${var.github_repo}"
    MetricType = "baseline_comparison"
  }

  tags = local.common_tags
}

# Application Auto Scaling for CI/CD resources
resource "aws_appautoscaling_target" "lambda_concurrency" {
  count = var.auto_scaling_enabled ? 1 : 0

  max_capacity       = var.max_capacity
  min_capacity       = var.min_capacity
  resource_id        = "function:${aws_lambda_function.pipeline_metrics_collector.function_name}:provisioned"
  scalable_dimension = "lambda:function:ProvisionedConcurrencyConfig:ProvisionedConcurrencyUtilization"
  service_namespace  = "lambda"

  tags = local.common_tags
}

resource "aws_appautoscaling_policy" "lambda_scale_up" {
  count = var.auto_scaling_enabled ? 1 : 0

  name               = "${local.name_prefix}-lambda-scale-up"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.lambda_concurrency[0].resource_id
  scalable_dimension = aws_appautoscaling_target.lambda_concurrency[0].scalable_dimension
  service_namespace  = aws_appautoscaling_target.lambda_concurrency[0].service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "LambdaProvisionedConcurrencyUtilization"
    }
    target_value = var.scaling_target_utilization
  }
}

# DynamoDB table for circuit breaker state management
resource "aws_dynamodb_table" "circuit_breaker_state" {
  name           = "${local.name_prefix}-circuit-breaker-state"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "service_name"

  attribute {
    name = "service_name"
    type = "S"
  }

  ttl {
    attribute_name = "ttl"
    enabled        = true
  }

  server_side_encryption {
    enabled     = true
    kms_key_arn = aws_kms_key.monitoring_encryption.arn
  }

  tags = local.common_tags
}

# EventBridge rule for cost optimization scheduling
resource "aws_cloudwatch_event_rule" "cost_optimization_schedule" {
  count = var.enable_cost_optimization ? 1 : 0

  name                = "${local.name_prefix}-cost-optimization-schedule"
  description         = "Trigger cost optimization analysis"
  schedule_expression = var.cost_optimization_schedule

  tags = local.common_tags
}

resource "aws_cloudwatch_event_target" "cost_optimization_target" {
  count = var.enable_cost_optimization ? 1 : 0

  rule      = aws_cloudwatch_event_rule.cost_optimization_schedule[0].name
  target_id = "CostOptimizer"
  arn       = aws_lambda_function.cost_optimizer.arn
}

resource "aws_lambda_permission" "allow_eventbridge_cost_optimizer" {
  count = var.enable_cost_optimization ? 1 : 0

  statement_id  = "AllowExecutionFromEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.cost_optimizer.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.cost_optimization_schedule[0].arn
}

# Lambda permission for CloudWatch alarms to invoke recovery manager
resource "aws_lambda_permission" "allow_cloudwatch_recovery_manager" {
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.recovery_manager.function_name
  principal     = "lambda.alarms.cloudwatch.amazonaws.com"
  source_arn    = aws_cloudwatch_metric_alarm.pipeline_failure_rate_high.arn
}

# Health check endpoint using API Gateway for external monitoring
resource "aws_api_gateway_rest_api" "monitoring_health" {
  name        = "${local.name_prefix}-monitoring-health"
  description = "Health check endpoint for CI/CD monitoring system"

  endpoint_configuration {
    types = ["REGIONAL"]
  }

  tags = local.common_tags
}

resource "aws_api_gateway_resource" "health" {
  rest_api_id = aws_api_gateway_rest_api.monitoring_health.id
  parent_id   = aws_api_gateway_rest_api.monitoring_health.root_resource_id
  path_part   = "health"
}

resource "aws_api_gateway_method" "health_get" {
  rest_api_id   = aws_api_gateway_rest_api.monitoring_health.id
  resource_id   = aws_api_gateway_resource.health.id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "health_integration" {
  rest_api_id = aws_api_gateway_rest_api.monitoring_health.id
  resource_id = aws_api_gateway_resource.health.id
  http_method = aws_api_gateway_method.health_get.http_method
  
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.health_check.invoke_arn
}

# Health check Lambda function
resource "aws_lambda_function" "health_check" {
  filename         = "health_check.zip"
  function_name    = "${local.name_prefix}-health-check"
  role            = aws_iam_role.lambda_execution_role.arn
  handler         = "index.handler"
  source_code_hash = data.archive_file.health_check.output_base64sha256
  runtime         = "python3.11"
  timeout         = 30
  memory_size     = 128

  environment {
    variables = {
      DYNAMODB_TABLE_METADATA       = aws_dynamodb_table.pipeline_metadata.name
      DYNAMODB_TABLE_CIRCUIT_BREAKER = aws_dynamodb_table.circuit_breaker_state.name
      S3_BUCKET                     = aws_s3_bucket.monitoring_data.id
      GITHUB_OWNER                  = var.github_owner
      GITHUB_REPO                   = var.github_repo
    }
  }

  tags = local.common_tags
}

data "archive_file" "health_check" {
  type        = "zip"
  output_path = "${path.module}/health_check.zip"
  source {
    content  = file("${path.module}/lambda/health_check.py")
    filename = "index.py"
  }
}

resource "aws_lambda_permission" "allow_api_gateway_health" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.health_check.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.monitoring_health.execution_arn}/*/*"
}

resource "aws_api_gateway_deployment" "monitoring_health" {
  depends_on = [
    aws_api_gateway_integration.health_integration,
  ]

  rest_api_id = aws_api_gateway_rest_api.monitoring_health.id
  stage_name  = var.environment

  lifecycle {
    create_before_destroy = true
  }
}

# Dead Letter Queue for failed Lambda invocations
resource "aws_sqs_queue" "lambda_dlq" {
  name                       = "${local.name_prefix}-lambda-dlq"
  message_retention_seconds  = 1209600  # 14 days
  visibility_timeout_seconds = 300

  kms_master_key_id = aws_kms_key.monitoring_encryption.id

  tags = local.common_tags
}

# Update Lambda functions to use DLQ
resource "aws_lambda_function_event_invoke_config" "pipeline_metrics_collector_dlq" {
  function_name = aws_lambda_function.pipeline_metrics_collector.function_name

  destination_config {
    on_failure {
      destination = aws_sqs_queue.lambda_dlq.arn
    }
  }

  maximum_retry_attempts = 2
}

resource "aws_lambda_function_event_invoke_config" "performance_monitor_dlq" {
  function_name = aws_lambda_function.performance_monitor.function_name

  destination_config {
    on_failure {
      destination = aws_sqs_queue.lambda_dlq.arn
    }
  }

  maximum_retry_attempts = 2
}