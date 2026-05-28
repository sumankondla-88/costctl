package provider

import (
	"fmt"
)

// GCPProvider implements the Provider interface for Google Cloud Platform
type GCPProvider struct {
	ProjectID string
}

// NewGCPProvider instantiates the GCP provider
func NewGCPProvider(projectID string) *GCPProvider {
	return &GCPProvider{ProjectID: projectID}
}

// GetName returns the provider type
func (g *GCPProvider) GetName() CloudType {
	return GCP
}

// GetCosts queries the GCP Cloud Billing APIs or BigQuery Billing exports
func (g *GCPProvider) GetCosts(days int) ([]CostItem, error) {
	// Plan for GCP Billing Integration:
	// ctx := context.Background()
	// billingService, err := cloudbilling.NewService(ctx, option.WithCredentialsFile(...))
	
	return nil, fmt.Errorf("GCP Cloud Billing API is currently unlinked. Please configure credentials and enable billing module")
}

// GetResources queries GCP Asset Inventory or Recommender API to locate waste
func (g *GCPProvider) GetResources() ([]Resource, error) {
	// Plan for GCP Asset / Recommender query:
	// recommenderService, err := recommender.NewService(ctx)
	// Query compute.Instances, compute.Disks, compute.Addresses to check status and metrics.
	
	return nil, fmt.Errorf("GCP Asset Inventory / Recommender API is currently unlinked. Please configure credentials and enable compute module")
}

// Ensure GCPProvider implements Provider
var _ Provider = (*GCPProvider)(nil)
