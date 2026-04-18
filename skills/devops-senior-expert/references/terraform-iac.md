# Terraform & IaC — Production Module Design

## Table of Contents
1. [Module Structure](#modules)
2. [State Management](#state)
3. [Workspace Strategy](#workspaces)
4. [Security Patterns](#security)
5. [Anti-Patterns](#antipatterns)

---

## 1. Module Structure

```
terraform/
├── modules/                    # Reusable modules (versioned separately)
│   ├── eks-cluster/
│   │   ├── main.tf
│   │   ├── variables.tf        # All inputs documented with type + validation
│   │   ├── outputs.tf          # Only expose what callers need
│   │   ├── versions.tf         # Required versions pinned
│   │   └── README.md           # Auto-generated via terraform-docs
│   ├── rds-postgres/
│   └── vpc/
├── environments/
│   ├── staging/
│   │   ├── main.tf             # Module calls only, no resources directly
│   │   ├── variables.tf
│   │   ├── terraform.tfvars    # Non-sensitive defaults
│   │   └── backend.tf          # S3 remote state
│   └── production/
│       └── ...
└── shared/                     # Cross-environment resources
    ├── route53-zones/
    └── ecr-repos/
```

### Module Contract — variables.tf

```hcl
variable "environment" {
  type        = string
  description = "Deployment environment (staging, production)"
  validation {
    condition     = contains(["staging", "production"], var.environment)
    error_message = "Environment must be 'staging' or 'production'."
  }
}

variable "cluster_version" {
  type        = string
  description = "EKS Kubernetes version"
  default     = "1.29"
  validation {
    condition     = can(regex("^1\\.(2[5-9]|[3-9][0-9])$", var.cluster_version))
    error_message = "Must be Kubernetes 1.25 or higher."
  }
}

variable "node_groups" {
  type = map(object({
    instance_types = list(string)
    min_size       = number
    max_size       = number
    desired_size   = number
    disk_size_gb   = optional(number, 50)
    labels         = optional(map(string), {})
    taints         = optional(list(object({
      key    = string
      value  = string
      effect = string
    })), [])
  }))
  description = "Map of node group configurations"
}
```

### outputs.tf — Expose Minimally

```hcl
output "cluster_endpoint" {
  description = "EKS cluster API endpoint"
  value       = aws_eks_cluster.this.endpoint
  # Don't expose: sensitive IDs, internal ARNs callers don't need
}

output "cluster_certificate_authority" {
  description = "Base64 encoded cluster CA certificate"
  value       = aws_eks_cluster.this.certificate_authority[0].data
  sensitive   = true   # marked sensitive — won't print in plan output
}
```

---

## 2. State Management

### S3 Backend with Locking

```hcl
# backend.tf
terraform {
  backend "s3" {
    bucket         = "company-terraform-state"
    key            = "environments/production/eks/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true                           # AES-256 at rest
    kms_key_id     = "arn:aws:kms:us-east-1:..."   # CMK, not AWS managed
    dynamodb_table = "terraform-state-lock"          # prevents concurrent applies
    
    # Versioning must be enabled on the bucket
    # Never delete state — version history is your undo button
  }
}
```

### State Isolation Strategy

```
One state file per:
  - Environment × Service boundary
  - Examples:
    production/network/terraform.tfstate   (VPC, subnets, etc.)
    production/eks/terraform.tfstate       (cluster, node groups)
    production/rds/terraform.tfstate       (databases)
    production/apps/terraform.tfstate      (IAM roles for apps)

WHY: Small blast radius. A bad apply to `apps` can't destroy your VPC.
     Use data sources to reference cross-state outputs.
```

### Cross-State References

```hcl
# In eks/main.tf — reference VPC from separate state
data "terraform_remote_state" "network" {
  backend = "s3"
  config = {
    bucket = "company-terraform-state"
    key    = "environments/production/network/terraform.tfstate"
    region = "us-east-1"
  }
}

resource "aws_eks_cluster" "this" {
  # ...
  vpc_config {
    subnet_ids = data.terraform_remote_state.network.outputs.private_subnet_ids
  }
}
```

---

## 3. Workspace Strategy

**Opinion**: Use directories over workspaces for environment separation.
Workspaces share backend config; easy to accidentally apply to wrong env.
Separate directories with separate state files = explicit, safe.

```bash
# GOOD: explicit directory per environment
cd environments/production && terraform apply

# BAD: workspace confusion
terraform workspace select production
terraform apply   # are you SURE you're in production?
```

---

## 4. Security Patterns

### IAM Least Privilege for Terraform CI

```hcl
# IAM policy for CI/CD Terraform runner — explicit, no wildcards
data "aws_iam_policy_document" "terraform_ci" {
  statement {
    sid    = "EKSReadWrite"
    effect = "Allow"
    actions = [
      "eks:CreateCluster",
      "eks:UpdateClusterConfig",
      "eks:DescribeCluster",
      "eks:DeleteCluster",
    ]
    resources = [
      "arn:aws:eks:us-east-1:${data.aws_caller_identity.current.account_id}:cluster/production-*"
    ]
  }

  statement {
    sid    = "DenyDeleteProductionResources"
    effect = "Deny"
    actions = [
      "rds:DeleteDBInstance",
      "dynamodb:DeleteTable",
      "s3:DeleteBucket",
    ]
    resources = ["*"]
    condition {
      test     = "StringEquals"
      variable = "aws:ResourceTag/Environment"
      values   = ["production"]
    }
  }
}
```

### Sensitive Values — Never in tfvars, Always in Secrets Manager

```hcl
# Read secret from AWS Secrets Manager at plan/apply time
data "aws_secretsmanager_secret_version" "db_password" {
  secret_id = "production/order-service/db-password"
}

resource "aws_db_instance" "this" {
  password = jsondecode(data.aws_secretsmanager_secret_version.db_password.secret_string)["password"]
}
```

---

## 5. Anti-Patterns

| Anti-Pattern | Risk | Fix |
|---|---|---|
| Local state | Team can't collaborate, lost on laptop | S3 backend + DynamoDB lock |
| No state locking | Two applies simultaneously = corruption | DynamoDB locking table |
| `count` for environment branching | Confusing index drift | Separate directories per environment |
| Hardcoded account IDs, region | Not reusable, security risk | `data "aws_caller_identity"` + variables |
| `terraform apply` in CI without plan review | Surprise infra destruction | Always `plan` → human review → `apply` |
| No resource tagging | Unattributed costs, audit fails | `local.common_tags` on everything |
| Unconstrained provider versions | Breaking changes silently | `version = "~> 5.0"` in all modules |
| Running with admin IAM | Blast radius = everything | Least-privilege role per environment |
| Secrets in `.tfvars` in git | Credentials in repo | Secrets Manager data sources |
| Giant monolithic state | One failed apply blocks all teams | Split by layer (network/compute/apps) |

---

## Terraform CI Pipeline

```yaml
# .github/workflows/terraform.yml
name: Terraform

on:
  pull_request:
    paths: ['terraform/**']

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: "~1.7"

      - name: Terraform Format Check
        run: terraform fmt -check -recursive terraform/

      - name: Validate all modules
        run: |
          for dir in terraform/modules/*/; do
            terraform -chdir="$dir" init -backend=false
            terraform -chdir="$dir" validate
          done

      - name: tflint
        uses: terraform-linters/setup-tflint@v4
      - run: tflint --recursive

      - name: tfsec (security scan)
        uses: aquasecurity/tfsec-action@v1.0.3
        with:
          soft_fail: false

  plan:
    needs: validate
    runs-on: ubuntu-latest
    environment: staging-plan   # OIDC role for read-only plan
    steps:
      - name: Terraform Plan
        run: |
          terraform -chdir=terraform/environments/staging init
          terraform -chdir=terraform/environments/staging plan \
            -out=tfplan \
            -var-file=terraform.tfvars

      - name: Comment plan on PR
        uses: borchero/terraform-plan-comment@v2
        with:
          plan-path: tfplan
