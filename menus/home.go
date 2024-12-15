package menus

import (
	"fmt"
	"os"
	"strconv"

	"github.com/JustinTimperio/onionsoup/crypt"
	"github.com/JustinTimperio/onionsoup/data"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const Version = "1.0.0"

func HomeView(w fyne.Window) fyne.CanvasObject {
	logo := canvas.NewImageFromResource(data.Logo)
	logo.SetMinSize(fyne.NewSize(512, 512))

	var screen *fyne.Container
	var privateKeyDialog *dialog.FileDialog

	// Generate RSA Keys
	rsaGenerateSaveDialog := dialog.NewFileSave(func(f fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		if f == nil {
			dialog.ShowError(fmt.Errorf("File not selected"), w)
			return
		}

		pass1 := widget.NewPasswordEntry()
		pass2 := widget.NewPasswordEntry()
		pass1.PlaceHolder = "Password"
		pass2.PlaceHolder = "Password"
		size := widget.NewEntry()
		size.Text = "4096"

		saveKey := func() {
			defer f.Close()
			if pass1.Text != "" || pass2.Text != "" {
				if pass1.Text != pass2.Text {
					dialog.ShowError(fmt.Errorf("Passwords do not match"), w)
					return
				}
			}

			s, err := strconv.Atoi(size.Text)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			SingleMessageRSAPrivateKey, _, err = crypt.RSAGenerateKeyPair(s)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			privateKeyPEM, err := crypt.RSAPrivateKeyToBytes(SingleMessageRSAPrivateKey, pass1.Text)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			_, err = f.Write(privateKeyPEM)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			PEM, err := crypt.RSAPublicKeyToBytes(&SingleMessageRSAPrivateKey.PublicKey)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			SingleMessagePrivateKeyPath = f.URI().Path()
			SingleMessagePublicKey = string(PEM)
			SingleMessageKeyType = "rsa"

			screen.Objects = []fyne.CanvasObject{
				container.NewCenter(container.NewVBox(
					logo,
					widget.NewLabelWithStyle(fmt.Sprintf("Version: %s", Version), fyne.TextAlignCenter, fyne.TextStyle{}),
				)), container.NewVBox(
					widget.NewLabel("Loaded private key path: "+SingleMessagePrivateKeyPath),
					widget.NewButton("Reload Private Key", privateKeyDialog.Show),
					widget.NewButton("Copy Public Key", func() {
						w.Clipboard().SetContent(SingleMessagePublicKey)
					}),
				),
			}

			screen.Refresh()
			dialog.ShowInformation("Keys Generated!", "Private Key Saved", w)
		}

		d := dialog.NewForm(
			"Add Password to Private Key File",
			"Password Added",
			"NO Password Added",
			[]*widget.FormItem{
				{Text: "Password", Widget: pass1},
				{Text: "Confirm Password", Widget: pass2},
				{Text: "Size", Widget: size},
			},
			nil,
			w,
		)
		d.SetOnClosed(saveKey)
		d.Show()

	}, w)

	pgpGenerateSaveDialog := dialog.NewFileSave(func(f fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		if f == nil {
			dialog.ShowError(fmt.Errorf("File not selected"), w)
			return
		}

		pass1 := widget.NewPasswordEntry()
		pass2 := widget.NewPasswordEntry()
		pass1.PlaceHolder = "Password"
		pass2.PlaceHolder = "Password"
		accountName := widget.NewEntry()
		accountName.PlaceHolder = "Username"

		saveKey := func() {
			defer f.Close()

			if pass1.Text != "" || pass2.Text != "" {
				if pass1.Text != pass2.Text {
					dialog.ShowError(fmt.Errorf("Passwords do not match"), w)
					return
				}
			}

			SingleMessagePGPPrivateKey, err = crypt.PGPGenerateKeyPair(accountName.Text)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			privateKeyPEM, err := crypt.PGPPrivateKeyToBytes(SingleMessagePGPPrivateKey, pass1.Text)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			_, err = f.Write(privateKeyPEM)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			PEM, err := crypt.PGPPublicKeyToBytes(SingleMessagePGPPrivateKey)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			SingleMessagePrivateKeyPath = f.URI().Path()
			SingleMessagePublicKey = string(PEM)
			SingleMessageKeyType = "pgp"

			screen.Objects = []fyne.CanvasObject{
				container.NewCenter(container.NewVBox(
					logo,
					widget.NewLabelWithStyle(fmt.Sprintf("Version: %s", Version), fyne.TextAlignCenter, fyne.TextStyle{}),
				)), container.NewVBox(
					widget.NewLabel("Loaded private key path: "+SingleMessagePrivateKeyPath),
					widget.NewButton("Reload Private Key", privateKeyDialog.Show),
					widget.NewButton("Copy Public Key", func() {
						w.Clipboard().SetContent(SingleMessagePublicKey)
					}),
				),
			}

			screen.Refresh()
			dialog.ShowInformation("Keys Generated!", "Private Key Saved", w)
		}

		d := dialog.NewForm(
			"Add Password to Private Key File",
			"Password Added",
			"NO Password Added",
			[]*widget.FormItem{
				{Text: "Password", Widget: pass1},
				{Text: "Confirm Password", Widget: pass2},
				{Text: "Username", Widget: accountName},
			},
			nil,
			w,
		)
		d.SetOnClosed(saveKey)
		d.Show()

	}, w)

	// Select Private Key
	privateKeyDialog = dialog.NewFileOpen(func(f fyne.URIReadCloser, err error) {
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

		pass1 := widget.NewPasswordEntry()
		pass1.PlaceHolder = "Password"
		kType := widget.NewSelect([]string{"pgp", "rsa"}, func(value string) {
			SingleMessageKeyType = value
		})

		unlock := func() {
			defer f.Close()

			switch SingleMessageKeyType {
			case "rsa":
				SingleMessageRSAPrivateKey, err = crypt.RSAPrivateKeyToMem(privateKeyPEM, pass1.Text)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				PEM, err := crypt.RSAPublicKeyToBytes(&SingleMessageRSAPrivateKey.PublicKey)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}

				SingleMessagePublicKey = string(PEM)
				SingleMessagePrivateKeyPath = f.URI().Path()

				screen.Objects = []fyne.CanvasObject{
					container.NewCenter(container.NewVBox(
						logo,
						widget.NewLabelWithStyle(fmt.Sprintf("Version: %s", Version), fyne.TextAlignCenter, fyne.TextStyle{}),
					)), container.NewVBox(
						widget.NewLabel("Loaded private key path: "+SingleMessagePrivateKeyPath),
						widget.NewButton("Reload Private Key", privateKeyDialog.Show),
						widget.NewButton("Copy Public Key", func() {
							w.Clipboard().SetContent(SingleMessagePublicKey)
						}),
					),
				}
				screen.Refresh()

			case "pgp":
				SingleMessagePGPPrivateKey, err = crypt.PGPPrivateKeyToMem(privateKeyPEM, pass1.Text)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}

				PEM, err := crypt.PGPPublicKeyToBytes(SingleMessagePGPPrivateKey)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				SingleMessagePublicKey = string(PEM)
				SingleMessagePrivateKeyPath = f.URI().Path()

				screen.Objects = []fyne.CanvasObject{
					container.NewCenter(container.NewVBox(
						logo,
						widget.NewLabelWithStyle(fmt.Sprintf("Version: %s", Version), fyne.TextAlignCenter, fyne.TextStyle{}),
					)), container.NewVBox(
						widget.NewLabel("Loaded private key path: "+SingleMessagePrivateKeyPath),
						widget.NewButton("Reload New Private Key", privateKeyDialog.Show),
						widget.NewButton("Copy Public Key", func() {
							w.Clipboard().SetContent(SingleMessagePublicKey)
						}),
					),
				}
				screen.Refresh()

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
				{Text: "Password", Widget: pass1},
				{Text: "Key Type", Widget: kType},
			},
			nil,
			w,
		)
		d.SetOnClosed(unlock)
		d.Show()

	}, w)

	if SingleMessageRSAPrivateKey == nil && SingleMessagePGPPrivateKey == nil {
		screen = container.NewVBox(
			container.NewCenter(container.NewVBox(
				logo,
				widget.NewLabelWithStyle(fmt.Sprintf("Version: %s", Version), fyne.TextAlignCenter, fyne.TextStyle{}),
			)), container.NewVBox(
				widget.NewButton("Generate RSA Keys", rsaGenerateSaveDialog.Show),
				widget.NewButton("Generate PGP Keys", pgpGenerateSaveDialog.Show),
				widget.NewButton("Load Private Key", privateKeyDialog.Show),
			),
		)

		return screen
	}

	screen = container.NewVBox(
		container.NewCenter(container.NewVBox(
			logo,
			widget.NewLabelWithStyle(fmt.Sprintf("Version: %s", Version), fyne.TextAlignCenter, fyne.TextStyle{}),
		)), container.NewVBox(
			widget.NewLabel("Loaded private key path: "+SingleMessagePrivateKeyPath),
			widget.NewButton("Reload Private Key", privateKeyDialog.Show),
			widget.NewButton("Copy Public Key", func() {
				w.Clipboard().SetContent(SingleMessagePublicKey)
			}),
		),
	)

	return screen
}
