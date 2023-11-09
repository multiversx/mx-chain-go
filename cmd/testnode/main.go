package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	factoryMarshalizer "github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/secp256k1"
	secp256k1SinglerSig "github.com/multiversx/mx-chain-crypto-go/signing/secp256k1/singlesig"
	"github.com/multiversx/mx-chain-go/cmd/node/factory"
	"github.com/multiversx/mx-chain-go/cmd/seednode/api"
	"github.com/multiversx/mx-chain-go/cmd/testnode/components"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/disabled"
	"github.com/multiversx/mx-chain-go/facade"
	cryptoFactory "github.com/multiversx/mx-chain-go/factory/crypto"
	"github.com/multiversx/mx-chain-go/p2p"
	p2pConfig "github.com/multiversx/mx-chain-go/p2p/config"
	p2pFactory "github.com/multiversx/mx-chain-go/p2p/factory"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-logger-go/file"
	"github.com/urfave/cli"
)

const (
	defaultLogsPath     = "logs"
	logFilePrefix       = "multiversx-seed"
	filePathPlaceholder = "[path]"
)

var (
	seedNodeHelpTemplate = `NAME:
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
	// port defines a flag for setting the port on which the node will listen for connections
	port = cli.StringFlag{
		Name: "port",
		Usage: "The `[p2p port]` number on which the application will start. Can use single values such as " +
			"`0, 10230, 15670` or range of ports such as `5000-10000`",
		Value: "10000",
	}
	// restApiInterfaceFlag defines a flag for the interface on which the rest API will try to bind with
	restApiInterfaceFlag = cli.StringFlag{
		Name: "rest-api-interface",
		Usage: "The interface `address and port` to which the REST API will attempt to bind. " +
			"To bind to all available interfaces, set this flag to :8080. If set to `off` then the API won't be available",
		Value: facade.DefaultRestInterface,
	}
	// logLevel defines the logger level
	logLevel = cli.StringFlag{
		Name: "log-level",
		Usage: "This flag specifies the logger `level(s)`. It can contain multiple comma-separated value. For example" +
			", if set to *:INFO the logs for all packages will have the INFO level. However, if set to *:INFO,api:DEBUG" +
			" the logs for all packages will have the INFO level, excepting the api package which will receive a DEBUG" +
			" log level.",
		Value: "*:" + logger.LogInfo.String(),
	}
	// logFile is used when the log output needs to be logged in a file
	logSaveFile = cli.BoolFlag{
		Name:  "log-save",
		Usage: "Boolean option for enabling log saving. If set, it will automatically save all the logs into a file.",
	}
	// configurationFile defines a flag for the path to the main toml configuration file
	configurationFile = cli.StringFlag{
		Name: "config",
		Usage: "The `" + filePathPlaceholder + "` for the main configuration file. This TOML file contain the main " +
			"configurations such as the marshalizer type",
		Value: "./config/config.toml",
	}
	// p2pKeyPemFile defines the flag for the path to the key pem file used for p2p signing
	p2pKeyPemFile = cli.StringFlag{
		Name:  "p2p-key-pem-file",
		Usage: "The `filepath` for the PEM file which contains the secret keys for the p2p key. If this is not specified a new key will be generated (internally) by default.",
		Value: "./config/p2pKey.pem",
	}

	p2pConfigurationFile = "./config/p2p.toml"
)

var log = logger.GetOrCreate("main")

func main() {
	app := cli.NewApp()
	cli.AppHelpTemplate = seedNodeHelpTemplate
	app.Name = "Test node CLI App"
	app.Usage = "This is the entry point for starting a new test node"
	app.Flags = []cli.Flag{
		port,
		restApiInterfaceFlag,
		logLevel,
		logSaveFile,
		configurationFile,
		p2pKeyPemFile,
	}
	app.Version = "v0.0.1"
	app.Authors = []cli.Author{
		{
			Name:  "The MultiversX Team",
			Email: "contact@multiversx.com",
		},
	}

	app.Action = func(c *cli.Context) error {
		return startNode(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startNode(ctx *cli.Context) error {
	var err error

	logLevelFlagValue := ctx.GlobalString(logLevel.Name)
	err = logger.SetLogLevel(logLevelFlagValue)
	if err != nil {
		return err
	}

	configurationFileName := ctx.GlobalString(configurationFile.Name)
	generalConfig, err := loadMainConfig(configurationFileName)
	if err != nil {
		return err
	}

	internalMarshalizer, err := factoryMarshalizer.NewMarshalizer(generalConfig.Marshalizer.Type)
	if err != nil {
		return fmt.Errorf("error creating marshalizer (internal): %s", err.Error())
	}

	withLogFile := ctx.GlobalBool(logSaveFile.Name)
	var fileLogging factory.FileLoggingHandler
	if withLogFile {
		workingDir := getWorkingDir(log)
		args := file.ArgsFileLogging{
			WorkingDir:      workingDir,
			DefaultLogsPath: defaultLogsPath,
			LogFilePrefix:   logFilePrefix,
		}
		fileLogging, err = file.NewFileLogging(args)
		if err != nil {
			return fmt.Errorf("%w creating a log file", err)
		}

		timeLogLifeSpan := time.Second * time.Duration(generalConfig.Logs.LogFileLifeSpanInSec)
		sizeLogLifeSpanInMB := uint64(generalConfig.Logs.LogFileLifeSpanInMB)
		err = fileLogging.ChangeFileLifeSpan(timeLogLifeSpan, sizeLogLifeSpanInMB)
		if err != nil {
			return err
		}
	}

	startRestServices(ctx, internalMarshalizer)

	log.Info("starting test node...")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	p2pCfg, err := common.LoadP2PConfig(p2pConfigurationFile)
	if err != nil {
		return err
	}
	log.Info("initialized with p2p config",
		"filename", p2pConfigurationFile,
	)
	if ctx.IsSet(port.Name) {
		p2pCfg.Node.Port = ctx.GlobalString(port.Name)
	}

	err = checkExpectedPeerCount(*p2pCfg)
	if err != nil {
		return err
	}

	p2pKeyPemFileName := ctx.GlobalString(p2pKeyPemFile.Name)
	messenger, err := createNode(*p2pCfg, internalMarshalizer, p2pKeyPemFileName)
	if err != nil {
		return err
	}

	err = messenger.Bootstrap()
	if err != nil {
		return err
	}

	log.Info("application is now running...")
	err = mainLoop(messenger, sigs)
	if err != nil {
		return err
	}

	log.Debug("closing test node")
	if !check.IfNil(fileLogging) {
		err = fileLogging.Close()
		log.LogIfError(err)
	}

	return nil
}

func mainLoop(messenger p2p.Messenger, stop chan os.Signal) error {
	durationDisplay := time.Second * 5
	timerDisplay := time.NewTimer(durationDisplay)
	defer timerDisplay.Stop()

	durationSend := time.Second
	timerSend := time.NewTimer(durationSend)
	defer timerSend.Stop()

	testTopic := "test"
	err := messenger.CreateTopic(testTopic, true)
	if err != nil {
		return err
	}

	instance := components.NewInterceptor()
	err = messenger.RegisterMessageProcessor(testTopic, "", instance)
	if err != nil {
		return err
	}

	message := []byte("ping")

	log.Info("Statistics",
		"num connections", len(messenger.ConnectedAddresses()),
		"num received messages", instance.GetNumMessages())
	for {
		select {
		case <-stop:
			log.Info("terminating at user's signal...")
			return nil
		case <-timerSend.C:
			messenger.Broadcast(testTopic, message)
			timerSend.Reset(durationSend)
		case <-timerDisplay.C:
			log.Info("Statistics",
				"num connections", len(messenger.ConnectedAddresses()),
				"num received messages", instance.GetNumMessages())

			timerDisplay.Reset(durationDisplay)
		}
	}
}

func loadMainConfig(filepath string) (*config.Config, error) {
	cfg := &config.Config{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func createNode(
	p2pConfig p2pConfig.P2PConfig,
	marshalizer marshal.Marshalizer,
	p2pKeyFileName string,
) (p2p.Messenger, error) {
	p2pSingleSigner := &secp256k1SinglerSig.Secp256k1Signer{}
	p2pKeyGen := signing.NewKeyGenerator(secp256k1.NewSecp256k1())

	p2pKey, _, err := cryptoFactory.CreateP2pKeyPair(p2pKeyFileName, p2pKeyGen, log)
	if err != nil {
		return nil, err
	}

	arg := p2pFactory.ArgsNetworkMessenger{
		Marshalizer:           marshalizer,
		P2pConfig:             p2pConfig,
		SyncTimer:             &p2pFactory.LocalSyncTimer{},
		PreferredPeersHolder:  disabled.NewPreferredPeersHolder(),
		NodeOperationMode:     p2p.NormalOperation,
		PeersRatingHandler:    disabled.NewDisabledPeersRatingHandler(),
		ConnectionWatcherType: "disabled",
		P2pPrivateKey:         p2pKey,
		P2pSingleSigner:       p2pSingleSigner,
		P2pKeyGenerator:       p2pKeyGen,
	}

	return p2pFactory.NewNetworkMessenger(arg)
}

func getWorkingDir(log logger.Logger) string {
	workingDir, err := os.Getwd()
	if err != nil {
		log.LogIfError(err)
		workingDir = ""
	}

	log.Trace("working directory", "path", workingDir)

	return workingDir
}

func checkExpectedPeerCount(p2pConfig p2pConfig.P2PConfig) error {
	maxExpectedPeerCount := p2pConfig.Node.MaximumExpectedPeerCount

	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return fmt.Errorf("%w while getting RLimits", err)
	}

	log.Info("file limits",
		"current", rLimit.Cur,
		"max", rLimit.Max,
		"expected", maxExpectedPeerCount,
	)

	if maxExpectedPeerCount > rLimit.Cur {
		return fmt.Errorf("provided maxExpectedPeerCount is less than the current OS configured value")
	}

	return nil
}

func startRestServices(ctx *cli.Context, marshalizer marshal.Marshalizer) {
	restApiInterface := ctx.GlobalString(restApiInterfaceFlag.Name)
	if restApiInterface != facade.DefaultRestPortOff {
		go startGinServer(restApiInterface, marshalizer)
	} else {
		log.Info("rest api is disabled")
	}
}

func startGinServer(restApiInterface string, marshalizer marshal.Marshalizer) {
	err := api.Start(restApiInterface, marshalizer)
	if err != nil {
		log.LogIfError(err)
	}
}
