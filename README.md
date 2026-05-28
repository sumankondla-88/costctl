# costctl

`costctl` is a unified Multi-Cloud DevOps Spend Optimizer and FinOps command-line tool written in Go. It aggregates cost data, maps resources to a normalized schema, and executes waste-detection heuristics across AWS, Azure, and GCP. 

The tool includes a premium glassmorphic Single-Page Application (SPA) HTML report featuring integrated OAuth/SSO simulations, team auto-provisioning directory sync, live spend stats scaling, and custom tag connectors. It also functions as a CI/CD compliance gate to automatically block pipelines when cloud waste limits are violated.

---

## Key Features

- **Multi-Cloud Billing Aggregation**: Maps diverse provider cost structures (AWS Cost Explorer, Azure Cost Management, GCP Cloud Billing) into a unified, queryable schema.
- **Waste Asset Detection Rules**:
  - **Compute (VMs)**: Idle virtual machines (CPU utilization under thresholds over 7 days).
  - **Storage (Disks)**: Unattached EBS volumes, Azure Managed Disks, and GCE Persistent Disks.
  - **Networking (IPs)**: Orphaned Elastic/Public IPs not associated with active instances or load balancers.
  - **Backups (Snapshots)**: Stale or orphaned database/volume snapshots older than retention policies.
  - **Databases**: Idle RDS or Cloud SQL instances with no active connections or low CPU.
- **Unified Tag Normalization**: Maps mismatched tag keys (e.g., `Owner`, `owner`, `owner-contact`) into standard target dimensions defined in a single configuration.
- **CI/CD Quality Gates**: Exit-code policy checks (`--fail-on-waste` and `--max-waste`) to automatically fail automated pipelines when waste thresholds are breached.
- **Interactive SPA HTML Dashboard**: Generates a gorgeous glassmorphic dashboard report with:
  - **SSO Gate Portal**: Google Workspace and Microsoft Entra ID login flows.
  - **Consolidated Invoicing**: Multi-cloud invoice aggregating linked accounts and active subscriptions.
  - **Timeline Scaling**: Recalculates spend metrics dynamically across **7D**, **30D**, **QTR**, **1Y**, and **Custom** date ranges.
  - **Account Connector Form**: Dynamically connect and disconnect cloud accounts with custom key-value tags.
  - **Admin Settings & Provisioning**: Invite teammates with scopes (Admin, Operator, Viewer) and trigger Google directory sync.

---

## Installation

### Prerequisites
- Go 1.18 or higher (Go 1.26+ recommended)

### Build from Source
```bash
# Clone the repository
git clone https://github.com/<username>/costctl.git
cd costctl

# Build the native binary
go build -o costctl main.go
```

---

## Usage

### 1. Configuration
`costctl` looks for a `.costctl.yaml` file in the current directory or a custom path:
```yaml
clouds:
  - aws
  - azure
  - gcp

tag_mappings:
  environment:
    - env
    - environment
    - stage
  owner:
    - owner
    - Creator
  project:
    - project
    - app

thresholds:
  cpu_utilization_percent: 5
  idle_days: 7
  snapshot_retention_days: 30
  disk_unattached_days: 3

budget:
  total_monthly: 10000
  fail_on_waste_threshold: 500
```

### 2. Basic Commands

#### Run a dry-run scan (Demo Mode)
```bash
# View active cost spend summary by provider
./costctl --demo cost summary

# View detailed console waste audit table
./costctl --demo waste find

# View detailed JSON waste findings for automation pipelines
./costctl --demo waste find --format json
```

#### Generate HTML SPA Dashboard
```bash
./costctl --demo waste find --format html
# Generates "report.html" in the current directory
```

#### Enforce CI/CD Budget Thresholds
```bash
# Fails with exit code 1 if potential monthly waste savings exceed default budget threshold ($500)
./costctl --demo waste find --fail-on-waste

# Override threshold dynamically ($1,000 threshold)
./costctl --demo waste find --fail-on-waste --max-waste 1000
```

---

## License & Trademark

© 2026 costctl. Author: **costctl contributors**. Trademark ™ All rights reserved.
