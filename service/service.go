package service

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/rivo/tview"
)

const (
	// DefaultProfile is the default AWS Profile
	DefaultProfile = "default"
)

// Config holds service settings
type Config struct {
	Profile *string
	Args    []string
}

// Service holds internal state
type Service struct {
	config    *Config
	svc       *ec2.EC2
	app       *tview.Application
	table     *tview.Table
	instances map[string]*Instance
}

// NewService returns a new service instance
func NewService(conf *Config) *Service {
	app := tview.NewApplication()

	table := tview.NewTable().
		SetFixed(1, 0).
		SetSelectable(true, false)
	table.SetBorderPadding(0, 0, 1, 1)

	app.SetRoot(table, true)

	if conf == nil {
		p := DefaultProfile
		conf = &Config{
			Profile: &p,
		}
	}

	s := Service{
		config:    conf,
		app:       app,
		table:     table,
		instances: make(map[string]*Instance),
	}

	return &s
}

// Run starts the application
func (s *Service) Run() {
	s.table.SetSelectedFunc(s.handleSelected)
	s.ec2svc()
	s.fetchInstances()
	s.updateTable()

	// for _, i := range s.instances {
	// 	fmt.Printf("%+v\n", i)
	// }
	s.app.Run()
}

func (s *Service) handleSelected(row int, col int) {
	cell := s.table.GetCell(row, col)
	ref := cell.GetReference()
	instance, ok := s.instances[ref.(string)]

	if ok {
		s.app.Suspend(func() {
			instance.runSSH()
		})
	}
}

func (s *Service) ec2svc() {
	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			// &credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{
				Profile: *s.config.Profile,
			},
		},
	)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            aws.Config{Credentials: creds},
		Profile:           *s.config.Profile,
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
			i := NewInstance(instance)
			s.instances[i.ID] = i
		}
	}
}

func (s *Service) updateTable() {
	tagsNames := selectedTags(s.instances)
	tagsCount := len(tagsNames)
	headers := []string{"ID", "IP", "State", "AZ", "Type"}
	row := 0

	for _, instance := range s.instances {
		tags := instance.TagValues(tagsNames)
		vals := []string{
			instance.ID,
			instance.IP,
			instance.State,
			instance.AZ,
			instance.Type,
		}
		values := append(tags, vals...)

		// render headers
		if row == 0 {
			for c, t := range tagsNames {
				tag := tview.NewTableCell("Tag:" + t).
					SetSelectable(false)
				s.table.SetCell(0, c, tag)
			}

			for c, h := range headers {
				head := tview.NewTableCell(h).
					SetSelectable(false)
				s.table.SetCell(0, c+tagsCount, head)
			}

			row++
		}

		for col, val := range values {
			cell := tview.NewTableCell(val).
				SetSelectable(true).
				SetReference(instance.ID)
			s.table.SetCell(row, col, cell)
		}

		row++
	}
}
