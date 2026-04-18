# Security Hardening — Zero-Trust Infrastructure

## Table of Contents
1. [Container Security](#containers)
2. [Secrets Management](#secrets)
3. [Network Security](#network)
4. [IAM & Least Privilege](#iam)
5. [Security Scanning Pipeline](#scanning)

---

## 1. Container Security

### Image Security Checklist

```dockerfile
# Distroless: no shell, no package manager, minimal attack surface
FROM gcr.io/distroless/static-debian12:nonroot

# Pin to digest (not tag) — immune to tag mutation attacks
FROM gcr.io/distroless/static-debian12@sha256:abc123...

# Never: root user
USER nonroot:nonroot

# Never: install unnecessary tools in runtime image
# Build tools belong in builder stage only
```

### Pod Security Standards — Enforce at Namespace Level

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: production
  labels:
    pod-security.kubernetes.io/enforce: restricted      # strictest
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

### OPA/Gatekeeper — Admission Control Policies

```yaml
# ConstraintTemplate: deny images not from approved registries
apiVersion: templates.gatekeeper.sh/v1
kind: ConstraintTemplate
metadata:
  name: requireapprovedregistry
spec:
  crd:
    spec:
      names:
        kind: RequireApprovedRegistry
  targets:
    - target: admission.k8s.gatekeeper.sh
      rego: |
        package requireapprovedregistry

        violation[{"msg": msg}] {
          container := input.review.object.spec.containers[_]
          not startswith(container.image, "registry.company.com/")
          not startswith(container.image, "gcr.io/distroless/")
          msg := sprintf("Image '%v' is not from an approved registry", [container.image])
        }

---
apiVersion: constraints.gatekeeper.sh/v1beta1
kind: RequireApprovedRegistry
metadata:
  name: prod-approved-registries
spec:
  match:
    namespaces: ["production", "staging"]
```

---

## 2. Secrets Management

### External Secrets Operator + AWS Secrets Manager

```yaml
# SecretStore — defines how to connect to secret backend
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: aws-secrets-manager
  namespace: production
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-east-1
      auth:
        jwt:
          serviceAccountRef:
            name: external-secrets-sa   # IRSA — no static credentials

---
# ExternalSecret — maps AWS secret → Kubernetes Secret
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: order-service-secrets
  namespace: production
spec:
  refreshInterval: 1h           # auto-rotate without pod restart
  secretStoreRef:
    name: aws-secrets-manager
    kind: SecretStore
  target:
    name: order-service-secrets  # creates this K8s Secret
    creationPolicy: Owner
    deletionPolicy: Retain       # keep secret if ExternalSecret deleted
  data:
    - secretKey: DB_PASSWORD     # K8s secret key
      remoteRef:
        key: production/order-service  # AWS secret name
        property: db_password          # JSON field within secret
    - secretKey: PAYMENT_API_KEY
      remoteRef:
        key: production/order-service
        property: payment_api_key
```

### HashiCorp Vault — Dynamic Secrets

```yaml
# Vault Agent Injector — inject secrets as files (not env vars)
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    metadata:
      annotations:
        vault.hashicorp.com/agent-inject: "true"
        vault.hashicorp.com/role: "order-service"
        vault.hashicorp.com/agent-inject-secret-config: "secret/data/order-service"
        vault.hashicorp.com/agent-inject-template-config: |
          {{- with secret "secret/data/order-service" -}}
          DB_PASSWORD={{ .Data.data.db_password }}
          API_KEY={{ .Data.data.api_key }}
          {{- end -}}
        # File mounted at /vault/secrets/config
        # Application reads file, not env var (harder to exfiltrate via env dump)
```

---

## 3. Network Security

### Zero-Trust with Istio mTLS

```yaml
# Enforce strict mTLS across all pods in namespace
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
  namespace: production
spec:
  mtls:
    mode: STRICT    # reject all non-mTLS traffic

---
# AuthorizationPolicy — only allow specific services to communicate
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: order-service-authz
  namespace: production
spec:
  selector:
    matchLabels:
      app: order-service
  rules:
    - from:
        - source:
            principals:
              - "cluster.local/ns/production/sa/api-gateway"
              - "cluster.local/ns/production/sa/order-worker"
      to:
        - operation:
            methods: ["POST", "GET"]
            paths: ["/api/orders*"]
```

---

## 4. IAM & Least Privilege

### IRSA — IAM Roles for Service Accounts (AWS)

```hcl
# Terraform: create IAM role for order-service
data "aws_iam_policy_document" "order_service_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRoleWithWebIdentity"]
    principals {
      type        = "Federated"
      identifiers = [aws_iam_openid_connect_provider.eks.arn]
    }
    condition {
      test     = "StringEquals"
      variable = "${local.oidc_issuer}:sub"
      values   = ["system:serviceaccount:production:order-service"]
    }
    condition {
      test     = "StringEquals"
      variable = "${local.oidc_issuer}:aud"
      values   = ["sts.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "order_service" {
  name               = "order-service-production"
  assume_role_policy = data.aws_iam_policy_document.order_service_assume.json
  # NO: AdministratorAccess
  # YES: only what the service needs
}

resource "aws_iam_role_policy" "order_service" {
  role = aws_iam_role.order_service.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["secretsmanager:GetSecretValue"]
        Resource = "arn:aws:secretsmanager:us-east-1:${data.aws_caller_identity.current.account_id}:secret:production/order-service*"
      },
      {
        Effect   = "Allow"
        Action   = ["sqs:SendMessage", "sqs:ReceiveMessage", "sqs:DeleteMessage"]
        Resource = aws_sqs_queue.order_events.arn
      }
    ]
  })
}
```

---

## 5. Security Scanning Pipeline

```yaml
# .github/workflows/security.yml
name: Security Scans

on:
  push:
    branches: [main]
  schedule:
    - cron: '0 2 * * *'   # nightly scan of main branch

jobs:
  container-scan:
    runs-on: ubuntu-latest
    steps:
      - name: Trivy — container vulnerabilities
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: image
          image-ref: ${{ env.IMAGE }}
          severity: CRITICAL,HIGH
          exit-code: 1
          ignore-unfixed: true   # skip CVEs with no patch yet

  sast:
    runs-on: ubuntu-latest
    steps:
      - name: Semgrep — static analysis
        uses: semgrep/semgrep-action@v1
        with:
          config: |
            p/golang
            p/secrets
            p/docker
            p/terraform

  dependency-audit:
    runs-on: ubuntu-latest
    steps:
      - name: Nancy — Go dependency vulnerabilities
        run: |
          go list -json -m all | nancy sleuth --exit-code

  secret-scan:
    runs-on: ubuntu-latest
    steps:
      - name: Gitleaks — find secrets in code
        uses: gitleaks/gitleaks-action@v2
        with:
          config-path: .gitleaks.toml
```

---

## Security Checklist — Pre-Production

- [ ] All images scanned (no CRITICAL CVEs)
- [ ] Images signed with Cosign
- [ ] No secrets in environment variables (use files or ESO)
- [ ] Pod runs as non-root
- [ ] `readOnlyRootFilesystem: true`
- [ ] All capabilities dropped
- [ ] Network policies: default-deny + explicit allows
- [ ] mTLS enabled between services (Istio/Linkerd)
- [ ] RBAC: service account with minimal permissions
- [ ] IRSA configured (no static AWS credentials)
- [ ] Admission controllers: OPA/Gatekeeper or Kyverno
- [ ] Audit logging enabled on API server
- [ ] Secrets rotation configured (ESO refresh interval)
