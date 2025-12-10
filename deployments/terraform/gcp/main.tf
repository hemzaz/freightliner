# GCP Infrastructure for Freightliner Container Registry Replication
# This configuration creates GCP Artifact Registry repositories and associated resources

terraform {
  required_version = ">= 1.0"
  
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
  
  # Configure remote state storage
  backend "gcs" {
    # Configure these values according to your setup
    # bucket = "your-terraform-state-bucket"
    # prefix = "freightliner/gcp/terraform.tfstate"
  }
}

# Configure GCP Provider
provider "google" {
  project = var.project_id
  region  = var.gcp_region
}

provider "google-beta" {
  project = var.project_id
  region  = var.gcp_region
}

# Local values
locals {
  common_labels = {
    project     = "freightliner"  
    environment = var.environment
    managed_by  = "terraform"
    component   = "container-registry"
    team        = "platform"
    cost_center = var.cost_center
  }
  
  # Repository naming convention
  source_repositories = [
    for repo in var.source_repository_names : "${var.repository_prefix}${repo}"
  ]
  
  destination_repositories = [
    for repo in var.destination_repository_names : "${var.repository_prefix}${repo}"
  ]
  
  # Kubernetes service accounts for Workload Identity
  k8s_service_accounts = [
    for ns in var.k8s_namespaces : "${ns}/${var.k8s_service_account_name}"
  ]
}

# GCR/GAR Module
module "gcr" {
  source = "../modules/gcp-gcr"
  
  project_id  = var.project_id
  name_prefix = var.name_prefix
  environment = var.environment
  location    = var.gcp_location
  
  # Registry configuration
  use_artifact_registry     = var.use_artifact_registry
  source_repositories      = local.source_repositories
  destination_repositories = local.destination_repositories
  
  # Image retention
  max_image_count         = var.max_image_count
  untagged_retention_days = var.untagged_retention_days
  
  # Service account
  create_service_account_key = var.create_service_account_key
  
  # GCR configuration (if not using Artifact Registry)
  gcr_storage_locations   = var.gcr_storage_locations
  gcr_image_retention_days = var.gcr_image_retention_days
  
  # Monitoring and logging
  enable_audit_logging    = var.enable_audit_logging
  enable_monitoring      = var.enable_monitoring
  alert_email_addresses  = var.alert_email_addresses
  repository_quota_threshold = var.repository_quota_threshold
  
  # Workload Identity
  k8s_service_accounts = local.k8s_service_accounts
  
  # Binary Authorization
  enable_binary_authorization = var.enable_binary_authorization
  gke_clusters               = var.gke_clusters
  pgp_public_key            = var.pgp_public_key
  
  common_labels = local.common_labels
}

# Cloud Storage bucket for replication checkpoints
resource "google_storage_bucket" "replication_checkpoints" {
  project  = var.project_id
  name     = "${var.name_prefix}-replication-checkpoints-${var.environment}-${random_id.bucket_suffix.hex}"
  location = var.gcp_region
  
  uniform_bucket_level_access = true
  
  versioning {
    enabled = true
  }
  
  lifecycle_rule {
    condition {
      age = var.checkpoint_retention_days
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
  
  lifecycle_rule {
    condition {
      num_newer_versions = 10
    }
    action {
      type = "Delete"
    }
  }

  labels = local.common_labels
}

# Random ID for unique bucket naming
resource "random_id" "bucket_suffix" {
  byte_length = 4
}

# IAM binding for checkpoint bucket access
resource "google_storage_bucket_iam_member" "checkpoint_bucket_access" {
  bucket = google_storage_bucket.replication_checkpoints.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${module.gcr.service_account_email}"
}

# Secret Manager secret for AWS credentials
resource "google_secret_manager_secret" "aws_credentials" {
  count     = var.create_aws_secret ? 1 : 0
  project   = var.project_id
  secret_id = "${var.name_prefix}-aws-credentials-${var.environment}"
  
  replication {
    auto {}
  }
  
  labels = local.common_labels
}

# Secret Manager secret version (placeholder)
resource "google_secret_manager_secret_version" "aws_credentials_version" {
  count  = var.create_aws_secret ? 1 : 0
  secret = google_secret_manager_secret.aws_credentials[0].id
  secret_data = jsonencode({
    access_key_id     = "PLACEHOLDER"
    secret_access_key = "PLACEHOLDER"
    region           = var.aws_region
    # These should be populated via terraform variables or external process
  })
  
  lifecycle {
    ignore_changes = [secret_data]
  }
}

# IAM binding for Secret Manager access
resource "google_secret_manager_secret_iam_member" "aws_credentials_access" {
  count     = var.create_aws_secret ? 1 : 0
  project   = var.project_id
  secret_id = google_secret_manager_secret.aws_credentials[0].secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${module.gcr.service_account_email}"
}

# Cloud Build trigger for CI/CD (optional)
resource "google_cloudbuild_trigger" "freightliner_build" {
  count   = var.enable_cloud_build ? 1 : 0
  project = var.project_id
  name    = "${var.name_prefix}-build-trigger"
  
  description = "Build trigger for Freightliner application"
  
  github {
    owner = var.github_owner
    name  = var.github_repository
    
    push {
      branch = var.github_branch
    }
  }
  
  build {
    step {
      name = "gcr.io/cloud-builders/docker"
      args = [
        "build",
        "-t", "gcr.io/${var.project_id}/${var.name_prefix}:$COMMIT_SHA",
        "-t", "gcr.io/${var.project_id}/${var.name_prefix}:latest",
        "."
      ]
    }
    
    step {
      name = "gcr.io/cloud-builders/docker"
      args = [
        "push",
        "gcr.io/${var.project_id}/${var.name_prefix}:$COMMIT_SHA"
      ]
    }
    
    step {
      name = "gcr.io/cloud-builders/docker"
      args = [
        "push",
        "gcr.io/${var.project_id}/${var.name_prefix}:latest"
      ]
    }
    
    images = [
      "gcr.io/${var.project_id}/${var.name_prefix}:$COMMIT_SHA",
      "gcr.io/${var.project_id}/${var.name_prefix}:latest"
    ]
    
    tags = [var.environment, "freightliner"]
  }
}

# Cloud Run service for health checks and admin interface (optional)
resource "google_cloud_run_service" "freightliner_admin" {
  count    = var.enable_cloud_run_admin ? 1 : 0
  project  = var.project_id
  name     = "${var.name_prefix}-admin"
  location = var.gcp_region

  template {
    spec {
      containers {
        image = "gcr.io/${var.project_id}/${var.name_prefix}:latest"
        
        env {
          name  = "ENVIRONMENT"
          value = var.environment
        }
        
        env {
          name  = "GCP_PROJECT_ID"
          value = var.project_id
        }
        
        env {
          name  = "GCP_REGION"
          value = var.gcp_region
        }
        
        resources {
          limits = {
            cpu    = "1000m"
            memory = "1Gi"
          }
        }
        
        ports {
          container_port = 8080
        }
      }
      
      service_account_name = module.gcr.service_account_email
      
      timeout_seconds = 300
    }
    
    metadata {
      annotations = {
        "autoscaling.knative.dev/minScale" = "0"
        "autoscaling.knative.dev/maxScale" = "5"
        "run.googleapis.com/cpu-throttling" = "false"
      }
      
      labels = local.common_labels
    }
  }
  
  traffic {
    percent         = 100
    latest_revision = true
  }
  
  depends_on = [module.gcr]
}

# Cloud Run IAM binding for public access (if enabled)
resource "google_cloud_run_service_iam_member" "admin_public_access" {
  count    = var.enable_cloud_run_admin && var.cloud_run_public_access ? 1 : 0
  project  = var.project_id
  location = google_cloud_run_service.freightliner_admin[0].location
  service  = google_cloud_run_service.freightliner_admin[0].name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# Cloud Scheduler job for periodic replication health checks
resource "google_cloud_scheduler_job" "health_check" {
  count    = var.enable_health_check_scheduler ? 1 : 0
  project  = var.project_id
  region   = var.gcp_region
  name     = "${var.name_prefix}-health-check"
  
  description = "Periodic health check for Freightliner replication"
  schedule    = var.health_check_schedule
  time_zone   = var.health_check_timezone
  
  http_target {
    http_method = "GET"
    uri         = var.enable_cloud_run_admin ? "${google_cloud_run_service.freightliner_admin[0].status[0].url}/health" : var.external_health_check_url
    
    headers = {
      "User-Agent" = "Google-Cloud-Scheduler"
    }
  }
  
  retry_config {
    retry_count = 3
  }
}

# Pub/Sub topic for replication events
resource "google_pubsub_topic" "replication_events" {
  count   = var.enable_pubsub_events ? 1 : 0
  project = var.project_id
  name    = "${var.name_prefix}-replication-events"
  
  labels = local.common_labels
}

# Pub/Sub subscription for replication events
resource "google_pubsub_subscription" "replication_events_sub" {
  count   = var.enable_pubsub_events ? 1 : 0
  project = var.project_id
  name    = "${var.name_prefix}-replication-events-sub"
  topic   = google_pubsub_topic.replication_events[0].name
  
  ack_deadline_seconds = 300
  
  expiration_policy {
    ttl = "2678400s" # 31 days
  }
  
  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }
  
  labels = local.common_labels
}

# IAM binding for Pub/Sub access
resource "google_pubsub_topic_iam_member" "replication_events_publisher" {
  count   = var.enable_pubsub_events ? 1 : 0
  project = var.project_id
  topic   = google_pubsub_topic.replication_events[0].name
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:${module.gcr.service_account_email}"
}