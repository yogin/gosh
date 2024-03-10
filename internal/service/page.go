package service

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/yogin/go-ec2/internal/config"
)

type Slider interface {
	Get(nextSlide func()) (title string, content tview.Primitive)
}

type Slide struct {
	app     *tview.Application
	name    string
	profile *config.Profile
}

func NewSlide(app *tview.Application, name string, profile *config.Profile) *Slide {
	return &Slide{
		app:     app,
		name:    name,
		profile: profile,
	}
}

func (s *Slide) Get(nextSlide func()) (title string, content tview.Primitive) {
	return s.name, tview.NewTextView().
		SetText(fmt.Sprintf("Welcome to GoSH: %s", s.name))
}
