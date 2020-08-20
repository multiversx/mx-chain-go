package main

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/gorilla/websocket"
	"github.com/urfave/cli"
)

const (
	defaultLogPath = "logs"
	wsLogPath      = "/log"
	ws             = "ws"
	wss            = "wss"
)

type config struct {
	workingDir       string
	address          string
	logLevelPatterns string
	logFile          bool
	useWss           bool
	withCorrelation  bool
	withLoggerName   bool
}

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
		Name:        "address",
		Usage:       "Address and port number on which the application will try to connect to the elrond-go node",
		Value:       "127.0.0.1:8080",
		Destination: &argsConfig.address,
	}

	// logLevelPatterns defines the logger levels and patterns
	logLevelPatterns = cli.StringFlag{
		Name:        "level",
		Usage:       "This flag specifies the logger levels and patterns",
		Value:       "*:" + logger.LogInfo.String(),
		Destination: &argsConfig.logLevelPatterns,
	}

	//logFile is used when the log output needs to be logged in a file
	logFile = cli.BoolFlag{
		Name:        "file",
		Usage:       "Will automatically log into a file",
		Destination: &argsConfig.logFile,
	}

	//useWss is used when the user require connection through wss
	useWss = cli.BoolFlag{
		Name:        "use-wss",
		Usage:       "Will use wss instead of ws when creating the web socket",
		Destination: &argsConfig.useWss,
	}

	// withCorrelation is used when the user wants to include the log correlation elements in logs
	withCorrelation = cli.BoolFlag{
		Name:        "correlation",
		Usage:       "Will include log correlation elements",
		Destination: &argsConfig.withCorrelation,
	}

	// withLoggerName is used when the user wants to include the logger name in logs
	withLoggerName = cli.BoolFlag{
		Name:        "logger-name",
		Usage:       "Will include logger name",
		Destination: &argsConfig.withLoggerName,
	}

	// workingDirectory defines a flag for the path for the working directory.
	workingDirectory = cli.StringFlag{
		Name:        "working-directory",
		Usage:       "The application will store here the logs in a subfolder.",
		Value:       "",
		Destination: &argsConfig.workingDir,
	}

	argsConfig = &config{}

	log           = logger.GetOrCreate("logviewer")
	cliApp        *cli.App
	webSocket     *websocket.Conn
	fileForLogs   *os.File
	marshalizer   marshal.Marshalizer
	retryDuration = time.Second * 10
)

func main() {
	initCliFlags()
	marshalizer = &marshal.GogoProtoMarshalizer{}

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
	cliApp.Name = "Elrond Logviewer App"
	cliApp.Version = fmt.Sprintf("%s/%s/%s-%s", "1.0.0", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	cliApp.Usage = "Logviewer application used to communicate with elrond-go node to log the message lines"
	cliApp.Flags = []cli.Flag{
		address,
		logLevelPatterns,
		logFile,
		workingDirectory,
		useWss,
		withCorrelation,
		withLoggerName,
	}
	cliApp.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}
}

func startLogViewer(ctx *cli.Context) error {
	log.Info("logviewer application started", "version", cliApp.Version)

	var err error
	logLevels, _, err := logger.ParseLogLevelAndMatchingString(argsConfig.logLevelPatterns)
	if err != nil {
		return err
	}

	if !ctx.IsSet(workingDirectory.Name) {
		argsConfig.workingDir, err = os.Getwd()
		if err != nil {
			log.LogIfError(err)
			argsConfig.workingDir = ""
		}
	}

	if argsConfig.logFile {
		err = prepareLogFile()
		if err != nil {
			return err
		}

		defer func() {
			_ = fileForLogs.Close()
		}()
	}

	profile := &logger.Profile{
		LogLevelPatterns: argsConfig.logLevelPatterns,
		WithCorrelation:  argsConfig.withCorrelation,
		WithLoggerName:   argsConfig.withLoggerName,
	}

	err = profile.Apply()
	log.LogIfError(err)

	go func() {
		for {
			webSocket, err = openWebSocket(argsConfig.address, profile)
			if err != nil {
				log.Error(fmt.Sprintf("logviewer websocket error, retrying in %v...", retryDuration), "error", err.Error())
				time.Sleep(retryDuration)
				continue
			}

			listeningOnWebSocket()
			time.Sleep(retryDuration)
		}
	}()

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
	logDirectory := filepath.Join(argsConfig.workingDir, defaultLogPath)
	logsFile, err := core.CreateFile(
		core.ArgCreateFileArgument{
			Prefix:        "logviewer",
			Directory:     logDirectory,
			FileExtension: "log",
		},
	)
	if err != nil {
		return err
	}

	return logger.AddLogObserver(logsFile, &logger.PlainFormatter{})
}

func openWebSocket(address string, profile *logger.Profile) (*websocket.Conn, error) {
	scheme := ws

	if argsConfig.useWss {
		scheme = wss
	}

	u := url.URL{
		Scheme: scheme,
		Host:   address,
		Path:   wsLogPath,
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

func listeningOnWebSocket() {
	for {
		msgType, message, err := webSocket.ReadMessage()
		if msgType == websocket.CloseMessage {
			return
		}
		if err == nil {
			outputMessage(message)
			continue
		}

		_, isConnectionClosed := err.(*websocket.CloseError)
		if !isConnectionClosed {
			log.Error("logviewer websocket error, retrying in %v...", "error", err.Error())
		} else {
			log.Error(fmt.Sprintf("logviewer websocket terminated by the server side, retrying in %v...", retryDuration), "error", err.Error())
		}
		return
	}
}

func waitForUserToTerminateApp(conn *websocket.Conn) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	log.Info("terminating logviewer app at user's signal...")
	if conn != nil {
		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		log.LogIfError(err)
		time.Sleep(time.Second)
	}

	log.Info("logviewer application stopped")
}

func outputMessage(message []byte) {
	logLine := &logger.LogLineWrapper{}

	err := marshalizer.Unmarshal(logLine, message)
	if err != nil {
		log.Debug("can not unmarshal received data", "data", hex.EncodeToString(message))
		return
	}

	recoveredLogLine := &logger.LogLine{
		LoggerName:  logLine.LoggerName,
		Correlation: logLine.Correlation,
		Message:     logLine.Message,
		LogLevel:    logger.LogLevel(logLine.LogLevel),
		Args:        make([]interface{}, len(logLine.Args)),
		Timestamp:   time.Unix(0, logLine.Timestamp),
	}
	for i, str := range logLine.Args {
		recoveredLogLine.Args[i] = str
	}

	log.Log(recoveredLogLine)
}
