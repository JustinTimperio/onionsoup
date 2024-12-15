package menus

import (
	"crypto/rsa"

	"fyne.io/fyne/v2"
	"github.com/ProtonMail/gopenpgp/v3/crypto"
)

var SingleMessageRSAPrivateKey *rsa.PrivateKey
var SingleMessagePGPPrivateKey *crypto.Key
var SingleMessagePublicKey string
var SingleMessagePrivateKeyPath string
var SingleMessageKeyType string

type Menu struct {
	Title string
	View  func(w fyne.Window) fyne.CanvasObject
}

var (
	Menus = map[string]Menu{
		"Home":          {"Home", HomeView},
		"Conversations": {"Conversations", ServerView},
		"Encrypt":       {"Encrypt", EncryptView},
		"Decrypt":       {"Decrypt", DecryptView},
	}

	MenuOrder = []string{"Home", "Conversations", "Encrypt", "Decrypt"}
)
