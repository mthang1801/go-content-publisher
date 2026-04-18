# CI/CD & GitOps — Production Pipeline Patterns

## Table of Contents
1. [GitHub Actions — Production Pipeline](#github-actions)
2. [Pipeline Security & Supply Chain](#security)
3. [ArgoCD GitOps](#argocd)
4. [Release Strategies](#release)
5. [Anti-Patterns](#antipatterns)

---

## 1. GitHub Actions — Production Pipeline

### Full Pipeline: Test → Build → Scan → Deploy

```yaml
# .github/workflows/deploy.yml
name: CI/CD Pipeline

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  test:
    name: Test & Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest

      - name: Run tests with race detector
        run: go test -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v4

  build:
    name: Build & Scan Image
    needs: test
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
      id-token: write      # for OIDC signing
    outputs:
      image-digest: ${{ steps.build.outputs.digest }}

    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to GHCR
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Build and push
        id: build
        uses: docker/build-push-action@v6
        with:
          context: .
          push: ${{ github.ref == 'refs/heads/main' }}
          tags: |
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:latest
          cache-from: type=gha
          cache-to: type=gha,mode=max
          provenance: true   # SLSA provenance
          sbom: true         # Software Bill of Materials

      - name: Scan with Trivy
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}
          format: sarif
          output: trivy-results.sarif
          severity: CRITICAL,HIGH
          exit-code: 1       # fail build on critical CVEs

      - name: Sign image with Cosign (keyless OIDC)
        uses: sigstore/cosign-installer@v3
      - run: |
          cosign sign --yes \
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}@${{ steps.build.outputs.digest }}

  deploy-staging:
    name: Deploy to Staging
    needs: build
    runs-on: ubuntu-latest
    environment: staging
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4

      - name: Update image tag in GitOps repo
        env:
          GITOPS_TOKEN: ${{ secrets.GITOPS_PAT }}
        run: |
          git clone https://x-access-token:${GITOPS_TOKEN}@github.com/org/gitops-repo.git
          cd gitops-repo
          # Use yq to update image tag
          yq e '.image.tag = "${{ github.sha }}"' -i \
            apps/order-service/overlays/staging/values.yaml
          git config user.email "ci@company.com"
          git config user.name "CI Bot"
          git commit -am "chore: deploy order-service ${{ github.sha }} to staging"
          git push

  deploy-production:
    name: Deploy to Production
    needs: deploy-staging
    runs-on: ubuntu-latest
    environment: production     # requires manual approval in GitHub
    steps:
      - name: Verify image signature
        run: |
          cosign verify \
            --certificate-identity "https://github.com/${{ github.workflow_ref }}" \
            --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}@${{ needs.build.outputs.image-digest }}

      - name: Deploy to production
        run: |
          # Same GitOps pattern — ArgoCD picks up the change
          # ... update production overlay
```

---

## 2. Pipeline Security & Supply Chain

### Secrets in CI — Never Hardcode

```yaml
# Use GitHub OIDC — no long-lived credentials
- name: Configure AWS credentials
  uses: aws-actions/configure-aws-credentials@v4
  with:
    role-to-assume: arn:aws:iam::123456789012:role/github-actions-role
    aws-region: us-east-1
    # NO access keys needed — OIDC token exchange

# The IAM role trust policy (Terraform):
# data "aws_iam_policy_document" "github_actions_trust" {
#   statement {
#     actions = ["sts:AssumeRoleWithWebIdentity"]
#     principals {
#       type = "Federated"
#       identifiers = [aws_iam_openid_connect_provider.github.arn]
#     }
#     condition {
#       test     = "StringEquals"
#       variable = "token.actions.githubusercontent.com:sub"
#       values   = ["repo:org/repo:ref:refs/heads/main"]
#     }
#   }
# }
```

### Pin ALL Actions to Full SHA (supply chain attack prevention)

```yaml
# BAD — can be hijacked by tag mutation
- uses: actions/checkout@v4

# GOOD — pinned to immutable SHA
- uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683  # v4.2.2
```

---

## 3. ArgoCD GitOps

### App of Apps Pattern

```yaml
# Root application that manages all other applications
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: root-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/org/gitops-repo
    targetRevision: main
    path: apps/root
  destination:
    server: https://kubernetes.default.svc
    namespace: argocd
  syncPolicy:
    automated:
      prune: true
      selfHeal: true    # reconcile drift automatically
    syncOptions:
      - CreateNamespace=true
      - PrunePropagationPolicy=foreground
      - ApplyOutOfSyncOnly=true
```

### ApplicationSet — Multi-Cluster / Multi-Environment

```yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: order-service
  namespace: argocd
spec:
  generators:
    - list:
        elements:
          - cluster: staging
            url: https://staging-k8s.example.com
            env: staging
          - cluster: production
            url: https://prod-k8s.example.com
            env: production
  template:
    metadata:
      name: "order-service-{{env}}"
    spec:
      project: "{{cluster}}"
      source:
        repoURL: https://github.com/org/gitops-repo
        targetRevision: main
        path: "apps/order-service/overlays/{{env}}"
      destination:
        server: "{{url}}"
        namespace: production
      syncPolicy:
        automated:
          prune: true
          selfHeal: "{{env == 'staging'}}"   # auto-heal staging, manual for prod
```

### Sync Waves — Ordered Deployment

```yaml
# Wave 0: namespace + RBAC
# Wave 1: ConfigMaps, Secrets (ESO)
# Wave 2: Databases, StatefulSets
# Wave 3: Applications
# Wave 4: Ingress / Routes
metadata:
  annotations:
    argocd.argoproj.io/sync-wave: "3"
```

---

## 4. Release Strategies

### Canary with Argo Rollouts

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: order-service
spec:
  replicas: 10
  strategy:
    canary:
      canaryService: order-service-canary
      stableService: order-service-stable
      steps:
        - setWeight: 5             # 5% traffic to canary
        - pause: {duration: 5m}   # observe 5 minutes
        - analysis:                # automated quality gate
            templates:
              - templateName: success-rate
        - setWeight: 20
        - pause: {duration: 10m}
        - setWeight: 50
        - pause: {duration: 10m}
        - setWeight: 100

      trafficRouting:
        nginx:
          stableIngress: order-service-stable

---
apiVersion: argoproj.io/v1alpha1
kind: AnalysisTemplate
metadata:
  name: success-rate
spec:
  metrics:
    - name: success-rate
      interval: 1m
      successCondition: result[0] >= 0.99   # 99% success rate required
      failureLimit: 3
      provider:
        prometheus:
          address: http://prometheus:9090
          query: |
            sum(rate(http_requests_total{service="order-service-canary",status!~"5.."}[2m]))
            /
            sum(rate(http_requests_total{service="order-service-canary"}[2m]))
```

---

## 5. Anti-Patterns

| Anti-Pattern | Risk | Fix |
|---|---|---|
| Secrets in env vars | Exposed in process list, logs | External Secrets Operator + Vault |
| `:latest` image tag | Non-reproducible deployments | Pin to SHA digest |
| Direct kubectl in CI | No audit trail, race conditions | GitOps — push to repo, not cluster |
| No rollback plan | Hours of downtime on bad deploy | ArgoCD rollback / Argo Rollouts abort |
| Manual prod deploys | Snowflake envs, human error | Fully automated GitOps pipeline |
| Actions pinned to tag | Supply chain attacks | Pin to full SHA |
| One big pipeline | Slow feedback, flaky = block all | Split: fast-feedback / slow-gates |
| Shared CI secrets | Blast radius of leaked cred | OIDC per repo/workflow |
