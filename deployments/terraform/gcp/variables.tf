# GCP Infrastructure Variables

variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "gcp_region" {
  description = "GCP region"
  type        = string
  default     = "us-central1"
}

variable "gcp_location" {
  description = "GCP location for Artifact Registry (can be region or multi-region)"
  type        = string
  default     = "us-central1"
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
  description = "Cost center for resource labeling"
  type        = string
  default     = "platform"
}

variable "repository_prefix" {
  description = "Prefix for repository names"
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

# Registry Configuration
variable "use_artifact_registry" {
  description = "Use Artifact Registry instead of Container Registry (GCR)"
  type        = bool
  default     = true
}

variable "max_image_count" {
  description = "Maximum number of images to keep in each repository"
  type        = number
  default     = 100
}

variable "untagged_retention_days" {
  description = "Number of days to retain untagged images"
  type        = number
  default     = 7
}

variable "create_service_account_key" {
  description = "Create a service account key for external authentication"
  type        = bool
  default     = false
}

# GCR Configuration (if not using Artifact Registry)
variable "gcr_storage_locations" {
  description = "Storage locations for GCR buckets"
  type        = list(string)
  default     = ["us", "eu", "asia"]
}

variable "gcr_image_retention_days" {
  description = "Number of days to retain GCR images"
  type        = number
  default     = 90
}

# Monitoring and Logging
variable "enable_audit_logging" {
  description = "Enable audit logging for registry operations"
  type        = bool
  default     = true
}

variable "enable_monitoring" {
  description = "Enable Cloud Monitoring alerts"
  type        = bool
  default     = true
}

variable "alert_email_addresses" {
  description = "Email addresses for monitoring alerts"
  type        = list(string)
  default     = []
}

variable "repository_quota_threshold" {
  description = "Threshold for repository quota alerts"
  type        = number
  default     = 80
}

# Kubernetes Integration
variable "k8s_namespaces" {
  description = "Kubernetes namespaces for Workload Identity binding"
  type        = list(string)
  default     = ["freightliner", "freightliner-staging"]
}

variable "k8s_service_account_name" {
  description = "Kubernetes service account name for Workload Identity"
  type        = string
  default     = "freightliner"
}

# Binary Authorization
variable "enable_binary_authorization" {
  description = "Enable Binary Authorization for container images"
  type        = bool
  default     = false
}

variable "gke_clusters" {
  description = "GKE clusters for Binary Authorization policies"
  type        = list(string)
  default     = []
}

variable "pgp_public_key" {
  description = "PGP public key for Binary Authorization attestor"
  type        = string
  default     = ""
}

# Cloud Storage Configuration
variable "checkpoint_retention_days" {
  description = "Number of days to retain checkpoint files in Cloud Storage"
  type        = number
  default     = 90
}

# Secrets Configuration
variable "create_aws_secret" {
  description = "Create Secret Manager secret for AWS credentials"
  type        = bool
  default     = true
}

variable "aws_region" {
  description = "AWS region for cross-cloud replication"
  type        = string
  default     = "us-west-2"
}

# Cloud Build Configuration
variable "enable_cloud_build" {
  description = "Enable Cloud Build trigger for CI/CD"
  type        = bool
  default     = false
}

variable "github_owner" {
  description = "GitHub repository owner"
  type        = string
  default     = "company"
}

variable "github_repository" {
  description = "GitHub repository name"
  type        = string
  default     = "freightliner"
}

variable "github_branch" {
  description = "GitHub branch to trigger builds"
  type        = string
  default     = "main"
}

# Cloud Run Configuration
variable "enable_cloud_run_admin" {
  description = "Enable Cloud Run service for admin interface"
  type        = bool
  default     = false
}

variable "cloud_run_public_access" {
  description = "Allow public access to Cloud Run admin service"
  type        = bool
  default     = false
}

# Cloud Scheduler Configuration
variable "enable_health_check_scheduler" {
  description = "Enable Cloud Scheduler for health checks"
  type        = bool
  default     = false
}

variable "health_check_schedule" {
  description = "Cron schedule for health checks"
  type        = string
  default     = "0 */6 * * *" # Every 6 hours
}

variable "health_check_timezone" {
  description = "Timezone for health check schedule"
  type        = string
  default     = "UTC"
}

variable "external_health_check_url" {
  description = "External URL for health checks (used if Cloud Run admin is disabled)"
  type        = string
  default     = ""
}

# Pub/Sub Configuration
variable "enable_pubsub_events" {
  description = "Enable Pub/Sub for replication events"
  type        = bool
  default     = false
}