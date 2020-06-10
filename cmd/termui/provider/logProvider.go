package provider

import (
	"encoding/hex"
	"net/url"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/gorilla/websocket"
)

var formatter = logger.PlainFormatter{}
var webSocket *websocket.Conn
var retryDuration = time.Second * 10

const (
	ws  = "ws"
	wss = "wss"
)

// InitLogHandler will open the websocket and set the log level
func InitLogHandler(nodeURL string, profile logger.Profile, useWss bool) error {
	var err error
	scheme := ws
	if useWss {
		scheme = wss
	}
	for {
		webSocket, err = openWebSocket(scheme, nodeURL, profile)
		if err == nil {
			break
		}
		log.Error("init logger", "error", err)

		log.Info("connection not available", "duration before retrying", retryDuration)
		time.Sleep(retryDuration)
	}

	return nil
}

func openWebSocket(scheme string, address string, profile logger.Profile) (*websocket.Conn, error) {
	u := url.URL{
		Scheme: scheme,
		Host:   address,
		Path:   "/log",
	}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	profileMessage, err := profile.Marshal()
	if err != nil {
		return nil, err
	}

	err = conn.WriteMessage(websocket.TextMessage, profileMessage)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// StartListeningOnWebSocket will listen if a new log message is received and will display it
func StartListeningOnWebSocket(presenter PresenterHandler) {
	go func() {
		for {
			msgType, message, err := webSocket.ReadMessage()
			if msgType == websocket.CloseMessage {
				return
			}
			if err == nil {
				writeMessage(presenter, message)
				continue
			}

			_, isConnectionClosed := err.(*websocket.CloseError)
			if !isConnectionClosed {
				log.Error("termui websocket error", "error", err.Error())
			} else {
				log.Debug("termui websocket terminated", "error", err.Error())
			}

			time.Sleep(retryDuration)
		}
	}()
}

func writeMessage(presenter PresenterHandler, message []byte) {
	if strings.Contains(string(message), "/node/status") {
		return
	}

	message = formatMessage(message)
	_, _ = presenter.Write(message)
}

func formatMessage(message []byte) []byte {
	logLine := &logger.LogLineWrapper{}

	marshalizer := &marshal.GogoProtoMarshalizer{}
	err := marshalizer.Unmarshal(logLine, message)
	if err != nil {
		log.Debug("can not unmarshal received data", "data", hex.EncodeToString(message))
		return nil
	}

	return formatter.Output(logLine)
}

// StopWebSocket will send notify the node that the app is closed
func StopWebSocket() {
	if webSocket != nil {
		err := webSocket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		log.LogIfError(err)
		time.Sleep(time.Second)
	}
}
