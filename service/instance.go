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
	raw      *ec2.Instance
	ID       string
	IP       string
	State    string
	AZ       string
	Launched string
	Type     string
	Tags     map[string]string
}

// NewInstance returns a new Instance from an AWS EC2 Instance
func NewInstance(i *ec2.Instance) *Instance {
	instance := &Instance{
		raw:   i,
		ID:    *i.InstanceId,
		IP:    *i.PrivateIpAddress,
		State: *i.State.Name,
		AZ:    *i.Placement.AvailabilityZone,
		Type:  *i.InstanceType,
	}

	instance.extractTags()

	return instance
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
	cmd := exec.Command("ssh", i.IP)
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
