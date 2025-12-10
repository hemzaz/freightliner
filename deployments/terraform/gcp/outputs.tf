# GCP Infrastructure Outputs

# GCR/GAR Module Outputs
output "gcr_source_repositories" {
  description = "Source repository information"
  value = {
    urls  = module.gcr.source_repository_urls
    names = module.gcr.source_repository_names
  }
}

output "gcr_destination_repositories" {
  description = "Destination repository information"
  value = {
    urls  = module.gcr.destination_repository_urls
    names = module.gcr.destination_repository_names
  }
}

output "gcr_all_repositories" {
  description = "All repository information"
  value = {
    urls = module.gcr.all_repository_urls
  }
}

# Service Account Information
output "service_account_email" {
  description = "Email of the Freightliner service account"
  value       = module.gcr.service_account_email
}

output "service_account_name" {
  description = "Name of the Freightliner service account"
  value       = module.gcr.service_account_name
}

output "service_account_unique_id" {
  description = "Unique ID of the Freightliner service account"
  value       = module.gcr.service_account_unique_id
}

output "service_account_key" {
  description = "Base64 encoded service account key (sensitive)"
  value       = module.gcr.service_account_key
  sensitive   = true
}

# Cloud Storage Resources
output "checkpoint_bucket_name" {
  description = "Name of the Cloud Storage bucket for replication checkpoints"
  value       = google_storage_bucket.replication_checkpoints.name
}

output "checkpoint_bucket_url" {
  description = "URL of the Cloud Storage bucket for replication checkpoints"
  value       = google_storage_bucket.replication_checkpoints.url
}

output "gcr_bucket_names" {
  description = "Names of the GCR storage buckets (if using GCR)"
  value       = module.gcr.gcr_bucket_names
}

# Registry Endpoints
output "registry_endpoint" {
  description = "Primary registry endpoint for Docker authentication"
  value       = module.gcr.registry_endpoint
}

output "artifact_registry_endpoint" {
  description = "Artifact Registry endpoint"
  value       = module.gcr.artifact_registry_endpoint
}

output "gcr_endpoint" {
  description = "GCR endpoint"
  value       = module.gcr.gcr_endpoint
}

# Monitoring Resources
output "notification_channel_ids" {
  description = "IDs of the monitoring notification channels"
  value       = module.gcr.notification_channel_ids
}

output "alert_policy_names" {
  description = "Names of the monitoring alert policies"
  value       = module.gcr.alert_policy_names
}

output "logging_sink_name" {
  description = "Name of the logging sink for audit logs"
  value       = module.gcr.logging_sink_name
}

# Secret Manager Resources
output "aws_credentials_secret_name" {
  description = "Name of the Secret Manager secret for AWS credentials"
  value       = var.create_aws_secret ? google_secret_manager_secret.aws_credentials[0].secret_id : null
}

output "aws_credentials_secret_id" {
  description = "ID of the Secret Manager secret for AWS credentials"
  value       = var.create_aws_secret ? google_secret_manager_secret.aws_credentials[0].id : null
}

# Binary Authorization Resources
output "binary_authorization_policy_name" {
  description = "Name of the Binary Authorization policy"
  value       = module.gcr.binary_authorization_policy_name
}

output "binary_authorization_attestor_name" {
  description = "Name of the Binary Authorization attestor"
  value       = module.gcr.binary_authorization_attestor_name
}

# Cloud Build Resources
output "cloud_build_trigger_id" {
  description = "ID of the Cloud Build trigger"
  value       = var.enable_cloud_build ? google_cloudbuild_trigger.freightliner_build[0].trigger_id : null
}

output "cloud_build_trigger_name" {
  description = "Name of the Cloud Build trigger"
  value       = var.enable_cloud_build ? google_cloudbuild_trigger.freightliner_build[0].name : null
}

# Cloud Run Resources
output "cloud_run_service_url" {
  description = "URL of the Cloud Run admin service"
  value       = var.enable_cloud_run_admin ? google_cloud_run_service.freightliner_admin[0].status[0].url : null
}

output "cloud_run_service_name" {
  description = "Name of the Cloud Run admin service"
  value       = var.enable_cloud_run_admin ? google_cloud_run_service.freightliner_admin[0].name : null
}

# Cloud Scheduler Resources
output "health_check_job_name" {
  description = "Name of the Cloud Scheduler health check job"
  value       = var.enable_health_check_scheduler ? google_cloud_scheduler_job.health_check[0].name : null
}

# Pub/Sub Resources
output "pubsub_topic_name" {
  description = "Name of the Pub/Sub topic for replication events"
  value       = var.enable_pubsub_events ? google_pubsub_topic.replication_events[0].name : null
}

output "pubsub_subscription_name" {
  description = "Name of the Pub/Sub subscription for replication events"
  value       = var.enable_pubsub_events ? google_pubsub_subscription.replication_events_sub[0].name : null
}

# Workload Identity Configuration
output "workload_identity_config" {
  description = "Workload Identity configuration for Kubernetes"
  value       = module.gcr.workload_identity_config
}

# Registry Configuration
output "registry_config" {
  description = "Complete registry configuration for Freightliner application"
  value = {
    gcp = {
      project_id = var.project_id
      region     = var.gcp_region
      location   = var.gcp_location
      
      registry_type     = var.use_artifact_registry ? "artifact-registry" : "container-registry"
      registry_endpoint = module.gcr.registry_endpoint
      
      source_repositories      = module.gcr.source_repository_urls
      destination_repositories = module.gcr.destination_repository_urls
      
      service_account = {
        email     = module.gcr.service_account_email
        name      = module.gcr.service_account_name
        unique_id = module.gcr.service_account_unique_id
      }
      
      storage = {
        checkpoint_bucket = google_storage_bucket.replication_checkpoints.name
      }
      
      secrets = {
        aws_credentials_secret = var.create_aws_secret ? google_secret_manager_secret.aws_credentials[0].secret_id : null
      }
      
      monitoring = {
        notification_channels = module.gcr.notification_channel_ids
        alert_policies       = module.gcr.alert_policy_names
        logging_sink         = module.gcr.logging_sink_name
      }
      
      cloud_run = {
        admin_service_url = var.enable_cloud_run_admin ? google_cloud_run_service.freightliner_admin[0].status[0].url : null
      }
      
      pubsub = {
        topic_name        = var.enable_pubsub_events ? google_pubsub_topic.replication_events[0].name : null
        subscription_name = var.enable_pubsub_events ? google_pubsub_subscription.replication_events_sub[0].name : null
      }
    }
  }
}

# Environment Information
output "environment_info" {
  description = "Environment information"
  value = {
    environment = var.environment
    project_id  = var.project_id
    region      = var.gcp_region
    location    = var.gcp_location
    name_prefix = var.name_prefix
  }
}

# Repository Summary
output "repository_summary" {
  description = "Summary of created repositories"
  value = merge(module.gcr.repository_summary, {
    gcp_project_id = var.project_id
    gcp_region     = var.gcp_region
    gcp_location   = var.gcp_location
  })
}

# Application Configuration
output "application_env_vars" {
  description = "Environment variables for Freightliner application"
  value = {
    GCP_PROJECT_ID              = var.project_id
    GCP_REGION                  = var.gcp_region
    GCP_LOCATION               = var.gcp_location
    REGISTRY_ENDPOINT          = module.gcr.registry_endpoint
    CHECKPOINT_GCS_BUCKET      = google_storage_bucket.replication_checkpoints.name
    AWS_CREDENTIALS_SECRET_ID  = var.create_aws_secret ? google_secret_manager_secret.aws_credentials[0].secret_id : ""
    PUBSUB_TOPIC_NAME         = var.enable_pubsub_events ? google_pubsub_topic.replication_events[0].name : ""
    CLOUD_RUN_SERVICE_URL     = var.enable_cloud_run_admin ? google_cloud_run_service.freightliner_admin[0].status[0].url : ""
  }
}

# Kubernetes ConfigMap data 
output "k8s_config_data" {
  description = "Configuration data for Kubernetes ConfigMap"
  value = {
    "gcp-config.yaml" = yamlencode({
      gcp = {
        project_id = var.project_id
        region     = var.gcp_region
        location   = var.gcp_location
        
        registry = {
          type     = var.use_artifact_registry ? "artifact-registry" : "container-registry"
          endpoint = module.gcr.registry_endpoint
          
          source_repositories      = keys(module.gcr.source_repository_urls)
          destination_repositories = keys(module.gcr.destination_repository_urls)
        }
        
        service_account = {
          email = module.gcr.service_account_email
        }
        
        storage = {
          checkpoint_bucket = google_storage_bucket.replication_checkpoints.name
        }
        
        secrets = {
          aws_credentials_secret = var.create_aws_secret ? google_secret_manager_secret.aws_credentials[0].secret_id : ""
        }
        
        monitoring = {
          notification_channels = module.gcr.notification_channel_ids
          logging_sink         = module.gcr.logging_sink_name
        }
        
        pubsub = {
          topic_name        = var.enable_pubsub_events ? google_pubsub_topic.replication_events[0].name : ""
          subscription_name = var.enable_pubsub_events ? google_pubsub_subscription.replication_events_sub[0].name : ""
        }
      }
    })
  }
}

# Terraform state information
output "terraform_state_info" {
  description = "Information about Terraform state and configuration"
  value = {
    backend_type = "gcs"
    environment  = var.environment
    project_id   = var.project_id
    region       = var.gcp_region
    
    resources_created = {
      repositories      = length(local.source_repositories) + length(local.destination_repositories)
      service_accounts  = 1
      storage_buckets   = 1 + (var.use_artifact_registry ? 0 : length(var.gcr_storage_locations))
      secrets          = var.create_aws_secret ? 1 : 0
      cloud_run_services = var.enable_cloud_run_admin ? 1 : 0
      scheduler_jobs    = var.enable_health_check_scheduler ? 1 : 0
      pubsub_topics     = var.enable_pubsub_events ? 1 : 0
    }
  }
}