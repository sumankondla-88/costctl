package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"costctl/pkg/provider"
)

// getProviders returns active cloud providers based on configurations and credentials
func getProviders() []provider.Provider {
	if DemoMode {
		return []provider.Provider{
			provider.NewMockProvider(provider.AWS),
			provider.NewMockProvider(provider.Azure),
			provider.NewMockProvider(provider.GCP),
		}
	}

	var activeProviders []provider.Provider
	for _, cloud := range Cfg.Clouds {
		switch strings.ToLower(cloud) {
		case "aws":
			if hasAWSCredentials() {
				activeProviders = append(activeProviders, provider.NewAWSProvider(""))
			}
		case "azure":
			if hasAzureCredentials() {
				activeProviders = append(activeProviders, provider.NewAzureProvider(""))
			}
		case "gcp":
			if hasGCPCredentials() {
				activeProviders = append(activeProviders, provider.NewGCPProvider(""))
			}
		}
	}

	return activeProviders
}

// hasAWSCredentials checks for AWS credential indicators
func hasAWSCredentials() bool {
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" || os.Getenv("AWS_PROFILE") != "" {
		return true
	}
	// Also check for ~/.aws/config or credentials
	home, err := os.UserHomeDir()
	if err == nil {
		if _, err := os.Stat(home + "/.aws/credentials"); err == nil {
			return true
		}
	}
	return false
}

// hasAzureCredentials checks for Azure credential indicators
func hasAzureCredentials() bool {
	if os.Getenv("AZURE_CLIENT_ID") != "" && os.Getenv("AZURE_TENANT_ID") != "" {
		return true
	}
	home, err := os.UserHomeDir()
	if err == nil {
		if _, err := os.Stat(home + "/.azure"); err == nil {
			return true
		}
	}
	return false
}

// hasGCPCredentials checks for GCP credential indicators
func hasGCPCredentials() bool {
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" {
		return true
	}
	home, err := os.UserHomeDir()
	if err == nil {
		if _, err := os.Stat(home + "/.config/gcloud"); err == nil {
			return true
		}
	}
	return false
}

// gatherCosts collects and normalizes costs across all active providers
func gatherCosts(days int) ([]provider.CostItem, error) {
	providersList := getProviders()
	if len(providersList) == 0 {
		return nil, fmt.Errorf("no cloud credentials detected. Please configure AWS, Azure, or GCP environment variables, or run with the '--demo' flag to use sample data (e.g. 'costctl cost list --demo')")
	}

	var allCosts []provider.CostItem
	for _, prov := range providersList {
		costs, err := prov.GetCosts(days)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to gather costs for %s: %v\n", prov.GetName(), err)
			continue
		}

		// Normalize tags
		for i := range costs {
			costs[i].NormalizedTags = Cfg.NormalizeTags(costs[i].Tags)
		}

		allCosts = append(allCosts, costs...)
	}
	return allCosts, nil
}

// gatherResources collects and normalizes resources across all active providers
func gatherResources() ([]provider.Resource, error) {
	providersList := getProviders()
	if len(providersList) == 0 {
		return nil, fmt.Errorf("no cloud credentials detected. Please configure AWS, Azure, or GCP environment variables, or run with the '--demo' flag to use sample data (e.g. 'costctl waste find --demo')")
	}

	var allResources []provider.Resource
	for _, prov := range providersList {
		resources, err := prov.GetResources()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to gather resources for %s: %v\n", prov.GetName(), err)
			continue
		}

		// Normalize tags
		for i := range resources {
			resources[i].NormalizedTags = Cfg.NormalizeTags(resources[i].Tags)
		}

		allResources = append(allResources, resources...)
	}
	return allResources, nil
}

// getOutputStream handles output redirection if `--output` is specified
func getOutputStream() (io.Writer, func(), error) {
	if OutputFile == "" {
		return os.Stdout, func() {}, nil
	}

	file, err := os.Create(OutputFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create output file: %w", err)
	}

	cleanup := func() {
		file.Close()
	}

	return file, cleanup, nil
}
