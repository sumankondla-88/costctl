package provider

import (
	"fmt"
)

// AWSProvider implements the Provider interface for Amazon Web Services
type AWSProvider struct {
	Region string
}

// NewAWSProvider instantiates the AWS provider
func NewAWSProvider(region string) *AWSProvider {
	if region == "" {
		region = "us-east-1"
	}
	return &AWSProvider{Region: region}
}

// GetName returns the provider type
func (a *AWSProvider) GetName() CloudType {
	return AWS
}

// GetCosts queries the AWS Cost Explorer API
func (a *AWSProvider) GetCosts(days int) ([]CostItem, error) {
	// Plan for AWS Cost Explorer Integration:
	// cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(a.Region))
	// client := costexplorer.NewFromConfig(cfg)
	// input := &costexplorer.GetCostAndUsageInput{ ... }
	// res, err := client.GetCostAndUsage(context.TODO(), input)
	
	// Check if we are running in a skeleton fallback mode
	return nil, fmt.Errorf("AWS Cost Explorer SDK is currently unlinked. Please install AWS SDK packages and connect active credentials")
}

// GetResources queries the AWS EC2/RDS APIs to check for waste
func (a *AWSProvider) GetResources() ([]Resource, error) {
	// Plan for AWS Resource Scan:
	// ec2Client := ec2.NewFromConfig(cfg)
	// rdsClient := rds.NewFromConfig(cfg)
	
	// 1. VM/EC2 Check:
	//    instances, err := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	//    Query CloudWatch for CPUUtilization metrics for running instances.
	
	// 2. Volume/EBS Check:
	//    volumes, err := ec2Client.DescribeVolumes(context.TODO(), &ec2.DescribeVolumesInput{})
	//    Flag volumes with State == "available" (which means detached).
	
	// 3. PublicIP/EIP Check:
	//    ips, err := ec2Client.DescribeAddresses(context.TODO(), &ec2.DescribeAddressesInput{})
	//    Flag EIPs with AssociationId == nil.
	
	// 4. NAT Gateway Check:
	//    natGates, err := ec2Client.DescribeNatGateways(context.TODO(), &ec2.DescribeNatGatewaysInput{})
	
	return nil, fmt.Errorf("AWS Resource Manager SDK is currently unlinked. Please install AWS SDK packages and connect active credentials")
}

// Ensure AWSProvider implements Provider
var _ Provider = (*AWSProvider)(nil)
