"""
CI/CD Pipeline Metrics Collector
Collects metrics from GitHub Actions and publishes to CloudWatch
"""

import json
import boto3
import os
import logging
from datetime import datetime, timedelta
from typing import Dict, List, Any, Optional
import requests
from botocore.exceptions import ClientError
import time

# Configure logging
logging.basicConfig(level=os.getenv('LOG_LEVEL', 'INFO'))
logger = logging.getLogger(__name__)

# AWS clients
cloudwatch = boto3.client('cloudwatch')
dynamodb = boto3.resource('dynamodb')
s3 = boto3.client('s3')
sns = boto3.client('sns')

# Configuration
GITHUB_TOKEN = os.getenv('GITHUB_TOKEN')
GITHUB_OWNER = os.getenv('GITHUB_OWNER')
GITHUB_REPO = os.getenv('GITHUB_REPO')
DYNAMODB_TABLE = os.getenv('DYNAMODB_TABLE')
S3_BUCKET = os.getenv('S3_BUCKET')
SNS_CRITICAL_TOPIC = os.getenv('SNS_CRITICAL_TOPIC')
SNS_WARNING_TOPIC = os.getenv('SNS_WARNING_TOPIC')

# Thresholds from Terraform template
SUCCESS_RATE_THRESHOLD = ${success_rate_threshold}
DURATION_THRESHOLD_MINUTES = ${duration_threshold}
FAILURE_RATE_THRESHOLD = ${failure_rate_threshold}

# Circuit breaker state
circuit_breaker_state = {
    'failures': 0,
    'last_failure': None,
    'state': 'CLOSED'  # CLOSED, OPEN, HALF_OPEN
}

def circuit_breaker(func):
    """Circuit breaker decorator for external API calls"""
    def wrapper(*args, **kwargs):
        global circuit_breaker_state
        
        if circuit_breaker_state['state'] == 'OPEN':
            # Check if we should try half-open
            if (circuit_breaker_state['last_failure'] and 
                datetime.now() - circuit_breaker_state['last_failure'] > timedelta(minutes=5)):
                circuit_breaker_state['state'] = 'HALF_OPEN'
                logger.info("Circuit breaker moved to HALF_OPEN state")
            else:
                logger.warning("Circuit breaker is OPEN, skipping external call")
                return None
        
        try:
            result = func(*args, **kwargs)
            # Success - reset circuit breaker
            if circuit_breaker_state['state'] == 'HALF_OPEN':
                circuit_breaker_state['state'] = 'CLOSED'
                circuit_breaker_state['failures'] = 0
                logger.info("Circuit breaker moved to CLOSED state")
            return result
            
        except Exception as e:
            circuit_breaker_state['failures'] += 1
            circuit_breaker_state['last_failure'] = datetime.now()
            
            if circuit_breaker_state['failures'] >= 5:
                circuit_breaker_state['state'] = 'OPEN'
                logger.error(f"Circuit breaker opened due to {circuit_breaker_state['failures']} failures")
            
            logger.error(f"Circuit breaker recorded failure: {str(e)}")
            raise
    
    return wrapper

@circuit_breaker
def get_github_workflow_runs(page: int = 1, per_page: int = 100) -> Optional[Dict]:
    """Fetch workflow runs from GitHub API with circuit breaker"""
    url = f"https://api.github.com/repos/{GITHUB_OWNER}/{GITHUB_REPO}/actions/runs"
    headers = {
        'Authorization': f'token {GITHUB_TOKEN}',
        'Accept': 'application/vnd.github.v3+json'
    }
    params = {
        'page': page,
        'per_page': per_page,
        'status': 'completed'
    }
    
    logger.info(f"Fetching workflow runs from GitHub API (page {page})")
    response = requests.get(url, headers=headers, params=params, timeout=30)
    response.raise_for_status()
    
    return response.json()

def retry_with_backoff(func, max_retries: int = 3, base_delay: float = 1.0):
    """Retry function with exponential backoff"""
    for attempt in range(max_retries):
        try:
            return func()
        except Exception as e:
            if attempt == max_retries - 1:
                raise
            
            delay = base_delay * (2 ** attempt)
            logger.warning(f"Attempt {attempt + 1} failed: {str(e)}. Retrying in {delay} seconds...")
            time.sleep(delay)

def calculate_pipeline_metrics(workflow_runs: List[Dict]) -> Dict[str, Any]:
    """Calculate pipeline health metrics from workflow runs"""
    if not workflow_runs:
        return {}
    
    total_runs = len(workflow_runs)
    successful_runs = sum(1 for run in workflow_runs if run['conclusion'] == 'success')
    failed_runs = sum(1 for run in workflow_runs if run['conclusion'] == 'failure')
    cancelled_runs = sum(1 for run in workflow_runs if run['conclusion'] == 'cancelled')
    
    # Calculate durations
    durations = []
    for run in workflow_runs:
        if run.get('created_at') and run.get('updated_at'):
            created = datetime.fromisoformat(run['created_at'].replace('Z', '+00:00'))
            updated = datetime.fromisoformat(run['updated_at'].replace('Z', '+00:00'))
            duration_minutes = (updated - created).total_seconds() / 60
            durations.append(duration_minutes)
    
    avg_duration = sum(durations) / len(durations) if durations else 0
    max_duration = max(durations) if durations else 0
    min_duration = min(durations) if durations else 0
    
    # Performance trends
    recent_runs = workflow_runs[:10]  # Last 10 runs
    recent_success_rate = (
        sum(1 for run in recent_runs if run['conclusion'] == 'success') / len(recent_runs)
        if recent_runs else 0
    )
    
    success_rate = successful_runs / total_runs if total_runs > 0 else 0
    failure_rate = failed_runs / total_runs if total_runs > 0 else 0
    
    return {
        'total_runs': total_runs,
        'successful_runs': successful_runs,
        'failed_runs': failed_runs,
        'cancelled_runs': cancelled_runs,
        'success_rate': success_rate,
        'failure_rate': failure_rate,
        'recent_success_rate': recent_success_rate,
        'avg_duration_minutes': avg_duration,
        'max_duration_minutes': max_duration,
        'min_duration_minutes': min_duration,
        'long_running_pipelines': sum(1 for d in durations if d > DURATION_THRESHOLD_MINUTES)
    }

def publish_cloudwatch_metrics(metrics: Dict[str, Any], timestamp: datetime):
    """Publish metrics to CloudWatch"""
    try:
        metric_data = [
            {
                'MetricName': 'PipelineSuccessRate',
                'Value': metrics.get('success_rate', 0),
                'Unit': 'Percent',
                'Timestamp': timestamp,
                'Dimensions': [
                    {'Name': 'Repository', 'Value': f"{GITHUB_OWNER}/{GITHUB_REPO}"},
                    {'Name': 'Environment', 'Value': os.getenv('ENVIRONMENT', 'unknown')}
                ]
            },
            {
                'MetricName': 'PipelineFailureRate',
                'Value': metrics.get('failure_rate', 0),
                'Unit': 'Percent',
                'Timestamp': timestamp,
                'Dimensions': [
                    {'Name': 'Repository', 'Value': f"{GITHUB_OWNER}/{GITHUB_REPO}"},
                    {'Name': 'Environment', 'Value': os.getenv('ENVIRONMENT', 'unknown')}
                ]
            },
            {
                'MetricName': 'AveragePipelineDuration',
                'Value': metrics.get('avg_duration_minutes', 0),
                'Unit': 'Count',
                'Timestamp': timestamp,
                'Dimensions': [
                    {'Name': 'Repository', 'Value': f"{GITHUB_OWNER}/{GITHUB_REPO}"},
                    {'Name': 'Environment', 'Value': os.getenv('ENVIRONMENT', 'unknown')}
                ]
            },
            {
                'MetricName': 'TotalPipelineRuns',
                'Value': metrics.get('total_runs', 0),
                'Unit': 'Count',
                'Timestamp': timestamp,
                'Dimensions': [
                    {'Name': 'Repository', 'Value': f"{GITHUB_OWNER}/{GITHUB_REPO}"},
                    {'Name': 'Environment', 'Value': os.getenv('ENVIRONMENT', 'unknown')}
                ]
            },
            {
                'MetricName': 'LongRunningPipelines',
                'Value': metrics.get('long_running_pipelines', 0),
                'Unit': 'Count',
                'Timestamp': timestamp,
                'Dimensions': [
                    {'Name': 'Repository', 'Value': f"{GITHUB_OWNER}/{GITHUB_REPO}"},
                    {'Name': 'Environment', 'Value': os.getenv('ENVIRONMENT', 'unknown')}
                ]
            }
        ]
        
        # Batch publish metrics
        cloudwatch.put_metric_data(
            Namespace='CI-CD/Pipeline',
            MetricData=metric_data
        )
        
        logger.info(f"Published {len(metric_data)} metrics to CloudWatch")
        
    except ClientError as e:
        logger.error(f"Failed to publish metrics to CloudWatch: {str(e)}")
        raise

def store_pipeline_metadata(metrics: Dict[str, Any], timestamp: datetime):
    """Store pipeline metadata in DynamoDB"""
    try:
        table = dynamodb.Table(DYNAMODB_TABLE)
        
        item = {
            'pipeline_id': f"{GITHUB_OWNER}/{GITHUB_REPO}",
            'execution_id': f"metrics-{int(timestamp.timestamp())}",
            'timestamp': timestamp.isoformat(),
            'metrics': metrics,
            'success_rate': metrics.get('success_rate', 0),
            'failure_rate': metrics.get('failure_rate', 0),
            'avg_duration': metrics.get('avg_duration_minutes', 0),
            'total_runs': metrics.get('total_runs', 0),
            'status': 'healthy' if metrics.get('success_rate', 0) >= SUCCESS_RATE_THRESHOLD else 'degraded',
            'ttl': int((timestamp + timedelta(days=90)).timestamp())  # Auto-expire after 90 days
        }
        
        table.put_item(Item=item)
        logger.info("Stored pipeline metadata in DynamoDB")
        
    except ClientError as e:
        logger.error(f"Failed to store pipeline metadata: {str(e)}")
        raise

def check_and_alert(metrics: Dict[str, Any]):
    """Check metrics against thresholds and send alerts if needed"""
    alerts = []
    
    # Check success rate
    success_rate = metrics.get('success_rate', 0)
    if success_rate < SUCCESS_RATE_THRESHOLD:
        alerts.append({
            'severity': 'CRITICAL' if success_rate < 0.8 else 'WARNING',
            'message': f"Pipeline success rate is {success_rate:.2%}, below threshold of {SUCCESS_RATE_THRESHOLD:.2%}",
            'metric': 'success_rate',
            'value': success_rate,
            'threshold': SUCCESS_RATE_THRESHOLD
        })
    
    # Check failure rate
    failure_rate = metrics.get('failure_rate', 0)
    if failure_rate > FAILURE_RATE_THRESHOLD:
        alerts.append({
            'severity': 'WARNING',
            'message': f"Pipeline failure rate is {failure_rate:.2%}, above threshold of {FAILURE_RATE_THRESHOLD:.2%}",
            'metric': 'failure_rate',
            'value': failure_rate,
            'threshold': FAILURE_RATE_THRESHOLD
        })
    
    # Check average duration
    avg_duration = metrics.get('avg_duration_minutes', 0)
    if avg_duration > DURATION_THRESHOLD_MINUTES:
        alerts.append({
            'severity': 'WARNING',
            'message': f"Average pipeline duration is {avg_duration:.1f} minutes, above threshold of {DURATION_THRESHOLD_MINUTES} minutes",
            'metric': 'avg_duration',
            'value': avg_duration,
            'threshold': DURATION_THRESHOLD_MINUTES
        })
    
    # Send alerts
    for alert in alerts:
        try:
            topic_arn = SNS_CRITICAL_TOPIC if alert['severity'] == 'CRITICAL' else SNS_WARNING_TOPIC
            
            message = {
                'alert_type': 'pipeline_health',
                'severity': alert['severity'],
                'timestamp': datetime.now().isoformat(),
                'repository': f"{GITHUB_OWNER}/{GITHUB_REPO}",
                'metric': alert['metric'],
                'current_value': alert['value'],
                'threshold': alert['threshold'],
                'message': alert['message'],
                'circuit_breaker_state': circuit_breaker_state['state']
            }
            
            sns.publish(
                TopicArn=topic_arn,
                Message=json.dumps(message, default=str),
                Subject=f"CI/CD Alert: {alert['message']}"
            )
            
            logger.warning(f"Sent {alert['severity']} alert: {alert['message']}")
            
        except ClientError as e:
            logger.error(f"Failed to send alert: {str(e)}")

def save_metrics_to_s3(metrics: Dict[str, Any], timestamp: datetime):
    """Save detailed metrics to S3 for historical analysis"""
    try:
        key = f"pipeline-metrics/{timestamp.strftime('%Y/%m/%d')}/metrics-{int(timestamp.timestamp())}.json"
        
        data = {
            'timestamp': timestamp.isoformat(),
            'repository': f"{GITHUB_OWNER}/{GITHUB_REPO}",
            'metrics': metrics,
            'circuit_breaker_state': circuit_breaker_state,
            'collection_metadata': {
                'collector_version': '1.0.0',
                'environment': os.getenv('ENVIRONMENT', 'unknown')
            }
        }
        
        s3.put_object(
            Bucket=S3_BUCKET,
            Key=key,
            Body=json.dumps(data, default=str),
            ContentType='application/json',
            ServerSideEncryption='aws:kms'
        )
        
        logger.info(f"Saved metrics to S3: s3://{S3_BUCKET}/{key}")
        
    except ClientError as e:
        logger.error(f"Failed to save metrics to S3: {str(e)}")

def handler(event, context):
    """Lambda handler for pipeline metrics collection"""
    logger.info("Starting pipeline metrics collection")
    
    try:
        timestamp = datetime.now()
        
        # Fetch workflow runs with circuit breaker protection
        workflow_data = get_github_workflow_runs()
        if not workflow_data:
            logger.warning("No workflow data received, possibly due to circuit breaker")
            return {
                'statusCode': 200,
                'body': json.dumps({
                    'message': 'Metrics collection skipped due to circuit breaker',
                    'circuit_breaker_state': circuit_breaker_state['state']
                })
            }
        
        workflow_runs = workflow_data.get('workflow_runs', [])
        logger.info(f"Fetched {len(workflow_runs)} workflow runs")
        
        # Calculate metrics
        metrics = calculate_pipeline_metrics(workflow_runs)
        logger.info(f"Calculated metrics: {json.dumps(metrics, default=str)}")
        
        # Store and publish metrics with retry
        retry_with_backoff(lambda: publish_cloudwatch_metrics(metrics, timestamp))
        retry_with_backoff(lambda: store_pipeline_metadata(metrics, timestamp))
        retry_with_backoff(lambda: save_metrics_to_s3(metrics, timestamp))
        
        # Check for alerts
        check_and_alert(metrics)
        
        return {
            'statusCode': 200,
            'body': json.dumps({
                'message': 'Pipeline metrics collected successfully',
                'metrics_summary': {
                    'total_runs': metrics.get('total_runs', 0),
                    'success_rate': f"{metrics.get('success_rate', 0):.2%}",
                    'avg_duration': f"{metrics.get('avg_duration_minutes', 0):.1f} minutes"
                },
                'circuit_breaker_state': circuit_breaker_state['state']
            })
        }
        
    except Exception as e:
        logger.error(f"Pipeline metrics collection failed: {str(e)}")
        
        # Send critical alert about collection failure
        try:
            sns.publish(
                TopicArn=SNS_CRITICAL_TOPIC,
                Message=json.dumps({
                    'alert_type': 'collector_failure',
                    'severity': 'CRITICAL',
                    'timestamp': datetime.now().isoformat(),
                    'error': str(e),
                    'function': 'pipeline_metrics_collector'
                }, default=str),
                Subject="CI/CD Monitoring: Pipeline Metrics Collection Failed"
            )
        except Exception as alert_error:
            logger.error(f"Failed to send failure alert: {str(alert_error)}")
        
        return {
            'statusCode': 500,
            'body': json.dumps({
                'error': 'Pipeline metrics collection failed',
                'details': str(e)
            })
        }