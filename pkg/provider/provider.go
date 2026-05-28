package provider

import "time"

// CloudType represents the cloud provider type
type CloudType string

const (
	AWS   CloudType = "AWS"
	Azure CloudType = "Azure"
	GCP   CloudType = "GCP"
	Mock  CloudType = "Mock"
)

// CostItem represents cost records returned by cloud spend APIs
type CostItem struct {
	Provider       CloudType         `json:"provider"`
	Region         string            `json:"region"`
	Service        string            `json:"service"`
	Cost           float64           `json:"cost"`
	Date           time.Time         `json:"date"`
	Tags           map[string]string `json:"tags"`
	NormalizedTags map[string]string `json:"normalized_tags"`
}

// ResourceType represents the resource categories we scan for waste
type ResourceType string

const (
	VM         ResourceType = "VM"
	Volume     ResourceType = "Volume"
	PublicIP   ResourceType = "PublicIP"
	Database   ResourceType = "Database"
	Snapshot   ResourceType = "Snapshot"
	NATGateway ResourceType = "NATGateway"
)

// Resource represents a cloud asset that generates cost
type Resource struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Type           ResourceType      `json:"type"`
	Provider       CloudType         `json:"provider"`
	Region         string            `json:"region"`
	CostPerMonth   float64           `json:"cost_per_month"`
	Tags           map[string]string `json:"tags"`
	NormalizedTags map[string]string `json:"normalized_tags"`
	Status         string            `json:"status"`
	LaunchTime     time.Time         `json:"launch_time"`
	Metrics        map[string]float64 `json:"metrics"` // e.g. "CPUUtilization": 2.4, "DatabaseConnections": 0
}

// Provider defines the standard interface for retrieving cost and asset details
type Provider interface {
	GetName() CloudType
	GetCosts(days int) ([]CostItem, error)
	GetResources() ([]Resource, error)
}
