# GCP Container Registry / Artifact Registry Module Outputs

# Artifact Registry repositories (source)
output "source_repository_urls" {
  description = "URLs of the source Artifact Registry repositories"
  value = var.use_artifact_registry ? {
    for name, repo in google_artifact_registry_repository.source_repositories : name => "${var.location}-docker.pkg.dev/${var.project_id}/${repo.repository_id}"
  } : {}
}

output "source_repository_names" {
  description = "Names of the source repositories"
  value = var.use_artifact_registry ? {
    for name, repo in google_artifact_registry_repository.source_repositories : name => repo.repository_id
  } : {}
}

# Artifact Registry repositories (destination)
output "destination_repository_urls" {
  description = "URLs of the destination Artifact Registry repositories"
  value = var.use_artifact_registry ? {
    for name, repo in google_artifact_registry_repository.destination_repositories : name => "${var.location}-docker.pkg.dev/${var.project_id}/${repo.repository_id}"
  } : {}
}

output "destination_repository_names" {
  description = "Names of the destination repositories"
  value = var.use_artifact_registry ? {
    for name, repo in google_artifact_registry_repository.destination_repositories : name => repo.repository_id
  } : {}
}

# All repositories combined
output "all_repository_urls" {
  description = "URLs of all repositories (source and destination)"
  value = var.use_artifact_registry ? merge(
    { for name, repo in google_artifact_registry_repository.source_repositories : name => "${var.location}-docker.pkg.dev/${var.project_id}/${repo.repository_id}" },
    { for name, repo in google_artifact_registry_repository.destination_repositories : name => "${var.location}-docker.pkg.dev/${var.project_id}/${repo.repository_id}" }
  ) : {}
}

# Service account information
output "service_account_email" {
  description = "Email of the Freightliner service account"
  value       = google_service_account.freightliner_sa.email
}

output "service_account_name" {
  description = "Name of the Freightliner service account"
  value       = google_service_account.freightliner_sa.name
}

output "service_account_unique_id" {
  description = "Unique ID of the Freightliner service account"
  value       = google_service_account.freightliner_sa.unique_id
}

# Service account key (if created)
output "service_account_key" {
  description = "Base64 encoded service account key (sensitive)"
  value       = var.create_service_account_key ? google_service_account_key.freightliner_key[0].private_key : null
  sensitive   = true
}

# GCR storage buckets (if using GCR)
output "gcr_bucket_names" {
  description = "Names of the GCR storage buckets"
  value = var.use_artifact_registry ? {} : {
    for location, bucket in google_storage_bucket.gcr_buckets : location => bucket.name
  }
}

output "gcr_bucket_urls" {
  description = "URLs of the GCR storage buckets"
  value = var.use_artifact_registry ? {} : {
    for location, bucket in google_storage_bucket.gcr_buckets : location => bucket.url
  }
}

# Registry endpoints
output "artifact_registry_endpoint" {
  description = "Artifact Registry endpoint for Docker authentication"
  value       = var.use_artifact_registry ? "${var.location}-docker.pkg.dev" : null
}

output "gcr_endpoint" {
  description = "GCR endpoint for Docker authentication"
  value       = var.use_artifact_registry ? null : "gcr.io"
}

output "registry_endpoint" {
  description = "Primary registry endpoint for Docker authentication"
  value       = var.use_artifact_registry ? "${var.location}-docker.pkg.dev" : "gcr.io"
}

# Project information
output "project_id" {
  description = "GCP project ID where resources are created"
  value       = var.project_id
}

output "location" {
  description = "Location where Artifact Registry repositories are created"
  value       = var.location
}

# Monitoring resources
output "notification_channel_ids" {
  description = "IDs of the monitoring notification channels"
  value       = google_monitoring_notification_channel.email_alerts[*].id
}

output "alert_policy_names" {
  description = "Names of the monitoring alert policies"
  value = compact([
    var.enable_monitoring && var.use_artifact_registry ? google_monitoring_alert_policy.repository_quota_alert[0].name : ""
  ])
}

# Logging resources
output "logging_sink_name" {
  description = "Name of the logging sink for audit logs"
  value       = var.enable_audit_logging && var.use_artifact_registry ? google_logging_project_sink.artifact_registry_logs[0].name : null
}

# Binary Authorization resources (if enabled)
output "binary_authorization_policy_name" {
  description = "Name of the Binary Authorization policy"
  value       = var.enable_binary_authorization ? google_binary_authorization_policy.policy[0].name : null
}

output "binary_authorization_attestor_name" {
  description = "Name of the Binary Authorization attestor"
  value       = var.enable_binary_authorization ? google_binary_authorization_attestor.attestor[0].name : null
}

# Registry configuration for Freightliner application
output "registry_config" {
  description = "Registry configuration for Freightliner application"
  value = {
    project_id = var.project_id
    location   = var.location
    type       = var.use_artifact_registry ? "artifact-registry" : "container-registry"
    endpoint   = var.use_artifact_registry ? "${var.location}-docker.pkg.dev" : "gcr.io"
    
    source_repositories = var.use_artifact_registry ? {
      for name, repo in google_artifact_registry_repository.source_repositories : name => {
        url  = "${var.location}-docker.pkg.dev/${var.project_id}/${repo.repository_id}"
        name = repo.repository_id
        id   = repo.id
      }
    } : {}
    
    destination_repositories = var.use_artifact_registry ? {
      for name, repo in google_artifact_registry_repository.destination_repositories : name => {
        url  = "${var.location}-docker.pkg.dev/${var.project_id}/${repo.repository_id}"
        name = repo.repository_id
        id   = repo.id
      }
    } : {}
    
    service_account = {
      email     = google_service_account.freightliner_sa.email
      name      = google_service_account.freightliner_sa.name
      unique_id = google_service_account.freightliner_sa.unique_id
    }
  }
}

# Repository count summary
output "repository_summary" {
  description = "Summary of created repositories"
  value = {
    source_count      = length(var.source_repositories)
    destination_count = length(var.destination_repositories)
    total_count       = length(var.source_repositories) + length(var.destination_repositories)
    registry_type     = var.use_artifact_registry ? "artifact-registry" : "container-registry"
    environment       = var.environment
    name_prefix       = var.name_prefix
    project_id        = var.project_id
    location          = var.location
  }
}

# Workload Identity configuration
output "workload_identity_config" {
  description = "Workload Identity configuration for Kubernetes service accounts"
  value = {
    service_account_email = google_service_account.freightliner_sa.email
    k8s_service_accounts  = var.k8s_service_accounts
    annotations = {
      "iam.gke.io/gcp-service-account" = google_service_account.freightliner_sa.email
    }
  }
}