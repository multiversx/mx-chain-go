package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/ElrondNetwork/elrond-go/cmd/termui/provider"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/statusHandler/presenter"
	"github.com/ElrondNetwork/elrond-go/statusHandler/view/termuic"
	"github.com/urfave/cli"
)

type config struct {
	address  string
	logLevel string
	interval int
	useWss   bool
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

	// logLevel defines the logger levels and patterns
	logLevel = cli.StringFlag{
		Name:        "log-level",
		Usage:       "This flag specifies the logger level",
		Value:       "*:" + logger.LogInfo.String(),
		Destination: &argsConfig.logLevel,
	}

	// logLevel defines the logger levels and patterns
	fetchInterval = cli.IntFlag{
		Name:        "interval",
		Usage:       "This flag specifies the duration in seconds until new data is fetched from the node",
		Value:       2,
		Destination: &argsConfig.interval,
	}

	//useWss is used when the user require connection through wss
	useWss = cli.BoolFlag{
		Name:        "use-wss",
		Usage:       "Will use wss instead of ws when creating the web socket",
		Destination: &argsConfig.useWss,
	}
	argsConfig = &config{}

	log    = logger.GetOrCreate("termui")
	cliApp *cli.App
)

func main() {
	initCliFlags()

	cliApp.Action = func(c *cli.Context) error {
		return startTermuiViewer()
	}

	err := cliApp.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startTermuiViewer() error {
	nodeAddress := argsConfig.address
	logLevel := argsConfig.logLevel
	fetchInterval := argsConfig.interval

	presenterStatusHandler := presenter.NewPresenterStatusHandler()
	statusMetricsProvider, err := provider.NewStatusMetricsProvider(presenterStatusHandler, nodeAddress, fetchInterval)
	if err != nil {
		return err
	}

	termuiConsole, err := termuic.NewTermuiConsole(presenterStatusHandler)
	if err != nil {
		return err
	}

	statusMetricsProvider.StartUpdatingData()

	err = provider.InitLogHandler(nodeAddress, logLevel, argsConfig.useWss)
	if err != nil {
		return err
	}

	err = termuiConsole.Start()
	if err != nil {
		return err
	}

	provider.StartListeningOnWebSocket(presenterStatusHandler)

	waitForUserToTerminateApp()

	return nil
}

func initCliFlags() {
	cliApp = cli.NewApp()
	cli.AppHelpTemplate = nodeHelpTemplate
	cliApp.Name = "Elrond Terminal UI App"
	cliApp.Version = fmt.Sprintf("%s/%s/%s-%s", "1.0.0", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	cliApp.Usage = "Terminal UI application used to display metrics from the node"
	cliApp.Flags = []cli.Flag{
		address,
		logLevel,
		fetchInterval,
		useWss,
	}
	cliApp.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}
}

func waitForUserToTerminateApp() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	log.Info("terminating terminal ui app at user's signal...")

	provider.StopWebSocket()
}
