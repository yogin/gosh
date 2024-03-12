package service

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/yogin/gosh/internal/config"
)

var (
	service *Service
)

type Service struct {
	config *config.Config
	app    *tview.Application
	status *Status
	devlog *DevLog
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

	// devlog must be started before any other component so it can receive log messages
	s.devlog = NewDevLog(s)

	if path := s.config.ConfigPath(); path != "" {
		s.Log("app", "Configuration loaded from %s", s.config.ConfigPath())
	}

	// status must be started before any other component so it can receive status updates (but after devlog)
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

	main := tview.NewFlex()
	main.SetDirection(tview.FlexColumn)
	main.AddItem(pages, 0, 1, true)                         // slides
	main.AddItem(s.devlog.Get(), 0, s.devlog.Size(), false) // page menu selector

	layout := tview.NewFlex()
	layout.SetDirection(tview.FlexRow)
	layout.AddItem(main, 0, 1, true)            // slides
	layout.AddItem(menu, 1, 1, false)           // page menu selector
	layout.AddItem(s.status.Get(), 1, 1, false) // input and status (time local/utc) line

	// global input capture, widgets can have their own input capture
	s.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		s.Log("app", "Key pressed Name=%s, Key=%d, Rune=%d", event.Name(), event.Key(), event.Rune())

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
				if err := s.config.Save(); err != nil {
					s.SetStatusText("app", "Error saving configuration: %s", err)
				} else {
					s.SetStatusText("app", "Configuration saved to %s", s.config.ConfigPath())
				}
				return nil

			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				slideID, _ := strconv.Atoi(string(event.Rune()))
				slideName := strconv.Itoa(slideID - 1)
				if pages.HasPage(slideName) {
					s.Log("app", "Page found: %s (%d)", slideName, slideID)
					pages.SwitchToPage(slideName)
					menu.Highlight(slideName)
					menu.ScrollToHighlight()
				} else {
					s.Log("app", "Page not found: %s (%d)", slideName, slideID)
				}

			case '~':
				main.ResizeItem(s.devlog.Get(), 0, s.devlog.Toggle())
				return nil
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

func (s Service) SetStatusText(prefix string, format string, a ...interface{}) {
	if s.status == nil {
		return
	}

	s.status.SetStatusText(format, a...)
	s.Log(prefix, format, a...)
}

func (s Service) Log(prefix string, format string, a ...interface{}) {
	if s.devlog == nil {
		return
	}

	l := fmt.Sprintf(format, a...)
	ts := time.Now().Format("15:04:05") // "2006-01-02 15:04:05"

	if len(prefix) > 0 {
		prefix = fmt.Sprintf("[yellow]%s[white] ", prefix)
	}

	s.devlog.Write(fmt.Sprintf("[gray][%s][white] %s%s\n", ts, prefix, l))
}
