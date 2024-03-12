package providers

import (
	"log"
	"sort"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/yogin/gosh/internal/config"
)

var AWSDefaultTags = []string{"name", "env", "environment", "stage", "role", "build", "version"}

type AWSProvider struct {
	profile   *config.Profile
	svc       *ec2.EC2
	instances map[string]*Instance
	mutex     sync.Mutex
}

func NewAWSProvider(profile *config.Profile) *AWSProvider {
	p := &AWSProvider{
		profile:   profile,
		instances: make(map[string]*Instance),
		mutex:     sync.Mutex{},
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
	// return []string{"ID", "Name", "Type", "State", "Public IP", "Private IP"}
	return []string{"ID", "Private IP", "Public IP", "State", "AZ", "Type", "AMI", "Running"}
}

func (p *AWSProvider) LoadInstances() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	res, err := p.svc.DescribeInstances(&ec2.DescribeInstancesInput{})
	if err != nil {
		log.Fatalln(err.Error())
	}

	insts := make(map[string]*Instance)
	for _, reservation := range res.Reservations {
		for _, instance := range reservation.Instances {
			i := NewInstance(instance)
			insts[i.ID] = i
		}
	}
	p.instances = insts

	return nil
}

func (p *AWSProvider) InstancesCount() int {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return len(p.instances)
}

func (p *AWSProvider) GetTags() []string {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	t := make(map[string]struct{})

	for _, i := range p.instances {
		for tag := range i.Tags {
			if _, ok := t[tag]; !ok {
				t[tag] = struct{}{}
			}
		}
	}

	keys := []string{}
	for _, tag := range AWSDefaultTags {
		if _, ok := t[tag]; ok {
			keys = append(keys, tag)
		}
	}

	return keys
}

func (p *AWSProvider) GetInstances() []*Instance {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	insts := []*Instance{}
	for _, i := range p.instances {
		insts = append(insts, i)
	}
	sort.Sort(AWSInstanceSorter(insts))

	return insts
}

func (p *AWSProvider) GetInstanceByID(id string) *Instance {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if i, ok := p.instances[id]; ok {
		return i
	}
	return nil
}

func (p *AWSProvider) GetInstanceIPByID(id string) string {
	instance := p.GetInstanceByID(id)
	if instance == nil {
		return ""
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.profile.PreferPublicIP && len(instance.PublicIP) > 0 {
		return instance.PublicIP
	}

	return instance.PrivateIP
}
