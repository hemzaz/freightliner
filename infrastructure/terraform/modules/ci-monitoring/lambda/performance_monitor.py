"""
CI/CD Performance Monitor
Tracks performance baselines, detects regressions, and optimizes resource utilization
"""

import json
import boto3
import os
import logging
from datetime import datetime, timedelta
from typing import Dict, List, Any, Optional, Tuple
import requests
from botocore.exceptions import ClientError
import statistics
import numpy as np
from dataclasses import dataclass

# Configure logging
logging.basicConfig(level=os.getenv('LOG_LEVEL', 'INFO'))
logger = logging.getLogger(__name__)

# AWS clients
cloudwatch = boto3.client('cloudwatch')
s3 = boto3.client('s3')
sns = boto3.client('sns')
secrets_manager = boto3.client('secretsmanager')

# Configuration
GITHUB_SECRET_ARN = os.getenv('GITHUB_SECRET_ARN')
GITHUB_OWNER = os.getenv('GITHUB_OWNER')
GITHUB_REPO = os.getenv('GITHUB_REPO')
S3_BUCKET = os.getenv('S3_BUCKET')
CLOUDWATCH_NAMESPACE = os.getenv('CLOUDWATCH_NAMESPACE', 'CI-CD/Performance')


class ConfigurationError(Exception):
    """Raised when configuration is invalid or missing"""
    pass


def validate_environment():
    """Validate required environment variables"""
    required_vars = {
        'GITHUB_SECRET_ARN': GITHUB_SECRET_ARN,
        'GITHUB_OWNER': GITHUB_OWNER,
        'GITHUB_REPO': GITHUB_REPO,
        'S3_BUCKET': S3_BUCKET
    }

    missing_vars = [name for name, value in required_vars.items() if not value]

    if missing_vars:
        raise ConfigurationError(f"Missing required environment variables: {', '.join(missing_vars)}")


def get_github_token() -> str:
    """Retrieve GitHub token from AWS Secrets Manager"""
    try:
        response = secrets_manager.get_secret_value(SecretId=GITHUB_SECRET_ARN)

        if 'SecretString' in response:
            secret_data = json.loads(response['SecretString'])
            token = secret_data.get('github_token')

            if not token:
                raise ConfigurationError("GitHub token not found in secret")

            return token
        else:
            raise ConfigurationError("Secret does not contain string data")

    except ClientError as e:
        error_code = e.response['Error']['Code']
        if error_code == 'ResourceNotFoundException':
            raise ConfigurationError(f"Secret not found: {GITHUB_SECRET_ARN}")
        elif error_code == 'AccessDeniedException':
            raise ConfigurationError(f"Access denied to secret: {GITHUB_SECRET_ARN}")
        else:
            raise ConfigurationError(f"Error retrieving secret: {str(e)}")


@dataclass
class PerformanceBaseline:
    """Performance baseline data structure"""
    metric_name: str
    mean: float
    std_dev: float
    percentile_95: float
    percentile_99: float
    sample_count: int
    last_updated: datetime
    confidence_interval: Tuple[float, float]


@dataclass
class PerformanceRegression:
    """Performance regression detection result"""
    metric_name: str
    current_value: float
    baseline_value: float
    regression_percent: float
    severity: str
    statistical_significance: float


class PerformanceAnalyzer:
    """Statistical analysis for performance metrics"""

    def __init__(self):
        self.regression_thresholds = {
            'minor': 0.1,    # 10% regression
            'major': 0.25,   # 25% regression
            'critical': 0.5  # 50% regression
        }
        self.min_samples_for_baseline = 10
        self.confidence_level = 0.95

    def calculate_baseline(self, values: List[float]) -> Optional[PerformanceBaseline]:
        """Calculate performance baseline from historical data"""
        if not values:
            logger.warning("Empty values list for baseline calculation")
            return None

        if len(values) < self.min_samples_for_baseline:
            logger.warning(f"Insufficient samples for baseline: {len(values)} < {self.min_samples_for_baseline}")
            return None

        try:
            # Remove outliers using IQR method
            q1 = np.percentile(values, 25)
            q3 = np.percentile(values, 75)
            iqr = q3 - q1
            lower_bound = q1 - 1.5 * iqr
            upper_bound = q3 + 1.5 * iqr

            filtered_values = [v for v in values if lower_bound <= v <= upper_bound]

            if len(filtered_values) < self.min_samples_for_baseline:
                logger.warning("Too many outliers removed, using original values")
                filtered_values = values

            mean = statistics.mean(filtered_values)
            std_dev = statistics.stdev(filtered_values) if len(filtered_values) > 1 else 0

            # Calculate confidence interval
            margin_of_error = 1.96 * (std_dev / np.sqrt(len(filtered_values)))  # 95% confidence
            confidence_interval = (mean - margin_of_error, mean + margin_of_error)

            return PerformanceBaseline(
                metric_name="",  # Will be set by caller
                mean=mean,
                std_dev=std_dev,
                percentile_95=np.percentile(filtered_values, 95),
                percentile_99=np.percentile(filtered_values, 99),
                sample_count=len(filtered_values),
                last_updated=datetime.now(),
                confidence_interval=confidence_interval
            )

        except (ValueError, TypeError, ZeroDivisionError) as e:
            logger.error(f"Error calculating baseline: {str(e)}")
            return None

    def detect_regression(self, current_value: float, baseline: PerformanceBaseline) -> Optional[PerformanceRegression]:
        """Detect performance regression against baseline"""
        if not baseline or baseline.mean == 0:
            return None

        # Calculate regression percentage
        regression_percent = (current_value - baseline.mean) / baseline.mean

        # Only consider it a regression if performance degraded
        if regression_percent <= 0:
            return None

        # Determine severity
        severity = 'info'
        if regression_percent >= self.regression_thresholds['critical']:
            severity = 'critical'
        elif regression_percent >= self.regression_thresholds['major']:
            severity = 'major'
        elif regression_percent >= self.regression_thresholds['minor']:
            severity = 'minor'

        # Calculate statistical significance (Z-score)
        if baseline.std_dev > 0:
            z_score = abs((current_value - baseline.mean) / baseline.std_dev)
            # Use normal distribution CDF for significance
            statistical_significance = min(1.0, z_score / 3.0)  # Normalize to 0-1
        else:
            statistical_significance = 1.0 if regression_percent > 0.01 else 0.0

        return PerformanceRegression(
            metric_name=baseline.metric_name,
            current_value=current_value,
            baseline_value=baseline.mean,
            regression_percent=regression_percent,
            severity=severity,
            statistical_significance=statistical_significance
        )


class GitHubMetricsCollector:
    """Collects performance metrics from GitHub Actions"""

    def __init__(self, github_token: str):
        self.session = requests.Session()
        self.session.headers.update({
            'Authorization': f'token {github_token}',
            'Accept': 'application/vnd.github.v3+json',
            'User-Agent': 'FreightlinerCI-PerformanceMonitor/1.0'
        })

    def get_workflow_runs(self, days_back: int = 30) -> List[Dict]:
        """Get workflow runs for performance analysis"""
        if not GITHUB_OWNER or not GITHUB_REPO:
            raise ConfigurationError("GitHub owner and repo must be configured")

        url = f"https://api.github.com/repos/{GITHUB_OWNER}/{GITHUB_REPO}/actions/runs"

        # Calculate date range
        end_date = datetime.now()
        start_date = end_date - timedelta(days=days_back)

        all_runs = []
        page = 1
        max_pages = 10  # Limit to prevent excessive API calls

        while len(all_runs) < 1000 and page <= max_pages:
            params = {
                'page': page,
                'per_page': 100,
                'status': 'completed',
                'created': f'>{start_date.isoformat()}'
            }

            try:
                response = self.session.get(url, params=params, timeout=30)
                response.raise_for_status()
                data = response.json()

                runs = data.get('workflow_runs', [])
                if not runs:
                    break

                all_runs.extend(runs)
                page += 1

                # Stop if we've gone beyond our date range
                if runs:
                    last_run_date = datetime.fromisoformat(runs[-1]['created_at'].replace('Z', '+00:00'))
                    if last_run_date < start_date:
                        break

            except requests.RequestException as e:
                logger.error(f"Error fetching workflow runs (page {page}): {str(e)}")
                break

        # Filter by date range
        filtered_runs = []
        for run in all_runs:
            try:
                created_date = datetime.fromisoformat(run['created_at'].replace('Z', '+00:00'))
                if start_date <= created_date <= end_date:
                    filtered_runs.append(run)
            except (ValueError, TypeError) as e:
                logger.warning(f"Invalid date format in run {run.get('id')}: {str(e)}")
                continue

        logger.info(f"Collected {len(filtered_runs)} workflow runs from last {days_back} days")
        return filtered_runs

    def get_job_details(self, run_id: str) -> List[Dict]:
        """Get detailed job information for a workflow run"""
        if not run_id:
            raise ValueError("Run ID cannot be empty")

        url = f"https://api.github.com/repos/{GITHUB_OWNER}/{GITHUB_REPO}/actions/runs/{run_id}/jobs"

        try:
            response = self.session.get(url, timeout=30)
            response.raise_for_status()
            data = response.json()
            return data.get('jobs', [])
        except requests.RequestException as e:
            logger.error(f"Error fetching job details for run {run_id}: {str(e)}")
            return []


class PerformanceMetricsExtractor:
    """Extracts performance metrics from workflow data"""

    def extract_pipeline_metrics(self, workflow_runs: List[Dict]) -> Dict[str, List[float]]:
        """Extract performance metrics from workflow runs"""
        metrics = {
            'total_duration': [],
            'queue_time': [],
            'success_rate_rolling': []
        }

        for i, run in enumerate(workflow_runs):
            try:
                # Calculate total duration
                if run.get('created_at') and run.get('updated_at'):
                    created = datetime.fromisoformat(run['created_at'].replace('Z', '+00:00'))
                    updated = datetime.fromisoformat(run['updated_at'].replace('Z', '+00:00'))
                    total_duration = (updated - created).total_seconds() / 60  # minutes
                    if total_duration > 0:  # Validate positive duration
                        metrics['total_duration'].append(total_duration)

                # Calculate queue time
                if run.get('created_at') and run.get('run_started_at'):
                    created = datetime.fromisoformat(run['created_at'].replace('Z', '+00:00'))
                    started = datetime.fromisoformat(run['run_started_at'].replace('Z', '+00:00'))
                    queue_time = (started - created).total_seconds() / 60  # minutes
                    if queue_time >= 0:  # Validate non-negative queue time
                        metrics['queue_time'].append(queue_time)

                # Calculate rolling success rate (last 10 runs)
                window_start = max(0, i - 9)
                window_runs = workflow_runs[window_start:i+1]
                success_count = sum(1 for r in window_runs if r.get('conclusion') == 'success')
                success_rate = success_count / len(window_runs) if window_runs else 0
                metrics['success_rate_rolling'].append(success_rate)

            except (ValueError, TypeError) as e:
                logger.error(f"Error extracting metrics from run {run.get('id', 'unknown')}: {str(e)}")
                continue

        return metrics

    def extract_resource_utilization(self, workflow_runs: List[Dict]) -> Dict[str, List[float]]:
        """Extract resource utilization metrics"""
        metrics = {
            'concurrent_jobs': []
        }

        time_buckets = {}
        for run in workflow_runs:
            if not (run.get('created_at') and run.get('updated_at')):
                continue

            try:
                created = datetime.fromisoformat(run['created_at'].replace('Z', '+00:00'))
                updated = datetime.fromisoformat(run['updated_at'].replace('Z', '+00:00'))

                # Bucket by hour for concurrent job analysis
                hour_bucket = created.replace(minute=0, second=0, microsecond=0)
                if hour_bucket not in time_buckets:
                    time_buckets[hour_bucket] = []
                time_buckets[hour_bucket].append((created, updated))

            except (ValueError, TypeError) as e:
                logger.error(f"Error processing run timing: {str(e)}")
                continue

        # Calculate concurrent jobs per hour
        for hour, runs in time_buckets.items():
            max_concurrent = 0
            for i, (start1, end1) in enumerate(runs):
                concurrent = 1  # Count the current run
                for j, (start2, end2) in enumerate(runs):
                    if i != j and start2 <= end1 and start1 <= end2:
                        concurrent += 1
                max_concurrent = max(max_concurrent, concurrent)

            metrics['concurrent_jobs'].append(max_concurrent)

        return metrics


def load_baselines_from_s3() -> Dict[str, PerformanceBaseline]:
    """Load existing performance baselines from S3"""
    try:
        key = "performance-baselines/current-baselines.json"
        response = s3.get_object(Bucket=S3_BUCKET, Key=key)
        data = json.loads(response['Body'].read())

        baselines = {}
        for metric_name, baseline_data in data.items():
            baselines[metric_name] = PerformanceBaseline(
                metric_name=metric_name,
                mean=baseline_data['mean'],
                std_dev=baseline_data['std_dev'],
                percentile_95=baseline_data['percentile_95'],
                percentile_99=baseline_data['percentile_99'],
                sample_count=baseline_data['sample_count'],
                last_updated=datetime.fromisoformat(baseline_data['last_updated']),
                confidence_interval=tuple(baseline_data['confidence_interval'])
            )

        logger.info(f"Loaded {len(baselines)} performance baselines from S3")
        return baselines

    except ClientError as e:
        if e.response['Error']['Code'] == 'NoSuchKey':
            logger.info("No existing baselines found, will create new ones")
            return {}
        else:
            logger.error(f"Error loading baselines from S3: {str(e)}")
            return {}
    except (json.JSONDecodeError, KeyError, ValueError) as e:
        logger.error(f"Error parsing baseline data: {str(e)}")
        return {}


def save_baselines_to_s3(baselines: Dict[str, PerformanceBaseline]):
    """Save performance baselines to S3"""
    try:
        baseline_data = {}
        for metric_name, baseline in baselines.items():
            baseline_data[metric_name] = {
                'mean': baseline.mean,
                'std_dev': baseline.std_dev,
                'percentile_95': baseline.percentile_95,
                'percentile_99': baseline.percentile_99,
                'sample_count': baseline.sample_count,
                'last_updated': baseline.last_updated.isoformat(),
                'confidence_interval': list(baseline.confidence_interval)
            }

        key = "performance-baselines/current-baselines.json"
        s3.put_object(
            Bucket=S3_BUCKET,
            Key=key,
            Body=json.dumps(baseline_data, default=str),
            ContentType='application/json',
            ServerSideEncryption='aws:kms'
        )

        # Also save a timestamped backup
        backup_key = f"performance-baselines/backups/baselines-{datetime.now().strftime('%Y%m%d-%H%M%S')}.json"
        s3.put_object(
            Bucket=S3_BUCKET,
            Key=backup_key,
            Body=json.dumps(baseline_data, default=str),
            ContentType='application/json',
            ServerSideEncryption='aws:kms'
        )

        logger.info(f"Saved {len(baselines)} performance baselines to S3")

    except ClientError as e:
        logger.error(f"Error saving baselines to S3: {str(e)}")


def publish_performance_metrics(metrics: Dict[str, List[float]], baselines: Dict[str, PerformanceBaseline]):
    """Publish performance metrics to CloudWatch"""
    timestamp = datetime.now()
    metric_data = []

    for metric_name, values in metrics.items():
        if not values:
            continue

        current_value = values[-1] if values else 0  # Most recent value
        avg_value = statistics.mean(values) if values else 0

        # Current value metric
        metric_data.append({
            'MetricName': f'{metric_name}_current',
            'Value': current_value,
            'Unit': 'None',
            'Timestamp': timestamp,
            'Dimensions': [
                {'Name': 'Repository', 'Value': f"{GITHUB_OWNER}/{GITHUB_REPO}"},
                {'Name': 'MetricType', 'Value': 'performance'}
            ]
        })

        # Average value metric
        metric_data.append({
            'MetricName': f'{metric_name}_average',
            'Value': avg_value,
            'Unit': 'None',
            'Timestamp': timestamp,
            'Dimensions': [
                {'Name': 'Repository', 'Value': f"{GITHUB_OWNER}/{GITHUB_REPO}"},
                {'Name': 'MetricType', 'Value': 'performance'}
            ]
        })

        # Baseline comparison if available
        if metric_name in baselines:
            baseline = baselines[metric_name]
            deviation_percent = ((current_value - baseline.mean) / baseline.mean * 100) if baseline.mean != 0 else 0

            metric_data.append({
                'MetricName': f'{metric_name}_baseline_deviation',
                'Value': deviation_percent,
                'Unit': 'Percent',
                'Timestamp': timestamp,
                'Dimensions': [
                    {'Name': 'Repository', 'Value': f"{GITHUB_OWNER}/{GITHUB_REPO}"},
                    {'Name': 'MetricType', 'Value': 'baseline_comparison'}
                ]
            })

    # Publish metrics in batches
    batch_size = 20
    for i in range(0, len(metric_data), batch_size):
        batch = metric_data[i:i + batch_size]
        try:
            cloudwatch.put_metric_data(
                Namespace=CLOUDWATCH_NAMESPACE,
                MetricData=batch
            )
        except ClientError as e:
            logger.error(f"Error publishing metrics batch: {str(e)}")

    logger.info(f"Published {len(metric_data)} performance metrics to CloudWatch")


def send_regression_alerts(regressions: List[PerformanceRegression]):
    """Send alerts for detected performance regressions"""
    if not regressions:
        return

    critical_regressions = [r for r in regressions if r.severity == 'critical']
    major_regressions = [r for r in regressions if r.severity == 'major']

    # Log critical regressions
    for regression in critical_regressions:
        logger.critical(f"CRITICAL REGRESSION: {regression.metric_name} degraded by {regression.regression_percent:.2%}")

    # Log major regressions
    if major_regressions:
        logger.warning(f"PERFORMANCE REGRESSIONS: {len(major_regressions)} metrics showing degradation")


def handler(event, context):
    """Lambda handler for performance monitoring"""
    logger.info("Starting performance monitoring")

    try:
        # Validate environment configuration
        validate_environment()

        # Get GitHub token from Secrets Manager
        github_token = get_github_token()

        # Initialize components
        collector = GitHubMetricsCollector(github_token)
        extractor = PerformanceMetricsExtractor()
        analyzer = PerformanceAnalyzer()

        # Collect workflow data
        workflow_runs = collector.get_workflow_runs(days_back=30)
        if not workflow_runs:
            logger.warning("No workflow runs found for analysis")
            return {
                'statusCode': 200,
                'body': json.dumps({'message': 'No workflow runs found for analysis'})
            }

        # Extract performance metrics
        pipeline_metrics = extractor.extract_pipeline_metrics(workflow_runs)
        resource_metrics = extractor.extract_resource_utilization(workflow_runs)

        # Combine all metrics
        all_metrics = {**pipeline_metrics, **resource_metrics}

        # Load existing baselines
        baselines = load_baselines_from_s3()

        # Update baselines and detect regressions
        updated_baselines = {}
        regressions = []

        for metric_name, values in all_metrics.items():
            if not values:
                continue

            # Calculate new baseline
            new_baseline = analyzer.calculate_baseline(values)
            if new_baseline:
                new_baseline.metric_name = metric_name
                updated_baselines[metric_name] = new_baseline

                # Check for regression using old baseline
                if metric_name in baselines:
                    current_value = values[-1]  # Most recent value
                    regression = analyzer.detect_regression(current_value, baselines[metric_name])
                    if regression:
                        regressions.append(regression)

        # Save updated baselines
        if updated_baselines:
            save_baselines_to_s3(updated_baselines)

        # Publish metrics to CloudWatch
        publish_performance_metrics(all_metrics, updated_baselines)

        # Send regression alerts
        send_regression_alerts(regressions)

        # Prepare response
        summary = {
            'workflow_runs_analyzed': len(workflow_runs),
            'metrics_collected': {name: len(values) for name, values in all_metrics.items()},
            'baselines_updated': len(updated_baselines),
            'regressions_detected': len(regressions),
            'critical_regressions': len([r for r in regressions if r.severity == 'critical']),
            'major_regressions': len([r for r in regressions if r.severity == 'major'])
        }

        logger.info(f"Performance monitoring completed: {json.dumps(summary)}")

        return {
            'statusCode': 200,
            'body': json.dumps({
                'message': 'Performance monitoring completed successfully',
                'summary': summary
            })
        }

    except ConfigurationError as e:
        logger.error(f"Configuration error: {str(e)}")
        return {
            'statusCode': 500,
            'body': json.dumps({
                'error': 'Configuration error',
                'details': str(e)
            })
        }

    except Exception as e:
        logger.error(f"Performance monitoring failed: {str(e)}", exc_info=True)
        return {
            'statusCode': 500,
            'body': json.dumps({
                'error': 'Performance monitoring failed',
                'details': str(e)
            })
        }
