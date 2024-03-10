package providers

import (
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/yogin/go-ec2/internal/utils"
)

var AWSDefaultTags = []string{"name", "env", "environment", "stage", "role", "build", "version"}

type AWSInstance struct {
	data *ec2.Instance

	ID        string
	PrivateIP string
	PublicIP  string
	State     string
	AZ        string
	Launched  time.Time
	Type      string
	AMI       string
	Tags      map[string]string
}

func NewAWSInstance(ins *ec2.Instance) *AWSInstance {
	i := &AWSInstance{
		data:      ins,
		ID:        *ins.InstanceId,
		PrivateIP: utils.SafeIP(ins.PrivateIpAddress),
		PublicIP:  utils.SafeIP(ins.PublicIpAddress),
		State:     *ins.State.Name,
		AZ:        *ins.Placement.AvailabilityZone,
		Type:      *ins.InstanceType,
		AMI:       *ins.ImageId,
		Launched:  *ins.LaunchTime,
		Tags:      make(map[string]string),
	}

	for _, t := range i.data.Tags {
		key := strings.ToLower(*t.Key)

		for _, tag := range AWSDefaultTags {
			if key == tag {
				i.Tags[tag] = *t.Value
				continue
			}
		}
	}

	return i
}
