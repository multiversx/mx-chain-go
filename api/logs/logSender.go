package logs

import (
	"bytes"
	"strings"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/gorilla/websocket"
)

const disconnectMessage = -1

type logSender struct {
	marshalizer marshal.Marshalizer
	conn        wsConn
	writer      *logWriter
	log         logger.Logger
	lastProfile logger.Profile
}

// NewLogSender returns a new component that is able to communicate with the log viewer application.
// After the correct handshake it will send all logs that come through the logger subsystem
func NewLogSender(marshalizer marshal.Marshalizer, conn wsConn, log logger.Logger) (*logSender, error) {
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(log) {
		return nil, ErrNilLogger
	}
	if conn == nil {
		return nil, ErrNilWsConn
	}

	ls := &logSender{
		marshalizer: marshalizer,
		log:         log,
		conn:        conn,
	}

	err := ls.registerLogWriter()
	if err != nil {
		return nil, err
	}

	return ls, nil
}

func (ls *logSender) registerLogWriter() error {
	w := NewLogWriter()
	formatter, err := logger.NewLogLineWrapperFormatter(ls.marshalizer)
	if err != nil {
		return err
	}

	err = logger.AddLogObserver(w, formatter)
	if err != nil {
		return err
	}

	ls.writer = w

	return nil
}

// StartSendingBlocking initialize the handshake by waiting the correct pattern and after that
// will start sending logs information and in the same time monitor the current connection.
// When the connection ends it will revert the previous log pattern.
func (ls *logSender) StartSendingBlocking() {
	ls.lastProfile = logger.GetCurrentProfile()

	defer func() {
		_ = ls.conn.Close()
		_ = ls.writer.Close()
		_ = logger.RemoveLogObserver(ls.writer)
		_ = ls.lastProfile.Apply()

		ls.log.Info("reverted log profile", "profile", ls.lastProfile.String())
	}()

	err := ls.waitForProfile()
	if err != nil {
		ls.log.Error(err.Error())
		return
	}

	go ls.monitorConnection()
	ls.doSendContinuously()
}

func (ls *logSender) waitForProfile() error {
	_, message, err := ls.conn.ReadMessage()
	if err != nil {
		return err
	}

	if bytes.Equal(message, []byte(core.DefaultLogProfileIdentifier)) {
		return nil
	}

	profile, err := logger.UnmarshalProfile(message)
	if err != nil {
		return err
	}

	ls.log.Info("websocket log profile received", "profile", profile.String())

	err = profile.Apply()
	if err != nil {
		return err
	}

	logger.NotifyProfileChange()
	return nil
}

func (ls *logSender) monitorConnection() {
	var err error
	var mt int

	defer func() {
		_ = ls.writer.Close()
	}()

	for {
		mt, _, err = ls.conn.ReadMessage()
		ls.log.Trace("message type", "value", mt)
		if mt == websocket.CloseMessage || mt == disconnectMessage {
			return
		}
		if err != nil {
			return
		}
	}
}

func (ls *logSender) doSendContinuously() {
	for {
		shouldStop := ls.sendMessage()
		if shouldStop {
			return
		}
	}
}

func (ls *logSender) sendMessage() (shouldStop bool) {
	data, ok := ls.writer.ReadBlocking()
	if !ok {
		return true
	}

	err := ls.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		isConnectionClosed := strings.Contains(err.Error(), "websocket: close sent")
		if !isConnectionClosed {
			ls.log.Error("web socket error", "error", err.Error())
		} else {
			ls.log.Info("web socket", "connection", "closed")
		}

		return true
	}

	return false
}
