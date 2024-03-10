package service

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/yogin/go-ec2/internal/config"
	"github.com/yogin/go-ec2/internal/providers"
)

type Slider interface {
	Get(nextSlide func()) (title string, content tview.Primitive)
}

type Slide struct {
	service      *Service
	profile      *config.Profile
	provider     providers.Provider
	table        *tview.Table
	view         *tview.Flex
	refreshTimer *time.Timer
}

func NewSlide(service *Service, profile *config.Profile) *Slide {
	s := &Slide{
		service: service,
		profile: profile,
	}

	if p := providers.NewProvider(profile.Provider, profile); p != nil {
		s.provider = p
	}

	table := tview.NewTable()
	table.SetFixed(1, 0)
	table.SetSelectable(true, false)
	table.SetBorderPadding(0, 0, 0, 0)
	table.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor) // tcell.ColorBlack.TrueColor()
	table.SetSelectedFunc(s.handleSelectedRow)                      // handles pressing ENTER key on table row
	s.table = table

	view := tview.NewFlex()
	view.SetDirection(tview.FlexRow)
	view.AddItem(table, 0, 1, true)
	s.view = view

	s.view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'r': // refresh
			s.update()
			return nil

		case 'R': // toggle auto-refresh (every minute)
			s.toggleAutoRefresh()
			return nil
		}

		return event
	})

	return s
}

func (s *Slide) toggleAutoRefresh() {
	if s.refreshTimer != nil {
		s.refreshTimer.Stop()
		s.refreshTimer = nil
	} else {
		s.update()                                            // update immediately before starting the timer
		s.refreshTimer = time.AfterFunc(time.Minute, func() { // TODO make the interval configurable
			s.service.Log("Auto-refreshing profile '%s'", s.profile.ID)
			s.update()
		})
		if s.refreshTimer != nil {
			s.service.SetStatusText("Auto-refreshing profile '%s' every minute", s.profile.ID)
		}
	}
}

func (s *Slide) handleSelectedRow(row int, col int) {
	s.service.Log("Selected row %d (profile:%s)", row, s.profile.ID)

	cell := s.table.GetCell(row, col)
	ref := cell.GetReference()
	instance := s.provider.GetInstanceByID(ref.(string))
	if instance == nil {
		s.service.Log("Instance not found for ID %s", ref)
		return
	}

	s.service.Log("Selected instance: %+v", instance)

	ip := s.provider.GetInstanceIPByID(instance.ID)
	if ip == "" {
		s.service.Log("IP address not found for instance %s", instance.ID)
		return
	}

	s.service.Log("Connecting to instance %s via %s", instance.ID, ip)
	s.service.GetApp().Suspend(func() {
		cmd := exec.Command("ssh", ip)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			s.service.SetStatusText("SSH to %s failed: %s", ip, err)
		}
	})
}

func (s *Slide) update() {
	if s.provider == nil {
		s.service.SetStatusText("Error: provider '%s' not found in profile '%s'", s.profile.Provider, s.profile.ID)
		return
	}

	if err := s.provider.LoadInstances(); err != nil {
		s.service.SetStatusText("Error fetching instances in profile '%s': %s", s.profile.ID, err)
		return
	}

	if s.provider.InstancesCount() == 0 {
		message := tview.NewTextView()
		message.SetText(fmt.Sprintf("No instances found in profile '%s'", s.profile.ID))
		s.view.Clear()
		s.view.AddItem(message, 0, 1, true)
		return
	}

	s.service.SetStatusText("Found %d instances in profile '%s'", s.provider.InstancesCount(), s.profile.ID)

	s.table.Clear()
	s.view.Clear()
	s.view.AddItem(s.table, 0, 1, true)

	tagsNames := s.provider.GetTags()
	tagsCount := len(tagsNames)
	instances := s.provider.GetInstances()

	row := 0
	for _, instance := range instances {
		// https://godoc.org/github.com/rivo/tview#hdr-Colors
		// https://pkg.go.dev/github.com/gdamore/tcell?tab=doc#Color
		// https://www.w3schools.com/colors/colors_names.asp
		color := tcell.ColorWhite.TrueColor()
		switch instance.State {
		case "terminated", "stopped":
			color = tcell.ColorGrey.TrueColor()
		case "pending", "stopping", "shutting-down":
			color = tcell.ColorCrimson.TrueColor()
		case "running":
			if instance.IsRunningLessThan(15) { // 15 minutes
				color = tcell.ColorPaleGreen.TrueColor()
			} else if instance.IsRunningMoreThan(129600) { //  129600 minutes = 90 days (1 quarter)
				color = tcell.ColorOrange.TrueColor()
			}
		}

		tags := instance.TagValues(tagsNames)
		vals := []string{
			instance.ID,
			instance.PrivateIP,
			instance.PublicIP,
			instance.State,
			instance.AZ,
			instance.Type,
			instance.AMI,
			instance.RunningDescription(),
		}
		values := append(tags, vals...)

		// headers
		if row == 0 {
			for c, t := range tagsNames {
				tag := tview.NewTableCell("Tag:" + t).
					SetSelectable(false).
					SetAttributes(tcell.AttrBold).
					SetBackgroundColor(tcell.ColorDimGrey.TrueColor())
				s.table.SetCell(0, c, tag)
			}

			for c, h := range s.provider.Headers() {
				head := tview.NewTableCell(h).
					SetSelectable(false).
					SetAttributes(tcell.AttrBold).
					SetBackgroundColor(tcell.ColorDimGrey.TrueColor())
				s.table.SetCell(0, c+tagsCount, head)
			}

			row++
		}

		// instances
		for col, val := range values {
			cell := tview.NewTableCell(val).
				SetSelectable(true).
				SetReference(instance.ID).
				SetTextColor(color).
				SetBackgroundColor(tcell.ColorBlack.TrueColor())
			s.table.SetCell(row, col, cell)
		}

		row++
	}
}

func (s *Slide) Get(nextSlide func()) (title string, content tview.Primitive) {
	s.update()
	return s.profile.ID, s.view
}
