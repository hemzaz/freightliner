# AWS ECR Module Variables

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

variable "source_repositories" {
  description = "List of source ECR repository names to create"
  type        = list(string)
  default     = []
}

variable "destination_repositories" {
  description = "List of destination ECR repository names to create"
  type        = list(string)
  default     = []
}

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

variable "common_tags" {
  description = "Common tags to apply to all resources"
  type        = map(string)
  default = {
    Project     = "freightliner"
    Component   = "container-registry"
    ManagedBy   = "terraform"
  }
}