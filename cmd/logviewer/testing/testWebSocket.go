package testing

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/gorilla/websocket"
)

var log = logger.GetOrCreate("testWebSocket")

// TestWebSocket is a testing web socket implementation used in logviewer development
// The purpose of this struct is that the logviewer can be developed/debugged separately from
// the elrond-go node
type TestWebSocket struct {
	address            string
	server             *http.Server
	upgrader           websocket.Upgrader
	mutLogLevelPattern sync.RWMutex
	logLevels          []logger.LogLevel
	patterns           []string
}

// StartTestWebSocket starts a test web socket server, initialize it, opens /log route and waits for
// clients to connect. After the connection is done, it waits for the pattern message and then it sends continuously
// log lines-type of messages
func StartTestWebSocket(address string) *TestWebSocket {
	tws := &TestWebSocket{
		address:   address,
		upgrader:  websocket.Upgrader{},
		logLevels: make([]logger.LogLevel, 0),
		patterns:  make([]string, 0),
	}

	go tws.init()
	log.Info("test web socket", "server", "open")

	return tws
}

func (tws *TestWebSocket) init() {
	m := http.NewServeMux()
	tws.server = &http.Server{Addr: tws.address, Handler: m}

	m.HandleFunc("/log", tws.handleLogRequest)
	err := tws.server.ListenAndServe()
	if err != nil {
		log.Error("test web socket error", "error", err.Error())
	}
}

func (tws *TestWebSocket) handleLogRequest(w http.ResponseWriter, r *http.Request) {
	conn, err := tws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err.Error())
		return
	}

	defer func() {
		_ = conn.Close()
	}()

	err = tws.waitForPatternMessage(conn)
	if err != nil {
		log.Error(err.Error())
		return
	}

	tws.sendContinously(conn)
}

func (tws *TestWebSocket) waitForPatternMessage(conn *websocket.Conn) error {
	_, message, err := conn.ReadMessage()
	if err != nil {
		return err
	}

	log.Info("pattern received", "pattern", string(message))
	logLevels, patterns, err := logger.ParseLogLevelAndMatchingString(string(message))
	if err != nil {
		return err
	}

	tws.mutLogLevelPattern.Lock()
	tws.logLevels = logLevels
	tws.patterns = patterns
	tws.mutLogLevelPattern.Unlock()

	return nil
}

func (tws *TestWebSocket) sendContinously(conn *websocket.Conn) {
	idxCrtLogLevel := 0
	count := 0

	chClose := make(chan struct{})
	formatter := &logger.ConsoleFormatter{}
	go func() {
		for {
			mt, _, err := conn.ReadMessage()
			if mt == websocket.CloseMessage {
				chClose <- struct{}{}
				return
			}
			if err != nil {
				return
			}
		}
	}()

	for {
		logLine := tws.generateLogLine(&idxCrtLogLevel, &count)
		if tws.shouldSendMessage(logLine.LogLevel) {
			shouldStop := tws.sendMessage(formatter, conn, logLine)
			if shouldStop {
				break
			}
		}

		select {
		case <-chClose:
			return
		case <-time.After(time.Millisecond * 100):
		}
	}
}

func (tws *TestWebSocket) shouldSendMessage(messageLogLevel logger.LogLevel) bool {
	tws.mutLogLevelPattern.RLock()
	defer tws.mutLogLevelPattern.RUnlock()

	lastLogLevelForWebsocket := logger.LogNone
	for i := 0; i < len(tws.logLevels); i++ {
		level := tws.logLevels[i]
		pattern := tws.patterns[i]

		if pattern == "*" || pattern == "websocket" {
			lastLogLevelForWebsocket = level
			continue
		}
	}

	return messageLogLevel >= lastLogLevelForWebsocket
}

func (tws *TestWebSocket) sendMessage(formatter logger.Formatter, conn *websocket.Conn, logLine *logger.LogLine) bool {
	message := formatter.Output(logLine)

	err := conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		isConnectionClosed := strings.Contains(err.Error(), "websocket: close sent")
		if !isConnectionClosed {
			log.Error("test web socket error", "error", err.Error())
		} else {
			log.Info("test web socket", "connection", "closed")
		}

		return true
	}

	return false
}

func (tws *TestWebSocket) generateLogLine(idxCrtLogLevel *int, count *int) *logger.LogLine {
	logLevel := logger.Levels[*idxCrtLogLevel]
	*idxCrtLogLevel = (*idxCrtLogLevel + 1) % len(logger.Levels)
	if logLevel == logger.LogNone {
		logLevel = logger.Levels[*idxCrtLogLevel]
		*idxCrtLogLevel = (*idxCrtLogLevel + 1) % len(logger.Levels)
	}

	logLine := &logger.LogLine{
		Message:   "websocket test message",
		LogLevel:  logLevel,
		Args:      []interface{}{"count", *count},
		Timestamp: time.Now(),
	}
	*count = *count + 1

	return logLine
}
