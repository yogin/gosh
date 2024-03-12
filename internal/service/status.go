package service

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rivo/tview"
	"github.com/yogin/gosh/internal/config"
)

type Status struct {
	service   *Service
	leftView  *tview.TextView
	rightView *tview.TextView
	view      *tview.Flex

	statusMutex      *sync.Mutex
	statusClearTimer *time.Timer
}

func NewStatus(service *Service) *Status {
	leftView := tview.NewTextView()
	leftView.SetWrap(false)
	leftView.SetTextAlign(tview.AlignLeft)

	rightView := tview.NewTextView()
	rightView.SetWrap(false)
	rightView.SetTextAlign(tview.AlignRight)

	view := tview.NewFlex()
	view.SetDirection(tview.FlexColumn)
	view.AddItem(leftView, 0, 1, false)
	view.AddItem(rightView, 0, 1, false)

	status := &Status{
		service:     service,
		view:        view,
		leftView:    leftView,
		rightView:   rightView,
		statusMutex: &sync.Mutex{},
	}
	status.SetStatusText("Gosh, it's a status bar!")
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
			s.service.GetApp().QueueUpdateDraw(func() {
				s.update()
			})
		}
	}()
}

func (s *Status) SetStatusText(format string, a ...interface{}) {
	s.statusMutex.Lock()
	defer s.statusMutex.Unlock()

	// clear the timer if it's running
	if s.statusClearTimer != nil {
		s.statusClearTimer.Stop()
		s.statusClearTimer = nil
	}

	s.leftView.SetText(fmt.Sprintf(format, a...))

	// set a new timer to clear the status text
	s.statusClearTimer = time.AfterFunc(3*time.Second, func() {
		s.leftView.SetText("")
	})
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
	return s.service.GetConfig().ShowLocalTime
}

func (s *Status) showUTCTime() bool {
	return s.service.GetConfig().ShowUTCTime
}

func (s *Status) timeFormat() string {
	if len(s.service.GetConfig().TimeFormat) == 0 {
		return config.DefaultTimeFormat
	}

	return s.service.GetConfig().TimeFormat
}
