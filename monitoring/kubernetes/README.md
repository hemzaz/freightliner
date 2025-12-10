# Grafana Kubernetes Deployment

Flexible Kubernetes deployment for Grafana with support for local, remote, and production environments using Kustomize overlays.

## Directory Structure

```
monitoring/kubernetes/
├── base/                          # Base Grafana configuration
│   ├── grafana-configmap.yaml     # Environment variables and settings
│   ├── grafana-deployment.yaml    # Deployment specification
│   ├── grafana-service.yaml       # ClusterIP service
│   ├── grafana-pvc.yaml           # Persistent volume claim
│   ├── grafana-secret.yaml        # Admin credentials (dev only)
│   └── kustomization.yaml         # Base kustomization
├── overlays/
│   ├── local/                     # Local development (NodePort)
│   │   ├── kustomization.yaml
│   │   └── service-nodeport.yaml
│   ├── remote/                    # Remote access (LoadBalancer)
│   │   ├── kustomization.yaml
│   │   └── service-loadbalancer.yaml
│   └── production/                # Production (Ingress + TLS)
│       ├── kustomization.yaml
│       ├── ingress.yaml
│       └── remove-default-secret.yaml
└── README.md                      # This file
```

## Deployment Modes

### 1. Local Development (NodePort)

**Features:**
- NodePort service on port 30300
- Reduced resource requirements
- Access via `http://localhost:30300`
- 5Gi storage

**Deploy:**
```bash
# Apply local configuration
kubectl apply -k monitoring/kubernetes/overlays/local

# Verify deployment
kubectl get pods -n monitoring -l app=grafana
kubectl get svc -n monitoring grafana

# Access Grafana
# URL: http://localhost:30300
# Username: admin
# Password: changeme
```

**Port-forward alternative:**
```bash
kubectl port-forward -n monitoring svc/grafana 3000:3000
# Access at http://localhost:3000
```

### 2. Remote Access (LoadBalancer)

**Features:**
- LoadBalancer service for external access
- Standard resource allocation
- Auto-assigned external IP
- 10Gi storage

**Deploy:**
```bash
# Update external IP in kustomization.yaml
# Edit: monitoring/kubernetes/overlays/remote/kustomization.yaml
# Set EXTERNAL_IP and GRAFANA_ROOT_URL

# Apply remote configuration
kubectl apply -k monitoring/kubernetes/overlays/remote

# Get external IP
kubectl get svc -n monitoring grafana-external
EXTERNAL_IP=$(kubectl get svc -n monitoring grafana-external -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

# Access Grafana
echo "Grafana URL: http://${EXTERNAL_IP}"
```

**Update configuration with actual IP:**
```bash
# After getting external IP, update configuration
kubectl patch configmap grafana-config -n monitoring \
  --patch "{\"data\":{\"EXTERNAL_IP\":\"${EXTERNAL_IP}\",\"GRAFANA_ROOT_URL\":\"http://${EXTERNAL_IP}:3000\"}}"

# Restart pods to apply changes
kubectl rollout restart deployment/grafana -n monitoring
```

### 3. Production (Ingress + TLS)

**Features:**
- NGINX Ingress with TLS termination
- Cert-Manager for Let's Encrypt certificates
- High availability (2 replicas)
- Pod anti-affinity
- 20Gi storage
- Manual secret creation required

**Prerequisites:**
```bash
# Install NGINX Ingress Controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/cloud/deploy.yaml

# Install Cert-Manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Create ClusterIssuer for Let's Encrypt
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@your-domain.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
      - http01:
          ingress:
            class: nginx
EOF
```

**Deploy:**
```bash
# Update domain in configuration files
# Edit: monitoring/kubernetes/overlays/production/kustomization.yaml
# Edit: monitoring/kubernetes/overlays/production/ingress.yaml
# Replace 'your-domain.com' with your actual domain

# Create admin secret (REQUIRED - no default in production)
kubectl create secret generic grafana-admin \
  --from-literal=username=admin \
  --from-literal=password=$(openssl rand -base64 32) \
  -n monitoring

# Apply production configuration
kubectl apply -k monitoring/kubernetes/overlays/production

# Verify deployment
kubectl get pods -n monitoring -l app=grafana
kubectl get ingress -n monitoring grafana
kubectl get certificate -n monitoring grafana-tls

# Access Grafana (after DNS is configured)
# URL: https://grafana.your-domain.com
```

**Retrieve admin password:**
```bash
kubectl get secret grafana-admin -n monitoring -o jsonpath='{.data.password}' | base64 -d
```

## Configuration Options

### Environment Variables (ConfigMap: grafana-config)

| Variable | Description | Default | Override in Overlay |
|----------|-------------|---------|---------------------|
| `DEPLOYMENT_MODE` | Deployment mode | `local` | Yes |
| `GRAFANA_ROOT_URL` | Root URL for Grafana | `http://localhost:3000` | Yes |
| `EXTERNAL_IP` | External IP address | ` ` | Yes (remote/production) |
| `API_BASE_URL` | Freightliner API endpoint | `http://freightliner-api:8080` | Yes |
| `GF_SERVER_DOMAIN` | Server domain | `localhost` | Yes |
| `GF_SERVER_PROTOCOL` | Protocol (http/https) | `http` | Yes (production: https) |
| `GF_INSTALL_PLUGINS` | Grafana plugins | `grafana-piechart-panel` | No |

### Resource Requirements

| Mode | CPU Request | CPU Limit | Memory Request | Memory Limit | Storage |
|------|-------------|-----------|----------------|--------------|---------|
| Local | 100m | 500m | 256Mi | 1Gi | 5Gi |
| Remote | 250m | 1000m | 512Mi | 2Gi | 10Gi |
| Production | 250m | 1000m | 512Mi | 2Gi | 20Gi |

## Common Tasks

### Switch Deployment Modes

```bash
# From local to remote
kubectl delete -k monitoring/kubernetes/overlays/local
kubectl apply -k monitoring/kubernetes/overlays/remote

# From remote to production
kubectl delete -k monitoring/kubernetes/overlays/remote
kubectl apply -k monitoring/kubernetes/overlays/production
```

### Update Configuration

```bash
# Edit ConfigMap
kubectl edit configmap grafana-config -n monitoring

# Restart to apply changes
kubectl rollout restart deployment/grafana -n monitoring
```

### Scale Replicas

```bash
# Scale up
kubectl scale deployment grafana -n monitoring --replicas=3

# Scale down
kubectl scale deployment grafana -n monitoring --replicas=1
```

### View Logs

```bash
# All pods
kubectl logs -n monitoring -l app=grafana --tail=100 -f

# Specific pod
kubectl logs -n monitoring grafana-xxxxxxxxxx-xxxxx -f
```

### Access Metrics

```bash
# Port-forward to metrics endpoint
kubectl port-forward -n monitoring svc/grafana 3000:3000

# Access metrics
curl http://localhost:3000/metrics
```

### Backup Data

```bash
# Get PVC name
PVC=$(kubectl get pvc -n monitoring grafana-data -o jsonpath='{.spec.volumeName}')

# Create snapshot (if supported by storage class)
kubectl create -f - <<EOF
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: grafana-backup-$(date +%Y%m%d-%H%M%S)
  namespace: monitoring
spec:
  volumeSnapshotClassName: csi-snapclass
  source:
    persistentVolumeClaimName: grafana-data
EOF
```

### Restore Data

```bash
# Create PVC from snapshot
kubectl create -f - <<EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: grafana-data-restored
  namespace: monitoring
spec:
  dataSource:
    name: grafana-backup-YYYYMMDD-HHMMSS
    kind: VolumeSnapshot
    apiGroup: snapshot.storage.k8s.io
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
EOF

# Update deployment to use restored PVC
kubectl patch deployment grafana -n monitoring \
  --patch '{"spec":{"template":{"spec":{"volumes":[{"name":"grafana-data","persistentVolumeClaim":{"claimName":"grafana-data-restored"}}]}}}}'
```

## Troubleshooting

### Pod Not Starting

```bash
# Check pod status
kubectl describe pod -n monitoring -l app=grafana

# Check events
kubectl get events -n monitoring --sort-by='.lastTimestamp'

# Check logs
kubectl logs -n monitoring -l app=grafana --previous
```

### PVC Issues

```bash
# Check PVC status
kubectl get pvc -n monitoring grafana-data

# Check PV
kubectl get pv

# Delete and recreate if needed
kubectl delete pvc -n monitoring grafana-data
kubectl apply -k monitoring/kubernetes/overlays/local
```

### Service Not Accessible

```bash
# Check service
kubectl get svc -n monitoring grafana
kubectl describe svc -n monitoring grafana

# Test from within cluster
kubectl run -n monitoring curl --image=curlimages/curl -i --tty --rm \
  -- curl http://grafana:3000/api/health

# Check endpoints
kubectl get endpoints -n monitoring grafana
```

### Ingress Issues (Production)

```bash
# Check ingress
kubectl describe ingress -n monitoring grafana

# Check certificate
kubectl describe certificate -n monitoring grafana-tls
kubectl get certificaterequest -n monitoring

# Check NGINX logs
kubectl logs -n ingress-nginx -l app.kubernetes.io/component=controller
```

### Configuration Issues

```bash
# View current configuration
kubectl get configmap -n monitoring grafana-config -o yaml

# Test configuration
kubectl exec -n monitoring -it deployment/grafana -- grafana-cli admin reset-admin-password newpassword
```

## Security Best Practices

1. **Always create unique credentials in production:**
   ```bash
   kubectl create secret generic grafana-admin \
     --from-literal=username=admin \
     --from-literal=password=$(openssl rand -base64 32) \
     -n monitoring
   ```

2. **Use TLS in production** (handled by Ingress)

3. **Enable RBAC** for Grafana users

4. **Regularly update Grafana image:**
   ```bash
   # Update image in base/grafana-deployment.yaml
   kubectl set image deployment/grafana -n monitoring \
     grafana=grafana/grafana:10.2.0
   ```

5. **Use NetworkPolicies** to restrict access:
   ```bash
   kubectl apply -f - <<EOF
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: grafana-network-policy
     namespace: monitoring
   spec:
     podSelector:
       matchLabels:
         app: grafana
     policyTypes:
       - Ingress
       - Egress
     ingress:
       - from:
           - namespaceSelector:
               matchLabels:
                 name: ingress-nginx
         ports:
           - protocol: TCP
             port: 3000
     egress:
       - to:
           - namespaceSelector:
               matchLabels:
                 name: monitoring
         ports:
           - protocol: TCP
             port: 9090  # Prometheus
   EOF
   ```

## Monitoring & Observability

Grafana exports metrics on `/metrics` endpoint that can be scraped by Prometheus.

**Key metrics:**
- `grafana_stat_totals_*` - Dashboard and user statistics
- `grafana_api_*` - API request metrics
- `grafana_database_*` - Database connection metrics
- `process_*` - Process metrics (CPU, memory, file descriptors)

## References

- [Grafana Documentation](https://grafana.com/docs/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Kustomize Documentation](https://kustomize.io/)
- [Cert-Manager Documentation](https://cert-manager.io/docs/)
- [NGINX Ingress Controller](https://kubernetes.github.io/ingress-nginx/)
