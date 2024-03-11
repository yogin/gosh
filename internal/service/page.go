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
	service       *Service
	profile       *config.Profile
	provider      providers.Provider
	table         *tview.Table
	view          *tview.Flex
	refreshTicker *time.Ticker
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

	if profile.Refresh.Enabled {
		s.toggleAutoRefresh()
	}

	return s
}

func (s *Slide) toggleAutoRefresh() {
	if s.refreshTicker != nil {
		s.service.SetStatusText(s.profile.ID, "Stopping profile auto-refresh")
		s.refreshTicker.Stop()
		s.profile.Refresh.Enabled = false
		return
	}

	if s.profile.Refresh.Interval <= 0 {
		s.service.SetStatusText(s.profile.ID, "Invalid Auto-refresh interval: %d seconds", s.profile.Refresh.Interval)
		return
	}

	s.update() // update immediately before starting the timer

	s.refreshTicker = time.NewTicker(time.Second * time.Duration(s.profile.Refresh.Interval))
	go s.startAutoRefresh()
	if s.refreshTicker != nil {
		s.profile.Refresh.Enabled = true
		s.service.SetStatusText(s.profile.ID, "Auto-refreshing every %d seconds", s.profile.Refresh.Interval)
	}
}

func (s *Slide) startAutoRefresh() {
	for range s.refreshTicker.C {
		s.service.Log(s.profile.ID, "Auto-refreshing profile")
		s.update()
	}
}

func (s *Slide) handleSelectedRow(row int, col int) {
	s.service.Log(s.profile.ID, "Selected row %d", row)

	cell := s.table.GetCell(row, col)
	ref := cell.GetReference()
	instance := s.provider.GetInstanceByID(ref.(string))
	if instance == nil {
		s.service.Log(s.profile.ID, "Instance not found for ID %s", ref)
		return
	}

	s.service.Log(s.profile.ID, "Selected instance: %+v", instance)

	ip := s.provider.GetInstanceIPByID(instance.ID)
	if ip == "" {
		s.service.Log(s.profile.ID, "IP address not found for instance %s", instance.ID)
		return
	}

	s.service.Log(s.profile.ID, "Connecting to instance %s via %s", instance.ID, ip)
	s.service.GetApp().Suspend(func() {
		cmd := exec.Command("ssh", ip)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			s.service.SetStatusText(s.profile.ID, "SSH to %s failed: %s", ip, err)
		}
	})
}

func (s *Slide) update() {
	if s.provider == nil {
		s.service.SetStatusText(s.profile.ID, "Invalid provider '%s'", s.profile.Provider)
		return
	}

	if err := s.provider.LoadInstances(); err != nil {
		s.service.SetStatusText(s.profile.ID, "Error fetching instances: %v", err)
		return
	}

	if s.provider.InstancesCount() == 0 {
		message := tview.NewTextView()
		message.SetText(fmt.Sprintf("No instances found in profile '%s'", s.profile.ID))
		s.view.Clear()
		s.view.AddItem(message, 0, 1, true)
		return
	}

	s.service.SetStatusText(s.profile.ID, "Found %d instances", s.provider.InstancesCount())

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
	if !s.profile.Refresh.Enabled {
		// update immediately if auto-refresh is disabled
		s.update()
	}

	return s.profile.ID, s.view
}
