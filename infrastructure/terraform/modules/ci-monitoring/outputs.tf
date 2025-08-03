# Outputs for CI/CD Monitoring Module

# Infrastructure Resource Outputs
output "s3_bucket_name" {
  description = "Name of the S3 bucket for monitoring data"
  value       = aws_s3_bucket.monitoring_data.bucket
}

output "s3_bucket_arn" {
  description = "ARN of the S3 bucket for monitoring data"
  value       = aws_s3_bucket.monitoring_data.arn
}

output "dynamodb_table_name" {
  description = "Name of the DynamoDB table for pipeline metadata"
  value       = aws_dynamodb_table.pipeline_metadata.name
}

output "dynamodb_table_arn" {
  description = "ARN of the DynamoDB table for pipeline metadata"
  value       = aws_dynamodb_table.pipeline_metadata.arn
}

output "circuit_breaker_table_name" {
  description = "Name of the DynamoDB table for circuit breaker state"
  value       = aws_dynamodb_table.circuit_breaker_state.name
}

output "circuit_breaker_table_arn" {
  description = "ARN of the DynamoDB table for circuit breaker state"
  value       = aws_dynamodb_table.circuit_breaker_state.arn
}

# KMS and Security Outputs
output "kms_key_id" {
  description = "ID of the KMS key used for monitoring data encryption"
  value       = aws_kms_key.monitoring_encryption.key_id
}

output "kms_key_arn" {
  description = "ARN of the KMS key used for monitoring data encryption"
  value       = aws_kms_key.monitoring_encryption.arn
}

output "kms_key_alias" {
  description = "Alias of the KMS key used for monitoring data encryption"
  value       = aws_kms_alias.monitoring_encryption.name
}

# SNS Topic Outputs
output "critical_alerts_topic_arn" {
  description = "ARN of the SNS topic for critical alerts"
  value       = aws_sns_topic.critical_alerts.arn
}

output "warning_alerts_topic_arn" {
  description = "ARN of the SNS topic for warning alerts"
  value       = aws_sns_topic.warning_alerts.arn
}

output "cost_alerts_topic_arn" {
  description = "ARN of the SNS topic for cost alerts"
  value       = aws_sns_topic.cost_alerts.arn
}

# Lambda Function Outputs
output "pipeline_metrics_collector_function_name" {
  description = "Name of the pipeline metrics collector Lambda function"
  value       = aws_lambda_function.pipeline_metrics_collector.function_name
}

output "pipeline_metrics_collector_function_arn" {
  description = "ARN of the pipeline metrics collector Lambda function"
  value       = aws_lambda_function.pipeline_metrics_collector.arn
}

output "performance_monitor_function_name" {
  description = "Name of the performance monitor Lambda function"
  value       = aws_lambda_function.performance_monitor.function_name
}

output "performance_monitor_function_arn" {
  description = "ARN of the performance monitor Lambda function"
  value       = aws_lambda_function.performance_monitor.arn
}

output "cost_optimizer_function_name" {
  description = "Name of the cost optimizer Lambda function"
  value       = aws_lambda_function.cost_optimizer.function_name
}

output "cost_optimizer_function_arn" {
  description = "ARN of the cost optimizer Lambda function"
  value       = aws_lambda_function.cost_optimizer.arn
}

output "recovery_manager_function_name" {
  description = "Name of the recovery manager Lambda function"
  value       = aws_lambda_function.recovery_manager.function_name
}

output "recovery_manager_function_arn" {
  description = "ARN of the recovery manager Lambda function"
  value       = aws_lambda_function.recovery_manager.arn
}

output "health_check_function_name" {
  description = "Name of the health check Lambda function"
  value       = aws_lambda_function.health_check.function_name
}

output "health_check_function_arn" {
  description = "ARN of the health check Lambda function"
  value       = aws_lambda_function.health_check.arn
}

# API Gateway Outputs
output "health_check_api_id" {
  description = "ID of the health check API Gateway"
  value       = aws_api_gateway_rest_api.monitoring_health.id
}

output "health_check_api_url" {
  description = "URL of the health check API endpoint"
  value       = "https://${aws_api_gateway_rest_api.monitoring_health.id}.execute-api.${data.aws_region.current.name}.amazonaws.com/${var.environment}/health"
}

# CloudWatch Outputs
output "log_group_pipeline_metrics" {
  description = "Name of the CloudWatch log group for pipeline metrics"
  value       = aws_cloudwatch_log_group.ci_pipeline_metrics.name
}

output "log_group_performance_metrics" {
  description = "Name of the CloudWatch log group for performance metrics"
  value       = aws_cloudwatch_log_group.ci_performance_metrics.name
}

# Dashboard Outputs
output "dashboard_urls" {
  description = "URLs for accessing monitoring dashboards"
  value       = local.dashboard_urls
}

output "pipeline_overview_dashboard_name" {
  description = "Name of the pipeline overview CloudWatch dashboard"
  value       = aws_cloudwatch_dashboard.ci_cd_pipeline_overview.dashboard_name
}

output "performance_analysis_dashboard_name" {
  description = "Name of the performance analysis CloudWatch dashboard"
  value       = aws_cloudwatch_dashboard.performance_analysis.dashboard_name
}

output "cost_analysis_dashboard_name" {
  description = "Name of the cost analysis CloudWatch dashboard"
  value       = aws_cloudwatch_dashboard.cost_analysis.dashboard_name
}

# Grafana Outputs (conditional)
output "grafana_workspace_id" {
  description = "ID of the Grafana workspace (if enabled)"
  value       = var.enable_grafana_dashboard ? aws_grafana_workspace.monitoring[0].id : null
}

output "grafana_workspace_endpoint" {
  description = "Endpoint URL of the Grafana workspace (if enabled)"
  value       = var.enable_grafana_dashboard ? aws_grafana_workspace.monitoring[0].endpoint : null
}

output "grafana_workspace_status" {
  description = "Status of the Grafana workspace (if enabled)"
  value       = var.enable_grafana_dashboard ? aws_grafana_workspace.monitoring[0].status : null
}

# Alarm Outputs
output "pipeline_failure_rate_alarm_name" {
  description = "Name of the pipeline failure rate alarm"
  value       = aws_cloudwatch_metric_alarm.pipeline_failure_rate_high.alarm_name
}

output "pipeline_duration_alarm_name" {
  description = "Name of the pipeline duration alarm"
  value       = aws_cloudwatch_metric_alarm.pipeline_duration_high.alarm_name
}

output "performance_regression_alarm_name" {
  description = "Name of the performance regression alarm"
  value       = aws_cloudwatch_metric_alarm.performance_regression_critical.alarm_name
}

output "system_health_composite_alarm_name" {
  description = "Name of the system health composite alarm"
  value       = aws_cloudwatch_composite_alarm.ci_cd_system_health.alarm_name
}

# IAM Role Outputs
output "lambda_execution_role_arn" {
  description = "ARN of the Lambda execution role"
  value       = aws_iam_role.lambda_execution_role.arn
}

output "lambda_execution_role_name" {
  description = "Name of the Lambda execution role"
  value       = aws_iam_role.lambda_execution_role.name
}

output "github_actions_role_arn" {
  description = "ARN of the GitHub Actions role (if OIDC is enabled)"
  value       = var.enable_github_oidc ? aws_iam_role.github_actions_role[0].arn : null
}

# Auto Scaling Outputs
output "lambda_autoscaling_target_resource_id" {
  description = "Resource ID of the Lambda auto scaling target (if enabled)"
  value       = var.auto_scaling_enabled ? aws_appautoscaling_target.lambda_concurrency[0].resource_id : null
}

output "lambda_autoscaling_policy_name" {
  description = "Name of the Lambda auto scaling policy (if enabled)"
  value       = var.auto_scaling_enabled ? aws_appautoscaling_policy.lambda_scale_up[0].name : null
}

# Cost Optimization Outputs
output "monthly_cost_threshold_usd" {
  description = "Monthly cost threshold for alerts (USD)"
  value       = var.cost_alert_threshold_usd
}

output "cost_optimization_schedule" {
  description = "Schedule expression for cost optimization"
  value       = var.enable_cost_optimization ? var.cost_optimization_schedule : null
}

# Circuit Breaker Configuration
output "circuit_breaker_enabled" {
  description = "Whether circuit breaker functionality is enabled"
  value       = var.enable_circuit_breaker
}

output "circuit_breaker_failure_threshold" {
  description = "Circuit breaker failure threshold"
  value       = var.circuit_breaker_failure_threshold
}

output "circuit_breaker_timeout_seconds" {
  description = "Circuit breaker timeout in seconds"
  value       = var.circuit_breaker_timeout_seconds
}

# Performance Monitoring Configuration
output "performance_baseline_enabled" {
  description = "Whether performance baseline tracking is enabled"
  value       = var.enable_performance_baseline
}

output "performance_regression_threshold" {
  description = "Performance regression threshold percentage"
  value       = var.performance_regression_threshold
}

# GitHub Integration Details
output "monitoring_repository" {
  description = "GitHub repository being monitored"
  value       = "${var.github_owner}/${var.github_repo}"
}

# Security Configuration
output "encryption_at_rest_enabled" {
  description = "Whether encryption at rest is enabled"
  value       = var.enable_encryption_at_rest
}

output "encryption_in_transit_enabled" {
  description = "Whether encryption in transit is enabled"
  value       = var.enable_encryption_in_transit
}

# EventBridge Schedule Outputs
output "pipeline_metrics_schedule_expression" {
  description = "Schedule expression for pipeline metrics collection"
  value       = aws_cloudwatch_event_rule.pipeline_metrics_schedule.schedule_expression
}

output "performance_monitor_schedule_expression" {
  description = "Schedule expression for performance monitoring"
  value       = aws_cloudwatch_event_rule.performance_monitor_schedule.schedule_expression
}

# Dead Letter Queue Output
output "lambda_dlq_url" {
  description = "URL of the Lambda dead letter queue"
  value       = aws_sqs_queue.lambda_dlq.url
}

output "lambda_dlq_arn" {
  description = "ARN of the Lambda dead letter queue"
  value       = aws_sqs_queue.lambda_dlq.arn
}

# Summary Output for Quick Reference
output "monitoring_system_summary" {
  description = "Summary of the deployed CI/CD monitoring system"
  value = {
    project_name              = var.project_name
    environment              = var.environment
    repository               = "${var.github_owner}/${var.github_repo}"
    health_check_url         = "https://${aws_api_gateway_rest_api.monitoring_health.id}.execute-api.${data.aws_region.current.name}.amazonaws.com/${var.environment}/health"
    dashboard_count          = 3
    lambda_functions_count   = 5
    alarm_count             = 4
    circuit_breaker_enabled = var.enable_circuit_breaker
    cost_optimization_enabled = var.enable_cost_optimization
    auto_scaling_enabled    = var.auto_scaling_enabled
    grafana_enabled         = var.enable_grafana_dashboard
    encryption_enabled      = var.enable_encryption_at_rest
    data_retention_days     = var.data_retention_days
    log_retention_days      = var.log_retention_days
  }
}

# Instructions for Next Steps
output "setup_instructions" {
  description = "Next steps for completing the monitoring setup"
  value = {
    configure_sns_subscriptions = "Add email/SMS subscriptions to SNS topics: ${aws_sns_topic.critical_alerts.arn}, ${aws_sns_topic.warning_alerts.arn}, ${aws_sns_topic.cost_alerts.arn}"
    access_dashboards = "View dashboards at: ${local.dashboard_urls.pipeline_overview}"
    health_check = "Monitor system health at: https://${aws_api_gateway_rest_api.monitoring_health.id}.execute-api.${data.aws_region.current.name}.amazonaws.com/${var.environment}/health"
    grafana_setup = var.enable_grafana_dashboard ? "Complete Grafana setup at: ${aws_grafana_workspace.monitoring[0].endpoint}" : "Grafana is disabled. Set enable_grafana_dashboard=true to enable."
    github_integration = var.enable_github_oidc ? "Use GitHub Actions role ARN: ${aws_iam_role.github_actions_role[0].arn}" : "GitHub OIDC is disabled. Set enable_github_oidc=true to enable secure CI/CD integration."
  }
}