package server

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ConversationRender struct {
	Fullscreen *fyne.Container
	Messages   *fyne.Container
	Header     *fyne.Container
}

func (cr *ConversationRender) UpdateHeader(c *Conversation, headerName string) error {

	if cr.Header == nil {
		return nil
	}

	header, err := CreateHeader(c, headerName)
	if err != nil {
		return nil
	}

	cr.Header.Objects = []fyne.CanvasObject{
		header,
	}

	cr.Header.Refresh()
	return nil
}

func CreateHeader(c *Conversation, headerName string) (*fyne.Container, error) {
	var header *fyne.Container
	switch headerName {
	case "pending":
		header = container.NewHBox(
			widget.NewLabel(fmt.Sprintf("Remote Address: PENDING")),
			widget.NewLabel(fmt.Sprintf("Conversation ID: %s...%s", c.ConversationID[:5], c.ConversationID[len(c.ConversationID)-5:])),
			widget.NewLabel(fmt.Sprintf("Token: %s...%s", c.SelfToken[:5], c.SelfToken[len(c.SelfToken)-5:])),
			widget.NewLabel(fmt.Sprintf("Remote Token: PENDING")),
			widget.NewLabel(fmt.Sprintf("Key Type: %s", c.KeyType)),
		)

	case "active":
		header = container.NewHBox(
			widget.NewLabel(fmt.Sprintf("Remote Address: %s...%s", c.RemoteAddress[:5], c.RemoteAddress[len(c.RemoteAddress)-15:])),
			widget.NewLabel(fmt.Sprintf("Conversation ID: %s...%s", c.ConversationID[:5], c.ConversationID[len(c.ConversationID)-5:])),
			widget.NewLabel(fmt.Sprintf("Token: %s...%s", c.SelfToken[:5], c.SelfToken[len(c.SelfToken)-5:])),
			widget.NewLabel(fmt.Sprintf("Remote Token: %s...%s", c.RemoteToken[:5], c.RemoteToken[len(c.RemoteToken)-5:])),
			widget.NewLabel(fmt.Sprintf("Key Type: %s", c.KeyType)),
		)

	case "ended":
		header = container.NewHBox(
			widget.NewLabel(fmt.Sprintf("Remote Address: ENDED")),
			widget.NewLabel(fmt.Sprintf("Conversation ID: %s...%s", c.ConversationID[:5], c.ConversationID[len(c.ConversationID)-5:])),
			widget.NewLabel(fmt.Sprintf("Token: ENDED")),
			widget.NewLabel(fmt.Sprintf("Remote Token: ENDED")),
			widget.NewLabel(fmt.Sprintf("Key Type: %s", c.KeyType)),
		)

	default:
		return nil, fmt.Errorf("Bad Header Requested")
	}

	return container.NewCenter(header), nil
}

func (cr *ConversationRender) AddMessage(msg Message) error {
	if cr.Messages == nil {
		return nil
	}

	t := time.Unix(int64(msg.Time), 0)
	cr.Messages.Add(newMessage(msg.Text, msg.Self, msg.FinalMessage, t))
	cr.Messages.Refresh()
	return nil
}

func (cr *ConversationRender) EndConversation() {
	cr.Fullscreen.Objects = []fyne.CanvasObject{
		container.NewVBox(),
	}
	cr.Fullscreen.Refresh()
}

func (c *Conversation) RenderMessages(w fyne.Window, rh *RouteHandler) {

	mContainer := container.NewVBox()
	for _, m := range c.Messages {
		mContainer.Add(newMessage(m.Text, m.Self, false, time.Now()))
	}

	msg := widget.NewEntry()

	send := widget.NewButtonWithIcon("",
		theme.MailSendIcon(), func() {

			if msg.Text == "" {
				dialog.ShowError(fmt.Errorf("Empty Message"), w)
				return
			}

			pmsg, err := c.PackMessage(msg.Text, false)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			err = rh.SendMessage(pmsg, c.RemoteAddress, MessagePath)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			message := Message{
				Text:         msg.Text,
				Time:         int(time.Now().Unix()),
				Self:         true,
				FinalMessage: false,
			}

			err = c.Render.AddMessage(message)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			msg.SetText("")
		},
	)

	cancel := widget.NewButtonWithIcon("",
		theme.CancelIcon(), func() {
			if !c.Established || c.Ended {
				dialog.ShowError(fmt.Errorf("Conversation is not active"), w)
				return
			}

			if msg.Text == "" {
				msg.Text = "Ended Conversation"
			}

			pmsg, err := c.PackMessage(msg.Text, true)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			err = rh.SendMessage(pmsg, c.RemoteAddress, MessagePath)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			message := Message{
				Text:         msg.Text,
				Time:         int(time.Now().Unix()),
				Self:         true,
				FinalMessage: true,
			}

			err = c.Render.AddMessage(message)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			c.Ended = true
			c.Render.UpdateHeader(c, "ended")
			msg.SetText("")
		},
	)

	destroy := widget.NewButtonWithIcon("",
		theme.DeleteIcon(), func() {
			if c.Established && !c.Ended {
				if msg.Text == "" {
					msg.Text = "Ended Conversation"
				}

				pmsg, err := c.PackMessage(msg.Text, true)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}

				rh.SendMessage(pmsg, c.RemoteAddress, MessagePath)
				c.Ended = true
			}

			message := Message{
				Text:         msg.Text,
				Time:         int(time.Now().Unix()),
				Self:         true,
				FinalMessage: true,
			}

			err := c.Render.AddMessage(message)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			c.Ended = true
			c.Render.EndConversation()
			rh.DeleteConversation(c.ConversationID)
		},
	)

	controlBar := container.NewBorder(nil, nil, container.NewHBox(destroy, cancel), send, msg)
	messages := container.NewVScroll(mContainer)

	var header *fyne.Container
	if c.Established {
		if c.Ended {
			header, _ = CreateHeader(c, "ended")
		} else {
			header, _ = CreateHeader(c, "active")
		}
	} else {
		header, _ = CreateHeader(c, "pending")
	}

	fullScreen := container.NewBorder(header, controlBar, nil, nil, messages)
	c.Render.Header = header
	c.Render.Messages = mContainer
	c.Render.Fullscreen = fullScreen

	return
}
