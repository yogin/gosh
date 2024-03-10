package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/rivo/tview"
	"github.com/yogin/go-ec2/internal/config"
)

type Status struct {
	config    *config.Config
	leftView  *tview.TextView
	rightView *tview.TextView
	view      *tview.Flex
	app       *tview.Application
}

func NewStatus(app *tview.Application, cfg *config.Config) *Status {
	leftView := tview.NewTextView()
	leftView.SetWrap(false)
	leftView.SetTextAlign(tview.AlignLeft)
	leftView.SetText("Gosh, it's a status bar!")

	rightView := tview.NewTextView()
	rightView.SetWrap(false)
	rightView.SetTextAlign(tview.AlignRight)

	view := tview.NewFlex()
	view.SetDirection(tview.FlexColumn)
	view.AddItem(leftView, 0, 1, false)
	view.AddItem(rightView, 0, 1, false)

	status := &Status{
		config:    cfg,
		view:      view,
		leftView:  leftView,
		rightView: rightView,
		app:       app,
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
	s.rightView.SetText(s.renderTime())
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
