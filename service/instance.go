package service

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
)

var (
	// DefaultTags list of tags to look for
	DefaultTags = []string{"name", "env", "environment", "stage", "role"}
)

// Instance holds an ec2 instance information
type Instance struct {
	raw       *ec2.Instance
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

// NewInstance returns a new Instance from an AWS EC2 Instance
func NewInstance(i *ec2.Instance) *Instance {

	instance := &Instance{
		raw:       i,
		ID:        *i.InstanceId,
		PrivateIP: safeIP(i.PrivateIpAddress),
		PublicIP:  safeIP(i.PublicIpAddress),
		State:     *i.State.Name,
		AZ:        *i.Placement.AvailabilityZone,
		Type:      *i.InstanceType,
		AMI:       *i.ImageId,
		Launched:  *i.LaunchTime,
	}

	instance.extractTags()

	return instance
}

func safeIP(ip *string) string {
	if ip == nil {
		return ""
	}

	return *ip
}

func (i *Instance) extractTags() {
	tags := make(map[string]string)

	for _, rawTag := range i.raw.Tags {
		key := strings.ToLower(*rawTag.Key)

		for _, tag := range DefaultTags {
			if key == tag {
				tags[tag] = *rawTag.Value
				continue
			}
		}
	}

	i.Tags = tags
}

func (i *Instance) runSSH() {
	cmd := exec.Command("ssh", i.PrivateIP)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf("Command failed: %s", err)
		time.Sleep(time.Second * 3)
	}
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

func selectedTags(instances map[string]*Instance) []string {
	// collect all tags from instances
	t := make(map[string]bool)
	for _, i := range instances {
		for tag := range i.Tags {
			if _, ok := t[tag]; !ok {
				t[tag] = true
			}
		}
	}

	// order collected tags
	keys := []string{}
	for _, tag := range DefaultTags {
		if _, ok := t[tag]; ok {
			keys = append(keys, tag)
		}
	}

	return keys
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
