package crypt_test

import (
	"testing"

	"github.com/JustinTimperio/onionsoup/crypt"
)

func TestPGP(t *testing.T) {

	message := []byte("Hello World")

	alice, err := crypt.PGPGenerateKeyPair("Alice")
	if err != nil {
		t.Error(err)
		return
	}

	bob, err := crypt.PGPGenerateKeyPair("Bob")
	if err != nil {
		t.Error(err)
		return
	}

	encryptedMessage, err := crypt.PGPEncryptAndSignMessage(alice, bob, message)
	if err != nil {
		t.Error(err)
		return
	}

	decryptedMessage, err := crypt.PGPDecryptAndVerifyMessage(bob, alice, encryptedMessage)
	if err != nil {
		t.Error(err)
		return
	}

	if string(message) != string(decryptedMessage) {
		t.Error("Message not equal")
		return
	}
	t.Logf("%s", decryptedMessage)

	return
}
