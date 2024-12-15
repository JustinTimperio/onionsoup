package crypt_test

import (
	"testing"

	"github.com/JustinTimperio/onionsoup/crypt"
)

func TestRSA(t *testing.T) {
	alicePrivate, alicePublic, err := crypt.RSAGenerateKeyPair(4096)
	if err != nil {
		t.Fatalf("Error generating key pair: %s", err)
	}

	bobPrivate, bobPublic, err := crypt.RSAGenerateKeyPair(4096)
	if err != nil {
		t.Fatalf("Error generating key pair: %s", err)
	}

	emsg, sig, err := crypt.RSAEncryptAndSignMessage(alicePrivate, bobPublic, []byte("Hello World!"))
	if err != nil {
		t.Fatalf("Error generating key pair: %s", err)
	}

	msg, err := crypt.RSADecryptAndVerifyMessage(bobPrivate, alicePublic, emsg, sig)
	if err != nil {
		t.Fatalf("Error generating key pair: %s", err)
	}

	if string(msg) != "Hello World!" {
		t.Fatalf("Message mismatch")
	}
}
