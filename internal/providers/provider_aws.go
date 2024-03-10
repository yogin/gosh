package providers

type AWSProvider struct {
}

func NewAWSProvider() *AWSProvider {
	return &AWSProvider{}
}

func (p *AWSProvider) Type() ProviderType {
	return ProviderTypeAWS
}

func (p *AWSProvider) Headers() []string {
	return []string{"ID", "Name", "Type", "State", "Public IP", "Private IP"}
}

func (p *AWSProvider) Instances() error {
	return nil
}
