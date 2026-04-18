# Kubernetes in Production — Battle-Tested Patterns

## Table of Contents
1. [Resource Management](#resources)
2. [High Availability](#ha)
3. [RBAC & Security](#rbac)
4. [Networking](#networking)
5. [Storage](#storage)
6. [Debugging Cheatsheet](#debug)

---

## 1. Resource Management

### Requests vs Limits — The Most Misunderstood K8s Concept

```
Requests = what the scheduler GUARANTEES you
Limits   = the ceiling you can BURST to (CPU: throttled, Memory: OOMKilled)
```

**CPU**: Never set CPU limits (controversial but Google SRE-recommended for latency-sensitive apps).
CPU limits cause throttling even when node has spare capacity. Use LimitRange at namespace level instead.

**Memory**: Always set memory limits = memory requests (Guaranteed QoS).

```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "100m"
  limits:
    memory: "256Mi"    # = request → Guaranteed QoS, never evicted under pressure
    # no cpu limit intentionally — avoid throttling
```

### Vertical Pod Autoscaler (VPA) — Right-sizing

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: order-service-vpa
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: order-service
  updatePolicy:
    updateMode: "Off"   # "Off" = recommend only; "Auto" = restart pods to apply
  resourcePolicy:
    containerPolicies:
      - containerName: order-service
        minAllowed:
          cpu: 50m
          memory: 64Mi
        maxAllowed:
          cpu: 2
          memory: 2Gi
```

### HPA — Horizontal Pod Autoscaler

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: order-service-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: order-service
  minReplicas: 3        # never go below 3 for HA across AZs
  maxReplicas: 50
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70    # scale up at 70% CPU
    - type: Pods
      pods:
        metric:
          name: http_requests_per_second  # custom metric via KEDA or prometheus-adapter
        target:
          type: AverageValue
          averageValue: "1000"
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300   # wait 5 min before scaling down
      policies:
        - type: Percent
          value: 10
          periodSeconds: 60             # scale down max 10% per minute
    scaleUp:
      stabilizationWindowSeconds: 0    # scale up immediately
      policies:
        - type: Percent
          value: 100
          periodSeconds: 15
```

### PodDisruptionBudget — Safe Rollouts and Node Drains

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: order-service-pdb
spec:
  minAvailable: 2       # always keep at least 2 pods (use this OR maxUnavailable)
  selector:
    matchLabels:
      app: order-service
```

---

## 2. High Availability

### Spread Pods Across Zones

```yaml
spec:
  template:
    spec:
      topologySpreadConstraints:
        - maxSkew: 1
          topologyKey: topology.kubernetes.io/zone
          whenUnsatisfiable: DoNotSchedule    # hard constraint
          labelSelector:
            matchLabels:
              app: order-service
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 100
              podAffinityTerm:
                labelSelector:
                  matchLabels:
                    app: order-service
                topologyKey: kubernetes.io/hostname  # prefer different nodes
```

### Zero-Downtime Rolling Update — Full Config

```yaml
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1          # spin up 1 extra pod before killing old
      maxUnavailable: 0    # never reduce capacity
  minReadySeconds: 10      # wait 10s after pod is ready before proceeding
  template:
    spec:
      terminationGracePeriodSeconds: 60   # give app 60s to drain connections
      containers:
        - lifecycle:
            preStop:
              exec:
                command: ["/bin/sh", "-c", "sleep 5"]  # wait for LB to deregister
```

---

## 3. RBAC & Security

### Least Privilege Service Account

```yaml
# ServiceAccount
apiVersion: v1
kind: ServiceAccount
metadata:
  name: order-service
  namespace: production
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/order-service-role  # IRSA

---
# Role — namespace-scoped
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: order-service-role
  namespace: production
rules:
  - apiGroups: [""]
    resources: ["configmaps"]
    resourceNames: ["order-service-config"]   # specific resource, not wildcard
    verbs: ["get", "watch"]
  - apiGroups: [""]
    resources: ["secrets"]
    resourceNames: ["order-service-secrets"]
    verbs: ["get"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: order-service-rolebinding
  namespace: production
subjects:
  - kind: ServiceAccount
    name: order-service
    namespace: production
roleRef:
  kind: Role
  name: order-service-role
  apiGroup: rbac.authorization.k8s.io
```

### Network Policy — Default Deny + Allow Explicit

```yaml
# Default deny all ingress in namespace
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-ingress
  namespace: production
spec:
  podSelector: {}
  policyTypes: ["Ingress"]

---
# Allow only from api-gateway to order-service on port 8080
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-api-gateway-to-order
  namespace: production
spec:
  podSelector:
    matchLabels:
      app: order-service
  policyTypes: ["Ingress"]
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: api-gateway
      ports:
        - protocol: TCP
          port: 8080
```

---

## 4. Networking

### Ingress with TLS + Rate Limiting (nginx)

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: order-service-ingress
  namespace: production
  annotations:
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-body-size: "1m"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "60"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  ingressClassName: nginx
  tls:
    - hosts: ["api.example.com"]
      secretName: api-tls-cert
  rules:
    - host: api.example.com
      http:
        paths:
          - path: /api/orders
            pathType: Prefix
            backend:
              service:
                name: order-service
                port:
                  number: 8080
```

---

## 5. Debugging Cheatsheet

```bash
# Pod not starting?
kubectl describe pod <pod-name> -n <ns>          # events section is key
kubectl logs <pod-name> -n <ns> --previous       # logs from crashed container

# OOMKilled diagnosis
kubectl get events -n <ns> --sort-by='.lastTimestamp' | grep OOM
kubectl top pods -n <ns> --sort-by=memory

# Node pressure
kubectl describe node <node-name> | grep -A 10 "Conditions:"
kubectl top nodes

# CrashLoopBackOff — check last exit code
kubectl get pod <pod> -n <ns> -o jsonpath='{.status.containerStatuses[0].lastState.terminated.exitCode}'
# 137 = OOMKilled, 143 = SIGTERM, 1 = app error

# Connectivity debugging
kubectl run debug --rm -it --image=nicolaka/netshoot -- bash
  # inside: curl, dig, netstat, tcpdump all available

# ConfigMap/Secret not mounted?
kubectl get events -n <ns> | grep FailedMount

# HPA not scaling?
kubectl describe hpa <name> -n <ns>   # check "Events" and "Conditions"
kubectl get --raw /apis/metrics.k8s.io/v1beta1/namespaces/<ns>/pods

# etcd and control plane health
kubectl get componentstatuses
kubectl cluster-info dump | head -100
```

---

## Production Checklist

- [ ] Resources (requests/limits) set on all containers
- [ ] Liveness and readiness probes configured
- [ ] PodDisruptionBudget defined
- [ ] TopologySpreadConstraints for zone distribution
- [ ] Dedicated ServiceAccount (not default)
- [ ] Network policies: default-deny + explicit allows
- [ ] Image pinned to digest, not tag
- [ ] `readOnlyRootFilesystem: true` where possible
- [ ] `runAsNonRoot: true`
- [ ] HPA configured with stable scale-down behavior
- [ ] `terminationGracePeriodSeconds` set appropriately
- [ ] Secrets from external store (ESO / Vault), not hardcoded
