package client

import (
	"strings"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/gorilla/websocket"
)

// Writer -
type Writer interface {
	ReadBlocking() ([]byte, bool)
	Close() error
}

const disconnectMessage = -1

var log = logger.GetOrCreate("websockets/client")

type webSocketsClient struct {
	writer Writer
	conn   *websocket.Conn
}

// NewWebSocketClient -
func NewWebSocketClient(writer Writer) (*webSocketsClient, error) {
	return &webSocketsClient{
		writer: writer,
	}, nil
}

// StartSendingBlocking -
func (wsc *webSocketsClient) StartSendingBlocking(conn *websocket.Conn) {
	wsc.conn = conn

	defer func() {
		_ = wsc.conn.Close()
		_ = wsc.writer.Close()
	}()

	go wsc.monitorConnection()

	wsc.doSendContinuously()
}

func (wsc *webSocketsClient) monitorConnection() {
	var err error
	var mt int

	defer func() {
		_ = wsc.writer.Close()
	}()

	for {
		mt, _, err = wsc.conn.ReadMessage()
		if mt == websocket.CloseMessage || mt == disconnectMessage {
			return
		}
		if err != nil {
			return
		}
	}
}

func (wsc *webSocketsClient) doSendContinuously() {
	for {
		shouldStop := wsc.sendMessage()
		if shouldStop {
			return
		}
	}
}

func (wsc *webSocketsClient) sendMessage() (shouldStop bool) {
	data, ok := wsc.writer.ReadBlocking()
	if !ok {
		return true
	}

	err := wsc.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		isConnectionClosed := strings.Contains(err.Error(), "websocket: close sent")
		if !isConnectionClosed {
			log.Error("web socket error", "error", err.Error())
		} else {
			log.Info("web socket", "connection", "closed")
		}

		return true
	}

	return false
}
