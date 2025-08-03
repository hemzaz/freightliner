# Variables for Development Environment CI/CD Monitoring

variable "aws_region" {
  description = "AWS region for resource deployment"
  type        = string
  default     = "us-west-2"
  validation {
    condition     = can(regex("^[a-z]{2}-[a-z]+-[0-9]+$", var.aws_region))
    error_message = "AWS region must be a valid region identifier."
  }
}

# GitHub Configuration
variable "github_token" {
  description = "GitHub personal access token for API access"
  type        = string
  sensitive   = true
  validation {
    condition     = can(regex("^(ghp_|gho_|ghu_|ghs_|ghr_)[A-Za-z0-9]{36}$", var.github_token))
    error_message = "GitHub token must be a valid personal access token format."
  }
}

variable "github_owner" {
  description = "GitHub repository owner/organization"
  type        = string
  default     = "hemzaz"
  validation {
    condition     = length(var.github_owner) > 0
    error_message = "GitHub owner cannot be empty."
  }
}

variable "github_repo" {
  description = "GitHub repository name"
  type        = string
  default     = "freightliner"
  validation {
    condition     = length(var.github_repo) > 0
    error_message = "GitHub repository name cannot be empty."
  }
}

# Alert Configuration
variable "alert_email_recipients" {
  description = "List of email addresses to receive monitoring alerts"
  type        = list(string)
  default     = []
  validation {
    condition = alltrue([
      for email in var.alert_email_recipients : can(regex("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", email))
    ])
    error_message = "All email addresses must be valid."
  }
}

variable "slack_webhook_url" {
  description = "Slack webhook URL for alert notifications (optional)"
  type        = string
  default     = ""
  sensitive   = true
  validation {
    condition     = var.slack_webhook_url == "" || can(regex("^https://hooks.slack.com/", var.slack_webhook_url))
    error_message = "Slack webhook URL must be empty or a valid Slack webhook URL."
  }
}

# Grafana Configuration
variable "grafana_admin_password" {
  description = "Admin password for Grafana workspace (leave empty for auto-generated)"
  type        = string
  default     = ""
  sensitive   = true
  validation {
    condition     = var.grafana_admin_password == "" || length(var.grafana_admin_password) >= 8
    error_message = "Grafana admin password must be empty or at least 8 characters long."
  }
}

# Development Environment Overrides
variable "enable_debug_logging" {
  description = "Enable debug logging for development environment"
  type        = bool
  default     = true
}

variable "dev_cost_limit_usd" {
  description = "Cost limit for development environment (USD)"
  type        = number
  default     = 100
  validation {
    condition     = var.dev_cost_limit_usd > 0 && var.dev_cost_limit_usd <= 1000
    error_message = "Development cost limit must be between $1 and $1000."
  }
}

variable "dev_data_retention_days" {
  description = "Data retention period for development environment (days)"
  type        = number
  default     = 30
  validation {
    condition     = var.dev_data_retention_days >= 7 && var.dev_data_retention_days <= 365
    error_message = "Development data retention must be between 7 and 365 days."
  }
}

# Feature Toggles for Development
variable "enable_experimental_features" {
  description = "Enable experimental monitoring features in development"
  type        = bool
  default     = true
}

variable "enable_detailed_metrics" {
  description = "Enable detailed metrics collection (may increase costs)"
  type        = bool
  default     = true
}

variable "enable_canary_deployments" {
  description = "Enable canary deployment monitoring"
  type        = bool
  default     = false
}

# Development Team Configuration
variable "team_notification_channels" {
  description = "Map of team notification channels"
  type = map(object({
    email           = optional(list(string), [])
    slack_webhook   = optional(string, "")
    severity_filter = optional(list(string), ["critical", "warning"])
  }))
  default = {
    devops = {
      email           = []
      slack_webhook   = ""
      severity_filter = ["critical", "warning", "info"]
    }
    backend = {
      email           = []
      slack_webhook   = ""
      severity_filter = ["critical", "warning"]
    }
    frontend = {
      email           = []
      slack_webhook   = ""
      severity_filter = ["critical"]
    }
  }
}

# Testing Configuration
variable "enable_load_testing_monitoring" {
  description = "Enable monitoring for load testing scenarios"
  type        = bool
  default     = true
}

variable "load_test_thresholds" {
  description = "Performance thresholds during load testing"
  type = object({
    max_response_time_ms = optional(number, 5000)
    max_error_rate       = optional(number, 0.05)
    max_cpu_utilization  = optional(number, 0.8)
  })
  default = {
    max_response_time_ms = 5000  # 5 seconds
    max_error_rate       = 0.05  # 5%
    max_cpu_utilization  = 0.8   # 80%
  }
}

# Integration Testing
variable "enable_integration_test_monitoring" {
  description = "Enable specialized monitoring for integration tests"
  type        = bool
  default     = true
}

variable "integration_test_environments" {
  description = "List of integration test environments to monitor"
  type        = list(string)
  default     = ["dev", "staging", "e2e"]
}

# Security Configuration for Development
variable "dev_security_settings" {
  description = "Security settings for development environment"
  type = object({
    enable_public_dashboards = optional(bool, false)
    enable_api_authentication = optional(bool, true)
    allowed_ip_ranges         = optional(list(string), [])
    enable_audit_logging      = optional(bool, true)
  })
  default = {
    enable_public_dashboards = false
    enable_api_authentication = true
    allowed_ip_ranges         = []
    enable_audit_logging      = true
  }
}

# Resource Sizing for Development
variable "dev_resource_sizing" {
  description = "Resource sizing for development environment"
  type = object({
    lambda_memory_mb     = optional(number, 256)
    lambda_timeout_sec   = optional(number, 300)
    dynamodb_read_capacity  = optional(number, 5)
    dynamodb_write_capacity = optional(number, 5)
    s3_lifecycle_days    = optional(number, 30)
  })
  default = {
    lambda_memory_mb     = 256   # Smaller memory for dev
    lambda_timeout_sec   = 300   # 5 minutes
    dynamodb_read_capacity  = 5     # Lower capacity for dev
    dynamodb_write_capacity = 5     # Lower capacity for dev
    s3_lifecycle_days    = 30    # Faster cleanup in dev
  }
}

# Development Workflow Configuration
variable "dev_workflow_settings" {
  description = "Development workflow specific settings"
  type = object({
    enable_pr_monitoring       = optional(bool, true)
    enable_branch_monitoring   = optional(bool, true)
    monitored_branches         = optional(list(string), ["main", "develop", "staging"])
    enable_commit_metrics      = optional(bool, true)
    enable_developer_metrics   = optional(bool, false)
  })
  default = {
    enable_pr_monitoring       = true
    enable_branch_monitoring   = true
    monitored_branches         = ["main", "develop", "staging"]
    enable_commit_metrics      = true
    enable_developer_metrics   = false  # Privacy consideration
  }
}

# Backup and Recovery Configuration
variable "dev_backup_settings" {
  description = "Backup and recovery settings for development"
  type = object({
    enable_automated_backups = optional(bool, false)
    backup_retention_days    = optional(number, 7)
    enable_cross_region_backup = optional(bool, false)
    backup_schedule          = optional(string, "cron(0 2 * * ? *)")
  })
  default = {
    enable_automated_backups = false  # Minimal backups in dev
    backup_retention_days    = 7
    enable_cross_region_backup = false
    backup_schedule          = "cron(0 2 * * ? *)"  # 2 AM daily
  }
}

# Compliance and Governance
variable "dev_compliance_settings" {
  description = "Compliance and governance settings for development"
  type = object({
    enable_compliance_monitoring = optional(bool, false)
    required_tags               = optional(list(string), ["Environment", "Project", "Owner"])
    enable_resource_inventory   = optional(bool, true)
    enable_cost_allocation_tags = optional(bool, true)
  })
  default = {
    enable_compliance_monitoring = false  # Simplified for dev
    required_tags               = ["Environment", "Project", "Owner"]
    enable_resource_inventory   = true
    enable_cost_allocation_tags = true
  }
}

# Local Development Support
variable "local_development_support" {
  description = "Settings to support local development and testing"
  type = object({
    enable_local_endpoints = optional(bool, true)
    local_test_data_size   = optional(string, "small")
    enable_mock_services   = optional(bool, true)
    debug_port_range       = optional(object({
      start = number
      end   = number
    }), { start = 9000, end = 9099 })
  })
  default = {
    enable_local_endpoints = true
    local_test_data_size   = "small"
    enable_mock_services   = true
    debug_port_range = {
      start = 9000
      end   = 9099
    }
  }
}