# Monitoring Dashboards and Visualization
# CloudWatch dashboards and Grafana configuration for comprehensive monitoring

# CloudWatch Dashboard for CI/CD Pipeline Metrics
resource "aws_cloudwatch_dashboard" "ci_cd_pipeline_overview" {
  dashboard_name = "${local.name_prefix}-pipeline-overview"

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
            [".", "PipelineFailureRate", ".", "."],
            [".", "AveragePipelineDuration", ".", "."]
          ]
          view    = "timeSeries"
          stacked = false
          region  = data.aws_region.current.name
          title   = "Pipeline Health Overview"
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
            ["CI-CD/Pipeline", "TotalPipelineRuns", "Repository", "${var.github_owner}/${var.github_repo}"],
            [".", "LongRunningPipelines", ".", "."]
          ]
          view    = "timeSeries"
          stacked = false
          region  = data.aws_region.current.name
          title   = "Pipeline Volume Metrics"
          period  = 300
          stat    = "Sum"
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 6
        width  = 8
        height = 6

        properties = {
          metrics = [
            ["CI-CD/Performance", "total_duration_current", "Repository", "${var.github_owner}/${var.github_repo}"],
            [".", "total_duration_average", ".", "."],
            [".", "total_duration_baseline_deviation", ".", ".", { "yAxis": "right" }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = data.aws_region.current.name
          title   = "Performance Metrics"
          period  = 300
          yAxis = {
            left = {
              min = 0
            }
            right = {
              min = -50
              max = 100
            }
          }
        }
      },
      {
        type   = "metric"
        x      = 8
        y      = 6
        width  = 8
        height = 6

        properties = {
          metrics = [
            ["CI-CD/Cost", "TotalCostUSD"],
            [".", "OptimizationPotentialUSD"],
            [".", "PotentialSavingsPercent", { "yAxis": "right" }]
          ]
          view    = "timeSeries"
          stacked = false
          region  = data.aws_region.current.name
          title   = "Cost Optimization"
          period  = 3600
          yAxis = {
            left = {
              min = 0
            }
            right = {
              min = 0
              max = 100
            }
          }
        }
      },
      {
        type   = "metric"
        x      = 16
        y      = 6
        width  = 8
        height = 6

        properties = {
          metrics = [
            ["AWS/Lambda", "Invocations", "FunctionName", "${local.name_prefix}-pipeline-metrics-collector"],
            [".", "Errors", ".", "."],
            [".", "Duration", ".", "."]
          ]
          view    = "timeSeries"
          stacked = false
          region  = data.aws_region.current.name
          title   = "Lambda Performance"
          period  = 300
        }
      },
      {
        type   = "log"
        x      = 0
        y      = 12
        width  = 24
        height = 6

        properties = {
          query   = "SOURCE '${aws_cloudwatch_log_group.ci_pipeline_metrics.name}'\n| fields @timestamp, @message\n| filter @message like /ERROR/\n| sort @timestamp desc\n| limit 20"
          region  = data.aws_region.current.name
          title   = "Recent Errors"
          view    = "table"
        }
      }
    ]
  })
}

# CloudWatch Dashboard for Performance Analysis
resource "aws_cloudwatch_dashboard" "performance_analysis" {
  dashboard_name = "${local.name_prefix}-performance-analysis"

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
            ["CI-CD/Performance", "queue_time_current", "Repository", "${var.github_owner}/${var.github_repo}"],
            [".", "setup_time_current", ".", "."],
            [".", "test_time_current", ".", "."],
            [".", "build_time_current", ".", "."]
          ]
          view    = "timeSeries"
          stacked = true
          region  = data.aws_region.current.name
          title   = "Pipeline Stage Durations"
          period  = 300
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
            ["CI-CD/Performance", "success_rate_rolling_current", "Repository", "${var.github_owner}/${var.github_repo}"],
            [".", "concurrent_jobs_current", ".", "."]
          ]
          view    = "timeSeries"
          stacked = false
          region  = data.aws_region.current.name
          title   = "Pipeline Efficiency"
          period  = 300
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 6
        width  = 24
        height = 6

        properties = {
          metrics = [
            ["CI-CD/Performance", "total_duration_baseline_deviation", "Repository", "${var.github_owner}/${var.github_repo}"],
            [".", "queue_time_baseline_deviation", ".", "."],
            [".", "test_time_baseline_deviation", ".", "."],
            [".", "build_time_baseline_deviation", ".", "."]
          ]
          view    = "timeSeries"
          stacked = false
          region  = data.aws_region.current.name
          title   = "Performance Baseline Deviations (%)"
          period  = 300
          annotations = {
            horizontal = [
              {
                label = "Regression Threshold"
                value = 20
              }
            ]
          }
        }
      }
    ]
  })
}

# CloudWatch Dashboard for Cost Analysis
resource "aws_cloudwatch_dashboard" "cost_analysis" {
  dashboard_name = "${local.name_prefix}-cost-analysis"

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
            ["CI-CD/Cost", "ServiceCostUSD", "Service", "AWS Lambda"],
            [".", ".", ".", "Amazon Elastic Compute Cloud - Compute"],
            [".", ".", ".", "Amazon CloudWatch"],
            [".", ".", ".", "Amazon Simple Storage Service"],
            [".", ".", ".", "Amazon DynamoDB"]
          ]
          view    = "timeSeries"
          stacked = true
          region  = data.aws_region.current.name
          title   = "Service Costs Breakdown"
          period  = 3600
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
            ["CI-CD/Cost", "ServiceCostChangePercent", "Service", "AWS Lambda"],
            [".", ".", ".", "Amazon Elastic Compute Cloud - Compute"],
            [".", ".", ".", "Amazon CloudWatch"],
            [".", ".", ".", "Amazon Simple Storage Service"]
          ]
          view    = "timeSeries"
          stacked = false
          region  = data.aws_region.current.name
          title   = "Cost Change Percentages"
          period  = 3600
          yAxis = {
            left = {
              min = -50
              max = 100
            }
          }
          annotations = {
            horizontal = [
              {
                label = "Alert Threshold"
                value = 25
              }
            ]
          }
        }
      },
      {
        type   = "number"
        x      = 0
        y      = 6
        width  = 6
        height = 3

        properties = {
          metrics = [
            ["CI-CD/Cost", "TotalCostUSD"]
          ]
          view    = "singleValue"
          region  = data.aws_region.current.name
          title   = "Total Monthly Cost (USD)"
          period  = 86400
          stat    = "Average"
        }
      },
      {
        type   = "number"
        x      = 6
        y      = 6
        width  = 6
        height = 3

        properties = {
          metrics = [
            ["CI-CD/Cost", "OptimizationPotentialUSD"]
          ]
          view    = "singleValue"
          region  = data.aws_region.current.name
          title   = "Optimization Potential (USD)"
          period  = 86400
          stat    = "Average"
        }
      },
      {
        type   = "number"
        x      = 12
        y      = 6
        width  = 6
        height = 3

        properties = {
          metrics = [
            ["CI-CD/Cost", "PotentialSavingsPercent"]
          ]
          view    = "singleValue"
          region  = data.aws_region.current.name
          title   = "Potential Savings (%)"
          period  = 86400
          stat    = "Average"
        }
      }
    ]
  })
}

# Grafana Dashboard Configuration (if enabled)
resource "aws_grafana_workspace" "monitoring" {
  count = var.enable_grafana_dashboard ? 1 : 0

  name                     = "${local.name_prefix}-monitoring"
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["AWS_SSO"]
  permission_type          = "SERVICE_MANAGED"
  
  data_sources = ["CLOUDWATCH", "PROMETHEUS"]

  notification_destinations = ["SNS"]

  workspace_notification_destinations = [
    aws_sns_topic.critical_alerts.arn
  ]

  tags = local.common_tags
}

# Grafana dashboard for comprehensive monitoring
resource "aws_grafana_dashboard" "comprehensive_monitoring" {
  count = var.enable_grafana_dashboard ? 1 : 0

  workspace_id = aws_grafana_workspace.monitoring[0].id
  
  config_json = jsonencode({
    dashboard = {
      id       = null
      title    = "CI/CD Comprehensive Monitoring"
      tags     = ["ci-cd", "monitoring", "freightliner"]
      timezone = "browser"
      panels = [
        {
          id          = 1
          title       = "Pipeline Success Rate"
          type        = "stat"
          gridPos     = { h = 8, w = 12, x = 0, y = 0 }
          
          targets = [
            {
              expr         = "avg(pipeline_success_rate)"
              legendFormat = "Success Rate"
            }
          ]
          
          fieldConfig = {
            defaults = {
              color = {
                mode = "thresholds"
              }
              thresholds = {
                steps = [
                  { color = "red", value = 0 },
                  { color = "yellow", value = 0.8 },
                  { color = "green", value = 0.95 }
                ]
              }
            }
          }
        },
        {
          id          = 2
          title       = "Pipeline Duration Trends"
          type        = "timeseries"
          gridPos     = { h = 8, w = 12, x = 12, y = 0 }
          
          targets = [
            {
              expr         = "avg(pipeline_duration_minutes)"
              legendFormat = "Average Duration"
            },
            {
              expr         = "max(pipeline_duration_minutes)"
              legendFormat = "Max Duration"
            }
          ]
        },
        {
          id          = 3
          title       = "Cost Analysis"
          type        = "barchart"
          gridPos     = { h = 8, w = 24, x = 0, y = 8 }
          
          targets = [
            {
              expr         = "sum by (service) (service_cost_usd)"
              legendFormat = "{{service}}"
            }
          ]
        },
        {
          id          = 4
          title       = "Performance Regression Detection"
          type        = "timeseries"
          gridPos     = { h = 8, w = 24, x = 0, y = 16 }
          
          targets = [
            {
              expr         = "baseline_deviation_percent"
              legendFormat = "Performance Deviation %"
            }
          ]
          
          alert = {
            conditions = [
              {
                query = {
                  queryType = ""
                  refId     = "A"
                }
                reducer = {
                  type   = "last"
                  params = []
                }
                evaluator = {
                  params = [20]
                  type   = "gt"
                }
              }
            ]
            executionErrorState = "alerting"
            noDataState        = "no_data"
            frequency          = "10s"
            handler            = 1
            name               = "Performance Regression Alert"
            message            = "Performance regression detected: {{$value}}% deviation from baseline"
          }
        }
      ]
      
      time = {
        from = "now-24h"
        to   = "now"
      }
      
      refresh = "30s"
      
      templating = {
        list = [
          {
            name  = "repository"
            type  = "constant"
            query = "${var.github_owner}/${var.github_repo}"
          },
          {
            name       = "environment"
            type       = "query"
            query      = "label_values(environment)"
            refresh    = 1
            multi      = false
            includeAll = false
          }
        ]
      }
    }
  })
}

# CloudWatch Composite Alarms for advanced alerting
resource "aws_cloudwatch_composite_alarm" "ci_cd_system_health" {
  alarm_name        = "${local.name_prefix}-system-health"
  alarm_description = "Composite alarm for overall CI/CD system health"

  alarm_rule = join(" OR ", [
    "ALARM('${aws_cloudwatch_metric_alarm.pipeline_failure_rate_high.alarm_name}')",
    "ALARM('${aws_cloudwatch_metric_alarm.pipeline_duration_high.alarm_name}')",
    "ALARM('${aws_cloudwatch_metric_alarm.performance_regression_critical.alarm_name}')"
  ])

  actions_enabled = true
  alarm_actions = [
    aws_sns_topic.critical_alerts.arn
  ]

  ok_actions = [
    aws_sns_topic.warning_alerts.arn
  ]

  tags = local.common_tags
}

# Custom Metric Filters for Log-based Alerting
resource "aws_cloudwatch_log_metric_filter" "error_rate" {
  name           = "${local.name_prefix}-error-rate"
  log_group_name = aws_cloudwatch_log_group.ci_pipeline_metrics.name
  pattern        = "[timestamp, request_id, ERROR, ...]"

  metric_transformation {
    name      = "ErrorRate"
    namespace = "CI-CD/Logs"
    value     = "1"
  }
}

resource "aws_cloudwatch_metric_alarm" "high_error_rate" {
  alarm_name          = "${local.name_prefix}-high-error-rate"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "ErrorRate"
  namespace           = "CI-CD/Logs"
  period              = "300"
  statistic           = "Sum"
  threshold           = "10"
  alarm_description   = "This metric monitors error rate in logs"
  alarm_actions       = [aws_sns_topic.warning_alerts.arn]

  tags = local.common_tags
}

# Performance Anomaly Detection
resource "aws_cloudwatch_anomaly_detector" "pipeline_duration_anomaly" {
  metric_math_anomaly_detector {
    metric_data_queries {
      id = "m1"
      metric_stat {
        metric {
          metric_name = "AveragePipelineDuration"
          namespace   = "CI-CD/Pipeline"
          dimensions = {
            Repository = "${var.github_owner}/${var.github_repo}"
          }
        }
        period = 300
        stat   = "Average"
      }
    }
  }
}

resource "aws_cloudwatch_metric_alarm" "pipeline_duration_anomaly" {
  alarm_name          = "${local.name_prefix}-pipeline-duration-anomaly"
  comparison_operator = "LessThanLowerOrGreaterThanUpperThreshold"
  evaluation_periods  = "2"
  threshold_metric_id = "ad1"
  alarm_description   = "This metric monitors pipeline duration anomalies"
  alarm_actions       = [aws_sns_topic.warning_alerts.arn]

  metric_query {
    id = "ad1"
    anomaly_detector {
      metric_math_anomaly_detector {
        metric_data_queries {
          id = "m1"
          metric_stat {
            metric {
              metric_name = "AveragePipelineDuration"
              namespace   = "CI-CD/Pipeline"
              dimensions = {
                Repository = "${var.github_owner}/${var.github_repo}"
              }
            }
            period = 300
            stat   = "Average"
          }
        }
      }
    }
  }

  tags = local.common_tags
}

# Dashboard permissions and sharing
resource "aws_grafana_workspace_api_key" "dashboard_api_key" {
  count = var.enable_grafana_dashboard ? 1 : 0

  key_name        = "${local.name_prefix}-dashboard-api-key"
  key_role        = "ADMIN"
  seconds_to_live = 3600 * 24 * 365  # 1 year
  workspace_id    = aws_grafana_workspace.monitoring[0].id
}

# Export dashboard URLs
locals {
  dashboard_urls = {
    pipeline_overview    = "https://${data.aws_region.current.name}.console.aws.amazon.com/cloudwatch/home?region=${data.aws_region.current.name}#dashboards/dashboard/${aws_cloudwatch_dashboard.ci_cd_pipeline_overview.dashboard_name}"
    performance_analysis = "https://${data.aws_region.current.name}.console.aws.amazon.com/cloudwatch/home?region=${data.aws_region.current.name}#dashboards/dashboard/${aws_cloudwatch_dashboard.performance_analysis.dashboard_name}"
    cost_analysis       = "https://${data.aws_region.current.name}.console.aws.amazon.com/cloudwatch/home?region=${data.aws_region.current.name}#dashboards/dashboard/${aws_cloudwatch_dashboard.cost_analysis.dashboard_name}"
    grafana_workspace   = var.enable_grafana_dashboard ? aws_grafana_workspace.monitoring[0].endpoint : null
  }
}