package logs

import (
	"strings"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/gorilla/websocket"
)

type logSender struct {
	marshalizer    marshal.Marshalizer
	conn           wsConn
	writer         *logWriter
	log            logger.Logger
	lastLogPattern string
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
	ls.lastLogPattern = logger.GetLogLevelPattern()

	defer func() {
		_ = ls.conn.Close()
		_ = ls.writer.Close()
		_ = logger.RemoveLogObserver(ls.writer)
		_ = logger.SetLogLevel(ls.lastLogPattern)
		ls.log.Info("reverted log pattern", "pattern", string(ls.lastLogPattern))
	}()

	err := ls.waitForPatternMessage()
	if err != nil {
		ls.log.Error(err.Error())
		return
	}

	go ls.monitorConnection()
	ls.doSendContinously()
}

func (ls *logSender) waitForPatternMessage() error {
	_, message, err := ls.conn.ReadMessage()
	if err != nil {
		return err
	}

	ls.log.Info("websocket log pattern received", "pattern", string(message))
	err = logger.SetLogLevel(string(message))
	if err != nil {
		return err
	}

	return nil
}

func (ls *logSender) monitorConnection() {
	for {
		mt, _, err := ls.conn.ReadMessage()
		if mt == websocket.CloseMessage {
			_ = ls.writer.Close()
			return
		}
		if err != nil {
			return
		}
	}
}

func (ls *logSender) doSendContinously() {
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
