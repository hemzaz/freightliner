# GCP Container Registry / Artifact Registry Module
# Creates and configures GCR/GAR repositories for container registry replication

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 5.0"
    }
  }
  required_version = ">= 1.0"
}

# Enable required APIs
resource "google_project_service" "required_apis" {
  for_each = toset([
    "artifactregistry.googleapis.com",
    "containerregistry.googleapis.com",
    "storage.googleapis.com",
    "cloudresourcemanager.googleapis.com",
    "iam.googleapis.com",
    "logging.googleapis.com",
    "monitoring.googleapis.com"
  ])
  
  project = var.project_id
  service = each.value
  
  disable_dependent_services = false
  disable_on_destroy         = false
}

# Artifact Registry repositories for source
resource "google_artifact_registry_repository" "source_repositories" {
  for_each = var.use_artifact_registry ? toset(var.source_repositories) : toset([])
  
  project       = var.project_id
  location      = var.location
  repository_id = each.value
  description   = "Source repository for ${each.value} - managed by Freightliner"
  format        = "DOCKER"
  mode          = "STANDARD_REPOSITORY"
  
  cleanup_policies {
    id     = "keep-minimum-versions"
    action = "KEEP"
    
    most_recent_versions {
      keep_count = var.max_image_count
    }
  }
  
  cleanup_policies {
    id     = "delete-untagged"
    action = "DELETE"
    
    condition {
      tag_state  = "UNTAGGED"
      older_than = "${var.untagged_retention_days}d"
    }
  }

  labels = merge(var.common_labels, {
    type        = "source"
    environment = var.environment
    managed_by  = "terraform"
  })

  depends_on = [google_project_service.required_apis]
}

# Artifact Registry repositories for destination
resource "google_artifact_registry_repository" "destination_repositories" {
  for_each = var.use_artifact_registry ? toset(var.destination_repositories) : toset([])
  
  project       = var.project_id
  location      = var.location
  repository_id = each.value
  description   = "Destination repository for ${each.value} - managed by Freightliner"
  format        = "DOCKER"
  mode          = "STANDARD_REPOSITORY"
  
  cleanup_policies {
    id     = "keep-minimum-versions"
    action = "KEEP"
    
    most_recent_versions {
      keep_count = var.max_image_count
    }
  }
  
  cleanup_policies {
    id     = "delete-untagged"
    action = "DELETE"
    
    condition {
      tag_state  = "UNTAGGED"
      older_than = "${var.untagged_retention_days}d"
    }
  }

  labels = merge(var.common_labels, {
    type        = "destination"
    environment = var.environment
    managed_by  = "terraform"
  })

  depends_on = [google_project_service.required_apis]
}

# Service account for Freightliner application
resource "google_service_account" "freightliner_sa" {
  project      = var.project_id
  account_id   = "${var.name_prefix}-freightliner-sa"
  display_name = "Freightliner Container Registry Replication Service Account"
  description  = "Service account for Freightliner application to access container registries"
}

# IAM binding for Artifact Registry repositories (source)
resource "google_artifact_registry_repository_iam_member" "source_repository_access" {
  for_each = var.use_artifact_registry ? toset(var.source_repositories) : toset([])
  
  project    = var.project_id
  location   = var.location
  repository = google_artifact_registry_repository.source_repositories[each.value].name
  role       = "roles/artifactregistry.repoAdmin"
  member     = "serviceAccount:${google_service_account.freightliner_sa.email}"
}

# IAM binding for Artifact Registry repositories (destination)
resource "google_artifact_registry_repository_iam_member" "destination_repository_access" {
  for_each = var.use_artifact_registry ? toset(var.destination_repositories) : toset([])
  
  project    = var.project_id
  location   = var.location
  repository = google_artifact_registry_repository.destination_repositories[each.value].name
  role       = "roles/artifactregistry.repoAdmin"
  member     = "serviceAccount:${google_service_account.freightliner_sa.email}"
}

# IAM binding for Container Registry (if using GCR)
resource "google_project_iam_member" "gcr_access" {
  count = var.use_artifact_registry ? 0 : 1
  
  project = var.project_id
  role    = "roles/storage.objectAdmin"  # Required for GCR
  member  = "serviceAccount:${google_service_account.freightliner_sa.email}"
}

# IAM binding for storage bucket access (GCR backend)
resource "google_project_iam_member" "storage_access" {
  count = var.use_artifact_registry ? 0 : 1
  
  project = var.project_id
  role    = "roles/storage.legacyBucketReader"
  member  = "serviceAccount:${google_service_account.freightliner_sa.email}"
}

# Service account key for external authentication
resource "google_service_account_key" "freightliner_key" {
  count              = var.create_service_account_key ? 1 : 0
  service_account_id = google_service_account.freightliner_sa.name
  public_key_type    = "TYPE_X509_PEM_FILE"
  private_key_type   = "TYPE_GOOGLE_CREDENTIALS_FILE"
}

# Cloud Storage buckets for GCR (if not using Artifact Registry)
resource "google_storage_bucket" "gcr_buckets" {
  for_each = var.use_artifact_registry ? toset([]) : toset(var.gcr_storage_locations)
  
  project  = var.project_id
  name     = "${each.value}.artifacts.${var.project_id}.appspot.com"
  location = each.value
  
  uniform_bucket_level_access = true
  
  versioning {
    enabled = true
  }
  
  lifecycle_rule {
    condition {
      age = var.gcr_image_retention_days
    }
    action {
      type = "Delete"
    }
  }
  
  lifecycle_rule {
    condition {
      matches_storage_class = ["STANDARD"]
      age                  = 30
    }
    action {
      type          = "SetStorageClass"
      storage_class = "NEARLINE"
    }
  }

  labels = merge(var.common_labels, {
    type        = "gcr-storage"
    environment = var.environment
    managed_by  = "terraform"
  })
}

# Cloud Logging sink for Artifact Registry audit logs
resource "google_logging_project_sink" "artifact_registry_logs" {
  count = var.enable_audit_logging && var.use_artifact_registry ? 1 : 0
  
  project     = var.project_id
  name        = "${var.name_prefix}-artifact-registry-logs"
  description = "Audit logs for Artifact Registry operations"
  
  destination = "logging.googleapis.com/projects/${var.project_id}/logs/${var.name_prefix}-artifact-registry-audit"
  
  filter = <<-EOT
    protoPayload.serviceName="artifactregistry.googleapis.com"
    AND (
      protoPayload.methodName:"CreateRepository"
      OR protoPayload.methodName:"DeleteRepository"
      OR protoPayload.methodName:"UpdateRepository"
      OR protoPayload.methodName:"GetRepository"
      OR protoPayload.methodName:"ListRepositories"
    )
  EOT
  
  unique_writer_identity = true
}

# Cloud Monitoring notification channel
resource "google_monitoring_notification_channel" "email_alerts" {
  count = length(var.alert_email_addresses) > 0 ? length(var.alert_email_addresses) : 0
  
  project     = var.project_id
  display_name = "Freightliner Alerts - ${var.alert_email_addresses[count.index]}"
  type         = "email"
  
  labels = {
    email_address = var.alert_email_addresses[count.index]
  }
}

# Cloud Monitoring alert policy for repository quota
resource "google_monitoring_alert_policy" "repository_quota_alert" {
  count = var.enable_monitoring && var.use_artifact_registry ? 1 : 0
  
  project     = var.project_id
  display_name = "${var.name_prefix} - Artifact Registry Repository Quota Alert"
  
  conditions {
    display_name = "Repository count approaching quota"
    
    condition_threshold {
      filter          = "resource.type=\"artifactregistry.googleapis.com/Repository\""
      comparison      = "COMPARISON_GREATER_THAN"
      threshold_value = var.repository_quota_threshold
      duration        = "300s"
      
      aggregations {
        alignment_period   = "300s"
        per_series_aligner = "ALIGN_COUNT"
      }
    }
  }
  
  notification_channels = google_monitoring_notification_channel.email_alerts[*].id
  
  alert_strategy {
    auto_close = "1800s"
  }
}

# Workload Identity binding for Kubernetes service accounts
resource "google_service_account_iam_member" "workload_identity_binding" {
  count = length(var.k8s_service_accounts)
  
  service_account_id = google_service_account.freightliner_sa.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${var.project_id}.svc.id.goog[${var.k8s_service_accounts[count.index]}]"
}

# Binary Authorization policy (if enabled)
resource "google_binary_authorization_policy" "policy" {
  count = var.enable_binary_authorization ? 1 : 0
  
  project = var.project_id
  
  default_admission_rule {
    evaluation_mode  = "REQUIRE_ATTESTATION"
    enforcement_mode = "ENFORCED_BLOCK_AND_AUDIT_LOG"
    
    require_attestations_by = [
      google_binary_authorization_attestor.attestor[0].name
    ]
  }
  
  dynamic "cluster_admission_rules" {
    for_each = var.gke_clusters
    content {
      cluster                = cluster_admission_rules.value
      evaluation_mode        = "REQUIRE_ATTESTATION"
      enforcement_mode       = "ENFORCED_BLOCK_AND_AUDIT_LOG"
      require_attestations_by = [
        google_binary_authorization_attestor.attestor[0].name
      ]
    }
  }
}

# Binary Authorization attestor
resource "google_binary_authorization_attestor" "attestor" {
  count = var.enable_binary_authorization ? 1 : 0
  
  project = var.project_id
  name    = "${var.name_prefix}-attestor"
  
  attestation_authority_note {
    note_reference = google_container_analysis_note.note[0].name
    public_keys {
      ascii_armored_pgp_public_key = var.pgp_public_key
    }
  }
}

# Container Analysis note for Binary Authorization
resource "google_container_analysis_note" "note" {
  count = var.enable_binary_authorization ? 1 : 0
  
  project = var.project_id
  name    = "${var.name_prefix}-attestor-note"
  
  attestation_authority {
    hint {
      human_readable_name = "Freightliner Attestor"
    }
  }
}