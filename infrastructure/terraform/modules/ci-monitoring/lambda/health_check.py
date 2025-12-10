"""
Health Check Endpoint for CI/CD Monitoring System
Provides comprehensive health status for all monitoring components
"""

import json
import boto3
import os
import logging
from datetime import datetime, timedelta
from typing import Dict, List, Any, Optional
from botocore.exceptions import ClientError

# Configure logging
logging.basicConfig(level=os.getenv('LOG_LEVEL', 'INFO'))
logger = logging.getLogger(__name__)

# AWS clients
dynamodb = boto3.resource('dynamodb')
s3 = boto3.client('s3')
lambda_client = boto3.client('lambda')
cloudwatch = boto3.client('cloudwatch')

# Configuration
DYNAMODB_TABLE_METADATA = os.getenv('DYNAMODB_TABLE_METADATA')
DYNAMODB_TABLE_CIRCUIT_BREAKER = os.getenv('DYNAMODB_TABLE_CIRCUIT_BREAKER')
S3_BUCKET = os.getenv('S3_BUCKET')
GITHUB_OWNER = os.getenv('GITHUB_OWNER')
GITHUB_REPO = os.getenv('GITHUB_REPO')

class HealthChecker:
    """Comprehensive health checker for monitoring system components"""
    
    def __init__(self):
        self.health_status = {
            'overall': 'unknown',
            'components': {},
            'timestamp': datetime.now().isoformat(),
            'uptime_seconds': 0
        }
    
    def check_all_components(self) -> Dict[str, Any]:
        """Check health of all monitoring components"""
        logger.info("Starting comprehensive health check")
        
        # Check each component
        self.health_status['components']['dynamodb'] = self._check_dynamodb_health()
        self.health_status['components']['s3'] = self._check_s3_health()
        self.health_status['components']['lambda_functions'] = self._check_lambda_health()
        self.health_status['components']['cloudwatch'] = self._check_cloudwatch_health()
        self.health_status['components']['circuit_breakers'] = self._check_circuit_breaker_health()
        self.health_status['components']['data_freshness'] = self._check_data_freshness()
        
        # Calculate overall health
        self.health_status['overall'] = self._calculate_overall_health()
        
        # Add summary statistics
        self.health_status['summary'] = self._generate_health_summary()
        
        return self.health_status
    
    def _check_dynamodb_health(self) -> Dict[str, Any]:
        """Check DynamoDB table health"""
        status = {
            'status': 'unknown',
            'details': {},
            'errors': []
        }
        
        try:
            # Check pipeline metadata table
            if DYNAMODB_TABLE_METADATA:
                table = dynamodb.Table(DYNAMODB_TABLE_METADATA)
                
                # Check table status
                table_info = table.meta.client.describe_table(TableName=DYNAMODB_TABLE_METADATA)
                table_status = table_info['Table']['TableStatus']
                
                status['details']['metadata_table_status'] = table_status
                
                # Check recent activity
                response = table.scan(
                    Limit=10,
                    FilterExpression='#ts > :recent_time',
                    ExpressionAttributeNames={'#ts': 'timestamp'},
                    ExpressionAttributeValues={':recent_time': (datetime.now() - timedelta(hours=1)).isoformat()}
                )
                
                status['details']['recent_records'] = response['Count']
                
                if table_status == 'ACTIVE' and response['Count'] > 0:
                    status['status'] = 'healthy'
                elif table_status == 'ACTIVE':
                    status['status'] = 'warning'
                    status['errors'].append('No recent activity in metadata table')
                else:
                    status['status'] = 'unhealthy'
                    status['errors'].append(f'Table status: {table_status}')
            
            # Check circuit breaker table if it exists
            if DYNAMODB_TABLE_CIRCUIT_BREAKER:
                try:
                    cb_table = dynamodb.Table(DYNAMODB_TABLE_CIRCUIT_BREAKER)
                    cb_info = cb_table.meta.client.describe_table(TableName=DYNAMODB_TABLE_CIRCUIT_BREAKER)
                    status['details']['circuit_breaker_table_status'] = cb_info['Table']['TableStatus']
                except ClientError:
                    status['details']['circuit_breaker_table_status'] = 'not_found'
        
        except ClientError as e:
            status['status'] = 'unhealthy'
            status['errors'].append(f'DynamoDB error: {str(e)}')
        
        return status
    
    def _check_s3_health(self) -> Dict[str, Any]:
        """Check S3 bucket health"""
        status = {
            'status': 'unknown',
            'details': {},
            'errors': []
        }
        
        try:
            # Check bucket accessibility
            s3.head_bucket(Bucket=S3_BUCKET)
            
            # Check recent uploads
            response = s3.list_objects_v2(
                Bucket=S3_BUCKET,
                MaxKeys=10
            )
            
            object_count = response.get('KeyCount', 0)
            status['details']['recent_objects'] = object_count
            
            # Check bucket size (approximate)
            total_size = sum(obj.get('Size', 0) for obj in response.get('Contents', []))
            status['details']['approximate_size_bytes'] = total_size
            
            # Check for recent monitoring data
            recent_data = s3.list_objects_v2(
                Bucket=S3_BUCKET,
                Prefix=f'pipeline-metrics/{datetime.now().strftime("%Y/%m/%d")}',
                MaxKeys=5
            )
            
            status['details']['todays_data_files'] = recent_data.get('KeyCount', 0)
            
            if object_count > 0:
                status['status'] = 'healthy'
            else:
                status['status'] = 'warning'
                status['errors'].append('No objects found in monitoring bucket')
        
        except ClientError as e:
            status['status'] = 'unhealthy'
            status['errors'].append(f'S3 error: {str(e)}')
        
        return status
    
    def _check_lambda_health(self) -> Dict[str, Any]:
        """Check Lambda functions health"""
        status = {
            'status': 'unknown',
            'details': {},
            'errors': []
        }
        
        expected_functions = [
            'pipeline-metrics-collector',
            'performance-monitor',
            'cost-optimizer',
            'recovery-manager',
            'health-check'
        ]
        
        function_statuses = {}
        healthy_count = 0
        
        for func_suffix in expected_functions:
            try:
                # Try both naming patterns
                possible_names = [
                    f"{GITHUB_OWNER}-{GITHUB_REPO}-{func_suffix}",
                    f"freightliner-{func_suffix}",
                    func_suffix
                ]
                
                function_found = False
                for func_name in possible_names:
                    try:
                        config = lambda_client.get_function_configuration(FunctionName=func_name)
                        
                        func_status = {
                            'state': config.get('State', 'Unknown'),
                            'last_modified': config.get('LastModified', 'Unknown'),
                            'memory_size': config.get('MemorySize', 0),
                            'timeout': config.get('Timeout', 0),
                            'runtime': config.get('Runtime', 'Unknown')
                        }
                        
                        # Check recent invocations
                        try:
                            end_time = datetime.now()
                            start_time = end_time - timedelta(hours=1)
                            
                            invocations = cloudwatch.get_metric_statistics(
                                Namespace='AWS/Lambda',
                                MetricName='Invocations',
                                Dimensions=[{'Name': 'FunctionName', 'Value': func_name}],
                                StartTime=start_time,
                                EndTime=end_time,
                                Period=3600,
                                Statistics=['Sum']
                            )
                            
                            total_invocations = sum(dp['Sum'] for dp in invocations['Datapoints'])
                            func_status['recent_invocations'] = total_invocations
                            
                            # Check for errors
                            errors = cloudwatch.get_metric_statistics(
                                Namespace='AWS/Lambda',
                                MetricName='Errors',
                                Dimensions=[{'Name': 'FunctionName', 'Value': func_name}],
                                StartTime=start_time,
                                EndTime=end_time,
                                Period=3600,
                                Statistics=['Sum']
                            )
                            
                            total_errors = sum(dp['Sum'] for dp in errors['Datapoints'])
                            func_status['recent_errors'] = total_errors
                            
                            if config.get('State') == 'Active' and total_errors == 0:
                                func_status['health'] = 'healthy'
                                healthy_count += 1
                            elif total_errors > 0:
                                func_status['health'] = 'warning'
                                func_status['error_details'] = f'{total_errors} errors in last hour'
                            else:
                                func_status['health'] = 'unknown'
                            
                        except ClientError:
                            func_status['health'] = 'unknown'
                            func_status['monitoring_note'] = 'Unable to fetch metrics'
                        
                        function_statuses[func_suffix] = func_status
                        function_found = True
                        break
                        
                    except ClientError as e:
                        if e.response['Error']['Code'] != 'ResourceNotFoundException':
                            logger.error(f"Error checking function {func_name}: {str(e)}")
                
                if not function_found:
                    function_statuses[func_suffix] = {
                        'health': 'missing',
                        'error': 'Function not found with any expected naming pattern'
                    }
            
            except Exception as e:
                function_statuses[func_suffix] = {
                    'health': 'error',
                    'error': str(e)
                }
        
        status['details']['functions'] = function_statuses
        status['details']['healthy_count'] = healthy_count
        status['details']['total_expected'] = len(expected_functions)
        
        if healthy_count == len(expected_functions):
            status['status'] = 'healthy'
        elif healthy_count > len(expected_functions) // 2:
            status['status'] = 'warning'
            status['errors'].append(f'Only {healthy_count}/{len(expected_functions)} functions are healthy')
        else:
            status['status'] = 'unhealthy'
            status['errors'].append(f'Only {healthy_count}/{len(expected_functions)} functions are healthy')
        
        return status
    
    def _check_cloudwatch_health(self) -> Dict[str, Any]:
        """Check CloudWatch metrics health"""
        status = {
            'status': 'unknown',
            'details': {},
            'errors': []
        }
        
        try:
            # Check for recent metrics in our namespaces
            namespaces = ['CI-CD/Pipeline', 'CI-CD/Performance', 'CI-CD/Cost']
            
            metrics_found = 0
            recent_datapoints = 0
            
            end_time = datetime.now()
            start_time = end_time - timedelta(hours=1)
            
            for namespace in namespaces:
                try:
                    # List metrics in namespace
                    metrics_response = cloudwatch.list_metrics(Namespace=namespace)
                    namespace_metrics = len(metrics_response.get('Metrics', []))
                    metrics_found += namespace_metrics
                    
                    # Check for recent datapoints
                    for metric in metrics_response.get('Metrics', [])[:5]:  # Check first 5 metrics
                        try:
                            stats = cloudwatch.get_metric_statistics(
                                Namespace=namespace,
                                MetricName=metric['MetricName'],
                                Dimensions=metric.get('Dimensions', []),
                                StartTime=start_time,
                                EndTime=end_time,
                                Period=300,
                                Statistics=['Average']
                            )
                            recent_datapoints += len(stats.get('Datapoints', []))
                        except ClientError:
                            continue
                
                except ClientError as e:
                    status['errors'].append(f'Error checking namespace {namespace}: {str(e)}')
            
            status['details']['total_metrics'] = metrics_found
            status['details']['recent_datapoints'] = recent_datapoints
            
            if metrics_found > 0 and recent_datapoints > 0:
                status['status'] = 'healthy'
            elif metrics_found > 0:
                status['status'] = 'warning'
                status['errors'].append('Metrics exist but no recent datapoints')
            else:
                status['status'] = 'unhealthy'
                status['errors'].append('No monitoring metrics found')
        
        except ClientError as e:
            status['status'] = 'unhealthy'
            status['errors'].append(f'CloudWatch error: {str(e)}')
        
        return status
    
    def _check_circuit_breaker_health(self) -> Dict[str, Any]:
        """Check circuit breaker health"""
        status = {
            'status': 'unknown',
            'details': {},
            'errors': []
        }
        
        try:
            if DYNAMODB_TABLE_CIRCUIT_BREAKER:
                table = dynamodb.Table(DYNAMODB_TABLE_CIRCUIT_BREAKER)
                
                # Scan for circuit breaker states
                response = table.scan(Limit=50)
                
                circuit_breakers = {}
                open_breakers = 0
                
                for item in response.get('Items', []):
                    service_name = item.get('service_name')
                    cb_state = item.get('state', 'UNKNOWN')
                    
                    circuit_breakers[service_name] = {
                        'state': cb_state,
                        'failure_count': item.get('failure_count', 0),
                        'last_failure': item.get('last_failure'),
                        'last_success': item.get('last_success')
                    }
                    
                    if cb_state == 'OPEN':
                        open_breakers += 1
                
                status['details']['circuit_breakers'] = circuit_breakers
                status['details']['total_breakers'] = len(circuit_breakers)
                status['details']['open_breakers'] = open_breakers
                
                if open_breakers == 0:
                    status['status'] = 'healthy'
                elif open_breakers <= len(circuit_breakers) // 2:
                    status['status'] = 'warning'
                    status['errors'].append(f'{open_breakers} circuit breakers are open')
                else:
                    status['status'] = 'unhealthy'
                    status['errors'].append(f'{open_breakers} circuit breakers are open')
            else:
                status['status'] = 'disabled'
                status['details']['note'] = 'Circuit breaker functionality is disabled'
        
        except ClientError as e:
            status['status'] = 'error'
            status['errors'].append(f'Circuit breaker check error: {str(e)}')
        
        return status
    
    def _check_data_freshness(self) -> Dict[str, Any]:
        """Check freshness of monitoring data"""
        status = {
            'status': 'unknown',
            'details': {},
            'errors': []
        }
        
        try:
            # Check pipeline metadata freshness
            if DYNAMODB_TABLE_METADATA:
                table = dynamodb.Table(DYNAMODB_TABLE_METADATA)
                
                # Get most recent record
                response = table.scan(
                    Limit=1,
                    ProjectionExpression='#ts, pipeline_id',
                    ExpressionAttributeNames={'#ts': 'timestamp'}
                )
                
                if response.get('Items'):
                    latest_record = response['Items'][0]
                    latest_timestamp = datetime.fromisoformat(latest_record['timestamp'])
                    age_minutes = (datetime.now() - latest_timestamp).total_seconds() / 60
                    
                    status['details']['latest_data_age_minutes'] = age_minutes
                    status['details']['latest_pipeline'] = latest_record.get('pipeline_id')
                    
                    if age_minutes <= 10:  # Data within 10 minutes
                        status['status'] = 'healthy'
                    elif age_minutes <= 60:  # Data within 1 hour
                        status['status'] = 'warning'
                        status['errors'].append(f'Latest data is {age_minutes:.1f} minutes old')
                    else:
                        status['status'] = 'stale'
                        status['errors'].append(f'Latest data is {age_minutes:.1f} minutes old')
                else:
                    status['status'] = 'no_data'
                    status['errors'].append('No pipeline metadata found')
            
            # Check S3 data freshness
            recent_s3_objects = s3.list_objects_v2(
                Bucket=S3_BUCKET,
                Prefix=f'pipeline-metrics/{datetime.now().strftime("%Y/%m/%d")}',
                MaxKeys=1
            )
            
            status['details']['todays_s3_objects'] = recent_s3_objects.get('KeyCount', 0)
            
        except ClientError as e:
            status['status'] = 'error'
            status['errors'].append(f'Data freshness check error: {str(e)}')
        
        return status
    
    def _calculate_overall_health(self) -> str:
        """Calculate overall system health based on component health"""
        component_healths = []
        
        for component, health_info in self.health_status['components'].items():
            health = health_info.get('status', 'unknown')
            component_healths.append(health)
        
        # Count health statuses
        healthy_count = component_healths.count('healthy')
        warning_count = component_healths.count('warning')
        unhealthy_count = component_healths.count('unhealthy')
        error_count = component_healths.count('error')
        
        total_components = len(component_healths)
        
        # Determine overall health
        if unhealthy_count > 0 or error_count > 0:
            return 'unhealthy'
        elif warning_count > total_components // 2:
            return 'degraded'
        elif healthy_count >= total_components // 2:
            return 'healthy'
        else:
            return 'unknown'
    
    def _generate_health_summary(self) -> Dict[str, Any]:
        """Generate summary statistics"""
        component_statuses = [
            comp['status'] for comp in self.health_status['components'].values()
        ]
        
        return {
            'total_components': len(component_statuses),
            'healthy_components': component_statuses.count('healthy'),
            'warning_components': component_statuses.count('warning'),
            'unhealthy_components': component_statuses.count('unhealthy'),
            'error_components': component_statuses.count('error'),
            'repository': f"{GITHUB_OWNER}/{GITHUB_REPO}" if GITHUB_OWNER and GITHUB_REPO else 'unknown'
        }

def handler(event, context):
    """Lambda handler for health check endpoint"""
    logger.info("Health check requested")
    
    try:
        # Perform health check
        health_checker = HealthChecker()
        health_status = health_checker.check_all_components()
        
        # Determine HTTP status code based on health
        if health_status['overall'] == 'healthy':
            status_code = 200
        elif health_status['overall'] in ['warning', 'degraded']:
            status_code = 200  # Still OK, but with warnings
        else:
            status_code = 503  # Service unavailable
        
        return {
            'statusCode': status_code,
            'headers': {
                'Content-Type': 'application/json',
                'Cache-Control': 'no-cache',
                'X-Health-Status': health_status['overall']
            },
            'body': json.dumps(health_status, default=str, indent=2)
        }
        
    except Exception as e:
        logger.error(f"Health check failed: {str(e)}")
        
        error_response = {
            'overall': 'error',
            'error': str(e),
            'timestamp': datetime.now().isoformat(),
            'message': 'Health check system failure'
        }
        
        return {
            'statusCode': 500,
            'headers': {
                'Content-Type': 'application/json',
                'X-Health-Status': 'error'
            },
            'body': json.dumps(error_response, default=str, indent=2)
        }