# costctl

**Multi-cloud spend visibility + waste detection in one CLI.**

Find orphaned resources, idle VMs, and unattached volumes across AWS, Azure, and GCP. 
No credentials needed to demo. No SaaS bloat. Just a single binary.

## Quick Start (2 minutes)

```bash
# Try it now with mock data (no setup required)
./costctl --demo cost summary
./costctl --demo waste find
./costctl --demo waste find --format html
```

Opens `report.html` with a real-time dashboard. See example below.

## Dashboard Preview

**Consolidated spend across 3 clouds with waste detection:**

![costctl Dashboard](https://github.com/sumankondla-88/costctl/raw/main/docs/screenshot-dashboard.png)

**Actionable waste findings by resource:**

![costctl Detailed Reports](https://github.com/sumankondla-88/costctl/raw/main/docs/screenshot-reports.png)

## Why costctl?

| Feature | SaaS FinOps Tools | costctl |
|---------|---|---|
| Cost | $50K+/year | Free & open-source |
| Multi-cloud visibility | ✓ (with friction) | ✓ (unified CLI) |
| Waste detection | Vague | Specific resources + savings |
| Setup time | Weeks | 2 minutes |
| CI/CD gates | Extra cost | Built-in |

## Features

- **Multi-Cloud Aggregation**: AWS, Azure, GCP → single queryable schema
- **Waste Detection**: Idle compute, unattached disks, orphaned IPs, stale snapshots, idle databases
- **Tag Normalization**: Map provider-specific tag variants to standard dimensions
- **HTML Dashboard**: Time filtering, cloud toggles, exportable findings
- **CI/CD Gates**: Fail pipelines when waste exceeds thresholds

## Installation

```bash
git clone https://github.com/sumankondla-88/costctl.git
cd costctl
go build -o costctl main.go
```

## Usage

### Demo Mode (No Credentials)
```bash
./costctl --demo cost summary
./costctl --demo waste find
./costctl --demo waste find --format html  # Opens report.html
```

### Real Cloud Data
```bash
./costctl cost summary
./costctl waste find --format json
```

### CI/CD Compliance
```bash
# Fail pipeline if waste > $500/month
./costctl waste find --fail-on-waste

# Custom threshold
./costctl waste find --fail-on-waste --max-waste 1000
```

## Configuration

See `.costctl.yaml.example` for tag mapping, thresholds, and cloud setup.

## License

MIT
