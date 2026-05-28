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

Opens `report.html` with a real-time dashboard showing:
- Consolidated spend across 3 clouds ($60K+/month example)
- **$1,086/month in detected waste** (idle VMs, unattached volumes, stale snapshots)
- Waste flagged per resource with resize/decommission recommendations

## Why costctl?

| Problem | Typical Cost | costctl |
|---------|---|---|
| CloudHealth / Densify | $50K+/year | Free & open-source |
| Multi-cloud visibility | Manual CSV exports | Unified CLI + HTML dashboard |
| Waste detection | Vague recommendations | Specific resources + savings estimates |
| CI/CD gates | Not included | Built-in (`--fail-on-waste`) |

## Features

- **Multi-Cloud Aggregation**: AWS Cost Explorer, Azure Cost Management, GCP Cloud Billing → single queryable schema
- **Waste Detection**: Idle compute, unattached disks, orphaned IPs, stale snapshots, idle databases
- **Tag Normalization**: Map provider-specific tag variants (Owner/owner/owner-contact) to standard dimensions
- **HTML Dashboard**: Glassmorphic SPA with time filtering, cloud toggles, exportable findings
- **CI/CD Gates**: Fail pipelines when waste exceeds thresholds (`--fail-on-waste`, `--max-waste`)

## Installation

```bash
git clone https://github.com/sumankondla-88/costctl.git
cd costctl
go build -o costctl main.go
```

## Usage

### Demo Mode (No Credentials)
```bash
./costctl --demo cost summary      # Show spend by cloud
./costctl --demo waste find         # Show waste findings (console table)
./costctl --demo waste find --format html  # Generate dashboard report
```

### Real Cloud Data
```bash
# Create .costctl.yaml with your cloud credentials (see docs/)
./costctl cost summary
./costctl waste find --format json  # Pipe to automation
```

### CI/CD Compliance Gate
```bash
# Fail pipeline if waste > $500/month
./costctl waste find --fail-on-waste

# Override threshold
./costctl waste find --fail-on-waste --max-waste 1000
```

## Configuration

See `.costctl.yaml.example` for tag mapping, thresholds, and budget settings.

## Contributing

PRs welcome. Help us add more waste heuristics, cloud providers, or integrations.

## License

MIT
