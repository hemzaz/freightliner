# Additional variables for GitHub OIDC integration

variable "enable_github_oidc" {
  description = "Enable GitHub Actions OIDC provider for secure authentication"
  type        = bool
  default     = false
}
