package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/appStatusPolling"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/core/serviceContainer"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/core/statistics/machine"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/facade"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/ntp"
	factoryVM "github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	factoryViews "github.com/ElrondNetwork/elrond-go/statusHandler/factory"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm/iele/elrond/node/endpoint"
	"github.com/google/gops/agent"
	"github.com/urfave/cli"
)

const (
	defaultLogPath      = "logs"
	defaultStatsPath    = "stats"
	defaultDBPath       = "db"
	defaultEpochString  = "Epoch"
	defaultShardString  = "Shard"
	metachainShardName  = "metachain"
	milisecondsInSecond = 1000
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

	// genesisFile defines a flag for the path of the bootstrapping file.
	genesisFile = cli.StringFlag{
		Name:  "genesis-file",
		Usage: "The node will extract bootstrapping info from the genesis.json",
		Value: "./config/genesis.json",
	}
	// nodesFile defines a flag for the path of the initial nodes file.
	nodesFile = cli.StringFlag{
		Name:  "nodesSetup-file",
		Usage: "The node will extract initial nodes info from the nodesSetup.json",
		Value: "./config/nodesSetup.json",
	}
	// txSignSk defines a flag for the path of the single sign private key used when starting the node
	txSignSk = cli.StringFlag{
		Name:  "tx-sign-sk",
		Usage: "Private key that the node will load on startup and will sign transactions - temporary until we have a wallet that can do that",
		Value: "",
	}
	// sk defines a flag for the path of the multi sign private key used when starting the node
	sk = cli.StringFlag{
		Name:  "sk",
		Usage: "Private key that the node will load on startup and will sign blocks",
		Value: "",
	}
	// configurationFile defines a flag for the path to the main toml configuration file
	configurationFile = cli.StringFlag{
		Name:  "config",
		Usage: "The main configuration file to load",
		Value: "./config/config.toml",
	}
	// p2pConfigurationFile defines a flag for the path to the toml file containing P2P configuration
	p2pConfigurationFile = cli.StringFlag{
		Name:  "p2pconfig",
		Usage: "The configuration file for P2P",
		Value: "./config/p2p.toml",
	}
	// p2pConfigurationFile defines a flag for the path to the toml file containing P2P configuration
	serversConfigurationFile = cli.StringFlag{
		Name:  "serversconfig",
		Usage: "The configuration file for servers confidential data",
		Value: "./config/server.toml",
	}
	// withUI defines a flag for choosing the option of starting with/without UI. If false, the node will start automatically
	withUI = cli.BoolTFlag{
		Name:  "with-ui",
		Usage: "If true, the application will be accompanied by a UI. The node will have to be manually started from the UI",
	}
	// port defines a flag for setting the port on which the node will listen for connections
	port = cli.IntFlag{
		Name:  "port",
		Usage: "Port number on which the application will start",
		Value: 0,
	}
	// profileMode defines a flag for profiling the binary
	// If enabled, it will open the pprof routes over the default gin rest webserver.
	// There are several routes that will be available for profiling (profiling can be analyzed with: go tool pprof):
	//  /debug/pprof/ (can be accessed in the browser, will list the available options)
	//  /debug/pprof/goroutine
	//  /debug/pprof/heap
	//  /debug/pprof/threadcreate
	//  /debug/pprof/block
	//  /debug/pprof/mutex
	//  /debug/pprof/profile (CPU profile)
	//  /debug/pprof/trace?seconds=5 (CPU trace) -> being a trace, can be analyzed with: go tool trace
	// Usage: go tool pprof http(s)://ip.of.the.server/debug/pprof/xxxxx
	profileMode = cli.BoolFlag{
		Name:  "profile-mode",
		Usage: "Boolean profiling mode option. If set to true, the /debug/pprof routes will be available on the node for profiling the application.",
	}
	// txSignSkIndex defines a flag that specifies the 0-th based index of the private key to be used from initialBalancesSk.pem file
	txSignSkIndex = cli.IntFlag{
		Name:  "tx-sign-sk-index",
		Usage: "Single sign private key index specifies the 0-th based index of the private key to be used from initialBalancesSk.pem file.",
		Value: 0,
	}
	// skIndex defines a flag that specifies the 0-th based index of the private key to be used from initialNodesSk.pem file
	skIndex = cli.IntFlag{
		Name:  "sk-index",
		Usage: "Private key index specifies the 0-th based index of the private key to be used from initialNodesSk.pem file.",
		Value: 0,
	}
	// gopsEn used to enable diagnosis of running go processes
	gopsEn = cli.BoolFlag{
		Name:  "gops-enable",
		Usage: "Enables gops over the process. Stack can be viewed by calling 'gops stack <pid>'",
	}
	// numOfNodes defines a flag that specifies the maximum number of nodes which will be used from the initialNodes
	numOfNodes = cli.Uint64Flag{
		Name:  "num-of-nodes",
		Usage: "Number of nodes specifies the maximum number of nodes which will be used from initialNodes list exposed in nodesSetup.json file",
		Value: math.MaxUint64,
	}
	// storageCleanup defines a flag for choosing the option of starting the node from scratch. If it is not set (false)
	// it starts from the last state stored on disk
	storageCleanup = cli.BoolFlag{
		Name:  "storage-cleanup",
		Usage: "If set the node will start from scratch, otherwise it starts from the last state stored on disk",
	}

	// restApiPort defines a flag for port on which the rest API will start on
	restApiPort = cli.StringFlag{
		Name:  "rest-api-port",
		Usage: "The port on which the rest API will start on",
		Value: "8080",
	}

	// networkID defines the version of the network. If set, will override the same parameter from config.toml
	networkID = cli.StringFlag{
		Name:  "network-id",
		Usage: "The network version, overriding the one from config.toml",
		Value: "",
	}

	// nodeDisplayName defines the friendly name used by a node in the public monitoring tools. If set, will override
	// the NodeDisplayName from config.toml
	nodeDisplayName = cli.StringFlag{
		Name:  "display-name",
		Usage: "This will represent the friendly name in the public monitoring tools. Will override the config.toml one",
		Value: "",
	}

	// usePrometheus joins the node for prometheus monitoring if set
	usePrometheus = cli.BoolFlag{
		Name:  "use-prometheus",
		Usage: "Will make the node available for prometheus and grafana monitoring",
	}

	//useLogView is used when termui interface is not needed.
	useLogView = cli.BoolFlag{
		Name:  "use-log-view",
		Usage: "will not enable the user-friendly terminal view of the node",
	}

	// initialBalancesSkPemFile defines a flag for the path to the ...
	initialBalancesSkPemFile = cli.StringFlag{
		Name:  "initialBalancesSkPemFile",
		Usage: "The file containing the secret keys which ...",
		Value: "./config/initialBalancesSk.pem",
	}

	// initialNodesSkPemFile defines a flag for the path to the ...
	initialNodesSkPemFile = cli.StringFlag{
		Name:  "initialNodesSkPemFile",
		Usage: "The file containing the secret keys which ...",
		Value: "./config/initialNodesSk.pem",
	}
	// logLevel defines the logger level
	logLevel = cli.StringFlag{
		Name:  "logLevel",
		Usage: "This flag specifies the logger level",
		Value: logger.LogInfo,
	}
	// bootstrapRoundIndex defines a flag that specifies the round index from which node should bootstrap from storage
	bootstrapRoundIndex = cli.UintFlag{
		Name:  "bootstrap-round-index",
		Usage: "Bootstrap round index specifies the round index from which node should bootstrap from storage",
		Value: math.MaxUint64,
	}
	// enableTxIndexing enables transaction indexing. There can be cases when it's too expensive to index all transactions
	//  so we provide the command line option to disable this behaviour
	enableTxIndexing = cli.BoolTFlag{
		Name:  "tx-indexing",
		Usage: "Enables transaction indexing. There can be cases when it's too expensive to index all transactions so we provide the command line option to disable this behaviour",
	}

	// workingDirectory defines a flag for the path for the working directory.
	workingDirectory = cli.StringFlag{
		Name:  "working-directory",
		Usage: "The node will store here DB, Logs and Stats",
		Value: "",
	}

	// destinationShardAsObserver defines a flag for the prefered shard to be assigned to as an observer.
	destinationShardAsObserver = cli.StringFlag{
		Name:  "destination-shard-as-observer",
		Usage: "The preferred shard as an observer",
		Value: "",
	}

	rm *statistics.ResourceMonitor
)

// dbIndexer will hold the database indexer. Defined globally so it can be initialised only in
//  certain conditions. If those conditions will not be met, it will stay as nil
var dbIndexer indexer.Indexer

// coreServiceContainer is defined globally so it can be injected with appropriate
//  params depending on the type of node we are starting
var coreServiceContainer serviceContainer.Core

// appVersion should be populated at build time using ldflags
// Usage examples:
// linux/mac:
//            go build -i -v -ldflags="-X main.appVersion=$(git describe --tags --long --dirty)"
// windows:
//            for /f %i in ('git describe --tags --long --dirty') do set VERS=%i
//            go build -i -v -ldflags="-X main.appVersion=%VERS%"
var appVersion = core.UnVersionedAppString

func main() {
	log := logger.DefaultLogger()

	app := cli.NewApp()
	cli.AppHelpTemplate = nodeHelpTemplate
	app.Name = "Elrond Node CLI App"
	app.Version = fmt.Sprintf("%s/%s/%s-%s", appVersion, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	app.Usage = "This is the entry point for starting a new Elrond node - the app will start after the genesis timestamp"
	app.Flags = []cli.Flag{
		genesisFile,
		nodesFile,
		port,
		configurationFile,
		p2pConfigurationFile,
		txSignSk,
		sk,
		profileMode,
		txSignSkIndex,
		skIndex,
		numOfNodes,
		storageCleanup,
		initialBalancesSkPemFile,
		initialNodesSkPemFile,
		gopsEn,
		serversConfigurationFile,
		networkID,
		nodeDisplayName,
		restApiPort,
		logLevel,
		usePrometheus,
		useLogView,
		bootstrapRoundIndex,
		enableTxIndexing,
		workingDirectory,
		destinationShardAsObserver,
	}
	app.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}

	//TODO: The next line should be removed when the write in batches is done
	// set the maximum allowed OS threads (not go routines) which can run in the same time (the default is 10000)
	debug.SetMaxThreads(100000)

	app.Action = func(c *cli.Context) error {
		return startNode(c, log, app.Version)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func getSuite(config *config.Config) (crypto.Suite, error) {
	switch config.Consensus.Type {
	case factory.BlsConsensusType:
		return kyber.NewSuitePairingBn256(), nil
	case factory.BnConsensusType:
		return kyber.NewBlakeSHA256Ed25519(), nil
	}

	return nil, errors.New("no consensus provided in config file")
}

func startNode(ctx *cli.Context, log *logger.Logger, version string) error {
	logLevel := ctx.GlobalString(logLevel.Name)
	log.SetLevel(logLevel)

	enableGopsIfNeeded(ctx, log)

	stop := make(chan bool, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Info(fmt.Sprintf("Starting node with version %s\n", version))
	log.Info(fmt.Sprintf("Process ID: %d\n", os.Getpid()))

	configurationFileName := ctx.GlobalString(configurationFile.Name)
	generalConfig, err := loadMainConfig(configurationFileName, log)
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Initialized with config from: %s", configurationFileName))

	p2pConfigurationFileName := ctx.GlobalString(p2pConfigurationFile.Name)
	p2pConfig, err := core.LoadP2PConfig(p2pConfigurationFileName)
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Initialized with p2p config from: %s", p2pConfigurationFileName))
	if ctx.IsSet(port.Name) {
		p2pConfig.Node.Port = ctx.GlobalInt(port.Name)
	}

	genesisConfig, err := sharding.NewGenesisConfig(ctx.GlobalString(genesisFile.Name))
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Initialized with genesis config from: %s", ctx.GlobalString(genesisFile.Name)))

	nodesConfig, err := sharding.NewNodesSetup(ctx.GlobalString(nodesFile.Name), ctx.GlobalUint64(numOfNodes.Name))
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Initialized with nodes config from: %s", ctx.GlobalString(nodesFile.Name)))

	syncer := ntp.NewSyncTime(generalConfig.NTPConfig, time.Hour, nil)
	go syncer.StartSync()

	log.Info(fmt.Sprintf("NTP average clock offset: %s", syncer.ClockOffset()))

	//TODO: The next 5 lines should be deleted when we are done testing from a precalculated (not hard coded) timestamp
	if nodesConfig.StartTime == 0 {
		time.Sleep(1000 * time.Millisecond)
		ntpTime := syncer.CurrentTime()
		nodesConfig.StartTime = (ntpTime.Unix()/60 + 1) * 60
	}

	startTime := time.Unix(nodesConfig.StartTime, 0)

	log.Info(fmt.Sprintf("Start time formatted: %s", startTime.Format("Mon Jan 2 15:04:05 MST 2006")))
	log.Info(fmt.Sprintf("Start time in seconds: %d", startTime.Unix()))

	suite, err := getSuite(generalConfig)
	if err != nil {
		return err
	}

	initialNodesSkPemFileName := ctx.GlobalString(initialNodesSkPemFile.Name)
	keyGen, privKey, pubKey, err := factory.GetSigningParams(
		ctx,
		log,
		sk.Name,
		skIndex.Name,
		initialNodesSkPemFileName,
		suite)
	if err != nil {
		return err
	}
	log.Info("Starting with public key: " + factory.GetPkEncoded(pubKey))

	if ctx.IsSet(destinationShardAsObserver.Name) {
		generalConfig.GeneralSettings.DestinationShardAsObserver = ctx.GlobalString(destinationShardAsObserver.Name)
	}

	if ctx.IsSet(networkID.Name) {
		generalConfig.GeneralSettings.NetworkID = ctx.GlobalString(networkID.Name)
	}

	if ctx.IsSet(nodeDisplayName.Name) {
		generalConfig.GeneralSettings.NodeDisplayName = ctx.GlobalString(nodeDisplayName.Name)
	}

	shardCoordinator, nodeType, err := createShardCoordinator(nodesConfig, pubKey, generalConfig.GeneralSettings, log)
	if err != nil {
		return err
	}

	var workingDir = ""
	if ctx.IsSet(workingDirectory.Name) {
		workingDir = ctx.GlobalString(workingDirectory.Name)
	} else {
		workingDir, err = os.Getwd()
		if err != nil {
			log.LogIfError(err)
			workingDir = ""
		}
	}

	var shardId = "metachain"
	if shardCoordinator.SelfId() != sharding.MetachainShardId {
		shardId = fmt.Sprintf("%d", shardCoordinator.SelfId())
	}

	uniqueDBFolder := filepath.Join(
		workingDir,
		defaultDBPath,
		fmt.Sprintf("%s_%d", defaultEpochString, 0),
		fmt.Sprintf("%s_%s", defaultShardString, shardId))

	storageCleanup := ctx.GlobalBool(storageCleanup.Name)
	if storageCleanup {
		err = os.RemoveAll(uniqueDBFolder)
		if err != nil {
			return err
		}
	}

	logDirectory := filepath.Join(workingDir, defaultLogPath)

	err = os.MkdirAll(logDirectory, os.ModePerm)
	if err != nil {
		return err
	}

	coreArgs := factory.NewCoreComponentsFactoryArgs(generalConfig, uniqueDBFolder)
	coreComponents, err := factory.CoreComponentsFactory(coreArgs)
	if err != nil {
		return err
	}

	stateArgs := factory.NewStateComponentsFactoryArgs(generalConfig, genesisConfig, shardCoordinator, coreComponents)
	stateComponents, err := factory.StateComponentsFactory(stateArgs)
	if err != nil {
		return err
	}

	err = initLogFileAndStatsMonitor(generalConfig, pubKey, log, workingDir)
	if err != nil {
		return err
	}

	var appStatusHandlers []core.AppStatusHandler
	var views []factoryViews.Viewer

	prometheusJoinUrl, usePrometheusBool := getPrometheusJoinURLIfAvailable(ctx)
	if usePrometheusBool {
		prometheusStatusHandler := statusHandler.NewPrometheusStatusHandler()
		appStatusHandlers = append(appStatusHandlers, prometheusStatusHandler)
	}

	presenterStatusHandler := factory.CreateStatusHandlerPresenter()

	useTermui := !ctx.GlobalBool(useLogView.Name)
	if useTermui {

		views, err = factory.CreateViews(presenterStatusHandler)
		if err != nil {
			return err
		}

		writer, ok := presenterStatusHandler.(io.Writer)
		if ok {
			err = log.ChangePrinterHookWriter(writer)
			if err != nil {
				return err
			}
		}

		appStatusHandler, ok := presenterStatusHandler.(core.AppStatusHandler)
		if ok {
			appStatusHandlers = append(appStatusHandlers, appStatusHandler)
		}
	}

	if views == nil {
		log.Warn("No views for current node")
	}

	if len(appStatusHandlers) > 0 {
		coreComponents.StatusHandler, err = statusHandler.NewAppStatusFacadeWithHandlers(appStatusHandlers...)
		if err != nil {
			log.Warn("Cannot init AppStatusFacade", err)
		}
	} else {
		coreComponents.StatusHandler = statusHandler.NewNilStatusHandler()
		log.Info("No AppStatusHandler used. Started with NilStatusHandler")
	}

	initMetrics(coreComponents.StatusHandler, pubKey, nodeType, shardCoordinator, nodesConfig, version)

	dataArgs := factory.NewDataComponentsFactoryArgs(generalConfig, shardCoordinator, coreComponents, uniqueDBFolder)
	dataComponents, err := factory.DataComponentsFactory(dataArgs)
	if err != nil {
		return err
	}

	cryptoArgs := factory.NewCryptoComponentsFactoryArgs(ctx, generalConfig, nodesConfig, shardCoordinator, keyGen,
		privKey, log, initialBalancesSkPemFile.Name, txSignSk.Name, txSignSkIndex.Name)
	cryptoComponents, err := factory.CryptoComponentsFactory(cryptoArgs)
	if err != nil {
		return err
	}

	output := fmt.Sprintf("%s:%s\n%s:%s\n%s:%s\n%s:%v\n%s:%s\n%s:%v\n",
		"PkBlockSign", factory.GetPkEncoded(pubKey),
		"PkAccount", factory.GetPkEncoded(cryptoComponents.TxSignPubKey),
		"ShardId", shardId,
		"TotalShards", shardCoordinator.NumberOfShards(),
		"AppVersion", version,
		"GenesisTimeStamp", startTime.Unix(),
	)

	txSignPk := factory.GetPkEncoded(cryptoComponents.TxSignPubKey)
	coreComponents.StatusHandler.SetStringValue(core.MetricPublicKeyTxSign, txSignPk)

	err = ioutil.WriteFile(filepath.Join(logDirectory, "session.info"), []byte(output), os.ModePerm)
	log.LogIfError(err)

	networkComponents, err := factory.NetworkComponentsFactory(p2pConfig, log, coreComponents)
	if err != nil {
		return err
	}

	tpsBenchmark, err := statistics.NewTPSBenchmark(shardCoordinator.NumberOfShards(), nodesConfig.RoundDuration/1000)
	if err != nil {
		return err
	}

	if generalConfig.Explorer.Enabled {
		serversConfigurationFileName := ctx.GlobalString(serversConfigurationFile.Name)
		dbIndexer, err = CreateElasticIndexer(
			ctx,
			serversConfigurationFileName,
			generalConfig.Explorer.IndexerURL,
			shardCoordinator,
			coreComponents.Marshalizer,
			coreComponents.Hasher,
			log)
		if err != nil {
			return err
		}

		err = setServiceContainer(shardCoordinator, tpsBenchmark)
		if err != nil {
			return err
		}
	}

	processArgs := factory.NewProcessComponentsFactoryArgs(genesisConfig, nodesConfig, syncer, shardCoordinator,
		dataComponents, coreComponents, cryptoComponents, stateComponents, networkComponents, coreServiceContainer)
	processComponents, err := factory.ProcessComponentsFactory(processArgs)
	if err != nil {
		return err
	}

	currentNode, err := createNode(
		generalConfig,
		nodesConfig,
		syncer,
		keyGen,
		privKey,
		pubKey,
		shardCoordinator,
		coreComponents,
		stateComponents,
		dataComponents,
		cryptoComponents,
		processComponents,
		networkComponents,
		uint64(ctx.GlobalUint(bootstrapRoundIndex.Name)),
		version,
	)
	if err != nil {
		return err
	}

	vmAccountsDB, err := hooks.NewVMAccountsDB(
		stateComponents.AccountsAdapter,
		stateComponents.AddressConverter,
	)
	if err != nil {
		return err
	}

	apiResolver, err := createApiResolver(vmAccountsDB)
	if err != nil {
		return err
	}

	err = startStatusPolling(
		currentNode.GetAppStatusHandler(),
		generalConfig.GeneralSettings.StatusPollingIntervalSec,
		networkComponents,
		processComponents,
	)
	if err != nil {
		return err
	}

	updateMachineStatisticsDurationSec := 1
	err = startMachineStatisticsPolling(coreComponents.StatusHandler, updateMachineStatisticsDurationSec)
	if err != nil {
		return err
	}

	restAPIServerDebugMode := !useTermui
	ef := facade.NewElrondNodeFacade(currentNode, apiResolver, restAPIServerDebugMode)

	efConfig := &config.FacadeConfig{
		RestApiPort:       ctx.GlobalString(restApiPort.Name),
		PprofEnabled:      ctx.GlobalBool(profileMode.Name),
		Prometheus:        usePrometheusBool,
		PrometheusJoinURL: prometheusJoinUrl,
		PrometheusJobName: generalConfig.GeneralSettings.NetworkID,
	}

	ef.SetLogger(log)
	ef.SetSyncer(syncer)
	ef.SetTpsBenchmark(tpsBenchmark)
	ef.SetConfig(efConfig)

	wg := sync.WaitGroup{}
	go ef.StartBackgroundServices(&wg)
	wg.Wait()

	if !ctx.Bool(withUI.Name) {
		log.Info("Bootstrapping node....")
		err = ef.StartNode()
		if err != nil {
			log.Error("starting node failed", err.Error())
			return err
		}
	}

	go func() {
		<-sigs
		log.Info("terminating at user's signal...")
		stop <- true
	}()

	log.Info("Application is now running...")
	<-stop

	if rm != nil {
		err = rm.Close()
		log.LogIfError(err)
	}
	return nil
}

func initMetrics(
	appStatusHandler core.AppStatusHandler,
	pubKey crypto.PublicKey,
	nodeType core.NodeType,
	shardCoordinator sharding.Coordinator,
	nodesConfig *sharding.NodesSetup,
	version string,
) {
	shardId := uint64(shardCoordinator.SelfId())
	roundDuration := nodesConfig.RoundDuration
	isSyncing := uint64(1)
	initUint := uint64(0)
	initString := ""

	appStatusHandler.SetStringValue(core.MetricPublicKeyBlockSign, factory.GetPkEncoded(pubKey))
	appStatusHandler.SetUInt64Value(core.MetricShardId, shardId)
	appStatusHandler.SetStringValue(core.MetricNodeType, string(nodeType))
	appStatusHandler.SetUInt64Value(core.MetricRoundTime, roundDuration/milisecondsInSecond)
	appStatusHandler.SetStringValue(core.MetricAppVersion, version)
	appStatusHandler.SetUInt64Value(core.MetricCountConsensus, initUint)
	appStatusHandler.SetUInt64Value(core.MetricCountLeader, initUint)
	appStatusHandler.SetUInt64Value(core.MetricCountAcceptedBlocks, initUint)
	appStatusHandler.SetUInt64Value(core.MetricNumTxInBlock, initUint)
	appStatusHandler.SetUInt64Value(core.MetricNumMiniBlocks, initUint)
	appStatusHandler.SetStringValue(core.MetricConsensusState, initString)
	appStatusHandler.SetStringValue(core.MetricConsensusRoundState, initString)
	appStatusHandler.SetStringValue(core.MetricCrossCheckBlockHeight, initString)
	appStatusHandler.SetUInt64Value(core.MetricIsSyncing, isSyncing)
	appStatusHandler.SetStringValue(core.MetricCurrentBlockHash, initString)
	appStatusHandler.SetUInt64Value(core.MetricNumProcessedTxs, initUint)
	appStatusHandler.SetUInt64Value(core.MetricCurrentRoundTimestamp, initUint)
	appStatusHandler.SetUInt64Value(core.MetricHeaderSize, initUint)
	appStatusHandler.SetUInt64Value(core.MetricMiniBlocksSize, initUint)
	appStatusHandler.SetUInt64Value(core.MetricNumShardHeadersFromPool, initUint)
	appStatusHandler.SetUInt64Value(core.MetricNumShardHeadersProcessed, initUint)
}

func startStatusPolling(
	ash core.AppStatusHandler,
	pollingInterval int,
	networkComponents *factory.Network,
	processComponents *factory.Process,
) error {

	if ash == nil {
		return errors.New("nil AppStatusHandler")
	}

	appStatusPollingHandler, err := appStatusPolling.NewAppStatusPolling(ash, pollingInterval)
	if err != nil {
		return errors.New("cannot init AppStatusPolling")
	}

	err = registerPollConnectedPeers(appStatusPollingHandler, networkComponents)
	if err != nil {
		return err
	}

	err = registerPollProbableHighestNonce(appStatusPollingHandler, processComponents)
	if err != nil {
		return err
	}

	appStatusPollingHandler.Poll()

	return nil
}

func registerPollConnectedPeers(
	appStatusPollingHandler *appStatusPolling.AppStatusPolling,
	networkComponents *factory.Network,
) error {

	numOfConnectedPeersHandlerFunc := func(appStatusHandler core.AppStatusHandler) {
		numOfConnectedPeers := uint64(len(networkComponents.NetMessenger.ConnectedAddresses()))
		appStatusHandler.SetUInt64Value(core.MetricNumConnectedPeers, numOfConnectedPeers)
	}

	err := appStatusPollingHandler.RegisterPollingFunc(numOfConnectedPeersHandlerFunc)
	if err != nil {
		return errors.New("cannot register handler func for num of connected peers")
	}

	return nil
}

//TODO: move out of main
func registerPollProbableHighestNonce(
	appStatusPollingHandler *appStatusPolling.AppStatusPolling,
	processComponents *factory.Process,
) error {

	probableHighestNonceHandlerFunc := func(appStatusHandler core.AppStatusHandler) {
		probableHigherNonce := processComponents.ForkDetector.ProbableHighestNonce()
		appStatusHandler.SetUInt64Value(core.MetricProbableHighestNonce, probableHigherNonce)
	}

	err := appStatusPollingHandler.RegisterPollingFunc(probableHighestNonceHandlerFunc)
	if err != nil {
		return errors.New("cannot register handler func for forkdetector's probable higher nonce")
	}

	return nil
}

func startMachineStatisticsPolling(ash core.AppStatusHandler, pollingInterval int) error {
	if ash == nil {
		return errors.New("nil AppStatusHandler")
	}

	appStatusPollingHandler, err := appStatusPolling.NewAppStatusPolling(ash, pollingInterval)
	if err != nil {
		return errors.New("cannot init AppStatusPolling")
	}

	err = registerCpuStatistics(appStatusPollingHandler)
	if err != nil {
		return err
	}

	err = registerMemStatistics(appStatusPollingHandler)
	if err != nil {
		return err
	}

	err = registeNetStatistics(appStatusPollingHandler)
	if err != nil {
		return err
	}

	appStatusPollingHandler.Poll()

	return nil
}

func registerMemStatistics(appStatusPollingHandler *appStatusPolling.AppStatusPolling) error {
	memStats := &machine.MemStatistics{}
	go func() {
		for {
			memStats.ComputeStatistics()
		}
	}()

	return appStatusPollingHandler.RegisterPollingFunc(func(appStatusHandler core.AppStatusHandler) {
		appStatusHandler.SetUInt64Value(core.MetricMemLoadPercent, memStats.MemPercentUsage())
		appStatusHandler.SetUInt64Value(core.MetricTotalMem, memStats.TotalMemory())
		appStatusHandler.SetUInt64Value(core.MetricMemoryUsedByNode, memStats.UsedMemory())
	})
}

func registeNetStatistics(appStatusPollingHandler *appStatusPolling.AppStatusPolling) error {
	netStats := &machine.NetStatistics{}
	go func() {
		for {
			netStats.ComputeStatistics()
		}
	}()

	return appStatusPollingHandler.RegisterPollingFunc(func(appStatusHandler core.AppStatusHandler) {
		appStatusHandler.SetUInt64Value(core.MetricNetworkRecvBps, netStats.BpsRecv())
		appStatusHandler.SetUInt64Value(core.MetricNetworkRecvBpsPeak, netStats.BpsRecvPeak())
		appStatusHandler.SetUInt64Value(core.MetricNetworkRecvPercent, netStats.PercentRecv())

		appStatusHandler.SetUInt64Value(core.MetricNetworkSentBps, netStats.BpsSent())
		appStatusHandler.SetUInt64Value(core.MetricNetworkSentBpsPeak, netStats.BpsSentPeak())
		appStatusHandler.SetUInt64Value(core.MetricNetworkSentPercent, netStats.PercentSent())
	})
}

func registerCpuStatistics(appStatusPollingHandler *appStatusPolling.AppStatusPolling) error {
	cpuStats := &machine.CpuStatistics{}
	go func() {
		for {
			cpuStats.ComputeStatistics()
		}
	}()

	return appStatusPollingHandler.RegisterPollingFunc(func(appStatusHandler core.AppStatusHandler) {
		appStatusHandler.SetUInt64Value(core.MetricCpuLoadPercent, cpuStats.CpuPercentUsage())
	})
}

func getPrometheusJoinURLIfAvailable(ctx *cli.Context) (string, bool) {
	prometheusURLAvailable := true
	prometheusJoinUrl, err := getPrometheusJoinURL(ctx.GlobalString(serversConfigurationFile.Name))
	if err != nil || prometheusJoinUrl == "" {
		prometheusURLAvailable = false
	}
	usePrometheusBool := ctx.GlobalBool(usePrometheus.Name) && prometheusURLAvailable

	return prometheusJoinUrl, usePrometheusBool
}

func getPrometheusJoinURL(serversConfigurationFileName string) (string, error) {
	serversConfig, err := core.LoadServersPConfig(serversConfigurationFileName)
	if err != nil {
		return "", err
	}
	baseURL := serversConfig.Prometheus.PrometheusBaseURL
	statusURL := baseURL + serversConfig.Prometheus.StatusRoute
	resp, err := http.Get(statusURL)
	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusNotFound {
		return "", errors.New("prometheus URL not available")
	}
	joinURL := baseURL + serversConfig.Prometheus.JoinRoute
	return joinURL, nil
}

func enableGopsIfNeeded(ctx *cli.Context, log *logger.Logger) {
	var gopsEnabled bool
	if ctx.IsSet(gopsEn.Name) {
		gopsEnabled = ctx.GlobalBool(gopsEn.Name)
	}

	if gopsEnabled {
		if err := agent.Listen(agent.Options{}); err != nil {
			log.Error(err.Error())
		}
	}
}

func loadMainConfig(filepath string, log *logger.Logger) (*config.Config, error) {
	cfg := &config.Config{}
	err := core.LoadTomlFile(cfg, filepath, log)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func createShardCoordinator(
	nodesConfig *sharding.NodesSetup,
	pubKey crypto.PublicKey,
	settingsConfig config.GeneralSettingsConfig,
	log *logger.Logger,
) (sharding.Coordinator, core.NodeType, error) {

	if pubKey == nil {
		return nil, "", errors.New("nil public key, could not create shard coordinator")
	}

	publicKey, err := pubKey.ToByteArray()
	if err != nil {
		return nil, "", err
	}

	selfShardId, err := nodesConfig.GetShardIDForPubKey(publicKey)
	nodeType := core.NodeTypeValidator
	if err == sharding.ErrPublicKeyNotFoundInGenesis {
		nodeType = core.NodeTypeObserver
		log.Info("Starting as observer node...")

		selfShardId, err = processDestinationShardAsObserver(settingsConfig)
	}
	if err != nil {
		return nil, "", err
	}

	var shardName string
	if selfShardId == sharding.MetachainShardId {
		shardName = metachainShardName
	} else {
		shardName = fmt.Sprintf("%d", selfShardId)
	}
	log.Info(fmt.Sprintf("Starting in shard: %s", shardName))

	shardCoordinator, err := sharding.NewMultiShardCoordinator(nodesConfig.NumberOfShards(), selfShardId)
	if err != nil {
		return nil, "", err
	}

	return shardCoordinator, nodeType, nil
}

func processDestinationShardAsObserver(settingsConfig config.GeneralSettingsConfig) (uint32, error) {
	destShard := strings.ToLower(settingsConfig.DestinationShardAsObserver)
	if len(destShard) == 0 {
		return 0, errors.New("option DestinationShardAsObserver is not set in config.toml")
	}
	if destShard == metachainShardName {
		return sharding.MetachainShardId, nil
	}

	val, err := strconv.ParseUint(destShard, 10, 32)
	if err != nil {
		return 0, errors.New("error parsing DestinationShardAsObserver option: " + err.Error())
	}

	return uint32(val), err
}

// CreateElasticIndexer creates a new elasticIndexer where the server listens on the url,
// authentication for the server is using the username and password
func CreateElasticIndexer(
	ctx *cli.Context,
	serversConfigurationFileName string,
	url string,
	coordinator sharding.Coordinator,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	log *logger.Logger,
) (indexer.Indexer, error) {
	serversConfig, err := core.LoadServersPConfig(serversConfigurationFileName)
	if err != nil {
		return nil, err
	}

	dbIndexer, err = indexer.NewElasticIndexer(
		url,
		serversConfig.ElasticSearch.Username,
		serversConfig.ElasticSearch.Password,
		coordinator,
		marshalizer,
		hasher,
		log,
		&indexer.Options{TxIndexingEnabled: ctx.GlobalBoolT(enableTxIndexing.Name)})
	if err != nil {
		return nil, err
	}

	return dbIndexer, nil
}

func getConsensusGroupSize(nodesConfig *sharding.NodesSetup, shardCoordinator sharding.Coordinator) (uint32, error) {
	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		return nodesConfig.MetaChainConsensusGroupSize, nil
	}
	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		return nodesConfig.ConsensusGroupSize, nil
	}

	return 0, state.ErrUnknownShardId
}

func createNode(
	config *config.Config,
	nodesConfig *sharding.NodesSetup,
	syncer ntp.SyncTimer,
	keyGen crypto.KeyGenerator,
	privKey crypto.PrivateKey,
	pubKey crypto.PublicKey,
	shardCoordinator sharding.Coordinator,
	core *factory.Core,
	state *factory.State,
	data *factory.Data,
	crypto *factory.Crypto,
	process *factory.Process,
	network *factory.Network,
	bootstrapRoundIndex uint64,
	version string,
) (*node.Node, error) {
	consensusGroupSize, err := getConsensusGroupSize(nodesConfig, shardCoordinator)
	if err != nil {
		return nil, err
	}

	nd, err := node.NewNode(
		node.WithMessenger(network.NetMessenger),
		node.WithHasher(core.Hasher),
		node.WithMarshalizer(core.Marshalizer),
		node.WithInitialNodesPubKeys(crypto.InitialPubKeys),
		node.WithAddressConverter(state.AddressConverter),
		node.WithAccountsAdapter(state.AccountsAdapter),
		node.WithBlockChain(data.Blkc),
		node.WithDataStore(data.Store),
		node.WithRoundDuration(nodesConfig.RoundDuration),
		node.WithConsensusGroupSize(int(consensusGroupSize)),
		node.WithSyncer(syncer),
		node.WithBlockProcessor(process.BlockProcessor),
		node.WithBlockTracker(process.BlockTracker),
		node.WithGenesisTime(time.Unix(nodesConfig.StartTime, 0)),
		node.WithRounder(process.Rounder),
		node.WithShardCoordinator(shardCoordinator),
		node.WithUint64ByteSliceConverter(core.Uint64ByteSliceConverter),
		node.WithSingleSigner(crypto.SingleSigner),
		node.WithMultiSigner(crypto.MultiSigner),
		node.WithKeyGen(keyGen),
		node.WithTxSignPubKey(crypto.TxSignPubKey),
		node.WithTxSignPrivKey(crypto.TxSignPrivKey),
		node.WithPubKey(pubKey),
		node.WithPrivKey(privKey),
		node.WithForkDetector(process.ForkDetector),
		node.WithInterceptorsContainer(process.InterceptorsContainer),
		node.WithResolversFinder(process.ResolversFinder),
		node.WithConsensusType(config.Consensus.Type),
		node.WithTxSingleSigner(crypto.TxSingleSigner),
		node.WithTxStorageSize(config.TxStorage.Cache.Size),
		node.WithBootstrapRoundIndex(bootstrapRoundIndex),
		node.WithAppStatusHandler(core.StatusHandler),
	)
	if err != nil {
		return nil, errors.New("error creating node: " + err.Error())
	}

	err = nd.StartHeartbeat(config.Heartbeat, version, config.GeneralSettings.NodeDisplayName)
	if err != nil {
		return nil, err
	}

	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		err = nd.ApplyOptions(
			node.WithInitialNodesBalances(state.InBalanceForShard),
			node.WithDataPool(data.Datapool),
		)
		if err != nil {
			return nil, errors.New("error creating node: " + err.Error())
		}
		err = nd.CreateShardedStores()
		if err != nil {
			return nil, err
		}
	}
	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		err = nd.ApplyOptions(node.WithMetaDataPool(data.MetaDatapool))
		if err != nil {
			return nil, errors.New("error creating meta-node: " + err.Error())
		}
	}
	return nd, nil
}

func initLogFileAndStatsMonitor(config *config.Config, pubKey crypto.PublicKey, log *logger.Logger,
	workingDir string) error {
	publicKey, err := pubKey.ToByteArray()
	if err != nil {
		return err
	}

	hexPublicKey := core.GetTrimmedPk(hex.EncodeToString(publicKey))
	err = log.ApplyOptions(
		logger.WithFileRotation(hexPublicKey, filepath.Join(workingDir, defaultLogPath), "log"),
		logger.WithStderrRedirect(),
	)
	if err != nil {
		return err
	}

	statsFile, err := core.CreateFile(hexPublicKey, filepath.Join(workingDir, defaultStatsPath), "txt")
	if err != nil {
		return err
	}
	err = startStatisticsMonitor(statsFile, config.ResourceStats, log)
	if err != nil {
		return err
	}
	return nil
}

func setServiceContainer(shardCoordinator sharding.Coordinator, tpsBenchmark *statistics.TpsBenchmark) error {
	var err error
	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		coreServiceContainer, err = serviceContainer.NewServiceContainer(serviceContainer.WithIndexer(dbIndexer))
		if err != nil {
			return err
		}
		return nil
	}
	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		coreServiceContainer, err = serviceContainer.NewServiceContainer(
			serviceContainer.WithIndexer(dbIndexer),
			serviceContainer.WithTPSBenchmark(tpsBenchmark))
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("could not init core service container")
}

func startStatisticsMonitor(file *os.File, config config.ResourceStatsConfig, log *logger.Logger) error {
	if !config.Enabled {
		return nil
	}

	if config.RefreshIntervalInSec < 1 {
		return errors.New("invalid RefreshIntervalInSec in section [ResourceStats]. Should be an integer higher than 1")
	}

	rm, err := statistics.NewResourceMonitor(file)
	if err != nil {
		return err
	}

	go func() {
		for {
			err = rm.SaveStatistics()
			log.LogIfError(err)
			time.Sleep(time.Second * time.Duration(config.RefreshIntervalInSec))
		}
	}()

	return nil
}

func createApiResolver(vmAccountsDB vmcommon.BlockchainHook) (facade.ApiResolver, error) {
	//TODO replace this with a vm factory
	cryptoHook := hooks.NewVMCryptoHook()
	ieleVM := endpoint.NewElrondIeleVM(factoryVM.IELEVirtualMachine, endpoint.ElrondTestnet, vmAccountsDB, cryptoHook)

	scDataGetter, err := smartContract.NewSCDataGetter(ieleVM)
	if err != nil {
		return nil, err
	}

	return external.NewNodeApiResolver(scDataGetter)
}
