package server

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/JustinTimperio/onionsoup/crypt"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/labstack/echo/v4"
)

func (rh *RouteHandler) SendMessage(packedMessage []byte, url, path string) error {

	r, err := http.NewRequest("POST", fmt.Sprintf("http://%s/%s", url, path), bytes.NewBuffer(packedMessage))
	if err != nil {
		return err
	}

	r.Header.Set("Content-Type", "application/json")

	resp, err := rh.Sender.Do(r)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Invalid Status Code: %d", resp.StatusCode)
	}

	return nil
}

func (rh *RouteHandler) Message(c echo.Context) error {
	var mw MessageWrapper
	if err := c.Bind(&mw); err != nil {
		return err
	}

	rh.mux.Lock()
	defer rh.mux.Unlock()

	convo, ok := rh.Conversations[mw.ID]
	if !ok {
		return echo.ErrUnauthorized
	}

	if convo.Ended {
		return echo.ErrUnauthorized
	}

	msg, err := convo.UnpackMessage(mw.Message, mw.Signature)
	if err != nil {
		return echo.ErrUnauthorized
	}
	msg.Self = false

	if msg.FinalMessage {
		convo.Ended = true
		convo.Render.UpdateHeader(convo, "ended")
	}

	err = convo.Render.AddMessage(*msg)
	if err != nil {
		return echo.ErrInternalServerError
	}

	return nil
}

func (rh *RouteHandler) Bootstrap(c echo.Context) error {
	var (
		messageJson []byte
		err         error
	)

	startConversation := &StartConversationWrapper{}
	if err := c.Bind(startConversation); err != nil {
		return err
	}

	rh.mux.Lock()
	defer rh.mux.Unlock()

	convo, ok := rh.Conversations[startConversation.ID]
	if !ok {
		return echo.ErrUnauthorized
	}

	switch convo.KeyType {
	case "rsa":
		messageJson, err = crypt.RSADecryptAndVerifyMessage(convo.SelfPrivateKey.(*rsa.PrivateKey), convo.RemotePublicKey.(*rsa.PublicKey), startConversation.Auth, startConversation.Signature)
	case "pgp":
		messageJson, err = crypt.PGPDecryptAndVerifyMessage(convo.SelfPrivateKey.(*crypto.Key), convo.RemotePublicKey.(*crypto.Key), startConversation.Auth)
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

	convo.Established = true
	convo.RemoteToken = auth.Token
	convo.RemoteAddress = auth.Address
	convo.Render.UpdateHeader(convo, "active")

	return nil
}
