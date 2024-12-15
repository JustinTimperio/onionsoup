package server

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"time"

	"github.com/JustinTimperio/onionsoup/crypt"

	"fyne.io/fyne/v2"
	"github.com/ProtonMail/gopenpgp/v3/crypto"
)

type Message struct {
	Text         string `json:"text"`
	Time         int    `json:"time"`
	Token        string `json:"token"`
	FinalMessage bool   `json:"final_message"`
	Self         bool   `json:"_"`
}

type MessageWrapper struct {
	Message   []byte `json:"message"`
	Signature []byte `json:"signature"`
	ID        string `json:"id"`
}

type StartConversation struct {
	Address string `json:"address"`
	Token   string `json:"token"`
}

type StartConversationWrapper struct {
	EncryptionType string `json:"encryption_type"`
	ID             string `json:"id"`
	Auth           []byte `json:"auth"`
	Signature      []byte `json:"signature"`
}

type Conversation struct {
	ConversationAlias string
	ConversationID    string
	Messages          []*Message
	Render            *ConversationRender

	// Self
	SelfToken      string
	SelfPublicKey  any
	SelfPrivateKey any

	// Internal
	KeyType     string
	Established bool
	Ended       bool

	// Remote
	RemoteAddress   string
	RemoteToken     string
	RemotePublicKey any
}

func NewConversationHandle(
	sAddress, rAddress, keyType, conversationAlias, remoteToken, selfToken, convoID string,
	receiverPublicKey, senderPrivateKey, senderPublicKey any,
	w fyne.Window, rh *RouteHandler) ([]byte, *Conversation, error) {

	var ac = &StartConversation{
		Address: sAddress,
		Token:   selfToken,
	}

	acJBytes, err := json.Marshal(ac)
	if err != nil {
		return nil, nil, err
	}

	var message, sig []byte
	switch keyType {
	case "rsa":
		message, sig, err = crypt.RSAEncryptAndSignMessage(senderPrivateKey.(*rsa.PrivateKey), receiverPublicKey.(*rsa.PublicKey), acJBytes)
	case "pgp":
		message, err = crypt.PGPEncryptAndSignMessage(senderPrivateKey.(*crypto.Key), receiverPublicKey.(*crypto.Key), acJBytes)
	default:
		err = fmt.Errorf("Invalid key type")
	}
	if err != nil {
		return nil, nil, err
	}

	var sc = &StartConversationWrapper{
		EncryptionType: keyType,
		ID:             convoID,
		Auth:           message,
		Signature:      sig,
	}

	scJBytes, err := json.Marshal(sc)
	if err != nil {
		return nil, nil, err
	}

	c := &Conversation{
		ConversationID:    convoID,
		ConversationAlias: conversationAlias,
		Messages:          make([]*Message, 0),
		Render:            &ConversationRender{},

		KeyType:     keyType,
		Established: false,
		Ended:       false,

		SelfToken:      ac.Token,
		SelfPrivateKey: senderPrivateKey,
		SelfPublicKey:  senderPublicKey,

		RemoteAddress:   rAddress,
		RemoteToken:     remoteToken,
		RemotePublicKey: receiverPublicKey,
	}

	return scJBytes, c, nil
}

func (h *Conversation) PackMessage(message string, end bool) ([]byte, error) {
	msg := Message{
		Text:         message,
		Time:         int(time.Now().Unix()),
		Token:        h.RemoteToken,
		FinalMessage: end,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	var messageJson, sig []byte
	switch h.KeyType {
	case "rsa":
		messageJson, sig, err = crypt.RSAEncryptAndSignMessage(h.SelfPrivateKey.(*rsa.PrivateKey), h.RemotePublicKey.(*rsa.PublicKey), msgBytes)
	case "pgp":
		messageJson, err = crypt.PGPEncryptAndSignMessage(h.SelfPrivateKey.(*crypto.Key), h.RemotePublicKey.(*crypto.Key), msgBytes)
	default:
		err = fmt.Errorf("Invalid key type")
	}
	if err != nil {
		return nil, err
	}

	var messageWrapper = &MessageWrapper{
		Message:   messageJson,
		Signature: sig,
		ID:        h.ConversationID,
	}

	mwJBytes, err := json.Marshal(messageWrapper)
	if err != nil {
		return nil, fmt.Errorf("Error Parsing Message: %e", err)
	}

	return mwJBytes, nil
}

func (h *Conversation) UnpackMessage(eMessage, sig []byte) (*Message, error) {
	var (
		err         error
		messageJson []byte
	)

	switch h.KeyType {
	case "rsa":
		messageJson, err = crypt.RSADecryptAndVerifyMessage(h.SelfPrivateKey.(*rsa.PrivateKey), h.RemotePublicKey.(*rsa.PublicKey), eMessage, sig)
	case "pgp":
		messageJson, err = crypt.PGPDecryptAndVerifyMessage(h.SelfPrivateKey.(*crypto.Key), h.RemotePublicKey.(*crypto.Key), eMessage)
	default:
		err = fmt.Errorf("Invalid key type")
	}
	if err != nil {
		return nil, err
	}

	var message Message
	err = json.Unmarshal(messageJson, &message)
	if err != nil {
		return nil, fmt.Errorf("Error Parsing Message")
	}

	if message.Token != h.SelfToken {
		return nil, fmt.Errorf("Invalid Token")
	}

	return &message, nil
}
