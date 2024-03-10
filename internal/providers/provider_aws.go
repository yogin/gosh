package providers

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/yogin/go-ec2/internal/config"
)

type AWSProvider struct {
	profile   *config.Profile
	svc       *ec2.EC2
	instances map[string]*AWSInstance
}

func NewAWSProvider(profile *config.Profile) *AWSProvider {
	p := &AWSProvider{
		profile:   profile,
		instances: make(map[string]*AWSInstance),
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
		// CredentialsChainVerboseErrors: aws.Bool(true),
		// HTTPClient:                    &http.Client{Timeout: 10 * time.Second},
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
	res, err := p.svc.DescribeInstances(&ec2.DescribeInstancesInput{})
	if err != nil {
		log.Fatalln(err.Error())
	}

	insts := make(map[string]*AWSInstance)
	for _, reservation := range res.Reservations {
		for _, instance := range reservation.Instances {
			i := NewAWSInstance(instance)
			insts[i.ID] = i
		}
	}
	p.instances = insts

	return nil
}

func (p *AWSProvider) InstancesCount() int {
	return len(p.instances)
}
