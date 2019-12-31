package service

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/rivo/tview"
)

// Instance holds an ec2 instance information
type Instance struct {
	ID       string
	IP       string
	State    string
	AZ       string
	Launched string
	Type     string
}

// Service holds internal state
type Service struct {
	svc       *ec2.EC2
	app       *tview.Application
	table     *tview.Table
	instances map[string]Instance
}

// NewService returns a new service instance
func NewService() *Service {
	app := tview.NewApplication()

	table := tview.NewTable().
		SetFixed(1, 0).
		SetSelectable(true, false)
	table.SetBorderPadding(0, 0, 1, 1)

	app.SetRoot(table, true)

	s := Service{
		app:       app,
		table:     table,
		instances: make(map[string]Instance),
	}

	return &s
}

// Run starts the application
func (s *Service) Run() {
	s.table.SetSelectedFunc(s.handleSelected)
	s.ec2svc()
	s.fetchInstances()
	s.updateTable()
	s.app.Run()
}

func (s *Service) handleSelected(row int, col int) {
	cell := s.table.GetCell(row, col)
	ref := cell.GetReference()
	instance, ok := s.instances[ref.(string)]

	if ok {
		s.app.Suspend(func() {
			s.sshInstance(instance.IP)
		})
	}
}

func (s *Service) sshInstance(ip string) {
	cmd := exec.Command("ssh", ip)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf("Command failed: %s", err)
		time.Sleep(time.Second * 3)
	}
}

func (s *Service) ec2svc() {
	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			// &credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{
				Profile: "default",
			},
		},
	)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            aws.Config{Credentials: creds},
		Profile:           "default",
	}))

	// s.svc = ec2.New(session.Must(session.NewSession(conf)))
	s.svc = ec2.New(sess)
}

func (s *Service) fetchInstances() {
	res, err := s.svc.DescribeInstances(&ec2.DescribeInstancesInput{})
	if err != nil {
		log.Fatalln(err.Error())
	}

	for _, reservation := range res.Reservations {
		for _, instance := range reservation.Instances {
			i := Instance{
				ID:    *instance.InstanceId,
				IP:    *instance.PrivateIpAddress,
				State: *instance.State.Name,
				AZ:    *instance.Placement.AvailabilityZone,
				Type:  *instance.InstanceType,
			}

			s.instances[i.ID] = i
		}
	}
}

func (s *Service) updateTable() {
	headers := []string{"ID", "IP", "State", "AZ", "Type"}
	row := 0

	for _, instance := range s.instances {
		values := []string{
			instance.ID,
			instance.IP,
			instance.State,
			instance.AZ,
			instance.Type,
		}

		for col, val := range values {
			if row == 0 {
				for c, h := range headers {
					head := tview.NewTableCell(h).
						SetSelectable(false)
					s.table.SetCell(0, c, head)
				}

				row++
			}

			cell := tview.NewTableCell(val).
				SetSelectable(true).
				SetReference(instance.ID)
			s.table.SetCell(row, col, cell)
		}

		row++
	}
}
