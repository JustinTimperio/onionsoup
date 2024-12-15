package menus

import (
	"crypto/rsa"
	"fmt"
	"os"
	"sort"

	"github.com/JustinTimperio/onionsoup/crypt"
	"github.com/JustinTimperio/onionsoup/server"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/labstack/echo/v4"
)

var (
	rh           *server.RouteHandler
	currentConvo string
	serverScreen *fyne.Container

	dialogPrivKey any
	dialogPubKey  any
	dialogKeyType string

	convoAlias     string
	convoPubString string
	convoToken     string

	wrp = widget.NewEntryWithData(binding.BindString(&convoPubString))
	wa  = widget.NewEntryWithData(binding.BindString(&convoAlias))
	wt  = widget.NewEntryWithData(binding.BindString(&convoToken))
)

func ServerView(w fyne.Window) fyne.CanvasObject {

	var (
		startServer fyne.Widget
		err         error
	)

	clearVars := func() {
		wrp.Text = ""
		wa.Text = ""
		wt.Text = ""

		convoToken = ""
		convoAlias = ""
		convoPubString = ""
		dialogKeyType = ""
		dialogPrivKey = nil
		dialogPubKey = nil
	}

	// Select Private Key
	privateKeyDialog := dialog.NewFileOpen(func(f fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		if f == nil {
			dialog.ShowError(fmt.Errorf("No file selected"), w)
			return
		}

		privateKeyPEM, err := os.ReadFile(f.URI().Path())
		if err != nil {
			dialog.ShowError(err, w)
			f.Close()
			return
		}

		pass := widget.NewPasswordEntry()
		pass.PlaceHolder = "Password"
		kType := widget.NewSelect([]string{"pgp", "rsa"}, func(value string) {
			dialogKeyType = value
		})

		unlock := func(b bool) {
			if !b {
				return
			}

			defer f.Close()

			switch dialogKeyType {
			case "rsa":
				dialogPrivKey, err = crypt.RSAPrivateKeyToMem(privateKeyPEM, pass.Text)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				dialogPubKey, err = crypt.RSAPublicKeyToBytes(&dialogPrivKey.(*rsa.PrivateKey).PublicKey)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}

			case "pgp":
				dialogPrivKey, err = crypt.PGPPrivateKeyToMem(privateKeyPEM, pass.Text)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}

				dialogPubKey, err = crypt.PGPPublicKeyToBytes(dialogPrivKey.(*crypto.Key))
				if err != nil {
					dialog.ShowError(err, w)
					return
				}

			default:
				dialog.ShowError(fmt.Errorf("No Key Type Selected"), w)
				return
			}
		}

		d := dialog.NewForm(
			"Unlock Private Key File",
			"Password",
			"NO Password",
			[]*widget.FormItem{
				{Text: "Password", Widget: pass},
				{Text: "Key Type", Widget: kType},
			},
			unlock,
			w,
		)
		d.Show()
	}, w)

	wpk := widget.NewButton("Select Private Key", privateKeyDialog.Show)
	wrp.MultiLine = true

	startConversation := func(b bool) {
		if !b {
			return
		}

		var rpk any
		var err error
		switch dialogKeyType {
		case "rsa":
			rpk, err = crypt.RSAPublicKeyToMem([]byte(convoPubString))
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
		case "pgp":
			rpk, err = crypt.PGPPublicKeyToMem([]byte(convoPubString))
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
		default:
			dialog.ShowError(fmt.Errorf("invalid key type"), w)
			return
		}

		err = rh.BootstrapConversation(convoToken, dialogPrivKey, dialogPubKey, rpk, convoAlias, w)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		clearVars()
	}

	generateConversation := func(b bool) {
		if !b {
			return
		}

		var rpk any
		var err error
		switch dialogKeyType {
		case "rsa":
			rpk, err = crypt.RSAPublicKeyToMem([]byte(convoPubString))
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

		case "pgp":
			rpk, err = crypt.PGPPublicKeyToMem([]byte(convoPubString))
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

		default:
			dialog.ShowError(fmt.Errorf("invalid key type"), w)
			return
		}

		t, err := rh.GenerateConversation(dialogPrivKey, dialogPubKey, rpk, dialogKeyType, convoAlias, w)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		clearVars()

		w.Clipboard().SetContent(t)
		dialog.NewInformation("Token Copied to Clipboard", "Token copied to clipboard.", w).Show()
		serverScreen.Refresh()
	}

	startConversationDialog := dialog.NewForm(
		"Load Conversation Token",
		"Start",
		"Cancel",
		[]*widget.FormItem{
			{Text: "Select Private Key File", Widget: wpk},
			{Text: "Enter Public Key", Widget: wrp},
			{Text: "Conversation Token", Widget: wt},
			{Text: "Conversation Alias", Widget: wa},
		},
		startConversation,
		w,
	)

	generateConversationDialog := dialog.NewForm(
		"Generate Conversation Token",
		"Generate",
		"Cancel",
		[]*widget.FormItem{
			{Text: "Select Private Key File", Widget: wpk},
			{Text: "Enter Public Key", Widget: wrp},
			{Text: "Conversation Alias", Widget: wa},
		},
		generateConversation,
		w,
	)

	startServer = widget.NewButton("Start Server", func() {
		convos := container.NewStack()

		setConvos := func(c *server.Conversation) {
			c.RenderMessages(w, rh)
			convos.Objects = []fyne.CanvasObject{
				c.Render.Fullscreen,
			}
			convos.Refresh()
		}

		rh, err = server.NewRouteHandler()
		if err != nil {
			dialog.ShowError(fmt.Errorf("Error starting route handler: %s", err), w)
			return
		}
		rh.ConversationScreen = convos

		e := echo.New()
		e.Listener = rh.Receiver
		e.POST("/message", rh.Message)
		e.POST("/bootstrap", rh.Bootstrap)
		go e.Start("")

		split := container.NewHSplit(makeConvos(setConvos), convos)
		split.Offset = 0.10

		serverScreen.Objects = []fyne.CanvasObject{container.NewBorder(container.NewCenter(
			container.NewHBox(
				widget.NewLabel(fmt.Sprintf("Running on %s...%s", rh.URL[:5], rh.URL[len(rh.URL)-15:])),
				widget.NewButton("Generate Conversation Token", generateConversationDialog.Show),
				widget.NewButton("Import Conversation Token", startConversationDialog.Show),
				widget.NewButton("Stop Server", func() {
					if rh != nil {
						e.Close()
						for _, convo := range rh.Conversations {
							if !convo.Ended && convo.Established {
								pmsg, err := convo.PackMessage("Ended Conversation", true)
								if err != nil {
									dialog.ShowError(err, w)
									return
								}

								rh.SendMessage(pmsg, convo.RemoteAddress, server.MessagePath)
								convo.Ended = true
								convo.Render.EndConversation()
								rh.DeleteConversation(convo.ConversationID)
							}
						}

						rh.Close()
						rh = nil
					}

					serverScreen.Objects = []fyne.CanvasObject{container.NewBorder(
						container.NewVBox(startServer),
						nil,
						nil,
						nil,
						nil,
					)}
					serverScreen.Refresh()
				}),
			)), nil, nil, nil, split)}

		serverScreen.Refresh()
	})

	if serverScreen == nil {
		serverScreen = container.NewBorder(
			container.NewVBox(startServer),
			nil,
			nil,
			nil,
			nil,
		)
	}

	return serverScreen
}

func makeConvos(setWindow func(c *server.Conversation)) fyne.CanvasObject {
	a := fyne.CurrentApp()

	tree := &widget.Tree{
		ChildUIDs: func(uid string) []string {
			var order []string
			if rh == nil {
				return order
			}
			for k := range rh.Conversations {
				order = append(order, k)
			}
			sort.Slice(order, func(i, j int) bool {
				return order[i] < order[j]
			})

			return order
		},
		IsBranch: func(uid string) bool {
			if uid == "" {
				return true
			} else {
				return false
			}
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Conversation")
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			t, ok := rh.Conversations[uid]
			if !ok {
				return
			}
			obj.(*widget.Label).SetText(t.ConversationAlias)
			obj.(*widget.Label).TextStyle = fyne.TextStyle{}
		},
		OnSelected: func(uid string) {
			if t, ok := rh.Conversations[uid]; ok {
				a.Preferences().SetString(currentConvo, uid)
				setWindow(t)
			}
		},
	}

	return container.NewBorder(nil, nil, nil, nil, tree)
}
