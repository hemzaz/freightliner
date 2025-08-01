# AWS Infrastructure Variables

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
  default     = "freightliner"
}

variable "environment" {
  description = "Environment name (e.g., dev, staging, prod)"
  type        = string
  default     = "prod"
}

variable "cost_center" {
  description = "Cost center for resource tagging"
  type        = string
  default     = "platform"
}

variable "repository_prefix" {
  description = "Prefix for ECR repository names"
  type        = string
  default     = ""
}

variable "source_repository_names" {
  description = "List of source repository names to create"
  type        = list(string)
  default = [
    "nginx",
    "alpine",
    "ubuntu",
    "postgres",
    "redis"
  ]
}

variable "destination_repository_names" {
  description = "List of destination repository names to create"
  type        = list(string)
  default = [
    "mirror/nginx",
    "mirror/alpine",
    "mirror/ubuntu",
    "mirror/postgres",
    "mirror/redis"
  ]
}

# ECR Configuration
variable "image_tag_mutability" {
  description = "The tag mutability setting for the repositories"
  type        = string
  default     = "MUTABLE"
  validation {
    condition     = contains(["MUTABLE", "IMMUTABLE"], var.image_tag_mutability)
    error_message = "Image tag mutability must be either MUTABLE or IMMUTABLE."
  }
}

variable "scan_on_push" {
  description = "Enable image scanning on push"
  type        = bool
  default     = true
}

variable "force_delete_repository" {
  description = "If true, will delete the repository even if it contains images"
  type        = bool
  default     = false
}

variable "max_image_count" {
  description = "Maximum number of images to keep in each repository"
  type        = number
  default     = 100
}

# Cross-account access
variable "enable_cross_account_access" {
  description = "Enable cross-account access to ECR repositories"
  type        = bool
  default     = false
}

variable "cross_account_arns" {
  description = "List of AWS account ARNs allowed to access ECR repositories"
  type        = list(string)
  default     = []
}

variable "assume_role_arns" {
  description = "List of ARNs that can assume the Freightliner IAM role"
  type        = list(string)
  default     = []
}

# Encryption
variable "enable_kms_encryption" {
  description = "Enable KMS encryption for ECR repositories"
  type        = bool
  default     = false
}

variable "kms_key_deletion_window" {
  description = "KMS key deletion window in days"
  type        = number
  default     = 7
}

# Monitoring and Logging
variable "enable_event_monitoring" {
  description = "Enable EventBridge monitoring for ECR events"
  type        = bool
  default     = true
}

variable "log_retention_days" {
  description = "CloudWatch log retention in days"
  type        = number
  default     = 30
}

variable "enable_monitoring_dashboard" {
  description = "Create CloudWatch dashboard for monitoring"
  type        = bool
  default     = true
}

variable "enable_sns_alerts" {
  description = "Enable SNS alerts for monitoring"
  type        = bool
  default     = true
}

variable "alert_email_addresses" {
  description = "Email addresses for monitoring alerts"
  type        = list(string)
  default     = []
}

variable "enable_repository_alarms" {
  description = "Enable CloudWatch alarms for repository metrics"
  type        = bool
  default     = true
}

variable "repository_count_threshold" {
  description = "Threshold for repository count alarm"
  type        = number
  default     = 50
}

# S3 Configuration
variable "checkpoint_retention_days" {
  description = "Number of days to retain checkpoint files in S3"
  type        = number
  default     = 90
}

# Secrets Management
variable "create_gcp_secret" {
  description = "Create AWS Secrets Manager secret for GCP credentials"
  type        = bool
  default     = true
}