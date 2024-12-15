package server

import (
	"context"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"github.com/cretz/bine/tor"
)

const (
	MessagePath   = "message"
	BootstrapPath = "bootstrap"
)

type RouteHandler struct {
	Conversations      map[string]*Conversation
	ConversationScreen *fyne.Container
	Receiver           *tor.OnionService
	Sender             http.Client
	Tor                *tor.Tor
	URL                string

	mux *sync.Mutex
}

func NewRouteHandler() (*RouteHandler, error) {

	t, err := tor.Start(nil, &tor.StartConf{
		DataDir: path.Join(os.TempDir(), "tor", randomString(10)),
	})
	if err != nil {
		return nil, err
	}

	listenCtx, listenCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer listenCancel()

	onion, err := t.Listen(listenCtx, &tor.ListenConf{RemotePorts: []int{80}, Version3: true})
	if err != nil {
		return nil, err
	}

	sendCtx, sendCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer sendCancel()

	d, err := t.Dialer(sendCtx, nil)
	if err != nil {
		return nil, err
	}

	t.DeleteDataDirOnClose = true
	t.StopProcessOnClose = true

	return &RouteHandler{
		Conversations: make(map[string]*Conversation),
		Tor:           t,
		Sender:        http.Client{Transport: &http.Transport{DialContext: d.DialContext}},
		Receiver:      onion,
		URL:           onion.String(),

		mux: &sync.Mutex{},
	}, nil
}

func (rh *RouteHandler) DeleteConversation(id string) {
	delete(rh.Conversations, id)
	rh.ConversationScreen.Refresh()
}

func (rh *RouteHandler) Close() {
	rh.Receiver.Close()
	rh.Tor.Close()
	rh.ConversationScreen.Refresh()
}
