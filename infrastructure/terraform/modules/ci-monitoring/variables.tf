# Variables for CI/CD Monitoring Module

variable "project_name" {
  description = "Name of the project"
  type        = string
  validation {
    condition     = can(regex("^[a-z0-9-]+$", var.project_name))
    error_message = "Project name must contain only lowercase letters, numbers, and hyphens."
  }
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be one of: dev, staging, prod."
  }
}

variable "vpc_id" {
  description = "VPC ID where monitoring resources will be deployed"
  type        = string
  validation {
    condition     = can(regex("^vpc-[a-z0-9]+$", var.vpc_id))
    error_message = "VPC ID must be a valid AWS VPC identifier."
  }
}

variable "subnet_ids" {
  description = "List of subnet IDs for Lambda functions"
  type        = list(string)
  validation {
    condition     = length(var.subnet_ids) >= 2
    error_message = "At least 2 subnet IDs must be provided for high availability."
  }
}

variable "github_token" {
  description = "GitHub token for API access"
  type        = string
  sensitive   = true
  validation {
    condition     = can(regex("^(ghp_|gho_|ghu_|ghs_|ghr_)[A-Za-z0-9]{36}$", var.github_token))
    error_message = "GitHub token must be a valid personal access token format."
  }
}

variable "github_owner" {
  description = "GitHub repository owner"
  type        = string
  validation {
    condition     = length(var.github_owner) > 0
    error_message = "GitHub owner cannot be empty."
  }
}

variable "github_repo" {
  description = "GitHub repository name"
  type        = string
  validation {
    condition     = length(var.github_repo) > 0
    error_message = "GitHub repository name cannot be empty."
  }
}

variable "tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}

variable "log_retention_days" {
  description = "Number of days to retain CloudWatch logs"
  type        = number
  default     = 30
  validation {
    condition     = contains([1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 3653], var.log_retention_days)
    error_message = "Log retention days must be one of the allowed CloudWatch values."
  }
}

variable "data_retention_days" {
  description = "Number of days to retain monitoring data in S3"
  type        = number
  default     = 365
  validation {
    condition     = var.data_retention_days >= 30 && var.data_retention_days <= 2555
    error_message = "Data retention days must be between 30 and 2555 days (7 years)."
  }
}

variable "kms_deletion_window" {
  description = "KMS key deletion window in days"
  type        = number
  default     = 7
  validation {
    condition     = var.kms_deletion_window >= 7 && var.kms_deletion_window <= 30
    error_message = "KMS deletion window must be between 7 and 30 days."
  }
}

variable "log_level" {
  description = "Log level for Lambda functions"
  type        = string
  default     = "INFO"
  validation {
    condition     = contains(["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"], var.log_level)
    error_message = "Log level must be one of: DEBUG, INFO, WARNING, ERROR, CRITICAL."
  }
}

# Alert configuration variables
variable "pipeline_success_rate_threshold" {
  description = "Minimum acceptable pipeline success rate (0.0-1.0)"
  type        = number
  default     = 0.95
  validation {
    condition     = var.pipeline_success_rate_threshold >= 0.0 && var.pipeline_success_rate_threshold <= 1.0
    error_message = "Pipeline success rate threshold must be between 0.0 and 1.0."
  }
}

variable "pipeline_duration_threshold_minutes" {
  description = "Maximum acceptable pipeline duration in minutes"
  type        = number
  default     = 30
  validation {
    condition     = var.pipeline_duration_threshold_minutes > 0 && var.pipeline_duration_threshold_minutes <= 180
    error_message = "Pipeline duration threshold must be between 1 and 180 minutes."
  }
}

variable "cost_alert_threshold_usd" {
  description = "Monthly cost threshold for alerts (USD)"
  type        = number
  default     = 100
  validation {
    condition     = var.cost_alert_threshold_usd > 0
    error_message = "Cost alert threshold must be greater than 0."
  }
}

# Notification configuration
variable "alert_email_recipients" {
  description = "List of email addresses for alert notifications"
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
  description = "Slack webhook URL for notifications"
  type        = string
  default     = ""
  sensitive   = true
  validation {
    condition     = var.slack_webhook_url == "" || can(regex("^https://hooks.slack.com/", var.slack_webhook_url))
    error_message = "Slack webhook URL must be empty or a valid Slack webhook URL."
  }
}

# Dashboard configuration
variable "enable_grafana_dashboard" {
  description = "Enable Grafana dashboard deployment"
  type        = bool
  default     = true
}

variable "grafana_admin_password" {
  description = "Admin password for Grafana"
  type        = string
  default     = ""
  sensitive   = true
  validation {
    condition     = var.grafana_admin_password == "" || length(var.grafana_admin_password) >= 8
    error_message = "Grafana admin password must be empty or at least 8 characters long."
  }
}

# Performance monitoring configuration
variable "enable_performance_baseline" {
  description = "Enable performance baseline tracking"
  type        = bool
  default     = true
}

variable "performance_regression_threshold" {
  description = "Performance regression threshold percentage (0.0-1.0)"
  type        = number
  default     = 0.2
  validation {
    condition     = var.performance_regression_threshold >= 0.0 && var.performance_regression_threshold <= 1.0
    error_message = "Performance regression threshold must be between 0.0 and 1.0."
  }
}

# Circuit breaker configuration
variable "enable_circuit_breaker" {
  description = "Enable circuit breaker for external dependencies"
  type        = bool
  default     = true
}

variable "circuit_breaker_failure_threshold" {
  description = "Number of failures before circuit breaker opens"
  type        = number
  default     = 5
  validation {
    condition     = var.circuit_breaker_failure_threshold > 0 && var.circuit_breaker_failure_threshold <= 20
    error_message = "Circuit breaker failure threshold must be between 1 and 20."
  }
}

variable "circuit_breaker_timeout_seconds" {
  description = "Circuit breaker timeout in seconds"
  type        = number
  default     = 60
  validation {
    condition     = var.circuit_breaker_timeout_seconds >= 10 && var.circuit_breaker_timeout_seconds <= 300
    error_message = "Circuit breaker timeout must be between 10 and 300 seconds."
  }
}

# Cost optimization configuration
variable "enable_cost_optimization" {
  description = "Enable automated cost optimization"
  type        = bool
  default     = true
}

variable "cost_optimization_schedule" {
  description = "Cron expression for cost optimization checks"
  type        = string
  default     = "cron(0 9 * * ? *)"  # Daily at 9 AM UTC
  validation {
    condition     = can(regex("^cron\\(", var.cost_optimization_schedule))
    error_message = "Cost optimization schedule must be a valid cron expression."
  }
}

# Resource scaling configuration
variable "auto_scaling_enabled" {
  description = "Enable auto-scaling for CI/CD resources"
  type        = bool
  default     = true
}

variable "scaling_target_utilization" {
  description = "Target utilization percentage for auto-scaling"
  type        = number
  default     = 70
  validation {
    condition     = var.scaling_target_utilization >= 50 && var.scaling_target_utilization <= 90
    error_message = "Scaling target utilization must be between 50 and 90 percent."
  }
}

variable "min_capacity" {
  description = "Minimum capacity for auto-scaling"
  type        = number
  default     = 1
  validation {
    condition     = var.min_capacity >= 1
    error_message = "Minimum capacity must be at least 1."
  }
}

variable "max_capacity" {
  description = "Maximum capacity for auto-scaling"
  type        = number
  default     = 10
  validation {
    condition     = var.max_capacity >= 1 && var.max_capacity <= 50
    error_message = "Maximum capacity must be between 1 and 50."
  }
}

# Security configuration
variable "enable_encryption_at_rest" {
  description = "Enable encryption at rest for all monitoring data"
  type        = bool
  default     = true
}

variable "enable_encryption_in_transit" {
  description = "Enable encryption in transit for all communications"
  type        = bool
  default     = true
}

variable "allowed_cidr_blocks" {
  description = "CIDR blocks allowed to access monitoring resources"
  type        = list(string)
  default     = []
  validation {
    condition = alltrue([
      for cidr in var.allowed_cidr_blocks : can(cidrhost(cidr, 0))
    ])
    error_message = "All CIDR blocks must be valid."
  }
}