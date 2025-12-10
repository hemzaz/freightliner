# Development Environment CI/CD Monitoring Deployment
# This configuration deploys the comprehensive CI/CD monitoring system

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

  backend "s3" {
    # Backend configuration should be provided via backend config file or CLI
    # bucket = "your-terraform-state-bucket"
    # key    = "environments/dev/ci-monitoring.tfstate"
    # region = "us-west-2"
    # encrypt = true
    # dynamodb_table = "terraform-state-lock"
  }
}

# Provider Configuration
provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "freightliner"
      Environment = "dev"
      ManagedBy   = "terraform"
      Owner       = "devops-team"
      CostCenter  = "engineering"
    }
  }
}

provider "github" {
  token = var.github_token
  owner = var.github_owner
}

# Data Sources
data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# Local Variables for Development Environment
locals {
  environment = "dev"
  common_tags = {
    Project     = "freightliner"
    Environment = local.environment
    ManagedBy   = "terraform"
    Owner       = "devops-team"
    Repository  = "${var.github_owner}/${var.github_repo}"
  }

  # Development-specific configuration
  dev_config = {
    log_retention_days           = 7    # Shorter retention for dev
    data_retention_days         = 90   # 3 months for dev
    cost_alert_threshold_usd    = 50   # Lower threshold for dev
    enable_cost_optimization    = true
    enable_performance_baseline = true
    enable_circuit_breaker     = true
    enable_grafana_dashboard   = true
    auto_scaling_enabled       = false # Disabled in dev to save costs
  }
}

# CI/CD Monitoring Module
module "ci_monitoring" {
  source = "../../modules/ci-monitoring"

  # Basic Configuration
  project_name = "freightliner"
  environment  = local.environment
  vpc_id       = data.aws_vpc.default.id
  subnet_ids   = data.aws_subnets.default.ids

  # GitHub Configuration
  github_token = var.github_token
  github_owner = var.github_owner
  github_repo  = var.github_repo

  # Development Environment Overrides
  log_retention_days        = local.dev_config.log_retention_days
  data_retention_days       = local.dev_config.data_retention_days
  cost_alert_threshold_usd  = local.dev_config.cost_alert_threshold_usd

  # Feature Flags for Development
  enable_cost_optimization    = local.dev_config.enable_cost_optimization
  enable_performance_baseline = local.dev_config.enable_performance_baseline
  enable_circuit_breaker     = local.dev_config.enable_circuit_breaker
  enable_grafana_dashboard   = local.dev_config.enable_grafana_dashboard
  auto_scaling_enabled       = local.dev_config.auto_scaling_enabled
  enable_github_oidc         = true # Enable for secure CI/CD integration

  # Pipeline Thresholds (Relaxed for Development)
  pipeline_success_rate_threshold      = 0.85  # 85% for dev (vs 95% for prod)
  pipeline_duration_threshold_minutes  = 45    # 45 minutes for dev
  performance_regression_threshold     = 0.3   # 30% regression threshold

  # Circuit Breaker Configuration
  circuit_breaker_failure_threshold = 3  # More lenient in dev
  circuit_breaker_timeout_seconds   = 30 # Shorter timeout for faster recovery

  # Cost Optimization
  cost_optimization_schedule = "cron(0 10 * * ? *)"  # Daily at 10 AM UTC

  # Auto Scaling (Disabled in dev but configuration provided)
  scaling_target_utilization = 60  # Lower target for dev
  min_capacity              = 1
  max_capacity              = 3    # Lower max for dev

  # Security Configuration
  enable_encryption_at_rest    = true
  enable_encryption_in_transit = true

  # Alert Configuration
  alert_email_recipients = var.alert_email_recipients
  slack_webhook_url     = var.slack_webhook_url

  # Grafana Configuration
  grafana_admin_password = var.grafana_admin_password

  # Tags
  tags = local.common_tags
}

# SNS Topic Subscriptions for Development Alerts
resource "aws_sns_topic_subscription" "dev_critical_alerts_email" {
  count = length(var.alert_email_recipients)

  topic_arn = module.ci_monitoring.critical_alerts_topic_arn
  protocol  = "email"
  endpoint  = var.alert_email_recipients[count.index]
}

resource "aws_sns_topic_subscription" "dev_warning_alerts_email" {
  count = length(var.alert_email_recipients)

  topic_arn = module.ci_monitoring.warning_alerts_topic_arn
  protocol  = "email"
  endpoint  = var.alert_email_recipients[count.index]
}

resource "aws_sns_topic_subscription" "dev_cost_alerts_email" {
  count = length(var.alert_email_recipients)

  topic_arn = module.ci_monitoring.cost_alerts_topic_arn
  protocol  = "email"
  endpoint  = var.alert_email_recipients[count.index]
}

# Slack Integration (if webhook URL is provided)
resource "aws_sns_topic_subscription" "dev_critical_alerts_slack" {
  count = var.slack_webhook_url != "" ? 1 : 0

  topic_arn = module.ci_monitoring.critical_alerts_topic_arn
  protocol  = "https"
  endpoint  = var.slack_webhook_url
}

# CloudWatch Log Insights Queries for Development Debugging
resource "aws_cloudwatch_query_definition" "dev_error_analysis" {
  name = "freightliner-dev-error-analysis"

  log_group_names = [
    module.ci_monitoring.log_group_pipeline_metrics,
    module.ci_monitoring.log_group_performance_metrics
  ]

  query_string = <<-EOT
fields @timestamp, @message
| filter @message like /ERROR/
| sort @timestamp desc
| limit 100
EOT
}

resource "aws_cloudwatch_query_definition" "dev_performance_analysis" {
  name = "freightliner-dev-performance-analysis"

  log_group_names = [
    module.ci_monitoring.log_group_performance_metrics
  ]

  query_string = <<-EOT
fields @timestamp, @message
| filter @message like /regression/
| sort @timestamp desc
| limit 50
EOT
}

# Development-specific alarms with different thresholds
resource "aws_cloudwatch_metric_alarm" "dev_lambda_cost_high" {
  alarm_name          = "freightliner-dev-lambda-cost-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "ServiceCostUSD"
  namespace           = "CI-CD/Cost"
  period              = "3600"
  statistic           = "Maximum"
  threshold           = "20"  # $20 for Lambda costs in dev
  alarm_description   = "Lambda costs are high in development environment"
  alarm_actions       = [module.ci_monitoring.cost_alerts_topic_arn]

  dimensions = {
    Service = "AWS Lambda"
  }

  tags = local.common_tags
}

# Create a simple monitoring dashboard for development team
resource "aws_cloudwatch_dashboard" "dev_team_dashboard" {
  dashboard_name = "freightliner-dev-team-overview"

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
            ["CI-CD/Pipeline", "PipelineSuccessRate", "Repository", "${var.github_owner}/${var.github_repo}"],
            [".", "AveragePipelineDuration", ".", "."]
          ]
          view    = "timeSeries"
          stacked = false
          region  = data.aws_region.current.name
          title   = "Dev Pipeline Health (Last 24h)"
          period  = 300
          stat    = "Average"
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 0
        width  = 12
        height = 6

        properties = {
          metrics = [
            ["CI-CD/Cost", "TotalCostUSD"],
            [".", "OptimizationPotentialUSD"]
          ]
          view    = "timeSeries"
          stacked = false
          region  = data.aws_region.current.name
          title   = "Dev Environment Costs"
          period  = 3600
        }
      }
    ]
  })
}

# Development Environment Health Check Script
resource "local_file" "dev_health_check_script" {
  content = templatefile("${path.module}/scripts/dev-health-check.sh.tpl", {
    health_check_url           = module.ci_monitoring.health_check_api_url
    pipeline_dashboard_url     = module.ci_monitoring.dashboard_urls.pipeline_overview
    performance_dashboard_url  = module.ci_monitoring.dashboard_urls.performance_analysis
    cost_dashboard_url        = module.ci_monitoring.dashboard_urls.cost_analysis
    grafana_url               = module.ci_monitoring.grafana_workspace_endpoint
  })
  filename = "${path.module}/scripts/dev-health-check.sh"

  file_permission = "0755"
}

# Output Development Environment Information
output "dev_monitoring_setup" {
  description = "Development environment monitoring setup information"
  value = {
    environment           = local.environment
    health_check_url     = module.ci_monitoring.health_check_api_url
    dashboard_urls       = module.ci_monitoring.dashboard_urls
    grafana_workspace    = module.ci_monitoring.grafana_workspace_endpoint
    alert_topics = {
      critical = module.ci_monitoring.critical_alerts_topic_arn
      warning  = module.ci_monitoring.warning_alerts_topic_arn
      cost     = module.ci_monitoring.cost_alerts_topic_arn
    }
    lambda_functions = {
      metrics_collector  = module.ci_monitoring.pipeline_metrics_collector_function_name
      performance_monitor = module.ci_monitoring.performance_monitor_function_name
      cost_optimizer     = module.ci_monitoring.cost_optimizer_function_name
      recovery_manager   = module.ci_monitoring.recovery_manager_function_name
      health_check       = module.ci_monitoring.health_check_function_name
    }
    storage = {
      s3_bucket     = module.ci_monitoring.s3_bucket_name
      dynamodb_table = module.ci_monitoring.dynamodb_table_name
    }
    security = {
      kms_key_id = module.ci_monitoring.kms_key_id
      github_role_arn = module.ci_monitoring.github_actions_role_arn
    }
    configuration = {
      cost_threshold_usd        = local.dev_config.cost_alert_threshold_usd
      data_retention_days       = local.dev_config.data_retention_days
      circuit_breaker_enabled   = local.dev_config.enable_circuit_breaker
      auto_scaling_enabled      = local.dev_config.auto_scaling_enabled
      grafana_enabled          = local.dev_config.enable_grafana_dashboard
    }
    next_steps = [
      "Subscribe to SNS topics for alerts",
      "Access dashboards for monitoring",
      "Configure Grafana workspace if enabled",
      "Run health check script: ./scripts/dev-health-check.sh",
      "Review cost optimization recommendations"
    ]
  }
}

# Summary for easy reference
output "quick_access_urls" {
  description = "Quick access URLs for development monitoring"
  value = {
    health_check = module.ci_monitoring.health_check_api_url
    main_dashboard = module.ci_monitoring.dashboard_urls.pipeline_overview
    team_dashboard = "https://${data.aws_region.current.name}.console.aws.amazon.com/cloudwatch/home?region=${data.aws_region.current.name}#dashboards/dashboard/${aws_cloudwatch_dashboard.dev_team_dashboard.dashboard_name}"
    grafana = module.ci_monitoring.grafana_workspace_endpoint
  }
}