package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/config"
	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/databasereader"
	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/dataindexer"
	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/elastic"
	nodeConfigPackage "github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/urfave/cli"
)

type flags struct {
	dbPathWithChainID  string
	configFilePath     string
	nodeConfigFilePath string
	nodesSetupFilePath string
	numShards          int
	timeout            int
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
	dbPathWithChainIDFlag = cli.StringFlag{
		Name:        "db-path",
		Usage:       "This string flag specifies the path for the database directory, the chain ID directory",
		Value:       "../../db",
		Destination: &flagsValues.dbPathWithChainID,
	}

	// nodeConfigFilePathFlag defines a flag which holds the configuration file path
	nodeConfigFilePathFlag = cli.StringFlag{
		Name:        "node-config",
		Usage:       "This string file specifies the `filepath` for the node's toml configuration file",
		Value:       "../node/config/config.toml",
		Destination: &flagsValues.nodeConfigFilePath,
	}

	// nodesSetupFilePathFlag defines a flag which holds the nodes setup json file path
	nodesSetupFilePathFlag = cli.StringFlag{
		Name:        "nodes-setup",
		Usage:       "This string file specifies the `filepath` for the node's nodes setup json configuration file",
		Value:       "../node/config/nodesSetup.json",
		Destination: &flagsValues.nodesSetupFilePath,
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

	log                      = logger.GetOrCreate("storer2elastic")
	cliApp                   *cli.App
	marshalizer              marshal.Marshalizer
	hasher                   hashing.Hasher
	shardCoordinator         sharding.Coordinator
	nodeConfig               nodeConfigPackage.Config
	addressPubKeyConverter   core.PubkeyConverter
	validatorPubKeyConverter core.PubkeyConverter
)

func main() {
	initCliFlags()
	marshalizer = &marshal.GogoProtoMarshalizer{}
	hasher = &blake2b.Blake2b{}
	addressPubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32)
	validatorPubKeyConverter, _ = pubkeyConverter.NewHexPubkeyConverter(96)

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
	cliApp.Name = "Elrond storer to Elastic Search App"
	cliApp.Version = fmt.Sprintf("%s/%s/%s-%s", "1.0.0", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	cliApp.Usage = "Elrond storer2elastic application is used to index all data inside storers to elastic search"
	cliApp.Flags = []cli.Flag{
		dbPathWithChainIDFlag,
		configFilePathFlag,
		nodeConfigFilePathFlag,
		nodesSetupFilePathFlag,
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
	log.Info("storer2elastic application started", "version", cliApp.Version)

	var err error
	shardCoordinator, err = sharding.NewMultiShardCoordinator(uint32(flagsValues.numShards), 0)
	if err != nil {
		return err
	}

	configuration := &config.Config{}
	err = core.LoadTomlFile(configuration, flagsValues.configFilePath)
	if err != nil {
		return err
	}
	if ctx.IsSet(dbPathWithChainIDFlag.Name) {
		configuration.General.DBPathWithChainID = ctx.GlobalString(dbPathWithChainIDFlag.Name)
	}
	if ctx.IsSet(timeoutFlag.Name) {
		configuration.General.Timeout = ctx.GlobalInt(timeoutFlag.Name)
	}
	if ctx.IsSet(nodeConfigFilePathFlag.Name) {
		configuration.General.NodeConfigFilePath = ctx.GlobalString(nodeConfigFilePathFlag.Name)
	}

	err = core.LoadTomlFile(&nodeConfig, configuration.General.NodeConfigFilePath)
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

	elasticIndexer, err := elasticConnectorFactory.Create()
	if err != nil {
		return fmt.Errorf("error connecting to elastic: %w", err)
	}

	// TODO: maybe use custom configs from node config instead of a general configuration
	generalDBConfig := config.DBConfig{
		Type:              string(storageUnit.LvlDBSerial),
		BatchDelaySeconds: 2,
		MaxBatchSize:      30000,
		MaxOpenFiles:      20,
	}
	persisterFactory := factory.NewPersisterFactory(nodeConfigPackage.DBConfig(generalDBConfig))
	dbReader, err := databasereader.NewDatabaseReader(configuration.General.DBPathWithChainID, marshalizer, persisterFactory)
	if err != nil {
		return err
	}
	genesisNodesConfig, err := sharding.NewNodesSetup(
		flagsValues.nodesSetupFilePath,
		addressPubKeyConverter,
		validatorPubKeyConverter,
	)
	if err != nil {
		return err
	}

	dataIndexerArgs := dataindexer.Args{
		ElasticIndexer:    elasticIndexer,
		DatabaseReader:    dbReader,
		ShardCoordinator:  shardCoordinator,
		Marshalizer:       marshalizer,
		Hasher:            hasher,
		GenesisNodesSetup: genesisNodesConfig,
	}
	dataIndexer, err := dataindexer.New(dataIndexerArgs)
	if err != nil {
		return err
	}

	err = dataIndexer.Index(configuration.General.Timeout)
	if err != nil {
		return err
	} else {
		log.Info("indexing finished")
	}

	waitForUserToTerminateApp()

	return nil
}

func waitForUserToTerminateApp() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	log.Info("terminating storer2elastic app at user's signal...")

	// TODO : close opened DBs here

	log.Info("storer2elastic application stopped")
}
