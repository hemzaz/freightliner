# GCP Container Registry / Artifact Registry Module Variables

variable "project_id" {
  description = "GCP project ID"
  type        = string
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

variable "location" {
  description = "Location for Artifact Registry repositories"
  type        = string
  default     = "us-central1"
}

variable "use_artifact_registry" {
  description = "Use Artifact Registry instead of Container Registry (GCR)"
  type        = bool
  default     = true
}

variable "source_repositories" {
  description = "List of source repository names to create"
  type        = list(string)
  default     = []
}

variable "destination_repositories" {
  description = "List of destination repository names to create"
  type        = list(string)
  default     = []
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

variable "gcr_storage_locations" {
  description = "Storage locations for GCR buckets (only used if use_artifact_registry is false)"
  type        = list(string)
  default     = ["us", "eu", "asia"]
}

variable "gcr_image_retention_days" {
  description = "Number of days to retain GCR images"
  type        = number
  default     = 90
}

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

variable "k8s_service_accounts" {
  description = "Kubernetes service accounts for Workload Identity binding (format: namespace/service-account-name)"
  type        = list(string)
  default     = []
}

variable "enable_binary_authorization" {
  description = "Enable Binary Authorization for container images"
  type        = bool
  default     = false
}

variable "gke_clusters" {
  description = "GKE clusters for Binary Authorization policies (format: projects/PROJECT_ID/locations/LOCATION/clusters/CLUSTER_NAME)"
  type        = list(string)
  default     = []
}

variable "pgp_public_key" {
  description = "PGP public key for Binary Authorization attestor"
  type        = string
  default     = ""
}

variable "common_labels" {
  description = "Common labels to apply to all resources"
  type        = map(string)
  default = {
    project     = "freightliner"
    component   = "container-registry"
    managed_by  = "terraform"
  }
}