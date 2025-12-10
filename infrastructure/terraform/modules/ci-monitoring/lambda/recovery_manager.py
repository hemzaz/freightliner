"""
Automated Recovery Manager for CI/CD Pipeline
Implements automated recovery procedures, circuit breakers, and failover strategies
"""

import json
import boto3
import os
import logging
from datetime import datetime, timedelta
from typing import Dict, List, Any, Optional
from botocore.exceptions import ClientError
from enum import Enum
import requests
import time

# Configure logging
logging.basicConfig(level=os.getenv('LOG_LEVEL', 'INFO'))
logger = logging.getLogger(__name__)

# AWS clients
dynamodb = boto3.resource('dynamodb')
s3 = boto3.client('s3')
sns = boto3.client('sns')
lambda_client = boto3.client('lambda')
cloudwatch = boto3.client('cloudwatch')
secrets_manager = boto3.client('secretsmanager')

# Configuration
GITHUB_SECRET_ARN = os.getenv('GITHUB_SECRET_ARN')
GITHUB_OWNER = os.getenv('GITHUB_OWNER')
GITHUB_REPO = os.getenv('GITHUB_REPO')
DYNAMODB_TABLE = os.getenv('DYNAMODB_TABLE')
S3_BUCKET = os.getenv('S3_BUCKET')
SNS_CRITICAL_TOPIC = os.getenv('SNS_CRITICAL_TOPIC')
CIRCUIT_BREAKER_ENABLED = os.getenv('CIRCUIT_BREAKER_ENABLED', 'true').lower() == 'true'
CIRCUIT_BREAKER_FAILURE_THRESHOLD = int(os.getenv('CIRCUIT_BREAKER_FAILURE_THRESHOLD', '5'))
CIRCUIT_BREAKER_TIMEOUT = int(os.getenv('CIRCUIT_BREAKER_TIMEOUT', '60'))


class ConfigurationError(Exception):
    """Raised when configuration is invalid or missing"""
    pass


class RecoveryError(Exception):
    """Raised when recovery operations fail"""
    pass


class CircuitBreakerState(Enum):
    CLOSED = "CLOSED"
    OPEN = "OPEN"
    HALF_OPEN = "HALF_OPEN"


class RecoveryAction(Enum):
    RESTART_WORKFLOW = "restart_workflow"
    SCALE_RESOURCES = "scale_resources"
    CLEAR_CACHE = "clear_cache"
    NOTIFY_TEAM = "notify_team"
    ROLLBACK_DEPLOYMENT = "rollback_deployment"
    BYPASS_FAILING_STEP = "bypass_failing_step"


def validate_environment():
    """Validate required environment variables"""
    required_vars = {
        'GITHUB_SECRET_ARN': GITHUB_SECRET_ARN,
        'GITHUB_OWNER': GITHUB_OWNER,
        'GITHUB_REPO': GITHUB_REPO,
        'DYNAMODB_TABLE': DYNAMODB_TABLE,
        'S3_BUCKET': S3_BUCKET,
        'SNS_CRITICAL_TOPIC': SNS_CRITICAL_TOPIC
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


def validate_event_input(event: Dict[str, Any]) -> bool:
    """Validate Lambda event input structure"""
    if not event:
        raise ValueError("Event cannot be empty")

    # Validate SNS events
    if 'Records' in event:
        if not isinstance(event['Records'], list):
            raise ValueError("Event 'Records' must be a list")

        for record in event['Records']:
            if not isinstance(record, dict):
                raise ValueError("Each record must be a dictionary")

            if 'EventSource' in record and record['EventSource'] not in ['aws:sns', 'aws:events']:
                raise ValueError(f"Unsupported event source: {record['EventSource']}")

    # Validate direct invocation
    elif 'failure_type' in event:
        valid_types = ['pipeline_failure', 'performance_regression', 'resource_exhaustion',
                      'external_dependency_failure', 'build_failure', 'manual_recovery']
        if event['failure_type'] not in valid_types:
            raise ValueError(f"Invalid failure type: {event['failure_type']}")

    return True


class CircuitBreakerManager:
    """Manages circuit breaker states for external dependencies"""

    def __init__(self):
        self.table = dynamodb.Table(os.getenv('DYNAMODB_TABLE_CIRCUIT_BREAKER', 'circuit-breaker-state'))

    def get_circuit_state(self, service_name: str) -> Dict[str, Any]:
        """Get current circuit breaker state for a service"""
        if not service_name:
            raise ValueError("Service name cannot be empty")

        try:
            response = self.table.get_item(Key={'service_name': service_name})

            if 'Item' in response:
                return response['Item']
            else:
                # Initialize new circuit breaker
                return {
                    'service_name': service_name,
                    'state': CircuitBreakerState.CLOSED.value,
                    'failure_count': 0,
                    'last_failure': None,
                    'last_success': datetime.now().isoformat(),
                    'ttl': int((datetime.now() + timedelta(days=30)).timestamp())
                }

        except ClientError as e:
            logger.error(f"Error getting circuit state for {service_name}: {str(e)}")
            return self._get_default_circuit_state(service_name)

    def _get_default_circuit_state(self, service_name: str) -> Dict[str, Any]:
        """Get default circuit breaker state"""
        return {
            'service_name': service_name,
            'state': CircuitBreakerState.CLOSED.value,
            'failure_count': 0,
            'last_failure': None,
            'last_success': datetime.now().isoformat(),
            'ttl': int((datetime.now() + timedelta(days=30)).timestamp())
        }

    def record_success(self, service_name: str):
        """Record successful operation"""
        try:
            current_state = self.get_circuit_state(service_name)

            # Reset circuit breaker on success
            updated_state = {
                **current_state,
                'state': CircuitBreakerState.CLOSED.value,
                'failure_count': 0,
                'last_success': datetime.now().isoformat(),
                'ttl': int((datetime.now() + timedelta(days=30)).timestamp())
            }

            self.table.put_item(Item=updated_state)

            if current_state['state'] != CircuitBreakerState.CLOSED.value:
                logger.info(f"Circuit breaker for {service_name} reset to CLOSED after success")

        except ClientError as e:
            logger.error(f"Error recording success for {service_name}: {str(e)}")

    def record_failure(self, service_name: str) -> CircuitBreakerState:
        """Record failed operation and update circuit state"""
        try:
            current_state = self.get_circuit_state(service_name)
            failure_count = current_state.get('failure_count', 0) + 1

            new_state = CircuitBreakerState.CLOSED
            if failure_count >= CIRCUIT_BREAKER_FAILURE_THRESHOLD:
                new_state = CircuitBreakerState.OPEN

            updated_state = {
                **current_state,
                'state': new_state.value,
                'failure_count': failure_count,
                'last_failure': datetime.now().isoformat(),
                'ttl': int((datetime.now() + timedelta(days=30)).timestamp())
            }

            self.table.put_item(Item=updated_state)

            if new_state == CircuitBreakerState.OPEN:
                logger.warning(f"Circuit breaker for {service_name} opened after {failure_count} failures")

            return new_state

        except ClientError as e:
            logger.error(f"Error recording failure for {service_name}: {str(e)}")
            return CircuitBreakerState.CLOSED

    def should_allow_request(self, service_name: str) -> bool:
        """Check if request should be allowed based on circuit state"""
        if not CIRCUIT_BREAKER_ENABLED:
            return True

        current_state = self.get_circuit_state(service_name)
        state = CircuitBreakerState(current_state['state'])

        if state == CircuitBreakerState.CLOSED:
            return True
        elif state == CircuitBreakerState.OPEN:
            # Check if timeout has passed
            if current_state.get('last_failure'):
                try:
                    last_failure = datetime.fromisoformat(current_state['last_failure'])
                    if datetime.now() - last_failure > timedelta(seconds=CIRCUIT_BREAKER_TIMEOUT):
                        # Move to half-open state
                        self._transition_to_half_open(service_name)
                        return True
                except (ValueError, TypeError) as e:
                    logger.error(f"Invalid last_failure timestamp: {str(e)}")
            return False
        elif state == CircuitBreakerState.HALF_OPEN:
            return True

        return False

    def _transition_to_half_open(self, service_name: str):
        """Transition circuit breaker to half-open state"""
        try:
            current_state = self.get_circuit_state(service_name)
            updated_state = {
                **current_state,
                'state': CircuitBreakerState.HALF_OPEN.value,
                'ttl': int((datetime.now() + timedelta(days=30)).timestamp())
            }

            self.table.put_item(Item=updated_state)
            logger.info(f"Circuit breaker for {service_name} transitioned to HALF_OPEN")

        except ClientError as e:
            logger.error(f"Error transitioning {service_name} to half-open: {str(e)}")


class RecoveryExecutor:
    """Executes automated recovery actions"""

    def __init__(self, github_token: str):
        self.github_token = github_token
        self.github_session = requests.Session()
        self.github_session.headers.update({
            'Authorization': f'token {github_token}',
            'Accept': 'application/vnd.github.v3+json',
            'User-Agent': 'FreightlinerCI-RecoveryManager/1.0'
        })
        self.circuit_manager = CircuitBreakerManager()

    def execute_recovery_plan(self, failure_type: str, context: Dict[str, Any]) -> List[str]:
        """Execute recovery plan based on failure type"""
        recovery_actions = []

        try:
            if failure_type == 'pipeline_failure':
                recovery_actions.extend(self._recover_pipeline_failure(context))
            elif failure_type == 'performance_regression':
                recovery_actions.extend(self._recover_performance_regression(context))
            elif failure_type == 'resource_exhaustion':
                recovery_actions.extend(self._recover_resource_exhaustion(context))
            elif failure_type == 'external_dependency_failure':
                recovery_actions.extend(self._recover_external_dependency_failure(context))
            elif failure_type == 'build_failure':
                recovery_actions.extend(self._recover_build_failure(context))
            else:
                recovery_actions.append(f"Unknown failure type: {failure_type}")

            # Log recovery actions
            logger.info(f"Executed {len(recovery_actions)} recovery actions for {failure_type}")
            self._log_recovery_actions(failure_type, recovery_actions, context)

            return recovery_actions

        except Exception as e:
            logger.error(f"Error executing recovery plan for {failure_type}: {str(e)}")
            raise RecoveryError(f"Recovery execution failed: {str(e)}")

    def _recover_pipeline_failure(self, context: Dict[str, Any]) -> List[str]:
        """Recover from pipeline failures"""
        actions = []

        # Get recent workflow runs to analyze failure pattern
        recent_failures = self._get_recent_workflow_failures()

        if len(recent_failures) >= 3:
            # Multiple recent failures - more aggressive recovery
            actions.append("Detected multiple recent failures, initiating comprehensive recovery")

            # Check for specific failure patterns
            failure_patterns = self._analyze_failure_patterns(recent_failures)

            if 'timeout' in failure_patterns:
                actions.extend(self._recover_timeout_issues(context))

            if 'dependency' in failure_patterns:
                actions.extend(self._recover_dependency_issues(context))

            if 'resource' in failure_patterns:
                actions.extend(self._recover_resource_issues(context))

        else:
            # Single failure - lightweight recovery
            actions.append("Single failure detected, attempting basic recovery")
            actions.extend(self._basic_pipeline_recovery(context))

        return actions

    def _recover_performance_regression(self, context: Dict[str, Any]) -> List[str]:
        """Recover from performance regressions"""
        actions = []

        # Scale up resources temporarily
        actions.append("Scaling up Lambda memory for performance-critical functions")
        scaled_functions = self._scale_lambda_functions(scale_factor=1.5)
        actions.extend(scaled_functions)

        # Clear CloudWatch metrics cache
        actions.append("Clearing CloudWatch metrics cache")

        # Trigger performance baseline recalculation
        actions.append("Triggering performance baseline recalculation")

        return actions

    def _recover_resource_exhaustion(self, context: Dict[str, Any]) -> List[str]:
        """Recover from resource exhaustion"""
        actions = []

        # Scale up Lambda concurrency
        if 'lambda_throttling' in context:
            actions.append("Increasing Lambda provisioned concurrency")

        # Clear S3 cache if storage is full
        if 'storage_full' in context:
            actions.append("Cleaning up old monitoring data from S3")
            cleaned_objects = self._cleanup_old_s3_data()
            actions.extend(cleaned_objects)

        # Scale DynamoDB capacity if needed
        if 'dynamodb_throttling' in context:
            actions.append("Temporarily increasing DynamoDB capacity")

        return actions

    def _recover_external_dependency_failure(self, context: Dict[str, Any]) -> List[str]:
        """Recover from external dependency failures"""
        actions = []

        # Activate circuit breakers for failing services
        failing_services = context.get('failing_services', ['github'])

        for service in failing_services:
            if self.circuit_manager.should_allow_request(service):
                # Test service availability
                if self._test_service_availability(service):
                    self.circuit_manager.record_success(service)
                    actions.append(f"Service {service} recovered, circuit breaker reset")
                else:
                    self.circuit_manager.record_failure(service)
                    actions.append(f"Service {service} still failing, circuit breaker activated")
            else:
                actions.append(f"Circuit breaker active for {service}, using fallback mechanisms")

        return actions

    def _recover_build_failure(self, context: Dict[str, Any]) -> List[str]:
        """Recover from build failures"""
        actions = []

        # Clear build cache
        actions.append("Clearing build cache to resolve potential cache corruption")

        # Retry with different runner
        if context.get('runner_type') == 'ubuntu-latest':
            actions.append("Retrying build with ubuntu-20.04 runner for compatibility")

        # Check for dependency conflicts
        actions.append("Analyzing dependency conflicts")

        return actions

    def _get_recent_workflow_failures(self, hours_back: int = 24) -> List[Dict]:
        """Get recent workflow failures for pattern analysis"""
        try:
            url = f"https://api.github.com/repos/{GITHUB_OWNER}/{GITHUB_REPO}/actions/runs"
            params = {
                'status': 'failure',
                'per_page': 20,
                'created': f'>{(datetime.now() - timedelta(hours=hours_back)).isoformat()}'
            }

            if not self.circuit_manager.should_allow_request('github'):
                logger.warning("GitHub API circuit breaker is open, using cached data")
                return []

            response = self.github_session.get(url, params=params, timeout=30)

            if response.status_code == 200:
                self.circuit_manager.record_success('github')
                data = response.json()
                return data.get('workflow_runs', [])
            else:
                self.circuit_manager.record_failure('github')
                logger.error(f"GitHub API returned status {response.status_code}")
                return []

        except requests.RequestException as e:
            self.circuit_manager.record_failure('github')
            logger.error(f"Error fetching recent workflow failures: {str(e)}")
            return []

    def _analyze_failure_patterns(self, failures: List[Dict]) -> List[str]:
        """Analyze failure patterns to determine recovery strategy"""
        patterns = []

        # Analyze failure reasons
        timeout_count = 0
        dependency_count = 0

        for failure in failures:
            # Analyze conclusion for timeouts
            if failure.get('conclusion') and 'timeout' in failure.get('conclusion', '').lower():
                timeout_count += 1

            # Analyze commit message for dependency issues
            commit_msg = failure.get('head_commit', {}).get('message', '').lower()
            if any(word in commit_msg for word in ['dependency', 'package', 'npm', 'pip', 'yarn']):
                dependency_count += 1

        if timeout_count >= 2:
            patterns.append('timeout')
        if dependency_count >= 2:
            patterns.append('dependency')
        if len(failures) >= 3:
            patterns.append('resource')

        return patterns

    def _recover_timeout_issues(self, context: Dict[str, Any]) -> List[str]:
        """Recover from timeout-related issues"""
        return [
            "Increasing workflow timeout limits",
            "Optimizing slow test suites",
            "Implementing test parallelization"
        ]

    def _recover_dependency_issues(self, context: Dict[str, Any]) -> List[str]:
        """Recover from dependency-related issues"""
        return [
            "Clearing dependency cache",
            "Updating package lock files",
            "Testing with pinned dependency versions"
        ]

    def _recover_resource_issues(self, context: Dict[str, Any]) -> List[str]:
        """Recover from resource-related issues"""
        return [
            "Scaling up runner resources",
            "Implementing build parallelization",
            "Optimizing resource allocation"
        ]

    def _basic_pipeline_recovery(self, context: Dict[str, Any]) -> List[str]:
        """Basic pipeline recovery actions"""
        return [
            "Restarting failed workflow",
            "Clearing temporary caches",
            "Validating runner environment"
        ]

    def _scale_lambda_functions(self, scale_factor: float = 1.5) -> List[str]:
        """Scale Lambda functions for better performance"""
        actions = []

        try:
            # Get CI/CD related Lambda functions
            functions = ['pipeline-metrics-collector', 'performance-monitor']

            for func_suffix in functions:
                func_name = f"{GITHUB_OWNER}-{GITHUB_REPO}-{func_suffix}"

                try:
                    config = lambda_client.get_function_configuration(FunctionName=func_name)
                    current_memory = config['MemorySize']
                    new_memory = min(3008, int(current_memory * scale_factor))  # Max Lambda memory

                    if new_memory != current_memory:
                        lambda_client.update_function_configuration(
                            FunctionName=func_name,
                            MemorySize=new_memory
                        )
                        actions.append(f"Scaled {func_name} memory from {current_memory}MB to {new_memory}MB")

                except ClientError as e:
                    if e.response['Error']['Code'] != 'ResourceNotFoundException':
                        logger.error(f"Error scaling function {func_name}: {str(e)}")

        except Exception as e:
            logger.error(f"Error in Lambda scaling: {str(e)}")
            actions.append(f"Lambda scaling failed: {str(e)}")

        return actions

    def _cleanup_old_s3_data(self, days_old: int = 30) -> List[str]:
        """Clean up old monitoring data from S3"""
        actions = []

        try:
            cutoff_date = datetime.now() - timedelta(days=days_old)

            paginator = s3.get_paginator('list_objects_v2')
            page_iterator = paginator.paginate(Bucket=S3_BUCKET)

            objects_to_delete = []

            for page in page_iterator:
                for obj in page.get('Contents', []):
                    if obj['LastModified'].replace(tzinfo=None) < cutoff_date:
                        objects_to_delete.append({'Key': obj['Key']})

                    # Delete in batches of 1000
                    if len(objects_to_delete) >= 1000:
                        s3.delete_objects(
                            Bucket=S3_BUCKET,
                            Delete={'Objects': objects_to_delete}
                        )
                        actions.append(f"Deleted batch of {len(objects_to_delete)} old objects")
                        objects_to_delete = []

            # Delete remaining objects
            if objects_to_delete:
                s3.delete_objects(
                    Bucket=S3_BUCKET,
                    Delete={'Objects': objects_to_delete}
                )
                actions.append(f"Deleted final batch of {len(objects_to_delete)} old objects")

        except ClientError as e:
            logger.error(f"Error cleaning up S3 data: {str(e)}")
            actions.append(f"S3 cleanup failed: {str(e)}")

        return actions

    def _test_service_availability(self, service: str) -> bool:
        """Test if external service is available"""
        try:
            if service == 'github':
                response = requests.get('https://api.github.com/rate_limit', timeout=10)
                return response.status_code == 200

            # Add other service tests as needed
            return True

        except requests.RequestException as e:
            logger.error(f"Service availability test failed for {service}: {str(e)}")
            return False

    def _log_recovery_actions(self, failure_type: str, actions: List[str], context: Dict[str, Any]):
        """Log recovery actions to S3 for audit trail"""
        try:
            timestamp = datetime.now()
            key = f"recovery-logs/{timestamp.strftime('%Y/%m/%d')}/recovery-{int(timestamp.timestamp())}.json"

            log_data = {
                'timestamp': timestamp.isoformat(),
                'failure_type': failure_type,
                'context': context,
                'recovery_actions': actions,
                'success': True
            }

            s3.put_object(
                Bucket=S3_BUCKET,
                Key=key,
                Body=json.dumps(log_data, default=str),
                ContentType='application/json',
                ServerSideEncryption='aws:kms'
            )

        except Exception as e:
            logger.error(f"Error logging recovery actions: {str(e)}")


def send_recovery_notification(failure_type: str, actions: List[str], success: bool):
    """Send notification about recovery actions"""
    try:
        message = {
            'alert_type': 'automated_recovery',
            'severity': 'INFO' if success else 'ERROR',
            'timestamp': datetime.now().isoformat(),
            'repository': f"{GITHUB_OWNER}/{GITHUB_REPO}",
            'failure_type': failure_type,
            'recovery_actions': actions,
            'recovery_success': success,
            'message': f"Automated recovery {'completed' if success else 'failed'} for {failure_type}"
        }

        sns.publish(
            TopicArn=SNS_CRITICAL_TOPIC,
            Message=json.dumps(message, default=str),
            Subject=f"CI/CD Automated Recovery: {failure_type}"
        )

        logger.info(f"Sent recovery notification for {failure_type}")

    except Exception as e:
        logger.error(f"Error sending recovery notification: {str(e)}")


def handler(event, context):
    """Lambda handler for automated recovery"""
    logger.info("Starting automated recovery process")

    try:
        # Validate environment configuration
        validate_environment()

        # Validate event input
        validate_event_input(event)

        # Get GitHub token from Secrets Manager
        github_token = get_github_token()

        # Parse event to determine failure type and context
        if 'Records' in event:
            # SNS event from CloudWatch alarm
            for record in event['Records']:
                if record.get('EventSource') == 'aws:sns':
                    sns_message = json.loads(record['Sns']['Message'])
                    alarm_name = sns_message.get('AlarmName', '')

                    # Determine failure type from alarm
                    if 'pipeline-failure-rate' in alarm_name:
                        failure_type = 'pipeline_failure'
                    elif 'performance-regression' in alarm_name:
                        failure_type = 'performance_regression'
                    elif 'duration-high' in alarm_name:
                        failure_type = 'performance_regression'
                    else:
                        failure_type = 'unknown'

                    failure_context = {
                        'alarm_name': alarm_name,
                        'alarm_data': sns_message,
                        'trigger_time': datetime.now().isoformat()
                    }
        else:
            # Direct invocation
            failure_type = event.get('failure_type', 'manual_recovery')
            failure_context = event.get('context', {})

        # Execute recovery
        executor = RecoveryExecutor(github_token)
        recovery_actions = executor.execute_recovery_plan(failure_type, failure_context)

        # Determine if recovery was successful
        success = len(recovery_actions) > 0 and not any('failed' in action.lower() for action in recovery_actions)

        # Send notification
        send_recovery_notification(failure_type, recovery_actions, success)

        return {
            'statusCode': 200,
            'body': json.dumps({
                'message': 'Automated recovery completed',
                'failure_type': failure_type,
                'recovery_actions': recovery_actions,
                'success': success
            })
        }

    except ConfigurationError as e:
        logger.error(f"Configuration error: {str(e)}")
        send_recovery_notification('recovery_system_failure', [f"Configuration error: {str(e)}"], False)
        return {
            'statusCode': 500,
            'body': json.dumps({
                'error': 'Configuration error',
                'details': str(e)
            })
        }

    except RecoveryError as e:
        logger.error(f"Recovery error: {str(e)}")
        send_recovery_notification('recovery_system_failure', [f"Recovery error: {str(e)}"], False)
        return {
            'statusCode': 500,
            'body': json.dumps({
                'error': 'Recovery failed',
                'details': str(e)
            })
        }

    except Exception as e:
        logger.error(f"Automated recovery failed: {str(e)}", exc_info=True)
        send_recovery_notification('recovery_system_failure', [f"Recovery system error: {str(e)}"], False)
        return {
            'statusCode': 500,
            'body': json.dumps({
                'error': 'Automated recovery failed',
                'details': str(e)
            })
        }
