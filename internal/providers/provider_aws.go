package providers

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/yogin/go-ec2/internal/config"
)

type AWSProvider struct {
	profile *config.Profile
	svc     *ec2.EC2
}

func NewAWSProvider(profile *config.Profile) *AWSProvider {
	p := &AWSProvider{
		profile: profile,
	}

	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			// &credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{
				Profile: p.profile.Name,
			},
		},
	)

	conf := aws.Config{
		Credentials: creds,
	}

	if len(p.profile.Region) > 0 {
		conf.Region = aws.String(p.profile.Region)
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            conf,
		Profile:           p.profile.Name,
	}))

	p.svc = ec2.New(sess)

	return p
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
