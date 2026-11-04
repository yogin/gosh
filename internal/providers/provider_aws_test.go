package providers

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestAWSInstanceSorterDeployerFirst(t *testing.T) {
	instances := AWSInstanceSorter{
		{ID: "other", Tags: map[string]string{"name": "Alpha"}},
		{ID: "deployer-staging", Tags: map[string]string{"name": "deployer", "env": "staging"}},
		{ID: "deployer-prod-b", Tags: map[string]string{"name": "deployer", "env": "production"}},
		{ID: "unnamed"},
		{ID: "deployer-prod-a", Tags: map[string]string{"name": "deployer", "env": "production"}},
		{ID: "similar", Tags: map[string]string{"name": "deployer-backup"}},
	}
	want := []string{"deployer-prod-a", "deployer-prod-b", "deployer-staging", "unnamed", "other", "similar"}
	sort.Sort(instances)
	var got []string
	for _, instance := range instances {
		got = append(got, instance.ID)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("sorted IDs = %v, want %v", got, want)
	}
	for i := range instances {
		for j := range instances {
			if instances.Less(i, j) != (i < j) {
				t.Errorf("Less(%d, %d) = %v, want %v", i, j, instances.Less(i, j), i < j)
			}
		}
	}
}

func TestAWSProviderLoadInstancesExcludesDeploy(t *testing.T) {
	svc := ec2.New(session.Must(session.NewSession(&aws.Config{Region: aws.String("us-west-1")})))
	names := []string{"Deploy", "deployer", "Deploy-backup", "deploy", ""}
	svc.Handlers.Clear()
	svc.Handlers.Send.PushBack(func(r *request.Request) {
		var instances []*ec2.Instance
		for _, name := range names {
			instances = append(instances, &ec2.Instance{
				InstanceId:   aws.String("id-" + name),
				State:        &ec2.InstanceState{Name: aws.String("stopped")},
				Placement:    &ec2.Placement{AvailabilityZone: aws.String("us-west-1a")},
				InstanceType: aws.String("t3.micro"),
				ImageId:      aws.String("ami-test"),
				LaunchTime:   aws.Time(time.Unix(0, 0)),
				Tags:         []*ec2.Tag{{Key: aws.String("Name"), Value: aws.String(name)}},
			})
		}
		*r.Data.(*ec2.DescribeInstancesOutput) = ec2.DescribeInstancesOutput{
			Reservations: []*ec2.Reservation{{Instances: instances}},
		}
	})
	p := &AWSProvider{svc: svc, instances: map[string]*Instance{"id-Deploy": {ID: "id-Deploy"}}}
	if err := p.LoadInstances(); err != nil {
		t.Fatal(err)
	}
	if p.GetInstanceByID("id-Deploy") != nil {
		t.Error("retired Deploy instance remains available")
	}
	if got := p.InstancesCount(); got != len(names)-1 {
		t.Errorf("instance count = %d, want %d", got, len(names)-1)
	}
	var got []string
	for _, instance := range p.GetInstances() {
		got = append(got, instance.Tags["name"])
	}
	want := []string{"deployer", "", "Deploy-backup", "deploy"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("visible names = %v, want %v", got, want)
	}
	names = []string{"Deploy"}
	if err := p.LoadInstances(); err != nil {
		t.Fatal(err)
	}
	if p.InstancesCount() != 0 || len(p.GetInstances()) != 0 {
		t.Error("refresh with only Deploy should leave an empty list")
	}
}
