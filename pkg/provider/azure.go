package provider

import (
	"fmt"
)

// AzureProvider implements the Provider interface for Microsoft Azure
type AzureProvider struct {
	SubscriptionID string
}

// NewAzureProvider instantiates the Azure provider
func NewAzureProvider(subID string) *AzureProvider {
	return &AzureProvider{SubscriptionID: subID}
}

// GetName returns the provider type
func (az *AzureProvider) GetName() CloudType {
	return Azure
}

// GetCosts queries the Azure Consumption / Cost Management APIs
func (az *AzureProvider) GetCosts(days int) ([]CostItem, error) {
	// Plan for Azure Cost Integration:
	// cred, err := azidentity.NewDefaultAzureCredential(nil)
	// client, err := armconsumption.NewUsageDetailsClient(az.SubscriptionID, cred, nil)
	
	return nil, fmt.Errorf("Azure Cost Management SDK is currently unlinked. Please configure credentials and enable armconsumption module")
}

// GetResources queries Azure Resource Graph to locate VM, disk, and DB waste
func (az *AzureProvider) GetResources() ([]Resource, error) {
	// Plan for Azure Resource Graph Scan:
	// client, err := armresourcegraph.NewClient(cred, nil)
	// query := "Resources | where type =~ 'microsoft.compute/virtualmachines' or type =~ 'microsoft.compute/disks'"
	// results, err := client.Resources(ctx, query, ...)
	
	return nil, fmt.Errorf("Azure Resource Graph SDK is currently unlinked. Please configure credentials and enable armresourcegraph module")
}

// Ensure AzureProvider implements Provider
var _ Provider = (*AzureProvider)(nil)
