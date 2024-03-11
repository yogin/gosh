package providers

import "github.com/yogin/go-ec2/internal/config"

type ProviderType string

const (
	ProviderTypeAWS ProviderType = "aws"
)

type Provider interface {
	Type() ProviderType                 // Type returns the provider type
	Headers() []string                  // Headers returns the table headers
	LoadInstances() error               // LoadInstances queries the provider for all instances
	InstancesCount() int                // InstancesCount returns the number of instances
	GetTags() []string                  // GetTags returns the list of tags across all instances
	GetInstances() []*Instance          // GetInstances returns the list of instances
	GetInstanceByID(string) *Instance   // GetInstanceByID returns an instance by ID
	GetInstanceIPByID(id string) string // GetInstanceIPByID returns an instance IP by ID (public or private)
}

func NewProvider(provider string, profile *config.Profile) Provider {
	switch provider {
	case string(ProviderTypeAWS):
		return NewAWSProvider(profile)
	default:
		return nil
	}
}
