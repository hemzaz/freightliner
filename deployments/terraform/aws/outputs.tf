# AWS Infrastructure Outputs

# ECR Module Outputs
output "ecr_source_repositories" {
  description = "Source ECR repository information"
  value = {
    urls  = module.ecr.source_repository_urls
    arns  = module.ecr.source_repository_arns
    names = module.ecr.source_repository_names
  }
}

output "ecr_destination_repositories" {
  description = "Destination ECR repository information"
  value = {
    urls  = module.ecr.destination_repository_urls
    arns  = module.ecr.destination_repository_arns
    names = module.ecr.destination_repository_names
  }
}

output "ecr_all_repositories" {
  description = "All ECR repository information"
  value = {
    urls = module.ecr.all_repository_urls
    arns = module.ecr.all_repository_arns
  }
}

# IAM Resources
output "freightliner_role_arn" {
  description = "ARN of the Freightliner IAM role"
  value       = module.ecr.freightliner_role_arn
}

output "freightliner_role_name" {
  description = "Name of the Freightliner IAM role"
  value       = module.ecr.freightliner_role_name
}

output "ecr_access_policy_arn" {
  description = "ARN of the ECR access policy"
  value       = module.ecr.ecr_access_policy_arn
}

# S3 Resources
output "checkpoint_bucket_name" {
  description = "Name of the S3 bucket for replication checkpoints"
  value       = aws_s3_bucket.replication_checkpoints.id
}

output "checkpoint_bucket_arn" {
  description = "ARN of the S3 bucket for replication checkpoints"
  value       = aws_s3_bucket.replication_checkpoints.arn
}

output "checkpoint_s3_policy_arn" {
  description = "ARN of the S3 checkpoint access policy"
  value       = aws_iam_policy.checkpoint_s3_access.arn
}

# KMS Resources (if enabled)
output "kms_key_id" {
  description = "ID of the KMS key for ECR encryption"
  value       = module.ecr.kms_key_id
}

output "kms_key_arn" {
  description = "ARN of the KMS key for ECR encryption"
  value       = module.ecr.kms_key_arn
}

output "kms_key_alias" {
  description = "Alias of the KMS key for ECR encryption"
  value       = module.ecr.kms_key_alias
}

# CloudWatch Resources
output "cloudwatch_log_group_name" {
  description = "Name of the CloudWatch log group for ECR events"
  value       = module.ecr.cloudwatch_log_group_name
}

output "cloudwatch_log_group_arn" {
  description = "ARN of the CloudWatch log group for ECR events"
  value       = module.ecr.cloudwatch_log_group_arn
}

output "cloudwatch_dashboard_url" {
  description = "URL of the CloudWatch dashboard"
  value = var.enable_monitoring_dashboard ? "https://console.aws.amazon.com/cloudwatch/home?region=${var.aws_region}#dashboards:name=${aws_cloudwatch_dashboard.freightliner_dashboard[0].dashboard_name}" : null
}

# SNS Resources (if enabled)
output "sns_topic_arn" {
  description = "ARN of the SNS topic for alerts"
  value       = var.enable_sns_alerts ? aws_sns_topic.freightliner_alerts[0].arn : null
}

# Secrets Manager Resources (if enabled)
output "gcp_credentials_secret_arn" {
  description = "ARN of the Secrets Manager secret for GCP credentials"
  value       = var.create_gcp_secret ? aws_secretsmanager_secret.gcp_credentials[0].arn : null
}

output "gcp_credentials_secret_name" {
  description = "Name of the Secrets Manager secret for GCP credentials"
  value       = var.create_gcp_secret ? aws_secretsmanager_secret.gcp_credentials[0].name : null
}

output "secrets_manager_policy_arn" {
  description = "ARN of the Secrets Manager access policy"
  value       = var.create_gcp_secret ? aws_iam_policy.secrets_manager_access[0].arn : null
}

# Registry Configuration
output "registry_config" {
  description = "Complete registry configuration for Freightliner application"
  value = {
    aws = {
      region           = var.aws_region
      account_id       = data.aws_caller_identity.current.account_id
      registry_endpoint = module.ecr.registry_endpoint
      
      source_repositories      = module.ecr.source_repository_urls
      destination_repositories = module.ecr.destination_repository_urls
      
      iam_role_arn    = module.ecr.freightliner_role_arn
      iam_role_name   = module.ecr.freightliner_role_name
      
      checkpoint_bucket = aws_s3_bucket.replication_checkpoints.id
      
      secrets = {
        gcp_credentials_secret = var.create_gcp_secret ? aws_secretsmanager_secret.gcp_credentials[0].name : null
      }
      
      monitoring = {
        log_group_name     = module.ecr.cloudwatch_log_group_name
        sns_topic_arn      = var.enable_sns_alerts ? aws_sns_topic.freightliner_alerts[0].arn : null
        dashboard_name     = var.enable_monitoring_dashboard ? aws_cloudwatch_dashboard.freightliner_dashboard[0].dashboard_name : null
      }
    }
  }
}

# Environment Information
output "environment_info" {
  description = "Environment information"
  value = {
    environment = var.environment
    region      = var.aws_region
    account_id  = data.aws_caller_identity.current.account_id
    name_prefix = var.name_prefix
  }
}

# Repository Summary
output "repository_summary" {
  description = "Summary of created repositories"
  value = merge(module.ecr.repository_summary, {
    aws_region   = var.aws_region
    aws_account_id = data.aws_caller_identity.current.account_id
  })
}

# Application Configuration
output "application_env_vars" {
  description = "Environment variables for Freightliner application"
  value = {
    AWS_REGION                    = var.aws_region
    AWS_ACCOUNT_ID               = data.aws_caller_identity.current.account_id
    ECR_REGISTRY_ENDPOINT        = module.ecr.registry_endpoint
    CHECKPOINT_S3_BUCKET         = aws_s3_bucket.replication_checkpoints.id
    GCP_CREDENTIALS_SECRET_NAME  = var.create_gcp_secret ? aws_secretsmanager_secret.gcp_credentials[0].name : ""
    CLOUDWATCH_LOG_GROUP         = module.ecr.cloudwatch_log_group_name
    SNS_ALERTS_TOPIC_ARN        = var.enable_sns_alerts ? aws_sns_topic.freightliner_alerts[0].arn : ""
  }
}

# Kubernetes ConfigMap data
output "k8s_config_data" {
  description = "Configuration data for Kubernetes ConfigMap"
  value = {
    "aws-config.yaml" = yamlencode({
      aws = {
        region            = var.aws_region
        account_id        = data.aws_caller_identity.current.account_id
        registry_endpoint = module.ecr.registry_endpoint
        
        ecr = {
          source_repositories      = keys(module.ecr.source_repository_urls)
          destination_repositories = keys(module.ecr.destination_repository_urls)
        }
        
        s3 = {
          checkpoint_bucket = aws_s3_bucket.replication_checkpoints.id
        }
        
        secrets_manager = {
          gcp_credentials_secret = var.create_gcp_secret ? aws_secretsmanager_secret.gcp_credentials[0].name : ""
        }
        
        monitoring = {
          cloudwatch_log_group = module.ecr.cloudwatch_log_group_name
          sns_topic_arn       = var.enable_sns_alerts ? aws_sns_topic.freightliner_alerts[0].arn : ""
        }
      }
    })
  }
}