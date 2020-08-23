package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go-logger/proto"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/gorilla/websocket"
	"github.com/urfave/cli"
)

var (
	nodeHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
VERSION:
   {{.Version}}
   {{end}}
`
	// addressFlag defines a flag for setting the address and port on which the node will listen for connections
	addressFlag = cli.StringFlag{
		Name:        "address",
		Usage:       "Address and port number on which the application will try to connect to the elrond-go node",
		Value:       "127.0.0.1:8080",
		Destination: &address,
	}

	log                = logger.GetOrCreate("testWebSocket")
	cliApp             *cli.App
	address            string
	server             *http.Server
	upgrader           websocket.Upgrader
	mutLogLevelPattern sync.RWMutex
	logLevels          []logger.LogLevel
	patterns           []string
	marshalizer        marshal.Marshalizer
)

// This application starts a test web socket server, initialize it, opens "/log" route and waits for
// clients to connect. After the connection is done, it waits for the pattern message and then it sends continuously
// log lines-type of messages
func main() {
	initCliFlags()
	marshalizer = &marshal.GogoProtoMarshalizer{}

	cliApp.Action = func(c *cli.Context) error {
		startTestLogViewer()

		return nil
	}

	err := cliApp.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func initCliFlags() {
	cliApp = cli.NewApp()
	cli.AppHelpTemplate = nodeHelpTemplate
	cliApp.Name = "Elrond Logviewer Testing App"
	cliApp.Version = fmt.Sprintf("%s/%s/%s-%s", "1.0.0", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	cliApp.Usage = "Testing application for Logwiever"
	cliApp.Flags = []cli.Flag{
		addressFlag,
	}
	cliApp.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}
}

func startTestLogViewer() {
	log.Info("test logviewer application", "status", "started", "version", cliApp.Version)

	upgrader = websocket.Upgrader{}
	logLevels = make([]logger.LogLevel, 0)
	patterns = make([]string, 0)

	go initWebSocketServer()
	log.Info("test web socket", "server", "open")

	waitForUserToTerminateApp()
}

func waitForUserToTerminateApp() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	log.Info("terminating test logviewer application at user's signal...")
}

func initWebSocketServer() {
	m := http.NewServeMux()
	server = &http.Server{Addr: address, Handler: m}

	m.HandleFunc("/log", handleLogRequest)
	err := server.ListenAndServe()
	if err != nil {
		log.Error("test web socket error", "error", err.Error())
	}
}

func handleLogRequest(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err.Error())
		return
	}

	defer func() {
		_ = conn.Close()
	}()

	err = waitForProfile(conn)
	if err != nil {
		log.Error(err.Error())
		return
	}

	sendContinouslyWithMonitor(conn)
}

func waitForProfile(conn *websocket.Conn) error {
	_, message, err := conn.ReadMessage()
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

	log.Info("websocket log profile received", "profile", profile.String())

	mutLogLevelPattern.Lock()
	logLevels, patterns, err = logger.ParseLogLevelAndMatchingString(profile.LogLevelPatterns)
	mutLogLevelPattern.Unlock()
	if err != nil {
		return err
	}

	return nil
}

func sendContinouslyWithMonitor(conn *websocket.Conn) {
	idxCrtLogLevel := 0
	count := 0
	chClose := make(chan struct{})

	go monitorConnection(conn, chClose)
	doSendContinously(conn, &idxCrtLogLevel, &count, chClose)
}

func monitorConnection(conn *websocket.Conn, chClose chan struct{}) {
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
}

func doSendContinously(conn *websocket.Conn, idxCrtLogLevel *int, count *int, chClose chan struct{}) {
	for {
		logLine := generateLogLine(idxCrtLogLevel, count)
		if shouldSendMessage(logLine.GetLogLevel()) {
			shouldStop := sendMessage(conn, logLine)
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

func shouldSendMessage(messageLogLevel int32) bool {
	mutLogLevelPattern.RLock()
	defer mutLogLevelPattern.RUnlock()

	lastLogLevelForWebsocket := logger.LogInfo
	for i := 0; i < len(logLevels); i++ {
		level := logLevels[i]
		pattern := patterns[i]

		if pattern == "*" || pattern == "websocket" {
			lastLogLevelForWebsocket = level
			continue
		}
	}

	return messageLogLevel >= int32(lastLogLevelForWebsocket)
}

func sendMessage(conn *websocket.Conn, logLine logger.LogLineHandler) bool {
	message, err := marshalizer.Marshal(logLine)
	if err != nil {
		log.Error("test web socket marshal", "error", err.Error())
		return false
	}

	err = conn.WriteMessage(websocket.TextMessage, message)
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

func generateLogLine(idxCrtLogLevel *int, count *int) logger.LogLineHandler {
	logLevel := logger.Levels[*idxCrtLogLevel]
	*idxCrtLogLevel = (*idxCrtLogLevel + 1) % len(logger.Levels)
	if logLevel == logger.LogNone {
		logLevel = logger.Levels[*idxCrtLogLevel]
		*idxCrtLogLevel = (*idxCrtLogLevel + 1) % len(logger.Levels)
	}

	logLine := &logger.LogLineWrapper{}
	logLine.LoggerName = "websocket test"
	logLine.Correlation = proto.LogCorrelationMessage{Shard: "foobar", Epoch: 42}
	logLine.Message = "websocket test message"
	logLine.LogLevel = int32(logLevel)
	logLine.Args = []string{"count", fmt.Sprintf("%v", *count)}
	logLine.Timestamp = time.Now().UnixNano()

	*count++

	return logLine
}
