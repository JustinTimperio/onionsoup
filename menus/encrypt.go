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

func EncryptView(w fyne.Window) fyne.CanvasObject {
	var recipientPublicKey string
	var rawMessage string

	// Input Message
	rawMessageInput := widget.NewEntryWithData(binding.BindString(&rawMessage))
	rawMessageInput.SetPlaceHolder("Paste or type your message here")
	rawMessageInput.MultiLine = true
	rawMessageInput.SetMinRowsVisible(15)

	// Recipient Public Key
	recipientPublicKeyInput := widget.NewEntryWithData(binding.BindString(&recipientPublicKey))
	recipientPublicKeyInput.SetPlaceHolder("Paste your recipient's public key here")
	recipientPublicKeyInput.MultiLine = true
	recipientPublicKeyInput.SetMinRowsVisible(5)

	encryptForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Recipient's Public Key", Widget: recipientPublicKeyInput},
			{Text: "Raw Message", Widget: rawMessageInput},
		},
		OnSubmit: func() {
			if rawMessage == "" || recipientPublicKey == "" {
				dialog.ShowError(fmt.Errorf("Please fill in all fields"), w)
				return
			}

			var emsg, sig []byte
			switch SingleMessageKeyType {
			case "rsa":
				if SingleMessageRSAPrivateKey == nil {
					dialog.ShowError(fmt.Errorf("No private key found"), w)
					return
				}

				pubkey, err := crypt.RSAPublicKeyToMem([]byte(recipientPublicKey))
				if err != nil {
					dialog.ShowError(err, w)
					return
				}

				emsg, sig, err = crypt.RSAEncryptAndSignMessage(SingleMessageRSAPrivateKey, pubkey, []byte(rawMessage))
				if err != nil {
					dialog.ShowError(err, w)
					return
				}

			case "pgp":
				if SingleMessagePGPPrivateKey == nil {
					dialog.ShowError(fmt.Errorf("No private key found"), w)
					return
				}

				pubkey, err := crypt.PGPPublicKeyToMem([]byte(recipientPublicKey))
				if err != nil {
					dialog.ShowError(err, w)
					return
				}

				emsg, err = crypt.PGPEncryptAndSignMessage(SingleMessagePGPPrivateKey, pubkey, []byte(rawMessage))
				if err != nil {
					dialog.ShowError(err, w)
					return
				}

			default:
				dialog.ShowError(fmt.Errorf("No private key found"), w)
				return
			}

			packedMessage, err := crypt.PackMessage(emsg, sig, []byte(SingleMessagePublicKey))
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			// Show encrypted message
			w.Clipboard().SetContent(packedMessage)
			dialog.ShowInformation("Successfully Encrypted", "Sent encrypted message to clipboard!", w)

			// Clear input fields
			rawMessageInput.SetText("")
			recipientPublicKeyInput.SetText("")
		},
	}

	return container.NewStack(
		encryptForm,
	)
}
