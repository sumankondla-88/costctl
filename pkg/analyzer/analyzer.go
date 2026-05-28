package analyzer

import (
	"fmt"
	"strings"
	"time"

	"costctl/pkg/config"
	"costctl/pkg/provider"
)

// Finding represents an identified optimization opportunity
type Finding struct {
	ResourceID       string            `json:"resource_id"`
	ResourceName     string            `json:"resource_name"`
	ResourceType     string            `json:"resource_type"`
	Provider         provider.CloudType `json:"provider"`
	Region           string            `json:"region"`
	CostPerMonth     float64           `json:"cost_per_month"`
	SavingsPotential float64           `json:"savings_potential"`
	Reason           string            `json:"reason"`
	Tags             map[string]string `json:"tags"`
	NormalizedTags   map[string]string `json:"normalized_tags"`
}

// AnalysisResult represents the aggregate outcome of the optimization audit
type AnalysisResult struct {
	Findings              []Finding `json:"findings"`
	TotalMonthlyCost      float64   `json:"total_monthly_cost"`
	TotalSavingsPotential float64   `json:"total_savings_potential"`
	WastePercent          float64   `json:"waste_percent"`
}

// AnalyzeResources runs FinOps heuristic rules over collected assets
func AnalyzeResources(resources []provider.Resource, cfg *config.Config) *AnalysisResult {
	var findings []Finding
	var totalCost float64
	var totalSavings float64

	for _, res := range resources {
		totalCost += res.CostPerMonth
		finding := checkWaste(res, cfg)
		if finding != nil {
			findings = append(findings, *finding)
			totalSavings += finding.SavingsPotential
		}
	}

	wastePercent := 0.0
	if totalCost > 0 {
		wastePercent = (totalSavings / totalCost) * 100.0
	}

	return &AnalysisResult{
		Findings:              findings,
		TotalMonthlyCost:      totalCost,
		TotalSavingsPotential: totalSavings,
		WastePercent:          wastePercent,
	}
}

func checkWaste(res provider.Resource, cfg *config.Config) *Finding {
	// Normalize status to lowercase for robust checks
	status := strings.ToLower(res.Status)
	
	switch res.Type {
	case provider.VM:
		// Check CPU Utilization if VM is running
		if (status == "running" || status == "vm running" || status == "ready") && res.Metrics != nil {
			if cpu, exists := res.Metrics["CPUUtilization"]; exists {
				if cpu < cfg.Thresholds.CPUUtilizationPercent {
					return &Finding{
						ResourceID:       res.ID,
						ResourceName:     res.Name,
						ResourceType:     string(res.Type),
						Provider:         res.Provider,
						Region:           res.Region,
						CostPerMonth:     res.CostPerMonth,
						SavingsPotential: res.CostPerMonth,
						Reason:           fmt.Sprintf("Idle VM: Avg CPU utilization is %.1f%% (threshold: %.1f%%)", cpu, cfg.Thresholds.CPUUtilizationPercent),
						Tags:             res.Tags,
						NormalizedTags:   res.NormalizedTags,
					}
				}
			}
		}

	case provider.Volume:
		// Check for detached disks
		if status == "available" || status == "unattached" || status == "detached" || status == "unused" {
			daysUnattached := int(time.Since(res.LaunchTime).Hours() / 24.0)
			// Ensure it has been unattached for at least the configured threshold
			if daysUnattached >= cfg.Thresholds.DiskUnattachedDays {
				return &Finding{
					ResourceID:       res.ID,
					ResourceName:     res.Name,
					ResourceType:     string(res.Type),
					Provider:         res.Provider,
					Region:           res.Region,
					CostPerMonth:     res.CostPerMonth,
					SavingsPotential: res.CostPerMonth,
					Reason:           fmt.Sprintf("Unattached Disk Volume: Status is %s for %d days (threshold: %d days)", res.Status, daysUnattached, cfg.Thresholds.DiskUnattachedDays),
					Tags:             res.Tags,
					NormalizedTags:   res.NormalizedTags,
				}
			}
		}

	case provider.Database:
		// Check database connections and CPU
		if res.Metrics != nil {
			conns, hasConns := res.Metrics["DatabaseConnections"]
			cpu, hasCPU := res.Metrics["CPUUtilization"]
			
			isIdle := false
			reason := ""
			
			if hasConns && conns == 0 {
				isIdle = true
				reason = "Idle Database: 0 active connections"
			} else if hasCPU && cpu < cfg.Thresholds.CPUUtilizationPercent {
				isIdle = true
				reason = fmt.Sprintf("Idle Database: Avg CPU utilization is %.1f%% (threshold: %.1f%%)", cpu, cfg.Thresholds.CPUUtilizationPercent)
			}

			if isIdle {
				return &Finding{
					ResourceID:       res.ID,
					ResourceName:     res.Name,
					ResourceType:     string(res.Type),
					Provider:         res.Provider,
					Region:           res.Region,
					CostPerMonth:     res.CostPerMonth,
					SavingsPotential: res.CostPerMonth,
					Reason:           reason,
					Tags:             res.Tags,
					NormalizedTags:   res.NormalizedTags,
				}
			}
		}

	case provider.PublicIP:
		// Check if IP is not associated with any running resource
		if status == "unassociated" || status == "detached" || status == "available" {
			return &Finding{
				ResourceID:       res.ID,
				ResourceName:     res.Name,
				ResourceType:     string(res.Type),
				Provider:         res.Provider,
				Region:           res.Region,
				CostPerMonth:     res.CostPerMonth,
				SavingsPotential: res.CostPerMonth,
				Reason:           "Unassociated IP Address: Incurring idle reservation charges",
				Tags:             res.Tags,
				NormalizedTags:   res.NormalizedTags,
			}
		}

	case provider.NATGateway:
		// Check data processed
		if res.Metrics != nil {
			if bytes, exists := res.Metrics["ProcessedBytes"]; exists && bytes == 0 {
				return &Finding{
					ResourceID:       res.ID,
					ResourceName:     res.Name,
					ResourceType:     string(res.Type),
					Provider:         res.Provider,
					Region:           res.Region,
					CostPerMonth:     res.CostPerMonth,
					SavingsPotential: res.CostPerMonth,
					Reason:           "Idle NAT Gateway: Zero data processed over the metric collection period",
					Tags:             res.Tags,
					NormalizedTags:   res.NormalizedTags,
				}
			}
		}

	case provider.Snapshot:
		// Check age
		ageDays := int(time.Since(res.LaunchTime).Hours() / 24.0)
		if ageDays > cfg.Thresholds.SnapshotRetentionDays {
			return &Finding{
				ResourceID:       res.ID,
				ResourceName:     res.Name,
				ResourceType:     string(res.Type),
				Provider:         res.Provider,
				Region:           res.Region,
				CostPerMonth:     res.CostPerMonth,
				SavingsPotential: res.CostPerMonth,
				Reason:           fmt.Sprintf("Stale Snapshot: Age is %d days (threshold: %d days)", ageDays, cfg.Thresholds.SnapshotRetentionDays),
				Tags:             res.Tags,
				NormalizedTags:   res.NormalizedTags,
			}
		}
	}

	return nil
}
