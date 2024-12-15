package server

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/JustinTimperio/onionsoup/crypt"

	"fyne.io/fyne/v2"
	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/google/uuid"
)

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func (rh *RouteHandler) BootstrapConversation(token string, privateKey, publicKey, remotePublicKey any, alias string, w fyne.Window) error {

	var (
		messageJson []byte
		err         error
	)

	jBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return err
	}

	request := &StartConversationWrapper{}
	err = json.Unmarshal(jBytes, request)
	if err != nil {
		return err
	}

	switch request.EncryptionType {
	case "rsa":
		messageJson, err = crypt.RSADecryptAndVerifyMessage(privateKey.(*rsa.PrivateKey), remotePublicKey.(*rsa.PublicKey), request.Auth, request.Signature)
	case "pgp":
		messageJson, err = crypt.PGPDecryptAndVerifyMessage(privateKey.(*crypto.Key), remotePublicKey.(*crypto.Key), request.Auth)
	default:
		err = fmt.Errorf("Invalid key type")
	}
	if err != nil {
		return err
	}

	var auth StartConversation
	err = json.Unmarshal(messageJson, &auth)
	if err != nil {
		return err
	}

	b, ch, err := NewConversationHandle(
		rh.URL,
		auth.Address,
		request.EncryptionType,
		alias,
		auth.Token,
		uuid.New().String(),
		request.ID,
		remotePublicKey,
		privateKey,
		publicKey,
		w,
		rh,
	)
	if err != nil {
		return err
	}

	err = rh.SendMessage(b, ch.RemoteAddress, BootstrapPath)
	if err != nil {
		return err
	}

	ch.Established = true

	rh.mux.Lock()
	defer rh.mux.Unlock()

	rh.Conversations[request.ID] = ch

	return nil
}

func (rh *RouteHandler) GenerateConversation(privateKey, publicKey, remotePublicKey any, keyType, alias string, w fyne.Window) (string, error) {

	var (
		messageJson []byte
		emsg        []byte
		sig         []byte
		err         error
	)

	var auth StartConversation
	auth.Address = rh.URL
	auth.Token = randomString(128)

	messageJson, err = json.Marshal(auth)
	if err != nil {
		return "", err
	}

	switch keyType {
	case "rsa":
		emsg, sig, err = crypt.RSAEncryptAndSignMessage(privateKey.(*rsa.PrivateKey), remotePublicKey.(*rsa.PublicKey), messageJson)
	case "pgp":
		emsg, err = crypt.PGPEncryptAndSignMessage(privateKey.(*crypto.Key), remotePublicKey.(*crypto.Key), messageJson)
	default:
		err = fmt.Errorf("Invalid key type")
	}
	if err != nil {
		return "", err
	}

	var authWrapper StartConversationWrapper
	authWrapper.Auth = emsg
	authWrapper.Signature = sig
	authWrapper.ID = uuid.New().String()
	authWrapper.EncryptionType = keyType

	authBytes, err := json.Marshal(authWrapper)
	if err != nil {
		return "", err
	}
	authB64 := base64.StdEncoding.EncodeToString(authBytes)

	_, ch, err := NewConversationHandle(
		rh.URL,
		"",
		keyType,
		alias,
		"",
		auth.Token,
		authWrapper.ID,
		remotePublicKey,
		privateKey,
		publicKey,
		w,
		rh,
	)
	if err != nil {
		return "", err
	}

	rh.mux.Lock()
	defer rh.mux.Unlock()

	rh.Conversations[authWrapper.ID] = ch

	return authB64, nil
}
