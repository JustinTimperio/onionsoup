package menus

import (
	"fmt"

	"github.com/JustinTimperio/onionsoup/crypt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func DecryptView(w fyne.Window) fyne.CanvasObject {
	var rawMessage string

	// Input Message
	rawMessageInput := widget.NewEntryWithData(binding.BindString(&rawMessage))
	rawMessageInput.SetPlaceHolder("Paste the encrypted base64 encoded message here")
	rawMessageInput.MultiLine = true
	rawMessageInput.SetMinRowsVisible(5)
	rawMessageInput.Wrapping = fyne.TextWrapWord

	// Decrypted Message
	decryptedMessage := widget.NewEntry()
	decryptedMessage.MultiLine = true
	decryptedMessage.SetMinRowsVisible(15)

	// Private Key
	decryptForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Raw Message", Widget: rawMessageInput},
		},
		OnSubmit: func() {
			if rawMessage == "" {
				dialog.ShowError(fmt.Errorf("No Message to Encrypt"), w)
				return
			}

			emsg, sig, senderPubkey, err := crypt.UnpackMessage(rawMessage)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			switch SingleMessageKeyType {
			case "pgp":
				pubkey, err := crypt.PGPPublicKeyToMem(senderPubkey)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}

				msg, err := crypt.PGPDecryptAndVerifyMessage(SingleMessagePGPPrivateKey, pubkey, emsg)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}

				decryptedMessage.SetText(string(msg))

			case "rsa":
				pubkey, err := crypt.RSAPublicKeyToMem(senderPubkey)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}

				msg, err := crypt.RSADecryptAndVerifyMessage(SingleMessageRSAPrivateKey, pubkey, emsg, sig)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}

				decryptedMessage.SetText(string(msg))

			default:
				dialog.ShowError(fmt.Errorf("Key Type Not Selected"), w)
				return
			}
		},
	}

	return container.NewVBox(
		decryptForm,
		widget.NewSeparator(),
		widget.NewLabel("Decrypted Message"),
		decryptedMessage,
	)
}
