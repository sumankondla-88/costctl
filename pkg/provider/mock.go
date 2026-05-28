package provider

import (
	"time"
)

type MockProvider struct {
	Name CloudType
}

func NewMockProvider(name CloudType) *MockProvider {
	return &MockProvider{Name: name}
}

func (m *MockProvider) GetName() CloudType {
	return m.Name
}

func (m *MockProvider) GetCosts(days int) ([]CostItem, error) {
	var items []CostItem
	now := time.Now()

	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -i)
		
		if m.Name == AWS {
			// AWS mock daily costs
			items = append(items, CostItem{
				Provider: AWS,
				Region:   "us-east-1",
				Service:  "EC2-Instances",
				Cost:     125.50 + float64(i%5)*12.20,
				Date:     date,
				Tags:     map[string]string{"Owner": "dev-team", "env": "staging", "Project": "frontend"},
			})
			items = append(items, CostItem{
				Provider: AWS,
				Region:   "us-west-2",
				Service:  "RDS",
				Cost:     85.00,
				Date:     date,
				Tags:     map[string]string{"owner-contact": "database-admins", "env": "production"},
			})
			items = append(items, CostItem{
				Provider: AWS,
				Region:   "us-east-1",
				Service:  "S3",
				Cost:     45.20 - float64(i%3)*2.50,
				Date:     date,
				Tags:     map[string]string{"Project": "frontend", "env": "production"},
			})
		}

		if m.Name == Azure {
			// Azure mock daily costs
			items = append(items, CostItem{
				Provider: Azure,
				Region:   "eastus",
				Service:  "Virtual Machines",
				Cost:     140.00 + float64(i%4)*8.50,
				Date:     date,
				Tags:     map[string]string{"CreatedBy": "Admin", "Stage": "dev"},
			})
			items = append(items, CostItem{
				Provider: Azure,
				Region:   "westeurope",
				Service:  "SQL Databases",
				Cost:     95.00,
				Date:     date,
				Tags:     map[string]string{"app": "billing-service", "Stage": "prod"},
			})
		}

		if m.Name == GCP {
			// GCP mock daily costs
			items = append(items, CostItem{
				Provider: GCP,
				Region:   "us-central1",
				Service:  "Compute Engine",
				Cost:     110.00 + float64(i%6)*5.00,
				Date:     date,
				Tags:     map[string]string{"owner": "alice", "environment": "production"},
			})
			items = append(items, CostItem{
				Provider: GCP,
				Region:   "us-central1",
				Service:  "Cloud Storage",
				Cost:     28.40,
				Date:     date,
				Tags:     map[string]string{"project": "data-analytics", "environment": "production"},
			})
		}
	}

	return items, nil
}

func (m *MockProvider) GetResources() ([]Resource, error) {
	now := time.Now()
	var resources []Resource
	
	if m.Name == AWS {
		// AWS Resources
		resources = append(resources, Resource{
			ID:           "arn:aws:ec2:us-east-1:123456789012:instance/i-0abcd1234efgh5678",
			Name:         "dev-web-server",
			Type:         VM,
			Provider:     AWS,
			Region:       "us-east-1",
			CostPerMonth: 120.00,
			Tags:         map[string]string{"Owner": "dev-team", "env": "staging", "Project": "web-frontend"},
			Status:       "running",
			LaunchTime:   now.AddDate(0, -3, 0),
			Metrics:      map[string]float64{"CPUUtilization": 0.8}, // Idle VM
		}, Resource{
			ID:           "arn:aws:ec2:us-east-1:123456789012:volume/vol-0987654321fedcba0",
			Name:         "temp-data-volume",
			Type:         Volume,
			Provider:     AWS,
			Region:       "us-east-1",
			CostPerMonth: 45.00,
			Tags:         map[string]string{"Project": "legacy-app"},
			Status:       "available", // Unattached
			LaunchTime:   now.AddDate(0, 0, -10),
			Metrics:      map[string]float64{},
		}, Resource{
			ID:           "arn:aws:rds:us-west-2:123456789012:db:prod-analytics-replica",
			Name:         "prod-analytics-replica",
			Type:         Database,
			Provider:     AWS,
			Region:       "us-west-2",
			CostPerMonth: 350.00,
			Tags:         map[string]string{"owner-contact": "db-team@company.com", "env": "production"},
			Status:       "available",
			LaunchTime:   now.AddDate(0, -6, 0),
			Metrics:      map[string]float64{"CPUUtilization": 1.2, "DatabaseConnections": 0.0}, // Idle database
		}, Resource{
			ID:           "arn:aws:ec2:us-east-1:123456789012:eip/eipalloc-0123456789abcdef0",
			Name:         "unassociated-ip",
			Type:         PublicIP,
			Provider:     AWS,
			Region:       "us-east-1",
			CostPerMonth: 3.60,
			Tags:         map[string]string{"env": "staging"},
			Status:       "unassociated", // Orphaned IP
			LaunchTime:   now.AddDate(0, 0, -30),
			Metrics:      map[string]float64{},
		}, Resource{
			ID:           "arn:aws:ec2:us-east-1:123456789012:natgateway/nat-04fa3e9c56bd23b82",
			Name:         "dev-nat-gateway",
			Type:         NATGateway,
			Provider:     AWS,
			Region:       "us-east-1",
			CostPerMonth: 32.40,
			Tags:         map[string]string{"Owner": "dev-team", "env": "staging"},
			Status:       "available",
			LaunchTime:   now.AddDate(0, -1, 0),
			Metrics:      map[string]float64{"ProcessedBytes": 0.0}, // Idle NAT Gateway
		}, Resource{
			ID:           "arn:aws:ec2:us-east-1:123456789012:snapshot/snap-01a2b3c4d5e6f7g8h",
			Name:         "backup-db-2024",
			Type:         Snapshot,
			Provider:     AWS,
			Region:       "us-east-1",
			CostPerMonth: 80.00,
			Tags:         map[string]string{"CreatedBy": "automation", "Project": "database"},
			Status:       "completed",
			LaunchTime:   now.AddDate(-1, -2, 0), // Very old snapshot (stale)
			Metrics:      map[string]float64{},
		}, Resource{
			ID:           "arn:aws:ec2:us-east-1:123456789012:instance/i-activewebprod",
			Name:         "prod-web-server-1",
			Type:         VM,
			Provider:     AWS,
			Region:       "us-east-1",
			CostPerMonth: 240.00,
			Tags:         map[string]string{"Owner": "ops-team", "env": "production", "Project": "web-frontend"},
			Status:       "running",
			LaunchTime:   now.AddDate(0, -10, 0),
			Metrics:      map[string]float64{"CPUUtilization": 45.2}, // High CPU, active
		}, Resource{
			ID:           "arn:aws:ec2:us-east-1:123456789012:volume/vol-activeprodvol",
			Name:         "prod-web-storage",
			Type:         Volume,
			Provider:     AWS,
			Region:       "us-east-1",
			CostPerMonth: 120.00,
			Tags:         map[string]string{"env": "production", "Project": "web-frontend"},
			Status:       "in-use", // Attached and active
			LaunchTime:   now.AddDate(0, -10, 0),
			Metrics:      map[string]float64{},
		})
	}

	if m.Name == Azure {
		// Azure Resources
		resources = append(resources, Resource{
			ID:           "/subscriptions/sub-1234/resourceGroups/rg-dev/providers/Microsoft.Compute/virtualMachines/vm-legacy-testing",
			Name:         "vm-legacy-testing",
			Type:         VM,
			Provider:     Azure,
			Region:       "eastus",
			CostPerMonth: 160.00,
			Tags:         map[string]string{"CreatedBy": "john", "Stage": "dev", "Project": "legacy-migration"},
			Status:       "VM running",
			LaunchTime:   now.AddDate(0, -2, -15),
			Metrics:      map[string]float64{"CPUUtilization": 1.5}, // Idle Azure VM
		}, Resource{
			ID:           "/subscriptions/sub-1234/resourceGroups/rg-dev/providers/Microsoft.Compute/disks/disk-detached-os-backup",
			Name:         "disk-detached-os-backup",
			Type:         Volume,
			Provider:     Azure,
			Region:       "eastus",
			CostPerMonth: 75.00,
			Tags:         map[string]string{"app": "billing-service"},
			Status:       "Unattached", // Detached disk
			LaunchTime:   now.AddDate(0, 0, -25),
			Metrics:      map[string]float64{},
		}, Resource{
			ID:           "/subscriptions/sub-1234/resourceGroups/rg-prod/providers/Microsoft.Sql/servers/sql-prod-costctl/databases/db-test-sandbox",
			Name:         "db-test-sandbox",
			Type:         Database,
			Provider:     Azure,
			Region:       "westeurope",
			CostPerMonth: 110.00,
			Tags:         map[string]string{"Stage": "dev", "Owner": "qa-team"},
			Status:       "Online",
			LaunchTime:   now.AddDate(0, -1, -5),
			Metrics:      map[string]float64{"CPUUtilization": 0.1, "DatabaseConnections": 0.0}, // Idle SQL Database
		})
	}

	if m.Name == GCP {
		// GCP Resources
		resources = append(resources, Resource{
			ID:           "projects/gcp-dev-project/zones/us-central1-a/instances/gce-dev-bastion",
			Name:         "gce-dev-bastion",
			Type:         VM,
			Provider:     GCP,
			Region:       "us-central1",
			CostPerMonth: 85.00,
			Tags:         map[string]string{"owner": "alice", "environment": "dev", "project": "bastion-hosts"},
			Status:       "RUNNING",
			LaunchTime:   now.AddDate(0, -4, 0),
			Metrics:      map[string]float64{"CPUUtilization": 0.4}, // Idle VM
		}, Resource{
			ID:           "projects/gcp-dev-project/zones/us-central1-a/disks/gce-unattached-pd",
			Name:         "gce-unattached-pd",
			Type:         Volume,
			Provider:     GCP,
			Region:       "us-central1",
			CostPerMonth: 25.00,
			Tags:         map[string]string{"project": "data-analytics"},
			Status:       "UNUSED", // Detached disk
			LaunchTime:   now.AddDate(0, 0, -8),
			Metrics:      map[string]float64{},
		})
	}

	return resources, nil
}

// Ensure MockProvider implements Provider interface
var _ Provider = (*MockProvider)(nil)
