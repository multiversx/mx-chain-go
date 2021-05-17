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
var marshalizer = &marshal.GogoProtoMarshalizer{}

const (
	ws  = "ws"
	wss = "wss"
)

// LogHandlerArgs hold the arguments needed to initialize log handling
type LogHandlerArgs struct {
	Presenter          PresenterHandler
	NodeURL            string
	Profile            *logger.Profile
	ChanNodeIsStarting chan struct{}
	UseWss             bool
	CustomLogProfile   bool
}

// InitLogHandler will open the websocket and start listening to logs
func InitLogHandler(args LogHandlerArgs) error {
	if args.Presenter == nil {
		return ErrNilTermuiPresenter
	}
	if args.ChanNodeIsStarting == nil {
		return ErrNilChanNodeIsStarting
	}
	if len(args.NodeURL) == 0 {
		return ErrEmptyNodeURL
	}

	var err error
	scheme := ws
	if args.UseWss {
		scheme = wss
	}
	go func() {
		for {
			webSocket, err = openWebSocket(scheme, args.NodeURL)
			if err != nil {
				_, _ = args.Presenter.Write([]byte(fmt.Sprintf("termui websocket error, retrying in %v...", retryDuration)))
				time.Sleep(retryDuration)
				continue
			}

			if args.CustomLogProfile {
				err = sendProfile(webSocket, args.Profile)
			} else {
				err = sendDefaultProfileIdentifier(webSocket)
			}
			log.LogIfError(err)

			startListeningOnWebSocket(args.Presenter, args.ChanNodeIsStarting)
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
func startListeningOnWebSocket(presenter PresenterHandler, chanNodeIsStarting chan struct{}) {
	for {
		msgType, message, err := webSocket.ReadMessage()
		if msgType == websocket.CloseMessage {
			return
		}
		if err == nil {
			writeMessage(presenter, message, chanNodeIsStarting)
			continue
		}

		_, isConnectionClosed := err.(*websocket.CloseError)
		if !isConnectionClosed {
			_, _ = presenter.Write([]byte(fmt.Sprintf("termui websocket error: %s", err.Error())))
		} else {
			_, _ = presenter.Write([]byte(fmt.Sprintf("termui websocket terminated: %s", err.Error())))
		}
		chanNodeIsStarting <- struct{}{}
		return
	}
}

func writeMessage(presenter PresenterHandler, message []byte, chanNodeIsStarting chan struct{}) {
	if strings.Contains(string(message), "/node/status") {
		return
	}
	if strings.Contains(string(message), "Shuffled out - soft restart") {
		chanNodeIsStarting <- struct{}{}
	}

	message = formatMessage(message)
	_, _ = presenter.Write(message)
}

func formatMessage(message []byte) []byte {
	logLine := &logger.LogLineWrapper{}

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
