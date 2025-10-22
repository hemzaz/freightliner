# Kubernetes Operations Skill

Expert skill for deploying and managing Freightliner on Kubernetes.

## Deployment Patterns

### Helm Chart Structure
```
deployments/helm/freightliner/
├── Chart.yaml              # Chart metadata
├── values.yaml             # Default values
├── values-dev.yaml         # Development overrides
├── values-staging.yaml     # Staging overrides
├── values-production.yaml  # Production overrides
└── templates/
    ├── deployment.yaml     # Deployment
    ├── service.yaml        # Service
    ├── ingress.yaml        # Ingress
    ├── configmap.yaml      # ConfigMap
    ├── secrets.yaml        # Secret templates
    ├── servicemonitor.yaml # Prometheus ServiceMonitor
    ├── hpa.yaml            # HorizontalPodAutoscaler
    └── networkpolicy.yaml  # NetworkPolicy
```

### Deployment Configuration

#### High Availability
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: freightliner
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: freightliner
  template:
    metadata:
      labels:
        app: freightliner
        version: v1.2.0
    spec:
      # Anti-affinity for pod distribution
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchLabels:
                  app: freightliner
              topologyKey: kubernetes.io/hostname

      # Security context
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000

      containers:
      - name: freightliner
        image: ghcr.io/hemzaz/freightliner:v1.2.0
        imagePullPolicy: IfNotPresent

        # Security context
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL

        # Ports
        ports:
        - name: http
          containerPort: 8080
          protocol: TCP
        - name: metrics
          containerPort: 2112
          protocol: TCP

        # Resource limits
        resources:
          requests:
            cpu: 500m
            memory: 512Mi
          limits:
            cpu: 2000m
            memory: 4Gi

        # Health checks
        livenessProbe:
          httpGet:
            path: /live
            port: http
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3

        readinessProbe:
          httpGet:
            path: /ready
            port: http
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3

        startupProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 0
          periodSeconds: 10
          timeoutSeconds: 3
          failureThreshold: 30

        # Environment variables
        env:
        - name: LOG_LEVEL
          value: "info"
        - name: METRICS_ENABLED
          value: "true"
        - name: PORT
          value: "8080"
        - name: METRICS_PORT
          value: "2112"

        # Environment from ConfigMap
        envFrom:
        - configMapRef:
            name: freightliner-config

        # Secrets mounted as volumes
        volumeMounts:
        - name: aws-credentials
          mountPath: /home/app/.aws
          readOnly: true
        - name: gcp-credentials
          mountPath: /home/app/.gcp
          readOnly: true
        - name: tmp
          mountPath: /tmp

      volumes:
      - name: aws-credentials
        secret:
          secretName: freightliner-aws-credentials
      - name: gcp-credentials
        secret:
          secretName: freightliner-gcp-credentials
      - name: tmp
        emptyDir: {}

      # Graceful termination
      terminationGracePeriodSeconds: 60
```

### Service Configuration
```yaml
apiVersion: v1
kind: Service
metadata:
  name: freightliner
  labels:
    app: freightliner
spec:
  type: ClusterIP
  ports:
  - name: http
    port: 80
    targetPort: 8080
    protocol: TCP
  - name: metrics
    port: 2112
    targetPort: 2112
    protocol: TCP
  selector:
    app: freightliner
```

### Ingress with TLS
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: freightliner
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - freightliner.example.com
    secretName: freightliner-tls
  rules:
  - host: freightliner.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: freightliner
            port:
              number: 80
```

## Monitoring Setup

### ServiceMonitor for Prometheus
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: freightliner
  labels:
    app: freightliner
spec:
  selector:
    matchLabels:
      app: freightliner
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
```

### HorizontalPodAutoscaler
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: freightliner
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: freightliner
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 100
        periodSeconds: 30
      - type: Pods
        value: 2
        periodSeconds: 30
      selectPolicy: Max
```

## Security Best Practices

### NetworkPolicy
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: freightliner
spec:
  podSelector:
    matchLabels:
      app: freightliner
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
      port: 8080
  - from:
    - namespaceSelector:
        matchLabels:
          name: monitoring
    ports:
    - protocol: TCP
      port: 2112
  egress:
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 53  # DNS
    - protocol: UDP
      port: 53
  - to:
    - podSelector: {}
    ports:
    - protocol: TCP
      port: 443  # HTTPS to registries
```

### PodSecurityPolicy
```yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: freightliner
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
  - ALL
  volumes:
  - configMap
  - emptyDir
  - projected
  - secret
  - downwardAPI
  - persistentVolumeClaim
  hostNetwork: false
  hostIPC: false
  hostPID: false
  runAsUser:
    rule: MustRunAsNonRoot
  seLinux:
    rule: RunAsAny
  supplementalGroups:
    rule: RunAsAny
  fsGroup:
    rule: RunAsAny
  readOnlyRootFilesystem: true
```

## Common Operations

### Deployment
```bash
# Deploy with Helm
helm upgrade --install freightliner \
  ./deployments/helm/freightliner \
  -f ./deployments/helm/freightliner/values-production.yaml \
  --namespace freightliner-production \
  --create-namespace

# Check status
kubectl rollout status deployment/freightliner -n freightliner-production

# View logs
kubectl logs -f deployment/freightliner -n freightliner-production

# View all logs from all pods
kubectl logs -f -l app=freightliner -n freightliner-production --max-log-requests=10
```

### Debugging
```bash
# Get pod details
kubectl describe pod <pod-name> -n freightliner-production

# Execute command in pod
kubectl exec -it <pod-name> -n freightliner-production -- /bin/sh

# Port forward for local access
kubectl port-forward deployment/freightliner 8080:8080 -n freightliner-production

# Get pod metrics
kubectl top pod -l app=freightliner -n freightliner-production
```

### Scaling
```bash
# Manual scaling
kubectl scale deployment/freightliner --replicas=5 -n freightliner-production

# Check HPA status
kubectl get hpa freightliner -n freightliner-production

# View HPA events
kubectl describe hpa freightliner -n freightliner-production
```

### Rollback
```bash
# View rollout history
kubectl rollout history deployment/freightliner -n freightliner-production

# Rollback to previous version
kubectl rollout undo deployment/freightliner -n freightliner-production

# Rollback to specific revision
kubectl rollout undo deployment/freightliner --to-revision=5 -n freightliner-production
```

### Secret Management
```bash
# Create secret from file
kubectl create secret generic freightliner-aws-credentials \
  --from-file=credentials=/path/to/.aws/credentials \
  --from-file=config=/path/to/.aws/config \
  -n freightliner-production

# Create secret from literal
kubectl create secret generic freightliner-gcp-credentials \
  --from-file=key.json=/path/to/service-account-key.json \
  -n freightliner-production

# Update secret
kubectl delete secret freightliner-aws-credentials -n freightliner-production
kubectl create secret generic freightliner-aws-credentials \
  --from-file=credentials=/path/to/.aws/credentials \
  -n freightliner-production

# Restart deployment to pick up new secrets
kubectl rollout restart deployment/freightliner -n freightliner-production
```
