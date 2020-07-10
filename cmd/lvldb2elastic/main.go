package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/cmd/lvldb2elastic/config"
	"github.com/ElrondNetwork/elrond-go/cmd/lvldb2elastic/elastic"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/urfave/cli"
)

type flags struct {
	dbPath         string
	configFilePath string
	timeout        int
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
	configFilePathFlag = cli.StringFlag{
		Name:        "config",
		Usage:       "This string file specifies the `filepath` for the toml configuration file",
		Value:       "./config.toml",
		Destination: &flagsValues.configFilePath,
	}

	// address defines a flag for setting the address and port on which the node will listen for connections
	dbPathFlag = cli.StringFlag{
		Name:        "db-path",
		Usage:       "This string flag specifies the path for the database directory",
		Value:       "../../db",
		Destination: &flagsValues.dbPath,
	}

	// logLevelPatterns defines the logger levels and patterns
	timeoutFlag = cli.IntFlag{
		Name:        "timeout",
		Usage:       "This integer flag specifies the maximum duration to wait for indexing the entire results",
		Value:       200,
		Destination: &flagsValues.timeout,
	}

	flagsValues = &flags{}

	log                      = logger.GetOrCreate("lvldb2elastic")
	cliApp                   *cli.App
	marshalizer              marshal.Marshalizer
	hasher                   hashing.Hasher
	addressPubKeyConverter   core.PubkeyConverter
	validatorPubKeyConverter core.PubkeyConverter
)

func main() {
	initCliFlags()
	marshalizer = &marshal.GogoProtoMarshalizer{}
	hasher = &blake2b.Blake2b{}
	addressPubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32)
	validatorPubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(96)

	cliApp.Action = func(c *cli.Context) error {
		return startLvlDb2Elastic(c)
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
	cliApp.Name = "Elrond LevelDB to Elastic Search App"
	cliApp.Version = fmt.Sprintf("%s/%s/%s-%s", "1.0.0", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	cliApp.Usage = "Elrond lvldb2elastic application is used to index all data inside serial level DBs to elastic search"
	cliApp.Flags = []cli.Flag{
		dbPathFlag,
		configFilePathFlag,
		timeoutFlag,
	}
	cliApp.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}
}

func startLvlDb2Elastic(_ *cli.Context) error {
	log.Info("lvldb2elastic application started", "version", cliApp.Version)

	configuration := &config.Config{}
	err := core.LoadTomlFile(configuration, flagsValues.configFilePath)
	if err != nil {
		return err
	}

	elasticConnectorFactoryArgs := elastic.ConnectorFactoryArgs{
		ElasticConfig:            configuration.ElasticSearch,
		Marshalizer:              marshalizer,
		Hasher:                   hasher,
		ValidatorPubKeyConverter: validatorPubKeyConverter,
		AddressPubKeyConverter:   addressPubKeyConverter,
	}

	elasticConnectorFactory, err := elastic.NewConnectorFactory(elasticConnectorFactoryArgs)
	if err != nil {
		return err
	}

	_, err = elasticConnectorFactory.Create()
	if err != nil {
		return fmt.Errorf("error connecting to elastic: %w", err)
	}

	waitForUserToTerminateApp()

	return nil
}

func waitForUserToTerminateApp() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	log.Info("terminating lvldb2elastic app at user's signal...")

	// TODO : close opened DBs here

	log.Info("lvldb2elastic application stopped")
}
