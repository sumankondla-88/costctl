package analyzer

import (
	"testing"
	"time"

	"costctl/pkg/config"
	"costctl/pkg/provider"
)

func TestAnalyzeResources(t *testing.T) {
	cfg := config.DefaultConfig()
	now := time.Now()

	resources := []provider.Resource{
		// 1. Idle VM (CPU = 1.2% < 5.0% threshold)
		{
			ID:           "id-1",
			Name:         "idle-instance",
			Type:         provider.VM,
			Provider:     provider.AWS,
			Region:       "us-east-1",
			CostPerMonth: 100.00,
			Status:       "running",
			LaunchTime:   now.AddDate(0, -1, 0),
			Metrics:      map[string]float64{"CPUUtilization": 1.2},
		},
		// 2. Active VM (CPU = 12.5% >= 5.0% threshold) - Should NOT flag
		{
			ID:           "id-2",
			Name:         "active-instance",
			Type:         provider.VM,
			Provider:     provider.AWS,
			Region:       "us-east-1",
			CostPerMonth: 200.00,
			Status:       "running",
			LaunchTime:   now.AddDate(0, -1, 0),
			Metrics:      map[string]float64{"CPUUtilization": 12.5},
		},
		// 3. Unattached Volume (Status = available for 10 days > 3 days threshold)
		{
			ID:           "id-3",
			Name:         "detached-disk",
			Type:         provider.Volume,
			Provider:     provider.Azure,
			Region:       "eastus",
			CostPerMonth: 50.00,
			Status:       "available",
			LaunchTime:   now.AddDate(0, 0, -10),
		},
		// 4. Stale Snapshot (Age = 40 days > 30 days threshold)
		{
			ID:           "id-4",
			Name:         "old-snap",
			Type:         provider.Snapshot,
			Provider:     provider.GCP,
			Region:       "us-central1",
			CostPerMonth: 20.00,
			Status:       "completed",
			LaunchTime:   now.AddDate(0, 0, -40),
		},
	}

	result := AnalyzeResources(resources, cfg)

	// We expect 3 waste findings (resources 1, 3, and 4)
	if len(result.Findings) != 3 {
		t.Errorf("Expected 3 findings, got %d", len(result.Findings))
	}

	// Verify total calculations
	expectedTotalMonthly := 370.00 // 100 + 200 + 50 + 20
	if result.TotalMonthlyCost != expectedTotalMonthly {
		t.Errorf("Expected total monthly cost to be %f, got %f", expectedTotalMonthly, result.TotalMonthlyCost)
	}

	expectedSavings := 170.00 // 100 + 50 + 20
	if result.TotalSavingsPotential != expectedSavings {
		t.Errorf("Expected total savings potential to be %f, got %f", expectedSavings, result.TotalSavingsPotential)
	}

	expectedWastePct := (expectedSavings / expectedTotalMonthly) * 100.0
	if result.WastePercent != expectedWastePct {
		t.Errorf("Expected waste percent to be %f, got %f", expectedWastePct, result.WastePercent)
	}
}
