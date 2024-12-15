package server

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type messageScreen struct {
	widget.BaseWidget
	text string
	time time.Time
	self bool
	end  bool
}
type messageRender struct {
	msg       *messageScreen
	container *fyne.Container
	bg        *canvas.Rectangle
}

func newMessage(text string, self bool, end bool, time time.Time) *messageScreen {
	m := &messageScreen{text: text, self: self, end: end, time: time}
	m.ExtendBaseWidget(m)
	return m
}

func (m *messageScreen) CreateRenderer() fyne.WidgetRenderer {
	messageText := widget.NewLabel(m.text)
	messageText.Wrapping = fyne.TextWrapWord

	timeText := widget.NewLabel(m.time.Format("15:04"))
	timeText.TextStyle = fyne.TextStyle{Italic: true}

	bg := &canvas.Rectangle{FillColor: color.Transparent}

	if m.self {
		messageText.Alignment = fyne.TextAlignTrailing
		timeText.Alignment = fyne.TextAlignTrailing
	} else {
		messageText.Alignment = fyne.TextAlignLeading
		timeText.Alignment = fyne.TextAlignLeading
	}

	c := container.NewStack(
		bg,
		container.New(
			layout.NewVBoxLayout(),
			messageText,
			timeText,
		),
	)

	return &messageRender{
		msg:       m,
		container: c,
		bg:        bg,
	}
}

func (r *messageRender) MaxSize() fyne.Size {
	return fyne.NewSize(0, 0)
}

func (r *messageRender) MinSize() fyne.Size {
	return r.container.MinSize()

}

func (r *messageRender) Layout(s fyne.Size) {
	size := r.MinSize()
	size = size.Max(fyne.NewSize(s.Width-(s.Width/2.3), s.Height))
	r.bg.FillColor = r.getBackgroundColor()
	pos := fyne.NewPos(0, 0)

	if r.msg.self {
		pos = fyne.NewPos(s.Width-size.Width, 0)
	}

	r.container.Resize(size)
	r.container.Move(pos)
	r.container.Refresh()

}

func (r *messageRender) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.container}
}

func (r *messageRender) Refresh() {
	r.container.Refresh()
}

func (r *messageRender) BackgroundColor() color.Color {
	return color.Transparent
}

func (r *messageRender) Destroy() {
}

func (r *messageRender) getBackgroundColor() color.Color {
	if r.msg.end {
		return theme.PrimaryColorNamed(theme.ColorRed)
	}
	if r.msg.self {
		return theme.PrimaryColorNamed(theme.ColorBlue)
	}
	return theme.PrimaryColorNamed(theme.ColorGreen)
}
