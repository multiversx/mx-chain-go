package provider

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
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
func InitLogHandler(presenter PresenterHandler, nodeURL string, profile *logger.Profile, useWss bool, customLogProfile bool) error {
	var err error
	scheme := ws
	if useWss {
		scheme = wss
	}
	go func() {
		for {
			webSocket, err = openWebSocket(scheme, nodeURL)
			if err != nil {
				_, _ = presenter.Write([]byte(fmt.Sprintf("termui websocket error, retrying in %v...", retryDuration)))
				time.Sleep(retryDuration)
				continue
			}

			if customLogProfile {
				err = sendProfile(webSocket, profile)
			} else {
				err = sendDefaultProfileIdentifier(webSocket)
			}
			log.LogIfError(err)

			startListeningOnWebSocket(presenter)
			time.Sleep(retryDuration)
		}
	}()

	return nil
}

func openWebSocket(scheme string, address string) (*websocket.Conn, error) {
	u := url.URL{
		Scheme: scheme,
		Host:   address,
		Path:   "/log",
	}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func sendProfile(conn *websocket.Conn, profile *logger.Profile) error {
	profileMessage, err := profile.Marshal()
	if err != nil {
		return err
	}

	return conn.WriteMessage(websocket.TextMessage, profileMessage)
}

func sendDefaultProfileIdentifier(conn *websocket.Conn) error {
	return conn.WriteMessage(websocket.TextMessage, []byte(core.DefaultLogProfileIdentifier))
}

// startListeningOnWebSocket will listen if a new log message is received and will display it
func startListeningOnWebSocket(presenter PresenterHandler) {
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
			_, _ = presenter.Write([]byte(fmt.Sprintf("termui websocket error: %s", err.Error())))
		} else {
			_, _ = presenter.Write([]byte(fmt.Sprintf("termui websocket terminated: %s", err.Error())))
		}
		return
	}
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
