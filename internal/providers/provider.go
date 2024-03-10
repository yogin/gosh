package providers

type ProviderType string

const (
	ProviderTypeAWS ProviderType = "aws"
)

type Provider interface {
	Type() ProviderType // Type returns the provider type
	Headers() []string  // Headers returns the table headers
	Instances() error   // Instances lists all instances
}
