package main

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/ElrondNetwork/elrond-go/cmd/logviewer/testing"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/gorilla/websocket"
	"github.com/urfave/cli"
)

const (
	defaultLogPath = "logs"
	ansi           = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
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

	// address defines a flag for setting the address and port on which the node will listen for connections
	address = cli.StringFlag{
		Name:  "address",
		Usage: "Address and port number on which the application will try to connect to the elrond-go node",
		Value: "127.0.0.1:8080",
	}

	// logLevelPatterns defines the logger levels and patterns
	logLevelPatterns = cli.StringFlag{
		Name:  "level",
		Usage: "This flag specifies the logger levels and patterns",
		Value: "*:" + logger.LogInfo.String(),
	}

	//logFile is used when the log output needs to be logged in a file
	logFile = cli.BoolFlag{
		Name:  "file",
		Usage: "will automatically log into a file",
	}

	// testServer is used when a test web socket is need for testing purposes
	testServer = cli.BoolFlag{
		Name:  "test-server",
		Usage: "starts a test server that will send log lines continuously. For development and debugging purposes only.",
	}

	ansiCleaner = regexp.MustCompile(ansi)
)

var log = logger.GetOrCreate("logviewer")
var cliApp *cli.App
var webSocket *websocket.Conn
var manualStop chan struct{}
var fileForLogs *os.File

func main() {
	initCliFlags()
	manualStop = make(chan struct{})
	log.Info("log viewer application started", "version", cliApp.Version)

	cliApp.Action = func(c *cli.Context) error {
		return startLogViewer(c)
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
	cliApp.Name = "Elrond Node CLI App"
	cliApp.Version = fmt.Sprintf("%s/%s/%s-%s", "1.0.0", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	cliApp.Usage = "This is the entry point for starting a new Elrond node - the app will start after the genesis timestamp"
	cliApp.Flags = []cli.Flag{
		address,
		logLevelPatterns,
		logFile,
		testServer,
	}
	cliApp.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}
}

func startLogViewer(ctx *cli.Context) error {
	var err error
	logLevelPattern := ctx.GlobalString(logLevelPatterns.Name)
	logLevels, _, err := logger.ParseLogLevelAndMatchingString(logLevelPattern)
	if err != nil {
		return err
	}

	if ctx.GlobalBool(testServer.Name) {
		_ = testing.StartTestWebSocket("127.0.0.1:8080")
		time.Sleep(time.Second * 2)
	}

	if ctx.GlobalBool(logFile.Name) {
		err = prepareLogFile()
		if err != nil {
			return err
		}

		defer func() {
			_ = fileForLogs.Close()
		}()
	}

	addrString := ctx.GlobalString(address.Name)
	webSocket, err = openWebSocket(addrString, logLevelPattern)
	if err != nil {
		return err
	}
	go listeningOnWebSocket()

	//set this log's level to the lowest desired log level that matches received logs from elrond-go
	lowestLogLevel := getLowestLogLevel(logLevels)
	log.SetLevel(lowestLogLevel)

	waitForUserToTerminateApp(webSocket)

	return nil
}

func getLowestLogLevel(logLevels []logger.LogLevel) logger.LogLevel {
	lowest := logLevels[0]
	for i := 1; i < len(logLevels); i++ {
		if lowest > logLevels[i] {
			lowest = logLevels[i]
		}
	}

	return lowest
}

func prepareLogFile() error {
	var workingDir = ""
	workingDir, err := os.Getwd()
	if err != nil {
		log.LogIfError(err)
		workingDir = ""
	}

	logDirectory := filepath.Join(workingDir, defaultLogPath)
	fileForLogs, err = core.CreateFile("logviewer", logDirectory, "log")
	if err != nil {
		return err
	}

	return nil
}

func openWebSocket(address string, logLevelPatterns string) (*websocket.Conn, error) {
	u := url.URL{
		Scheme: "ws",
		Host:   address,
		Path:   "/log",
	}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	err = conn.WriteMessage(websocket.TextMessage, []byte(logLevelPatterns))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func listeningOnWebSocket() {
	for {
		msgType, message, err := webSocket.ReadMessage()
		if msgType == websocket.CloseMessage {
			return
		}
		if err != nil {
			isConnectionClosed := strings.Contains(err.Error(), "websocket: close")
			if !isConnectionClosed {
				log.Error("logviewer websocket error", "error", err.Error())
				manualStop <- struct{}{}

			}
			return
		}

		outputMessage(message)
	}
}

func waitForUserToTerminateApp(conn *websocket.Conn) {
	stop := make(chan struct{}, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Info("terminating logviewer app at user's signal...")
		stop <- struct{}{}
	}()

	select {
	case <-stop:
		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		log.LogIfError(err)
		time.Sleep(time.Second)
	case <-manualStop:
	}

	log.Info("log viewer application stopped")
}

func outputMessage(message []byte) {
	fmt.Printf(string(message))

	if fileForLogs == nil {
		return
	}

	_, _ = fileForLogs.Write(stripAnsiColors(message))
	err := fileForLogs.Sync()
	if err != nil {
		log.Error("logviewer file sync error", "error", err.Error())
	}
}

func stripAnsiColors(input []byte) []byte {
	cleanedString := ansiCleaner.ReplaceAllString(string(input), "")

	return []byte(cleanedString)
}
