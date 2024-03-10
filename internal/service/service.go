package service

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/yogin/go-ec2/internal/config"
)

var (
	service *Service
)

type Service struct {
	config *config.Config
	app    *tview.Application
	status *Status
}

func NewService(cfg *config.Config) *Service {
	if service != nil {
		return service
	}

	service = &Service{
		config: cfg,
	}

	return service
}

func (s *Service) Run() error {
	s.app = tview.NewApplication()

	// status must be started before any other component so it can receive status updates
	s.status = NewStatus(s)
	s.status.Start()

	slides := make([]Slider, 0, len(s.config.Profiles))
	for _, profile := range s.config.Profiles {
		slides = append(slides, NewSlide(s, profile))
	}

	pages := tview.NewPages()

	menu := tview.NewTextView()
	menu.SetDynamicColors(true)
	menu.SetRegions(true)
	menu.SetWrap(false)
	menu.SetHighlightedFunc(func(added, removed, remaining []string) {
		pages.SwitchToPage(added[0])
	})

	previousSlide := func() {
		slide, _ := strconv.Atoi(menu.GetHighlights()[0])
		slide = (slide - 1 + len(slides)) % len(slides)
		menu.Highlight(strconv.Itoa(slide))
		menu.ScrollToHighlight()
	}

	nextSlide := func() {
		slide, _ := strconv.Atoi(menu.GetHighlights()[0])
		slide = (slide + 1) % len(slides)
		menu.Highlight(strconv.Itoa(slide))
		menu.ScrollToHighlight()
	}

	for idx, slide := range slides {
		title, primitive := slide.Get(nextSlide)
		pages.AddPage(strconv.Itoa(idx), primitive, true, idx == 0)
		fmt.Fprintf(menu, `%d ["%d"][darkcyan]%s[white][""]  `, idx+1, idx, title)
	}
	menu.Highlight("0")

	layout := tview.NewFlex()
	layout.SetDirection(tview.FlexRow)
	layout.AddItem(pages, 0, 1, true)           // slides
	layout.AddItem(menu, 1, 1, false)           // page menu selector
	layout.AddItem(s.status.Get(), 1, 1, false) // input and status (time local/utc) line

	// global input capture, widgets can have their own input capture
	s.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlN, tcell.KeyTab:
			nextSlide()
			return nil

		case tcell.KeyCtrlP:
			previousSlide()
			return nil

		default:
			switch event.Rune() {
			case 'q', 'Q':
				s.app.Stop()
				return nil

			case 'w', 'W':
				// TODO write configuation file
				return nil

			case '`':
				// TODO toggle debug mode
				return nil

				// default:
				// 	s.SetStatusText("Key Pressed Name: %s, Key: %d, Rune: %d", event.Name(), event.Key(), event.Rune())
			}
		}
		return event
	})

	s.app.SetRoot(layout, true)
	s.app.EnableMouse(true)
	return s.app.Run()
}

func (s *Service) GetConfig() *config.Config {
	return s.config
}

func (s *Service) GetApp() *tview.Application {
	return s.app
}

func (s Service) SetStatusText(format string, a ...interface{}) {
	s.status.SetStatusText(format, a...)
}
