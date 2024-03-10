package service

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/yogin/go-ec2/internal/config"
	"github.com/yogin/go-ec2/internal/providers"
)

type Slider interface {
	Get(nextSlide func()) (title string, content tview.Primitive)
}

type Slide struct {
	service  *Service
	profile  *config.Profile
	provider providers.Provider
	view     *tview.Table
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
	s.view = table

	return s
}

func (s *Slide) handleSelectedRow(row int, col int) {
}

func (s *Slide) update() {
	if s.provider == nil {
		// s.view.SetText(fmt.Sprintf("Provider '%s' not found in profile '%s'", s.profile.Provider, s.profile.ID))
		return
	}

	if err := s.provider.LoadInstances(); err != nil {
		// s.view.SetText(fmt.Sprintf("Error fetching instances in profile '%s': %s", s.profile.ID, err))
		return
	}

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
				s.view.SetCell(0, c, tag)
			}

			for c, h := range s.provider.Headers() {
				head := tview.NewTableCell(h).
					SetSelectable(false).
					SetAttributes(tcell.AttrBold).
					SetBackgroundColor(tcell.ColorDimGrey.TrueColor())
				s.view.SetCell(0, c+tagsCount, head)
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
			s.view.SetCell(row, col, cell)
		}

		row++
	}

	// s.view.SetText(fmt.Sprintf("Welcome to GoSH: %s (%d instances)", s.profile.ID, s.provider.InstancesCount()))
}

func (s *Slide) Get(nextSlide func()) (title string, content tview.Primitive) {
	s.update()
	return s.profile.ID, s.view
}
