package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// actionButton extends widget.Button to support hover detection for status/info descriptions.
type actionButton struct {
	widget.Button
	onHoverIn  func()
	onHoverOut func()
}

func newActionButton(label string, icon fyne.Resource, tapped func(), hoverIn func(), hoverOut func()) *actionButton {
	b := &actionButton{
		onHoverIn:  hoverIn,
		onHoverOut: hoverOut,
	}
	b.Text = label
	b.Icon = icon
	b.OnTapped = tapped
	b.ExtendBaseWidget(b)
	return b
}

func (b *actionButton) MouseIn(e *desktop.MouseEvent) {
	b.Button.MouseIn(e)
	if b.onHoverIn != nil {
		b.onHoverIn()
	}
}

func (b *actionButton) MouseMoved(e *desktop.MouseEvent) {
	b.Button.MouseMoved(e)
}

func (b *actionButton) MouseOut() {
	b.Button.MouseOut()
	if b.onHoverOut != nil {
		b.onHoverOut()
	}
}
