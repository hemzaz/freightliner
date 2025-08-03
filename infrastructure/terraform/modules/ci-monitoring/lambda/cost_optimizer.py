"""
CI/CD Cost Optimizer
Tracks resource usage, identifies cost optimization opportunities, and implements automated scaling
"""

import json
import boto3
import os
import logging
from datetime import datetime, timedelta
from typing import Dict, List, Any, Optional, Tuple
from botocore.exceptions import ClientError
from dataclasses import dataclass
import statistics

# Configure logging
logging.basicConfig(level=os.getenv('LOG_LEVEL', 'INFO'))
logger = logging.getLogger(__name__)

# AWS clients
ce = boto3.client('ce')  # Cost Explorer
cloudwatch = boto3.client('cloudwatch')
lambda_client = boto3.client('lambda')
ec2 = boto3.client('ec2')
s3 = boto3.client('s3')
sns = boto3.client('sns')
application_autoscaling = boto3.client('application-autoscaling')

# Configuration
S3_BUCKET = os.getenv('S3_BUCKET')
SNS_COST_TOPIC = os.getenv('SNS_COST_TOPIC')
COST_THRESHOLD_USD = float(os.getenv('COST_THRESHOLD_USD', '100'))
AUTO_SCALING_ENABLED = os.getenv('AUTO_SCALING_ENABLED', 'true').lower() == 'true'
SCALING_TARGET_UTIL = float(os.getenv('SCALING_TARGET_UTIL', '70'))

@dataclass
class CostAnalysis:
    """Cost analysis result"""
    service: str
    current_cost: float
    previous_cost: float
    cost_change_percent: float
    optimization_potential: float
    recommendations: List[str]

@dataclass
class ResourceUtilization:
    """Resource utilization metrics"""
    resource_type: str
    resource_id: str
    avg_utilization: float
    max_utilization: float
    cost_per_hour: float
    optimization_score: float

class CostAnalyzer:
    """Analyzes AWS costs and identifies optimization opportunities"""
    
    def __init__(self):
        self.services_to_analyze = [
            'Amazon Elastic Compute Cloud - Compute',
            'AWS Lambda',
            'Amazon Simple Storage Service',
            'Amazon CloudWatch',
            'Amazon DynamoDB',
            'AWS CodeBuild',
            'GitHub Actions'  # Third-party service tracking
        ]
        
    def get_cost_and_usage(self, start_date: datetime, end_date: datetime) -> Dict[str, Any]:
        """Get cost and usage data from AWS Cost Explorer"""
        try:
            response = ce.get_cost_and_usage(
                TimePeriod={
                    'Start': start_date.strftime('%Y-%m-%d'),
                    'End': end_date.strftime('%Y-%m-%d')
                },
                Granularity='DAILY',
                Metrics=['BlendedCost', 'UsageQuantity'],
                GroupBy=[
                    {
                        'Type': 'DIMENSION',
                        'Key': 'SERVICE'
                    }
                ]
            )
            
            return response
            
        except ClientError as e:
            logger.error(f"Error getting cost and usage data: {str(e)}")
            return {}
    
    def analyze_service_costs(self, days_back: int = 30) -> List[CostAnalysis]:
        """Analyze costs by service and identify trends"""
        end_date = datetime.now().date()
        start_date = end_date - timedelta(days=days_back)
        mid_date = start_date + timedelta(days=days_back // 2)
        
        # Get current period data
        current_data = self.get_cost_and_usage(mid_date, end_date)
        # Get previous period data for comparison
        previous_data = self.get_cost_and_usage(start_date, mid_date)
        
        analyses = []
        
        if not current_data or not previous_data:
            logger.warning("Unable to retrieve cost data for analysis")
            return analyses
        
        # Process current period costs
        current_costs = self._extract_service_costs(current_data)
        previous_costs = self._extract_service_costs(previous_data)
        
        for service in self.services_to_analyze:
            current_cost = current_costs.get(service, 0.0)
            previous_cost = previous_costs.get(service, 0.0)
            
            if previous_cost > 0:
                cost_change_percent = ((current_cost - previous_cost) / previous_cost) * 100
            else:
                cost_change_percent = 100.0 if current_cost > 0 else 0.0
            
            # Generate optimization recommendations
            recommendations = self._generate_recommendations(service, current_cost, cost_change_percent)
            
            # Calculate optimization potential (estimated savings)
            optimization_potential = self._calculate_optimization_potential(service, current_cost)
            
            analysis = CostAnalysis(
                service=service,
                current_cost=current_cost,
                previous_cost=previous_cost,
                cost_change_percent=cost_change_percent,
                optimization_potential=optimization_potential,
                recommendations=recommendations
            )
            
            analyses.append(analysis)
        
        return analyses
    
    def _extract_service_costs(self, cost_data: Dict[str, Any]) -> Dict[str, float]:
        """Extract service costs from Cost Explorer response"""
        service_costs = {}
        
        for result in cost_data.get('ResultsByTime', []):
            for group in result.get('Groups', []):
                service_name = group.get('Keys', ['Unknown'])[0]
                amount = float(group.get('Metrics', {}).get('BlendedCost', {}).get('Amount', '0'))
                
                if service_name in service_costs:
                    service_costs[service_name] += amount
                else:
                    service_costs[service_name] = amount
        
        return service_costs
    
    def _generate_recommendations(self, service: str, cost: float, change_percent: float) -> List[str]:
        """Generate cost optimization recommendations for a service"""
        recommendations = []
        
        if 'Lambda' in service:
            if cost > 50:  # High Lambda costs
                recommendations.extend([
                    "Consider optimizing Lambda memory allocation based on actual usage",
                    "Review Lambda timeout settings to prevent unnecessary charges",
                    "Implement caching to reduce Lambda invocation frequency",
                    "Consider using Provisioned Concurrency only during peak hours"
                ])
            
            if change_percent > 20:
                recommendations.append("Lambda costs increased significantly - review recent deployment changes")
        
        elif 'Compute Cloud' in service:
            if cost > 100:
                recommendations.extend([
                    "Consider using Spot Instances for CI/CD workloads",
                    "Implement auto-shutdown for development environments",
                    "Right-size instances based on actual CPU/memory usage",
                    "Consider Reserved Instances for predictable workloads"
                ])
        
        elif 'CloudWatch' in service:
            if cost > 20:
                recommendations.extend([
                    "Review log retention policies - reduce retention for debug logs",
                    "Consider sampling metrics for high-volume applications",
                    "Use CloudWatch Insights queries efficiently",
                    "Archive old logs to S3 for long-term storage"
                ])
        
        elif 'Simple Storage Service' in service:
            if cost > 30:
                recommendations.extend([
                    "Implement S3 lifecycle policies to transition to cheaper storage classes",
                    "Clean up incomplete multipart uploads",
                    "Use S3 Intelligent Tiering for unpredictable access patterns",
                    "Compress artifacts before storing in S3"
                ])
        
        elif 'DynamoDB' in service:
            if cost > 25:
                recommendations.extend([
                    "Consider on-demand billing for unpredictable workloads",
                    "Optimize table design to reduce read/write operations",
                    "Use DynamoDB auto scaling for consistent workloads",
                    "Archive old data to S3 for cost savings"
                ])
        
        # General recommendations for cost increases
        if change_percent > 50:
            recommendations.append(f"Cost increased by {change_percent:.1f}% - investigate recent changes")
        elif change_percent > 25:
            recommendations.append(f"Moderate cost increase of {change_percent:.1f}% detected")
        
        return recommendations
    
    def _calculate_optimization_potential(self, service: str, current_cost: float) -> float:
        """Calculate potential cost savings for a service"""
        # Conservative estimates of potential savings
        optimization_rates = {
            'AWS Lambda': 0.15,  # 15% through memory optimization and caching
            'Amazon Elastic Compute Cloud - Compute': 0.30,  # 30% through spot instances and right-sizing
            'Amazon CloudWatch': 0.25,  # 25% through log retention and sampling
            'Amazon Simple Storage Service': 0.20,  # 20% through lifecycle policies
            'Amazon DynamoDB': 0.15,  # 15% through capacity optimization
            'AWS CodeBuild': 0.10   # 10% through build optimization
        }
        
        rate = optimization_rates.get(service, 0.10)  # Default 10% potential savings
        return current_cost * rate

class ResourceOptimizer:
    """Optimizes resource allocation and implements auto-scaling"""
    
    def __init__(self):
        self.optimization_actions = []
    
    def analyze_lambda_utilization(self) -> List[ResourceUtilization]:
        """Analyze Lambda function utilization and costs"""
        utilizations = []
        
        try:
            # Get list of Lambda functions
            paginator = lambda_client.get_paginator('list_functions')
            
            for page in paginator.paginate():
                for function in page['Functions']:
                    function_name = function['FunctionName']
                    
                    # Skip if not related to CI/CD monitoring
                    if not any(keyword in function_name.lower() for keyword in ['ci', 'cd', 'pipeline', 'monitor']):
                        continue
                    
                    utilization = self._analyze_lambda_function(function_name)
                    if utilization:
                        utilizations.append(utilization)
        
        except ClientError as e:
            logger.error(f"Error analyzing Lambda utilization: {str(e)}")
        
        return utilizations
    
    def _analyze_lambda_function(self, function_name: str) -> Optional[ResourceUtilization]:
        """Analyze individual Lambda function utilization"""
        try:
            # Get function configuration
            config = lambda_client.get_function_configuration(FunctionName=function_name)
            memory_mb = config['MemorySize']
            timeout_seconds = config['Timeout']
            
            # Get CloudWatch metrics for the function
            end_time = datetime.now()
            start_time = end_time - timedelta(days=7)  # Last 7 days
            
            # Get invocation count
            invocations_response = cloudwatch.get_metric_statistics(
                Namespace='AWS/Lambda',
                MetricName='Invocations',
                Dimensions=[{'Name': 'FunctionName', 'Value': function_name}],
                StartTime=start_time,
                EndTime=end_time,
                Period=3600,  # 1 hour
                Statistics=['Sum']
            )
            
            # Get duration metrics
            duration_response = cloudwatch.get_metric_statistics(
                Namespace='AWS/Lambda',
                MetricName='Duration',
                Dimensions=[{'Name': 'FunctionName', 'Value': function_name}],
                StartTime=start_time,
                EndTime=end_time,
                Period=3600,
                Statistics=['Average', 'Maximum']
            )
            
            if not duration_response['Datapoints']:
                return None
            
            # Calculate utilization metrics
            avg_duration = statistics.mean([dp['Average'] for dp in duration_response['Datapoints']])
            max_duration = max([dp['Maximum'] for dp in duration_response['Datapoints']])
            
            # Calculate utilization as percentage of timeout used
            avg_utilization = (avg_duration / (timeout_seconds * 1000)) * 100  # Convert to percentage
            max_utilization = (max_duration / (timeout_seconds * 1000)) * 100
            
            # Estimate cost (rough calculation)
            total_invocations = sum([dp['Sum'] for dp in invocations_response['Datapoints']])
            gb_seconds = (memory_mb / 1024) * (avg_duration / 1000) * total_invocations
            estimated_cost = gb_seconds * 0.0000166667  # AWS Lambda pricing (approximate)
            
            # Calculate optimization score (lower is better for optimization)
            optimization_score = self._calculate_lambda_optimization_score(
                avg_utilization, memory_mb, avg_duration, estimated_cost
            )
            
            return ResourceUtilization(
                resource_type='lambda',
                resource_id=function_name,
                avg_utilization=avg_utilization,
                max_utilization=max_utilization,
                cost_per_hour=estimated_cost / (7 * 24),  # Weekly cost divided by hours
                optimization_score=optimization_score
            )
            
        except ClientError as e:
            logger.error(f"Error analyzing Lambda function {function_name}: {str(e)}")
            return None
    
    def _calculate_lambda_optimization_score(self, avg_util: float, memory_mb: int, avg_duration: float, cost: float) -> float:
        """Calculate optimization score for Lambda function"""
        score = 0
        
        # Memory over-provisioning penalty
        if avg_util < 30:  # Low utilization
            score += 3
        elif avg_util < 50:
            score += 2
        elif avg_util < 70:
            score += 1
        
        # High memory allocation penalty for low-utilization functions
        if memory_mb > 1024 and avg_util < 50:
            score += 2
        
        # High cost penalty
        if cost > 10:  # Weekly cost > $10
            score += 2
        elif cost > 5:
            score += 1
        
        return score
    
    def implement_optimizations(self, utilizations: List[ResourceUtilization]) -> List[str]:
        """Implement resource optimizations based on utilization analysis"""
        actions = []
        
        for util in utilizations:
            if util.resource_type == 'lambda' and util.optimization_score >= 3:
                action = self._optimize_lambda_function(util)
                if action:
                    actions.append(action)
        
        return actions
    
    def _optimize_lambda_function(self, util: ResourceUtilization) -> Optional[str]:
        """Optimize Lambda function configuration"""
        try:
            current_config = lambda_client.get_function_configuration(FunctionName=util.resource_id)
            current_memory = current_config['MemorySize']
            
            # Calculate optimal memory based on utilization
            if util.avg_utilization < 30:
                # Significantly over-provisioned
                new_memory = max(128, int(current_memory * 0.7))  # Reduce by 30%
            elif util.avg_utilization < 50:
                # Moderately over-provisioned
                new_memory = max(128, int(current_memory * 0.85))  # Reduce by 15%
            else:
                return None  # No optimization needed
            
            # Only update if the change is significant
            if abs(current_memory - new_memory) < 64:  # Less than 64MB difference
                return None
            
            # Update function configuration
            lambda_client.update_function_configuration(
                FunctionName=util.resource_id,
                MemorySize=new_memory
            )
            
            estimated_savings = util.cost_per_hour * 24 * 30 * 0.15  # Estimated 15% monthly savings
            
            action = f"Optimized {util.resource_id}: reduced memory from {current_memory}MB to {new_memory}MB (estimated monthly savings: ${estimated_savings:.2f})"
            logger.info(action)
            return action
            
        except ClientError as e:
            logger.error(f"Error optimizing Lambda function {util.resource_id}: {str(e)}")
            return None

def save_cost_analysis_to_s3(analyses: List[CostAnalysis], utilizations: List[ResourceUtilization]):
    """Save cost analysis results to S3"""
    try:
        timestamp = datetime.now()
        key = f"cost-analysis/{timestamp.strftime('%Y/%m/%d')}/analysis-{int(timestamp.timestamp())}.json"
        
        data = {
            'timestamp': timestamp.isoformat(),
            'cost_analyses': [
                {
                    'service': a.service,
                    'current_cost': a.current_cost,
                    'previous_cost': a.previous_cost,
                    'cost_change_percent': a.cost_change_percent,
                    'optimization_potential': a.optimization_potential,
                    'recommendations': a.recommendations
                }
                for a in analyses
            ],
            'resource_utilizations': [
                {
                    'resource_type': u.resource_type,
                    'resource_id': u.resource_id,
                    'avg_utilization': u.avg_utilization,
                    'max_utilization': u.max_utilization,
                    'cost_per_hour': u.cost_per_hour,
                    'optimization_score': u.optimization_score
                }
                for u in utilizations
            ],
            'total_current_cost': sum(a.current_cost for a in analyses),
            'total_optimization_potential': sum(a.optimization_potential for a in analyses)
        }
        
        s3.put_object(
            Bucket=S3_BUCKET,
            Key=key,
            Body=json.dumps(data, default=str),
            ContentType='application/json',
            ServerSideEncryption='aws:kms'
        )
        
        logger.info(f"Saved cost analysis to S3: s3://{S3_BUCKET}/{key}")
        
    except ClientError as e:
        logger.error(f"Error saving cost analysis to S3: {str(e)}")

def publish_cost_metrics(analyses: List[CostAnalysis]):
    """Publish cost metrics to CloudWatch"""
    try:
        timestamp = datetime.now()
        metric_data = []
        
        total_cost = sum(a.current_cost for a in analyses)
        total_optimization_potential = sum(a.optimization_potential for a in analyses)
        
        # Overall cost metrics
        metric_data.extend([
            {
                'MetricName': 'TotalCostUSD',
                'Value': total_cost,
                'Unit': 'None',
                'Timestamp': timestamp
            },
            {
                'MetricName': 'OptimizationPotentialUSD',
                'Value': total_optimization_potential,
                'Unit': 'None',
                'Timestamp': timestamp
            },
            {
                'MetricName': 'PotentialSavingsPercent',
                'Value': (total_optimization_potential / total_cost * 100) if total_cost > 0 else 0,
                'Unit': 'Percent',
                'Timestamp': timestamp
            }
        ])
        
        # Service-specific metrics
        for analysis in analyses:
            if analysis.current_cost > 0:  # Only include services with costs
                metric_data.extend([
                    {
                        'MetricName': 'ServiceCostUSD',
                        'Value': analysis.current_cost,
                        'Unit': 'None',
                        'Timestamp': timestamp,
                        'Dimensions': [{'Name': 'Service', 'Value': analysis.service}]
                    },
                    {
                        'MetricName': 'ServiceCostChangePercent',
                        'Value': analysis.cost_change_percent,
                        'Unit': 'Percent',
                        'Timestamp': timestamp,
                        'Dimensions': [{'Name': 'Service', 'Value': analysis.service}]
                    }
                ])
        
        # Publish metrics in batches
        batch_size = 20
        for i in range(0, len(metric_data), batch_size):
            batch = metric_data[i:i + batch_size]
            cloudwatch.put_metric_data(
                Namespace='CI-CD/Cost',
                MetricData=batch
            )
        
        logger.info(f"Published {len(metric_data)} cost metrics to CloudWatch")
        
    except ClientError as e:
        logger.error(f"Error publishing cost metrics: {str(e)}")

def send_cost_alerts(analyses: List[CostAnalysis], optimization_actions: List[str]):
    """Send cost alerts and optimization notifications"""
    total_cost = sum(a.current_cost for a in analyses)
    total_savings_potential = sum(a.optimization_potential for a in analyses)
    
    alerts = []
    
    # Check total cost threshold
    if total_cost > COST_THRESHOLD_USD:
        alerts.append({
            'severity': 'WARNING',
            'message': f"Total CI/CD costs (${total_cost:.2f}) exceed threshold (${COST_THRESHOLD_USD})",
            'type': 'cost_threshold'
        })
    
    # Check for significant cost increases
    high_increase_services = [a for a in analyses if a.cost_change_percent > 50]
    if high_increase_services:
        alerts.append({
            'severity': 'WARNING',
            'message': f"{len(high_increase_services)} services show significant cost increases",
            'type': 'cost_increase',
            'details': [f"{s.service}: +{s.cost_change_percent:.1f}%" for s in high_increase_services]
        })
    
    # Send optimization summary if significant savings are possible
    if total_savings_potential > 20:  # More than $20 potential savings
        message = {
            'alert_type': 'cost_optimization',
            'severity': 'INFO',
            'timestamp': datetime.now().isoformat(),
            'total_cost': total_cost,
            'savings_potential': total_savings_potential,
            'savings_percent': (total_savings_potential / total_cost * 100) if total_cost > 0 else 0,
            'optimization_actions': optimization_actions,
            'service_breakdown': [
                {
                    'service': a.service,
                    'cost': a.current_cost,
                    'savings_potential': a.optimization_potential,
                    'recommendations': a.recommendations[:3]  # Top 3 recommendations
                }
                for a in analyses if a.optimization_potential > 5
            ]
        }
        
        try:
            sns.publish(
                TopicArn=SNS_COST_TOPIC,
                Message=json.dumps(message, default=str),
                Subject=f"CI/CD Cost Optimization Report - ${total_savings_potential:.2f} potential savings"
            )
            logger.info(f"Sent cost optimization report with ${total_savings_potential:.2f} potential savings")
        except ClientError as e:
            logger.error(f"Error sending cost optimization report: {str(e)}")
    
    # Send alerts
    for alert in alerts:
        try:
            sns.publish(
                TopicArn=SNS_COST_TOPIC,
                Message=json.dumps(alert, default=str),
                Subject=f"CI/CD Cost Alert: {alert['message']}"
            )
            logger.warning(f"Sent cost alert: {alert['message']}")
        except ClientError as e:
            logger.error(f"Error sending cost alert: {str(e)}")

def handler(event, context):
    """Lambda handler for cost optimization"""
    logger.info("Starting cost optimization analysis")
    
    try:
        # Initialize analyzers
        cost_analyzer = CostAnalyzer()
        resource_optimizer = ResourceOptimizer()
        
        # Analyze service costs
        cost_analyses = cost_analyzer.analyze_service_costs(days_back=30)
        logger.info(f"Analyzed costs for {len(cost_analyses)} services")
        
        # Analyze resource utilization
        utilizations = resource_optimizer.analyze_lambda_utilization()
        logger.info(f"Analyzed utilization for {len(utilizations)} Lambda functions")
        
        # Implement optimizations if enabled
        optimization_actions = []
        if AUTO_SCALING_ENABLED:
            optimization_actions = resource_optimizer.implement_optimizations(utilizations)
            logger.info(f"Implemented {len(optimization_actions)} optimization actions")
        
        # Save analysis results
        save_cost_analysis_to_s3(cost_analyses, utilizations)
        
        # Publish metrics
        publish_cost_metrics(cost_analyses)
        
        # Send alerts and recommendations
        send_cost_alerts(cost_analyses, optimization_actions)
        
        # Prepare summary response
        total_cost = sum(a.current_cost for a in cost_analyses)
        total_savings_potential = sum(a.optimization_potential for a in cost_analyses)
        
        summary = {
            'total_cost': total_cost,
            'savings_potential': total_savings_potential,
            'services_analyzed': len(cost_analyses),
            'resources_analyzed': len(utilizations),
            'optimization_actions': len(optimization_actions),
            'high_cost_services': [
                a.service for a in cost_analyses 
                if a.current_cost > 20
            ],
            'optimization_candidates': [
                u.resource_id for u in utilizations 
                if u.optimization_score >= 3
            ]
        }
        
        logger.info(f"Cost optimization completed: {json.dumps(summary, default=str)}")
        
        return {
            'statusCode': 200,
            'body': json.dumps({
                'message': 'Cost optimization analysis completed successfully',
                'summary': summary
            })
        }
        
    except Exception as e:
        logger.error(f"Cost optimization failed: {str(e)}")
        
        # Send failure alert
        try:
            sns.publish(
                TopicArn=SNS_COST_TOPIC,
                Message=json.dumps({
                    'alert_type': 'cost_optimizer_failure',
                    'severity': 'ERROR',
                    'timestamp': datetime.now().isoformat(),
                    'error': str(e)
                }, default=str),
                Subject="CI/CD Cost Optimizer Failed"
            )
        except Exception as alert_error:
            logger.error(f"Failed to send failure alert: {str(alert_error)}")
        
        return {
            'statusCode': 500,
            'body': json.dumps({
                'error': 'Cost optimization failed',
                'details': str(e)
            })
        }