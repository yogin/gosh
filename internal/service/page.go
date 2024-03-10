package service

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/yogin/go-ec2/internal/config"
	"github.com/yogin/go-ec2/internal/providers"
)

type Slider interface {
	Get(nextSlide func()) (title string, content tview.Primitive)
}

type Slide struct {
	app      *tview.Application
	profile  *config.Profile
	provider providers.Provider
	view     *tview.TextView
}

func NewSlide(app *tview.Application, profile *config.Profile) *Slide {
	s := &Slide{
		app:     app,
		profile: profile,
	}

	if p := providers.NewProvider(profile.Provider, profile); p != nil {
		s.provider = p
	}

	view := tview.NewTextView()
	s.view = view

	return s
}

func (s *Slide) update() {
	if s.provider == nil {
		s.view.SetText(fmt.Sprintf("Provider '%s' not found in profile '%s'", s.profile.Provider, s.profile.ID))
		return
	}

	s.view.SetText(fmt.Sprintf("Welcome to GoSH: %s", s.profile.ID))
}

func (s *Slide) Get(nextSlide func()) (title string, content tview.Primitive) {
	s.update()
	return s.profile.ID, s.view
}
