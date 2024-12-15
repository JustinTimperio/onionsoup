package crypt_test

import (
	"testing"

	"github.com/JustinTimperio/onionsoup/crypt"
)

func TestAES(t *testing.T) {
	key := []byte("0123456789abcdef")
	plaintext := []byte("Hello World!")

	ebytes, err := crypt.AESEncrypt(key, plaintext)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	dbytes, err := crypt.AESDecrypt(key, ebytes)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	if string(plaintext) != string(dbytes) {
		t.Errorf("Expected %s, got %s", plaintext, dbytes)
	}

}
