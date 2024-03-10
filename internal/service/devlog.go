package service

import (
	"fmt"

	"github.com/rivo/tview"
)

type DevLog struct {
	service     *Service
	view        *tview.TextView
	currentSize int
}

func NewDevLog(service *Service) *DevLog {
	d := &DevLog{
		service:     service,
		currentSize: 0,
	}

	view := tview.NewTextView()
	view.SetWrap(true)
	view.SetDynamicColors(true)
	view.SetBorder(true)
	view.SetTitle(" Dev Log ")
	view.SetChangedFunc(func() {
		view.ScrollToEnd()
	})
	d.view = view

	return d
}

func (d *DevLog) Get() tview.Primitive {
	return d.view
}

func (d *DevLog) Size() int {
	return d.currentSize
}

func (d *DevLog) Toggle() int {
	if d.currentSize == 0 {
		d.currentSize = 1
	} else {
		d.currentSize = 0
	}

	return d.currentSize
}

func (d *DevLog) Write(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	d.view.Write([]byte(s))
}
