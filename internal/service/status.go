package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/rivo/tview"
	"github.com/yogin/go-ec2/internal/config"
)

type Status struct {
	config *config.Config
	view   *tview.TextView
	app    *tview.Application
}

func NewStatus(app *tview.Application, cfg *config.Config) *Status {
	view := tview.NewTextView()
	view.SetWrap(false)
	view.SetTextAlign(tview.AlignRight)

	status := &Status{
		config: cfg,
		view:   view,
		app:    app,
	}
	status.update()

	return status
}

func (s *Status) Get() tview.Primitive {
	return s.view
}

func (s *Status) Start() {
	// refresh every second
	go func() {
		for range time.Tick(time.Second) {
			s.app.QueueUpdateDraw(func() {
				s.update()
			})
		}
	}()
}

func (s *Status) update() {
	s.view.SetText(s.renderTime())
}

func (s *Status) renderTime() string {
	now := time.Now()
	times := make([]string, 0, 2)
	format := s.timeFormat()

	if s.showLocalTime() {
		times = append(times, fmt.Sprintf("Local: %s", now.Format(format)))
	}

	if s.showUTCTime() {
		times = append(times, fmt.Sprintf("UTC: %s", now.UTC().Format(format)))
	}

	return strings.Join(times, " | ")
}

func (s *Status) showLocalTime() bool {
	return s.config.ShowLocalTime
}

func (s *Status) showUTCTime() bool {
	return s.config.ShowUTCTime
}

func (s *Status) timeFormat() string {
	if len(s.config.TimeFormat) == 0 {
		return config.DefaultTimeFormat
	}

	return s.config.TimeFormat
}
