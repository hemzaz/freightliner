# Runtime Security Monitoring

This package provides interfaces and types for runtime security monitoring and threat detection in containerized environments. It is designed to integrate with runtime security tools like Falco, but provides a generic abstraction that can work with multiple backends.

## Overview

Runtime security monitoring involves detecting and responding to threats in real-time by observing system calls, network activity, file access, and other runtime behaviors. This package provides:

- **RuntimeMonitor**: Interface for monitoring runtime security events
- **PolicyEngine**: Interface for managing and evaluating security policies
- **AlertManager**: Interface for managing security alerts and notifications
- **SecurityEvent**: Type representing a detected security event
- **Policy**: Type representing a runtime security policy
- **Alert**: Type representing a security alert

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                         │
│  (Security Dashboard, CLI Tools, Alert Handlers)            │
└─────────────────┬───────────────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────────────┐
│                  Runtime Security Interfaces                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │RuntimeMonitor│  │PolicyEngine  │  │AlertManager  │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────┬───────────────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────────────┐
│                  Implementation Layer                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │Falco         │  │Custom Rules  │  │Alert Routing │      │
│  │Integration   │  │Engine        │  │System        │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────┬───────────────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────────────┐
│              Runtime Security Backends                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │Falco         │  │eBPF          │  │Audit         │      │
│  │(Kernel       │  │Programs      │  │Framework     │      │
│  │Module/eBPF)  │  │              │  │              │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

## Falco Integration Approach

### What is Falco?

Falco is an open-source runtime security tool designed to detect anomalous activity in applications. It uses system calls to secure and monitor a system by:

- Parsing Linux system calls from the kernel at runtime
- Asserting the stream against a powerful rules engine
- Alerting when a rule is violated

### Integration Strategy

#### 1. Deployment Models

**DaemonSet Deployment** (Recommended for Kubernetes):
```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: falco
  namespace: falco
spec:
  selector:
    matchLabels:
      app: falco
  template:
    spec:
      serviceAccountName: falco
      hostNetwork: true
      hostPID: true
      containers:
      - name: falco
        image: falcosecurity/falco:latest
        securityContext:
          privileged: true
        volumeMounts:
        - name: dev
          mountPath: /host/dev
        - name: proc
          mountPath: /host/proc
          readOnly: true
        - name: boot
          mountPath: /host/boot
          readOnly: true
        - name: lib-modules
          mountPath: /host/lib/modules
          readOnly: true
        - name: usr
          mountPath: /host/usr
          readOnly: true
        - name: etc
          mountPath: /host/etc
          readOnly: true
        - name: falco-config
          mountPath: /etc/falco
      volumes:
      - name: dev
        hostPath:
          path: /dev
      - name: proc
        hostPath:
          path: /proc
      - name: boot
        hostPath:
          path: /boot
      - name: lib-modules
        hostPath:
          path: /lib/modules
      - name: usr
        hostPath:
          path: /usr
      - name: etc
        hostPath:
          path: /etc
      - name: falco-config
        configMap:
          name: falco-config
```

**Sidecar Deployment** (For specific workloads):
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: app-with-falco
spec:
  containers:
  - name: app
    image: myapp:latest
  - name: falco
    image: falcosecurity/falco:latest
    securityContext:
      privileged: true
    volumeMounts:
    - name: shared-logs
      mountPath: /var/log/app
  volumes:
  - name: shared-logs
    emptyDir: {}
```

#### 2. Event Collection

**gRPC Output** (Recommended):
```go
package falco

import (
    "context"
    "fmt"

    "github.com/falcosecurity/client-go/pkg/api/outputs"
    "github.com/falcosecurity/client-go/pkg/client"
    "google.golang.org/grpc"
)

type FalcoMonitor struct {
    client   *client.Client
    handlers []runtime.EventHandler
}

func NewFalcoMonitor(endpoint string) (*FalcoMonitor, error) {
    c, err := client.NewForConfig(&client.Config{
        Hostname:   "falco.falco.svc.cluster.local",
        Port:       5060,
        OutputType: "grpc",
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create Falco client: %w", err)
    }

    return &FalcoMonitor{
        client:   c,
        handlers: make([]runtime.EventHandler, 0),
    }, nil
}

func (m *FalcoMonitor) Start(ctx context.Context) error {
    outputClient, err := m.client.Outputs()
    if err != nil {
        return fmt.Errorf("failed to get outputs client: %w", err)
    }

    // Subscribe to Falco events
    stream, err := outputClient.Sub(ctx, &outputs.Request{})
    if err != nil {
        return fmt.Errorf("failed to subscribe to events: %w", err)
    }

    go m.processEvents(ctx, stream)
    return nil
}

func (m *FalcoMonitor) processEvents(ctx context.Context, stream outputs.Service_SubClient) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            resp, err := stream.Recv()
            if err != nil {
                // Handle error
                continue
            }

            // Convert Falco event to SecurityEvent
            event := m.convertToSecurityEvent(resp)

            // Notify all handlers
            for _, handler := range m.handlers {
                go handler.HandleEvent(ctx, event)
            }
        }
    }
}
```

**HTTP Webhook**:
```yaml
# falco.yaml
json_output: true
json_include_output_property: true
http_output:
  enabled: true
  url: "http://alert-handler.security.svc.cluster.local:8080/falco-events"
  insecure: false
  mtls: true
  client_cert: /etc/falco/certs/client.crt
  client_key: /etc/falco/certs/client.key
  ca_cert: /etc/falco/certs/ca.crt
```

#### 3. Custom Rule Development

**Rule Structure**:
```yaml
# custom-rules.yaml
- rule: Unauthorized Process in Container
  desc: Detect processes running that aren't in the approved list
  condition: >
    container
    and spawned_process
    and not proc.name in (approved_processes)
  output: >
    Unauthorized process started in container
    (user=%user.name process=%proc.cmdline
    container=%container.name image=%container.image.repository)
  priority: WARNING
  tags: [container, process, runtime]
  source: syscall

- rule: Sensitive File Access
  desc: Detect access to sensitive files
  condition: >
    open_read
    and container
    and fd.name in (sensitive_files)
    and not proc.name in (trusted_programs)
  output: >
    Sensitive file accessed
    (file=%fd.name process=%proc.name
    user=%user.name container=%container.name)
  priority: WARNING
  tags: [file, sensitive]

- rule: Container Privilege Escalation
  desc: Detect privilege escalation attempts
  condition: >
    container
    and spawned_process
    and (proc.name in (su, sudo, setuid_programs)
         or proc.cmdline contains "sudo"
         or proc.pname in (su, sudo))
  output: >
    Privilege escalation attempt detected
    (user=%user.name process=%proc.cmdline
    container=%container.name)
  priority: CRITICAL
  tags: [privilege_escalation, runtime]

- rule: Unexpected Network Connection
  desc: Detect outbound connections to unexpected destinations
  condition: >
    outbound
    and container
    and not fd.sip in (allowed_destinations)
    and not fd.sport in (allowed_ports)
  output: >
    Unexpected network connection
    (destination=%fd.rip:%fd.rport
    process=%proc.name container=%container.name)
  priority: WARNING
  tags: [network, exfiltration]

# Lists for rule conditions
- list: approved_processes
  items: [
    node, java, python, nginx, httpd,
    postgres, mysqld, redis-server
  ]

- list: sensitive_files
  items: [
    /etc/shadow, /etc/passwd, /etc/sudoers,
    /root/.ssh/*, /home/*/.ssh/*,
    /etc/pki/*, /etc/ssl/*
  ]

- list: trusted_programs
  items: [systemd, sshd, cron, logrotate]

- list: allowed_destinations
  items: [
    "10.0.0.0/8",
    "172.16.0.0/12",
    "192.168.0.0/16"
  ]

- list: allowed_ports
  items: [80, 443, 8080, 8443]

- list: setuid_programs
  items: [sudo, su, passwd, chsh, chfn, newgrp]

# Macros for reusable conditions
- macro: container
  condition: container.id != host

- macro: spawned_process
  condition: evt.type = execve and evt.dir=<

- macro: open_read
  condition: (evt.type in (open,openat,openat2) and evt.is_open_read=true and fd.typechar='f')

- macro: outbound
  condition: (evt.type=connect and evt.dir=< and fd.typechar=4)
```

**Rule Testing**:
```bash
# Test rules with falco-event-generator
kubectl run falco-event-generator \
  --image=falcosecurity/event-generator \
  -- run --loop

# Validate custom rules
falco --validate /etc/falco/rules.d/custom-rules.yaml

# Test specific rule
falco --test-rules /etc/falco/rules.d/custom-rules.yaml \
  --rule "Unauthorized Process in Container"
```

#### 4. Alert Routing and Response

**Alert Processing Pipeline**:
```go
package falco

import (
    "context"
    "encoding/json"

    "github.com/yourorg/freightliner/pkg/security/runtime"
)

type AlertProcessor struct {
    policyEngine runtime.PolicyEngine
    alertManager runtime.AlertManager
    enrichers    []EventEnricher
}

type EventEnricher interface {
    Enrich(ctx context.Context, event *runtime.SecurityEvent) error
}

func (p *AlertProcessor) ProcessEvent(ctx context.Context, event *runtime.SecurityEvent) error {
    // 1. Enrich event with additional context
    for _, enricher := range p.enrichers {
        if err := enricher.Enrich(ctx, event); err != nil {
            // Log error but continue processing
        }
    }

    // 2. Evaluate against policies
    decision, err := p.policyEngine.EvaluateEvent(ctx, event)
    if err != nil {
        return fmt.Errorf("policy evaluation failed: %w", err)
    }

    // 3. Generate alert if needed
    if decision.GenerateAlert {
        alert := p.createAlert(event, decision)
        if err := p.alertManager.SendAlert(ctx, alert); err != nil {
            return fmt.Errorf("failed to send alert: %w", err)
        }
    }

    // 4. Execute automated response actions
    if len(decision.Actions) > 0 {
        go p.executeActions(ctx, event, decision.Actions)
    }

    return nil
}

func (p *AlertProcessor) createAlert(event *runtime.SecurityEvent, decision *runtime.PolicyDecision) *runtime.Alert {
    return &runtime.Alert{
        ID:          generateAlertID(),
        EventID:     event.ID,
        Timestamp:   time.Now(),
        Title:       fmt.Sprintf("Security Event: %s", event.RuleName),
        Description: event.Description,
        Severity:    event.Severity,
        Status:      "new",
        Category:    event.Category,
        Tags:        event.Tags,
        Source:      event.Source,
        Affected:    extractAffectedResources(event),
        RecommendedActions: decision.Actions,
        Evidence:    map[string]interface{}{
            "event": event,
            "policy_decision": decision,
        },
    }
}
```

**Multi-Channel Alert Routing**:
```go
package alerting

import (
    "context"
    "fmt"

    "github.com/yourorg/freightliner/pkg/security/runtime"
)

type MultiChannelAlerter struct {
    channels map[string]AlertChannel
    router   *AlertRouter
}

type AlertChannel interface {
    Send(ctx context.Context, alert *runtime.Alert) error
    GetChannelType() string
}

type AlertRouter struct {
    rules []runtime.RoutingRule
}

func (m *MultiChannelAlerter) SendAlert(ctx context.Context, alert *runtime.Alert) error {
    // Determine which channels to use based on routing rules
    channels := m.router.RouteAlert(alert)

    // Send to all matched channels in parallel
    errChan := make(chan error, len(channels))
    for _, channelName := range channels {
        channel, ok := m.channels[channelName]
        if !ok {
            continue
        }

        go func(ch AlertChannel) {
            errChan <- ch.Send(ctx, alert)
        }(channel)
    }

    // Collect errors
    var errors []error
    for i := 0; i < len(channels); i++ {
        if err := <-errChan; err != nil {
            errors = append(errors, err)
        }
    }

    if len(errors) > 0 {
        return fmt.Errorf("failed to send to some channels: %v", errors)
    }

    return nil
}

// Example channel implementations
type SlackChannel struct {
    webhookURL string
}

type PagerDutyChannel struct {
    integrationKey string
}

type WebhookChannel struct {
    url     string
    headers map[string]string
}

type EmailChannel struct {
    smtpServer   string
    smtpPort     int
    fromAddress  string
    toAddresses  []string
}
```

## Usage Examples

### Setting Up Runtime Monitoring

```go
package main

import (
    "context"
    "log"

    "github.com/yourorg/freightliner/pkg/security/runtime"
    "github.com/yourorg/freightliner/pkg/security/runtime/falco"
)

func main() {
    ctx := context.Background()

    // Initialize Falco monitor
    monitor, err := falco.NewFalcoMonitor("falco.falco.svc.cluster.local:5060")
    if err != nil {
        log.Fatalf("Failed to create monitor: %v", err)
    }

    // Initialize policy engine
    policyEngine := NewCustomPolicyEngine()

    // Load policies
    policies, err := LoadPoliciesFromDirectory("./policies")
    if err != nil {
        log.Fatalf("Failed to load policies: %v", err)
    }

    for _, policy := range policies {
        if err := policyEngine.LoadPolicy(ctx, policy); err != nil {
            log.Printf("Failed to load policy %s: %v", policy.Name, err)
        }
    }

    // Initialize alert manager
    alertManager := NewAlertManager(AlertManagerConfig{
        Channels: map[string]AlertChannel{
            "slack":     NewSlackChannel(slackWebhookURL),
            "pagerduty": NewPagerDutyChannel(pdIntegrationKey),
            "webhook":   NewWebhookChannel(webhookURL),
        },
        RoutingRules: loadRoutingRules(),
    })

    // Register event handler
    handler := NewSecurityEventHandler(policyEngine, alertManager)
    monitor.RegisterEventHandler(handler)

    // Start monitoring
    if err := monitor.Start(ctx); err != nil {
        log.Fatalf("Failed to start monitor: %v", err)
    }

    log.Println("Runtime security monitoring started")

    // Wait for shutdown signal
    <-ctx.Done()

    // Graceful shutdown
    if err := monitor.Stop(ctx); err != nil {
        log.Printf("Error during shutdown: %v", err)
    }
}
```

### Creating Custom Policies

```go
policy := &runtime.Policy{
    ID:          "custom-001",
    Name:        "Detect Cryptocurrency Mining",
    Description: "Detects processes commonly associated with cryptocurrency mining",
    Version:     "1.0.0",
    Enabled:     true,
    Priority:    1,
    Severity:    runtime.SeverityHigh,
    Category:    "resource_abuse",
    Rules: []runtime.Rule{
        {
            ID:          "rule-001",
            Name:        "Mining Process Detection",
            Description: "Detects known mining process names",
            Condition:   "proc.name in (xmrig, ethminer, claymore, phoenixminer)",
            Output:      "Cryptocurrency mining detected (process=%proc.name container=%container.name)",
            Priority:    runtime.SeverityCritical,
            Enabled:     true,
        },
    },
    Actions: []runtime.Action{
        {
            Type:     "alert",
            Target:   "all",
            Enabled:  true,
            Priority: 1,
        },
        {
            Type:     "kill",
            Target:   "process",
            Enabled:  true,
            Priority: 2,
        },
    },
    OnViolation: runtime.OnViolationAction{
        GenerateAlert: true,
        BlockAction:   true,
        LogEvent:      true,
        Notify:        []string{"slack", "pagerduty"},
        Terminate:     true,
    },
}

err := policyEngine.LoadPolicy(ctx, policy)
```

## Best Practices

### 1. Rule Development
- Start with Falco's default rules
- Add custom rules incrementally
- Test rules in non-production first
- Use specific conditions to avoid false positives
- Document rule intent and rationale
- Version control all custom rules

### 2. Performance Optimization
- Filter events at source (Falco) when possible
- Use efficient rule conditions
- Implement event batching for high-volume environments
- Monitor system overhead (CPU, memory)
- Tune syscall filters to reduce noise

### 3. Alert Management
- Implement alert deduplication
- Use severity-based routing
- Set appropriate alert thresholds
- Configure alert suppression for maintenance
- Regularly review and tune alert rules
- Track MTTA (Mean Time To Acknowledge) and MTTR

### 4. Incident Response
- Define runbooks for common alert types
- Automate initial response actions
- Integrate with SOAR platforms
- Maintain audit trail of all actions
- Conduct post-incident reviews

### 5. Compliance
- Map rules to compliance frameworks (CIS, PCI-DSS, etc.)
- Generate compliance reports
- Maintain evidence for audits
- Document policy exceptions

## Integration Points

### Kubernetes Integration
- Deploy as DaemonSet for node-level monitoring
- Use ServiceAccount with appropriate RBAC
- Store configuration in ConfigMaps
- Use Secrets for sensitive credentials
- Integrate with K8s audit logs

### Observability Integration
- Export metrics to Prometheus
- Send events to SIEM (Splunk, ELK)
- Integrate with distributed tracing
- Stream events to data lakes
- Create Grafana dashboards

### CI/CD Integration
- Scan container images during build
- Enforce policies in admission controllers
- Test rules in staging environments
- Automate policy deployment
- Version control all configurations

## Troubleshooting

### Common Issues

**High CPU Usage**:
```yaml
# Tune syscall collection
syscall_event_drops:
  actions:
    - ignore
    - log
  max_burst: 100
  rate: 10
```

**False Positives**:
```yaml
# Add exceptions to rules
- rule: Sensitive File Access
  exceptions:
    - name: Backup Process
      comps:
        - proc.name=backup-agent
```

**Missing Events**:
```bash
# Check Falco logs
kubectl logs -n falco daemonset/falco

# Verify kernel module or eBPF probe
falco-driver-loader

# Check dropped events
kubectl exec -it -n falco falco-xxxx -- cat /sys/kernel/debug/tracing/trace_pipe | grep falco
```

## Future Enhancements

- [ ] Machine learning-based anomaly detection
- [ ] Behavioral baselining
- [ ] Threat intelligence integration
- [ ] Automated remediation workflows
- [ ] Multi-cluster correlation
- [ ] Advanced forensics capabilities
- [ ] Custom eBPF program support
- [ ] Real-time policy updates

## References

- [Falco Documentation](https://falco.org/docs/)
- [Falco Rules](https://github.com/falcosecurity/rules)
- [MITRE ATT&CK Framework](https://attack.mitre.org/)
- [Kubernetes Security Best Practices](https://kubernetes.io/docs/concepts/security/)
- [CIS Kubernetes Benchmark](https://www.cisecurity.org/benchmark/kubernetes)
