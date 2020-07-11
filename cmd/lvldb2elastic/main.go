package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/cmd/lvldb2elastic/config"
	"github.com/ElrondNetwork/elrond-go/cmd/lvldb2elastic/databasereader"
	"github.com/ElrondNetwork/elrond-go/cmd/lvldb2elastic/elastic"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/urfave/cli"
)

type flags struct {
	dbPath         string
	configFilePath string
	numShards      int
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
	// configFilePathFlag defines a flag which holds the configuration file path
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

	// numOfShardsFlag defines the flag which holds the number of shards
	numOfShardsFlag = cli.IntFlag{
		Name:        "num-shards",
		Usage:       "This int flag specifies the number of shards",
		Value:       2,
		Destination: &flagsValues.numShards,
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
	shardCoordinator         sharding.Coordinator
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
		numOfShardsFlag,
		timeoutFlag,
	}
	cliApp.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}
}

func startLvlDb2Elastic(ctx *cli.Context) error {
	log.Info("lvldb2elastic application started", "version", cliApp.Version)

	var err error
	shardCoordinator, err = sharding.NewMultiShardCoordinator(uint32(flagsValues.numShards), 0)
	if err != nil {
		panic("cannot create shard coordinator: " + err.Error())
	}
	configuration := &config.Config{}
	err = core.LoadTomlFile(configuration, flagsValues.configFilePath)
	if err != nil {
		return err
	}
	if ctx.IsSet(dbPathFlag.Name) {
		configuration.General.DBPath = ctx.GlobalString(dbPathFlag.Name)
	}
	if ctx.IsSet(timeoutFlag.Name) {
		configuration.General.Timeout = ctx.GlobalInt(timeoutFlag.Name)
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

	_, _ = elasticConnectorFactory.Create()
	//if err != nil {
	//	return fmt.Errorf("error connecting to elastic: %w", err)
	//}

	dbReader, err := databasereader.NewDatabaseReader(configuration.General.DBPath, marshalizer)
	if err != nil {
		return err
	}

	records, err := dbReader.GetDatabaseInfos()
	if err != nil {
		return err
	}

	for _, db := range records {
		log.Info("database info", "epoch", db.Epoch, "shard", db.Shard)
	}

	hdrs, err := dbReader.GetHeaders(records[0])
	if err != nil {
		return err
	}
	for _, hdr := range hdrs {
		fmt.Println(hdr.GetEpoch())
	}

	fmt.Println("mini blocks")
	mbs, err := dbReader.GetMiniBlocks(records[0])
	if err != nil {
		return err
	}
	for _, mb := range mbs {
		fmt.Println(mb.Type)
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
