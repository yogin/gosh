package providers

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/yogin/gosh/internal/utils"
)

// type Instance interface {
// 	GetID() string
// }

type Instance struct {
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

func NewInstance(ins *ec2.Instance) *Instance {
	i := &Instance{
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

		i.Tags[key] = *t.Value
		// for _, tag := range defaultTags {
		// 	if key == tag {
		// i.Tags[tag] = *t.Value
		// 		continue
		// 	}
		// }
	}

	return i
}

func (i *Instance) GetID() string {
	return i.ID
}

// RunningDescription returns how old the instance is as a string
// eg. 1 day ago, 10 minutes ago, ...
func (i *Instance) RunningDescription() string {
	if i.State == "terminated" {
		return ""
	}

	elasped := time.Since(i.Launched)

	minutes := int(elasped.Minutes())
	hours := minutes / 60
	minutes = minutes % 60

	days := hours / 24
	hours = hours % 24

	parts := []string{}

	if days > 0 {
		unit := "day"
		if days > 1 {
			unit += "s"
		}

		parts = append(parts, fmt.Sprintf("%d %s", days, unit))
	}

	if hours > 0 {
		unit := "hour"
		if hours > 1 {
			unit += "s"
		}

		parts = append(parts, fmt.Sprintf("%d %s", hours, unit))
	}

	unit := "min"
	if minutes > 1 {
		unit += "s"
	}

	parts = append(parts, fmt.Sprintf("%d %s", minutes, unit))

	return strings.Join(parts, " ")
}

// TagValues returns the values for the requested tag names
func (i *Instance) TagValues(names []string) []string {
	values := []string{}

	for _, key := range names {
		value := ""

		if v, ok := i.Tags[key]; ok {
			value = v
		}

		values = append(values, value)
	}

	return values
}

// IsRunningLessThan indicates if the instance was started less than X minutes ago
func (i *Instance) IsRunningLessThan(mins int) bool {
	elapsed := time.Since(i.Launched)
	return int(elapsed.Minutes()) < mins
}

// IsRunningMoreThan indicates if the instance was started more than X minutes ago
func (i *Instance) IsRunningMoreThan(mins int) bool {
	elapsed := time.Since(i.Launched)
	return int(elapsed.Minutes()) > mins
}

// AWSInstanceSorter implements sort.Interface
type AWSInstanceSorter []*Instance

func (a AWSInstanceSorter) Len() int {
	return len(a)
}

func (a AWSInstanceSorter) Less(i, j int) bool {
	for _, key := range AWSDefaultTags {
		if a[i].Tags[key] < a[j].Tags[key] {
			return true
		}

		if a[i].Tags[key] > a[j].Tags[key] {
			return false
		}
	}

	return a[i].ID < a[j].ID
}

func (a AWSInstanceSorter) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
