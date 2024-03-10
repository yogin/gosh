package providers

import "github.com/yogin/go-ec2/internal/config"

type ProviderType string

const (
	ProviderTypeAWS ProviderType = "aws"
)

type Provider interface {
	Type() ProviderType  // Type returns the provider type
	Headers() []string   // Headers returns the table headers
	Instances() error    // Instances lists all instances
	InstancesCount() int // InstancesCount returns the number of instances
}

func NewProvider(provider string, profile *config.Profile) Provider {
	switch provider {
	case string(ProviderTypeAWS):
		return NewAWSProvider(profile)
	default:
		return nil
	}
}
