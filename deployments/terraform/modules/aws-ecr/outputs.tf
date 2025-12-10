# AWS ECR Module Outputs

# Source repositories
output "source_repository_urls" {
  description = "URLs of the source ECR repositories"
  value = {
    for name, repo in aws_ecr_repository.source_repositories : name => repo.repository_url
  }
}

output "source_repository_arns" {
  description = "ARNs of the source ECR repositories"
  value = {
    for name, repo in aws_ecr_repository.source_repositories : name => repo.arn
  }
}

output "source_repository_names" {
  description = "Names of the source ECR repositories"
  value = {
    for name, repo in aws_ecr_repository.source_repositories : name => repo.name
  }
}

# Destination repositories
output "destination_repository_urls" {
  description = "URLs of the destination ECR repositories"
  value = {
    for name, repo in aws_ecr_repository.destination_repositories : name => repo.repository_url
  }
}

output "destination_repository_arns" {
  description = "ARNs of the destination ECR repositories"
  value = {
    for name, repo in aws_ecr_repository.destination_repositories : name => repo.arn
  }
}

output "destination_repository_names" {
  description = "Names of the destination ECR repositories"
  value = {
    for name, repo in aws_ecr_repository.destination_repositories : name => repo.name
  }
}

# All repositories combined
output "all_repository_urls" {
  description = "URLs of all ECR repositories (source and destination)"
  value = merge(
    { for name, repo in aws_ecr_repository.source_repositories : name => repo.repository_url },
    { for name, repo in aws_ecr_repository.destination_repositories : name => repo.repository_url }
  )
}

output "all_repository_arns" {
  description = "ARNs of all ECR repositories (source and destination)"
  value = merge(
    { for name, repo in aws_ecr_repository.source_repositories : name => repo.arn },
    { for name, repo in aws_ecr_repository.destination_repositories : name => repo.arn }
  )
}

# IAM resources
output "freightliner_role_arn" {
  description = "ARN of the Freightliner IAM role"
  value       = aws_iam_role.freightliner_role.arn
}

output "freightliner_role_name" {
  description = "Name of the Freightliner IAM role"
  value       = aws_iam_role.freightliner_role.name
}

output "ecr_access_policy_arn" {
  description = "ARN of the ECR access policy"
  value       = aws_iam_policy.ecr_access_policy.arn
}

# KMS resources (if enabled)
output "kms_key_id" {
  description = "ID of the KMS key for ECR encryption"
  value       = var.enable_kms_encryption ? aws_kms_key.ecr_kms_key[0].key_id : null
}

output "kms_key_arn" {
  description = "ARN of the KMS key for ECR encryption"
  value       = var.enable_kms_encryption ? aws_kms_key.ecr_kms_key[0].arn : null
}

output "kms_key_alias" {
  description = "Alias of the KMS key for ECR encryption"
  value       = var.enable_kms_encryption ? aws_kms_alias.ecr_kms_key_alias[0].name : null
}

# CloudWatch resources
output "cloudwatch_log_group_name" {
  description = "Name of the CloudWatch log group for ECR events"
  value       = aws_cloudwatch_log_group.ecr_events.name
}

output "cloudwatch_log_group_arn" {
  description = "ARN of the CloudWatch log group for ECR events"
  value       = aws_cloudwatch_log_group.ecr_events.arn
}

# EventBridge resources (if enabled)
output "eventbridge_rule_name" {
  description = "Name of the EventBridge rule for ECR events"
  value       = var.enable_event_monitoring ? aws_cloudwatch_event_rule.ecr_image_push[0].name : null
}

output "eventbridge_rule_arn" {
  description = "ARN of the EventBridge rule for ECR events"
  value       = var.enable_event_monitoring ? aws_cloudwatch_event_rule.ecr_image_push[0].arn : null
}

# Registry configuration for Freightliner application
output "registry_config" {
  description = "Registry configuration for Freightliner application"
  value = {
    region = data.aws_region.current.name
    source_repositories = {
      for name, repo in aws_ecr_repository.source_repositories : name => {
        url  = repo.repository_url
        arn  = repo.arn
        name = repo.name
      }
    }
    destination_repositories = {
      for name, repo in aws_ecr_repository.destination_repositories : name => {
        url  = repo.repository_url
        arn  = repo.arn
        name = repo.name
      }
    }
    iam_role_arn = aws_iam_role.freightliner_role.arn
  }
}

# Current AWS region
data "aws_region" "current" {}

# Registry endpoint for Docker authentication
output "registry_endpoint" {
  description = "ECR registry endpoint for Docker authentication"
  value       = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${data.aws_region.current.name}.amazonaws.com"
}

# Current AWS account ID
data "aws_caller_identity" "current" {}

output "aws_account_id" {
  description = "AWS account ID where resources are created"
  value       = data.aws_caller_identity.current.account_id
}

# Repository count summary
output "repository_summary" {
  description = "Summary of created repositories"
  value = {
    source_count      = length(var.source_repositories)
    destination_count = length(var.destination_repositories)
    total_count       = length(var.source_repositories) + length(var.destination_repositories)
    environment       = var.environment
    name_prefix       = var.name_prefix
  }
}