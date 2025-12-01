# Monitoring Stack Deployment Checklist

## Pre-Deployment Preparation

### Infrastructure Requirements
- [ ] Kubernetes cluster 1.24+ running
- [ ] kubectl configured and authenticated
- [ ] 100GB+ persistent storage available
- [ ] LoadBalancer support (optional, for external access)
- [ ] Ingress controller installed (optional)

### Network Requirements
- [ ] Prometheus port 9090 accessible
- [ ] Grafana port 3000 accessible
- [ ] AlertManager port 9093 accessible
- [ ] Freightliner metrics port 2112 accessible
- [ ] Outbound HTTPS for Slack webhooks
- [ ] Outbound HTTPS for PagerDuty API

### Access Credentials
- [ ] Slack webhook URL obtained
- [ ] PagerDuty integration key obtained
- [ ] Grafana admin password generated (use: `openssl rand -base64 32`)
- [ ] Registry credentials for image pulls (if private registry)

## Deployment Steps

### 1. Namespace Creation (2 minutes)
```bash
kubectl create namespace monitoring
kubectl label namespace monitoring monitoring=enabled
```
- [ ] Namespace created
- [ ] Namespace labeled

### 2. Secrets Configuration (5 minutes)
```bash
# Grafana admin credentials
kubectl create secret generic grafana-admin \
  --from-literal=username=admin \
  --from-literal=password=YOUR_SECURE_PASSWORD \
  -n monitoring

# AlertManager notification credentials
kubectl create secret generic alertmanager-secrets \
  --from-literal=slack-webhook=https://hooks.slack.com/services/YOUR/WEBHOOK \
  --from-literal=pagerduty-key=YOUR_PAGERDUTY_INTEGRATION_KEY \
  -n monitoring
```
- [ ] Grafana admin secret created
- [ ] AlertManager secrets created
- [ ] Passwords stored in password manager

### 3. ConfigMap Creation (5 minutes)
```bash
# Prometheus configuration
kubectl create configmap prometheus-config \
  --from-file=prometheus.yml=monitoring/prometheus/prometheus.yml \
  --from-file=recording-rules.yml=monitoring/prometheus/recording-rules.yml \
  --from-file=alert-rules.yml=monitoring/prometheus/alert-rules.yml \
  -n monitoring

# AlertManager configuration (update with your webhook URLs first!)
kubectl create configmap alertmanager-config \
  --from-file=alertmanager.yml=monitoring/alertmanager/alertmanager.yml \
  -n monitoring

# Grafana dashboards
kubectl create configmap grafana-dashboards \
  --from-file=monitoring/grafana/dashboards/ \
  -n monitoring
```
- [ ] Prometheus ConfigMap created
- [ ] AlertManager ConfigMap created (with updated webhooks)
- [ ] Grafana dashboards ConfigMap created

### 4. Storage Provisioning (3 minutes)
```bash
# Verify StorageClass exists
kubectl get storageclass

# PVCs will be created automatically by deployments
```
- [ ] StorageClass verified
- [ ] Storage capacity sufficient (100GB+ recommended)

### 5. Deploy Prometheus (5 minutes)
```bash
kubectl apply -f monitoring/kubernetes/prometheus-deployment.yaml

# Wait for pod to be ready
kubectl wait --for=condition=ready pod -l app=prometheus -n monitoring --timeout=300s

# Verify deployment
kubectl get pods -n monitoring -l app=prometheus
kubectl logs -n monitoring deployment/prometheus --tail=50
```
- [ ] Prometheus deployment created
- [ ] Prometheus pod running
- [ ] Prometheus logs show no errors
- [ ] PVC bound successfully

### 6. Deploy AlertManager (5 minutes)
```bash
kubectl apply -f monitoring/kubernetes/alertmanager-deployment.yaml

# Wait for pod to be ready
kubectl wait --for=condition=ready pod -l app=alertmanager -n monitoring --timeout=300s

# Verify deployment
kubectl get pods -n monitoring -l app=alertmanager
kubectl logs -n monitoring deployment/alertmanager --tail=50
```
- [ ] AlertManager deployment created
- [ ] AlertManager pod running
- [ ] AlertManager logs show no errors
- [ ] Configuration loaded successfully

### 7. Deploy Grafana (5 minutes)
```bash
kubectl apply -f monitoring/kubernetes/grafana-deployment.yaml

# Wait for pod to be ready
kubectl wait --for=condition=ready pod -l app=grafana -n monitoring --timeout=300s

# Verify deployment
kubectl get pods -n monitoring -l app=grafana
kubectl logs -n monitoring deployment/grafana --tail=50
```
- [ ] Grafana deployment created
- [ ] Grafana pod running
- [ ] Grafana logs show no errors
- [ ] Datasource provisioned

### 8. Deploy ServiceMonitors (3 minutes)
```bash
kubectl apply -f monitoring/kubernetes/servicemonitor-freightliner.yaml

# Verify ServiceMonitor
kubectl get servicemonitor -n freightliner
kubectl describe servicemonitor freightliner -n freightliner
```
- [ ] ServiceMonitor created
- [ ] ServiceMonitor targeting correct labels
- [ ] Freightliner service exists with metrics port

## Post-Deployment Verification

### 9. Access Verification (10 minutes)

#### Prometheus
```bash
# Port forward
kubectl port-forward -n monitoring svc/prometheus 9090:9090

# Open browser to http://localhost:9090
# Verify:
# - Status → Targets shows freightliner as UP
# - Graph shows freightliner metrics when querying: freightliner_replications_total
# - Status → Rules shows recording and alert rules loaded
```
- [ ] Prometheus UI accessible
- [ ] Freightliner target is UP
- [ ] Metrics being collected
- [ ] Rules loaded (recording + alert)
- [ ] No scrape errors

#### AlertManager
```bash
# Port forward
kubectl port-forward -n monitoring svc/alertmanager 9093:9093

# Open browser to http://localhost:9093
# Verify:
# - Status page loads
# - Configuration shows correct receivers
# - Routes configured properly
```
- [ ] AlertManager UI accessible
- [ ] Configuration loaded
- [ ] Routes defined
- [ ] Receivers configured

#### Grafana
```bash
# Port forward
kubectl port-forward -n monitoring svc/grafana 3000:3000

# Get admin password
kubectl get secret grafana-admin -n monitoring -o jsonpath='{.data.password}' | base64 -d

# Open browser to http://localhost:3000
# Login with admin / <password>
# Verify:
# - Datasource "Prometheus" is working (Configuration → Data Sources → Test)
# - Dashboards are imported (Dashboards → Browse → Freightliner folder)
# - Dashboard shows data (open any dashboard)
```
- [ ] Grafana UI accessible
- [ ] Login successful
- [ ] Prometheus datasource connected
- [ ] All 4 dashboards imported
- [ ] Dashboards showing metrics

### 10. Metrics Flow Test (5 minutes)
```bash
# Query for recent metrics
kubectl exec -n monitoring deployment/prometheus -- \
  wget -qO- 'http://localhost:9090/api/v1/query?query=up{job="freightliner"}' | jq

# Should return value: "1"
```
- [ ] Metrics query successful
- [ ] Freightliner is reporting as up
- [ ] Recent timestamps on metrics

### 11. Alert Testing (10 minutes)

#### Test Alert Routing
```bash
# Create a test alert (optional - only if you want to verify notifications)
kubectl exec -n monitoring deployment/alertmanager -- \
  amtool alert add test_alert \
    severity=warning \
    component=test \
    summary="Test alert" \
    --alertmanager.url=http://localhost:9093

# Check if alert appears in AlertManager UI
# Check if Slack notification received (if configured)
```
- [ ] Test alert created
- [ ] Alert visible in AlertManager UI
- [ ] Slack notification received (if configured)
- [ ] Cleanup test alert after verification

### 12. Dashboard Validation (15 minutes)

For each dashboard, verify:

#### Replication Overview Dashboard
- [ ] Success rate stat panel shows percentage
- [ ] Active replications stat shows rate
- [ ] Data transfer rate shows bytes/sec
- [ ] P95 latency shows seconds
- [ ] Time series charts display data
- [ ] Error rate chart shows (or empty if no errors)
- [ ] Top replication pairs table populated

#### Infrastructure Metrics Dashboard
- [ ] Service uptime shows percentage
- [ ] Memory usage shows MB
- [ ] Goroutine count shows number
- [ ] Worker pool utilization shows percentage
- [ ] HTTP request rate time series
- [ ] HTTP duration percentiles
- [ ] All gauges functional

#### Business Metrics Dashboard
- [ ] Total replications (24h) shows count
- [ ] Data replicated shows GB
- [ ] Success rate shows percentage
- [ ] Estimated cost calculated
- [ ] Top repositories tables populated
- [ ] Hourly trends showing

#### Error Rate & Latency Dashboard
- [ ] Error budget remaining calculated
- [ ] Current error rate shows
- [ ] Latency SLO compliance shows
- [ ] HTTP error rate tracked
- [ ] Latency distribution chart
- [ ] Burn rate calculated

### 13. Alert Rule Validation (10 minutes)
```bash
# View loaded alert rules
kubectl exec -n monitoring deployment/prometheus -- \
  wget -qO- 'http://localhost:9090/api/v1/rules' | jq '.data.groups[].rules[] | select(.type=="alerting") | .name'

# Verify key alerts loaded:
# - HighReplicationFailureRate
# - ReplicationStopped
# - FreightlinerDown
# - HighHTTP5xxRate
# - HighMemoryUsage
# - ApplicationPanics
```
- [ ] All critical alerts loaded
- [ ] All warning alerts loaded
- [ ] No alert rule syntax errors
- [ ] Alerts in "Inactive" state (good sign)

### 14. Recording Rule Validation (5 minutes)
```bash
# Query a recording rule
kubectl exec -n monitoring deployment/prometheus -- \
  wget -qO- 'http://localhost:9090/api/v1/query?query=freightliner:replication:success_rate:5m' | jq

# Should return computed success rate values
```
- [ ] Recording rules computing
- [ ] Values look reasonable
- [ ] No recording rule errors

## Configuration Tuning

### 15. Update Alert Thresholds (if needed)
Based on baseline metrics, adjust thresholds in `monitoring/prometheus/alert-rules.yml`:
- [ ] Reviewed baseline metrics for 24 hours
- [ ] Adjusted thresholds if too sensitive
- [ ] Updated ConfigMap
- [ ] Reloaded Prometheus configuration

### 16. Configure External Access (if needed)

#### Option A: LoadBalancer
```bash
# Services already configured with LoadBalancer type
kubectl get svc -n monitoring | grep LoadBalancer

# Get external IPs
kubectl get svc grafana-external -n monitoring -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
```
- [ ] LoadBalancer IPs assigned
- [ ] Services accessible externally
- [ ] Firewall rules updated

#### Option B: Ingress
```yaml
# Create ingress resources for:
# - grafana.example.com
# - prometheus.example.com
# - alertmanager.example.com
```
- [ ] Ingress resources created
- [ ] DNS records configured
- [ ] TLS certificates obtained
- [ ] Services accessible via domain names

### 17. Configure Authentication (Production)
- [ ] Enable Grafana OAuth/LDAP
- [ ] Enable Prometheus basic auth
- [ ] Configure AlertManager with auth
- [ ] Update network policies
- [ ] Document access procedures

### 18. Set Up Monitoring of Monitoring
```bash
# Add alerts for monitoring stack health
# - PrometheusDown
# - GrafanaDown
# - AlertManagerDown
# - HighPrometheusScrapeErrors
```
- [ ] Meta-monitoring alerts added
- [ ] Monitoring stack health tracked
- [ ] Separate escalation path defined

## Documentation & Handoff

### 19. Update Documentation
- [ ] Add team-specific runbook links
- [ ] Update Slack channel names
- [ ] Update PagerDuty rotation details
- [ ] Document custom thresholds
- [ ] Update wiki with access info

### 20. Team Training
- [ ] Conduct dashboard walkthrough
- [ ] Review alert response procedures
- [ ] Practice runbook execution
- [ ] Document escalation paths
- [ ] Schedule on-call training

### 21. Operational Readiness
- [ ] On-call rotation configured
- [ ] PagerDuty schedules set
- [ ] Runbooks accessible to team
- [ ] Team has monitoring access
- [ ] Backup procedures documented

## Maintenance Schedule

### Daily
- [ ] Check monitoring stack health
- [ ] Review active alerts
- [ ] Verify metric collection

### Weekly
- [ ] Review alert trends
- [ ] Check storage usage
- [ ] Validate backup success
- [ ] Review dashboard usage

### Monthly
- [ ] Update alert thresholds based on trends
- [ ] Review and tune recording rules
- [ ] Check for component updates
- [ ] Capacity planning review

### Quarterly
- [ ] Review SLOs with stakeholders
- [ ] Update documentation
- [ ] Disaster recovery drill
- [ ] Security audit

## Rollback Procedure

If issues occur during deployment:

```bash
# Remove all monitoring resources
kubectl delete namespace monitoring

# Remove ServiceMonitors from freightliner namespace
kubectl delete servicemonitor -n freightliner --all

# Clean up any PVCs that might be stuck
kubectl get pvc -n monitoring
kubectl delete pvc -n monitoring --all
```

## Success Criteria

Deployment is successful when:
- [ ] All pods in monitoring namespace are Running
- [ ] All PVCs are Bound
- [ ] Prometheus is collecting metrics from Freightliner
- [ ] Grafana dashboards display real data
- [ ] AlertManager is configured and accessible
- [ ] Test alert successfully routed to Slack
- [ ] All documentation updated
- [ ] Team trained on usage
- [ ] On-call rotation configured

## Emergency Contacts

- **Monitoring Lead**: [Name/Slack/Phone]
- **Platform Team**: [Slack channel]
- **On-Call Engineer**: PagerDuty rotation
- **Escalation**: [Manager contact]

## Additional Resources

- Setup Guide: `monitoring/docs/MONITORING_SETUP.md`
- Runbooks: `monitoring/docs/SLO_RUNBOOKS.md`
- Main README: `monitoring/README.md`
- Team Wiki: [URL]
- Slack Channel: #freightliner-monitoring

## Sign-Off

- [ ] Deployment completed by: _________________ Date: _______
- [ ] Verified by: _________________ Date: _______
- [ ] Production ready approval: _________________ Date: _______
