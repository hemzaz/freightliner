# Replicate Debug Command

Debug a failing replication operation by analyzing logs, checking credentials, and testing connectivity.

## What This Command Does

1. Checks replication service configuration
2. Validates AWS ECR and GCP GCR credentials
3. Tests registry connectivity
4. Analyzes recent error logs
5. Provides actionable debugging steps

## Usage

```bash
/replicate-debug [source] [destination]
```

## Example

```bash
/replicate-debug ecr/my-repo gcr/my-repo
```

## Tasks

- Validate AWS credentials (`aws sts get-caller-identity`)
- Validate GCP credentials (`gcloud auth list`)
- Test ECR connectivity with `aws ecr describe-repositories`
- Test GCR connectivity with appropriate API calls
- Check recent application logs for errors
- Verify network connectivity to registry endpoints
- Check if repositories exist and are accessible
- Validate IAM/service account permissions
- Provide step-by-step troubleshooting recommendations
