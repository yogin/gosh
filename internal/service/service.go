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

	slides := make([]Slider, 0, len(s.config.Profiles))
	for name, profile := range s.config.Profiles {
		slides = append(slides, NewSlide(s.app, name, &profile))
	}

	pages := tview.NewPages()

	info := tview.NewTextView()
	info.SetDynamicColors(true)
	info.SetRegions(true)
	info.SetWrap(false)
	info.SetHighlightedFunc(func(added, removed, remaining []string) {
		pages.SwitchToPage(added[0])
	})

	previousSlide := func() {
		slide, _ := strconv.Atoi(info.GetHighlights()[0])
		slide = (slide - 1 + len(slides)) % len(slides)
		info.Highlight(strconv.Itoa(slide))
		info.ScrollToHighlight()
	}

	nextSlide := func() {
		slide, _ := strconv.Atoi(info.GetHighlights()[0])
		slide = (slide + 1) % len(slides)
		info.Highlight(strconv.Itoa(slide))
		info.ScrollToHighlight()
	}

	for idx, slide := range slides {
		title, primitive := slide.Get(nextSlide)
		pages.AddPage(strconv.Itoa(idx), primitive, true, idx == 0)
		fmt.Fprintf(info, `%d ["%d"][darkcyan]%s[white][""]  `, idx+1, idx, title)
	}
	info.Highlight("0")

	status := NewStatus(s.app, s.config)
	status.Start()

	layout := tview.NewFlex()
	layout.SetDirection(tview.FlexRow)
	layout.AddItem(pages, 0, 1, false)        // slides
	layout.AddItem(info, 1, 1, false)         // page selector
	layout.AddItem(status.Get(), 1, 1, false) // input and status (time local/utc) line

	s.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlN, tcell.KeyTab:
			nextSlide()
			return nil
		case tcell.KeyCtrlP:
			previousSlide()
			return nil
		}
		return event
	})

	s.app.SetRoot(layout, true)
	s.app.EnableMouse(true)
	return s.app.Run()
}
