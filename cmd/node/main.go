package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/config"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/round"
	"github.com/ElrondNetwork/elrond-go-sandbox/core"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/genesis"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/indexer"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/partitioning"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/serviceContainer"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/statistics"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber/singlesig"
	blsMultiSig "github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber/multisig"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/multisig"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/addressConverters"
	factoryState "github.com/ElrondNetwork/elrond-go-sandbox/data/state/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/factory/containers"
	metafactoryDataRetriever "github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/factory/metachain"
	shardfactoryDataRetriever "github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/factory/shard"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/requestHandlers"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/facade"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/node/external"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p"
	factoryP2P "github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/loadBalancer"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/smartContract"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/smartContract/hooks"
	processSync "github.com/ElrondNetwork/elrond-go-sandbox/process/sync"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/track"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/storageUnit"
	"github.com/btcsuite/btcd/btcec"
	"github.com/google/gops/agent"
	crypto2 "github.com/libp2p/go-libp2p-crypto"
	"github.com/pkg/profile"
	"github.com/urfave/cli"
)

const (
	defaultLogPath     = "logs"
	defaultStatsPath   = "stats"
	metachainShardName = "metachain"
	blsHashSize        = 16
	blsConsensusType   = "bls"
	bnConsensusType    = "bn"
	maxTxsToRequest    = 100
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
		Value: "genesis.json",
	}
	// nodesFile defines a flag for the path of the initial nodes file.
	nodesFile = cli.StringFlag{
		Name:  "nodesSetup-file",
		Usage: "The node will extract initial nodes info from the nodesSetup.json",
		Value: "nodesSetup.json",
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
		Value: "./config/servers.toml",
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
	profileMode = cli.StringFlag{
		Name:  "profile-mode",
		Usage: "Profiling mode. Available options: cpu, mem, mutex, block",
		Value: "",
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

	//TODO remove uniqueID
	uniqueID = ""

	rm *statistics.ResourceMonitor
)

type seedRandReader struct {
	index int
	seed  []byte
}

// NewSeedRandReader will return a new instance of a seed-based reader
func NewSeedRandReader(seed []byte) *seedRandReader {
	return &seedRandReader{seed: seed, index: 0}
}

func (srr *seedRandReader) Read(p []byte) (n int, err error) {
	if srr.seed == nil {
		return 0, errors.New("nil seed")
	}

	if len(srr.seed) == 0 {
		return 0, errors.New("empty seed")
	}

	if p == nil {
		return 0, errors.New("nil buffer")
	}

	if len(p) == 0 {
		return 0, errors.New("empty buffer")
	}

	for i := 0; i < len(p); i++ {
		p[i] = srr.seed[srr.index]

		srr.index++
		srr.index = srr.index % len(srr.seed)
	}

	return len(p), nil
}

type nullChronologyValidator struct {
}

// ValidateReceivedBlock should validate if parameters to be checked are valid
// In this implementation it just returns nil
func (*nullChronologyValidator) ValidateReceivedBlock(shardID uint32, epoch uint32, nonce uint64, round uint32) error {
	//TODO when implementing a workable variant take into account to receive headers "from future" (nonce or round > current round)
	// as this might happen when clocks are slightly de-synchronized
	return nil
}

// TODO - remove this mock and replace with a valid implementation
type mockProposerResolver struct {
}

// ResolveProposer computes a block proposer. For now, this is mocked.
func (mockProposerResolver) ResolveProposer(shardId uint32, roundIndex uint32, prevRandomSeed []byte) ([]byte, error) {
	return []byte("mocked proposer"), nil
}

// dbIndexer will hold the database indexer. Defined globally so it can be initialised only in
//  certain conditions. If those conditions will not be met, it will stay as nil
var dbIndexer indexer.Indexer

// coreServiceContainer is defined globally so it can be injected with appropriate
//  params depending on the type of node we are starting
var coreServiceContainer serviceContainer.Core

func main() {
	log := logger.DefaultLogger()
	log.SetLevel(logger.LogInfo)

	app := cli.NewApp()
	cli.AppHelpTemplate = nodeHelpTemplate
	app.Name = "Elrond Node CLI App"
	app.Version = "v0.0.1"
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
		return startNode(c, log)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func getSuite(config *config.Config) (crypto.Suite, error) {
	switch config.Consensus.Type {
	case blsConsensusType:
		return kyber.NewSuitePairingBn256(), nil
	case bnConsensusType:
		return kyber.NewBlakeSHA256Ed25519(), nil
	}

	return nil, errors.New("no consensus provided in config file")
}

func startNode(ctx *cli.Context, log *logger.Logger) error {
	profileMode := ctx.GlobalString(profileMode.Name)
	switch profileMode {
	case "cpu":
		p := profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook)
		defer p.Stop()
	case "mem":
		p := profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook)
		defer p.Stop()
	case "mutex":
		p := profile.Start(profile.MutexProfile, profile.ProfilePath("."), profile.NoShutdownHook)
		defer p.Stop()
	case "block":
		p := profile.Start(profile.BlockProfile, profile.ProfilePath("."), profile.NoShutdownHook)
		defer p.Stop()
	}

	enableGopsIfNeeded(ctx, log)

	log.Info("Starting node...")

	stop := make(chan bool, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

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
	keyGen, privKey, pubKey, err := getSigningParams(
		ctx,
		log,
		sk.Name,
		skIndex.Name,
		initialNodesSkPemFileName,
		suite)
	if err != nil {
		return err
	}
	log.Info("Starting with public key: " + getPkEncoded(pubKey))

	shardCoordinator, err := createShardCoordinator(nodesConfig, pubKey, generalConfig.GeneralSettings, log)
	if err != nil {
		return err
	}

	publicKey, err := pubKey.ToByteArray()
	if err != nil {
		return err
	}

	uniqueID = core.GetTrimmedPk(hex.EncodeToString(publicKey))

	storageCleanup := ctx.GlobalBool(storageCleanup.Name)
	if storageCleanup {
		err = os.RemoveAll(config.DefaultPath() + uniqueID)
		if err != nil {
			return err
		}
	}

	var currentNode *node.Node
	var tpsBenchmark *statistics.TpsBenchmark
	var externalResolver *external.ExternalResolver

	currentNode, externalResolver, tpsBenchmark, err = createNode(
		ctx,
		generalConfig,
		genesisConfig,
		nodesConfig,
		p2pConfig,
		syncer,
		keyGen,
		privKey,
		pubKey,
		shardCoordinator,
		log)

	if err != nil {
		return err
	}

	ef := facade.NewElrondNodeFacade(currentNode, externalResolver)

	ef.SetLogger(log)
	ef.SetSyncer(syncer)
	ef.SetTpsBenchmark(tpsBenchmark)

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
) (shardCoordinator sharding.Coordinator,
	err error) {
	if pubKey == nil {
		return nil, errors.New("nil public key, could not create shard coordinator")
	}

	publicKey, err := pubKey.ToByteArray()
	if err != nil {
		return nil, err
	}

	selfShardId, err := nodesConfig.GetShardIDForPubKey(publicKey)
	if err == sharding.ErrNoValidPublicKey {
		log.Info("Starting as observer node...")
		selfShardId, err = processDestinationShardAsObserver(settingsConfig)
	}
	if err != nil {
		return nil, err
	}

	var shardName string
	if selfShardId == sharding.MetachainShardId {
		shardName = metachainShardName
	} else {
		shardName = fmt.Sprintf("%d", selfShardId)
	}
	log.Info(fmt.Sprintf("Starting in shard: %s", shardName))

	shardCoordinator, err = sharding.NewMultiShardCoordinator(nodesConfig.NumberOfShards(), selfShardId)
	if err != nil {
		return nil, err
	}

	return shardCoordinator, nil
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

func CreateElasticIndexer(
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
		hasher, log)
	if err != nil {
		return nil, err
	}

	return dbIndexer, nil
}

func createNode(
	ctx *cli.Context,
	config *config.Config,
	genesisConfig *sharding.Genesis,
	nodesConfig *sharding.NodesSetup,
	p2pConfig *config.P2PConfig,
	syncer ntp.SyncTimer,
	keyGen crypto.KeyGenerator,
	privKey crypto.PrivateKey,
	pubKey crypto.PublicKey,
	shardCoordinator sharding.Coordinator,
	log *logger.Logger,
) (*node.Node, *external.ExternalResolver, *statistics.TpsBenchmark, error) {
	coreComponents, err := coreComponentsFactory(config)
	if err != nil {
		return nil, nil, nil, err
	}

	stateComponents, err := stateComponentsFactory(config, shardCoordinator, coreComponents)
	if err != nil {
		return nil, nil, nil, err
	}

	initialPubKeys := nodesConfig.InitialNodesPubKeys()
	publicKey, err := pubKey.ToByteArray()
	if err != nil {
		return nil, nil, nil, err
	}

	hexPublicKey := core.GetTrimmedPk(hex.EncodeToString(publicKey))
	logFile, err := core.CreateFile(hexPublicKey, defaultLogPath, "log")
	if err != nil {
		return nil, nil, nil, err
	}

	err = log.ApplyOptions(logger.WithFile(logFile))
	if err != nil {
		return nil, nil, nil, err
	}

	statsFile, err := core.CreateFile(hexPublicKey, defaultStatsPath, "txt")
	if err != nil {
		return nil, nil, nil, err
	}
	err = startStatisticsMonitor(statsFile, config.ResourceStats, log)
	if err != nil {
		return nil, nil, nil, err
	}

	dataComponents, err := dataComponentsFactory(config, shardCoordinator, coreComponents)
	if err != nil {
		return nil, nil, nil, err
	}

	cryptoComponents, err := cryptoComponentsFactory(ctx, config, nodesConfig, shardCoordinator, keyGen, privKey, log)
	if err != nil {
		return nil, nil, nil, err
	}

	var randReader io.Reader
	if p2pConfig.Node.Seed != "" {
		randReader = NewSeedRandReader(coreComponents.hasher.Compute(p2pConfig.Node.Seed))
	} else {
		randReader = rand.Reader
	}

	netMessenger, err := createNetMessenger(p2pConfig, log, randReader)
	if err != nil {
		return nil, nil, nil, err
	}

	tpsBenchmark, err := statistics.NewTPSBenchmark(shardCoordinator.NumberOfShards(), nodesConfig.RoundDuration/1000)
	if err != nil {
		return nil, nil, nil, err
	}

	log.Info("Starting with tx sign public key: " + getPkEncoded(cryptoComponents.txSignPubKey))

	interceptorContainerFactory, resolversContainerFactory, err := getInterceptorAndResolverContainerFactory(
		shardCoordinator, netMessenger, dataComponents, coreComponents, cryptoComponents, stateComponents)
	if err != nil {
		return nil, nil, nil, err
	}

	//TODO refactor all these factory calls
	interceptorsContainer, err := interceptorContainerFactory.Create()
	if err != nil {
		return nil, nil, nil, err
	}

	resolversContainer, err := resolversContainerFactory.Create()
	if err != nil {
		return nil, nil, nil, err
	}

	resolversFinder, err := containers.NewResolversFinder(resolversContainer, shardCoordinator)
	if err != nil {
		return nil, nil, nil, err
	}

	rounder, err := round.NewRound(
		time.Unix(nodesConfig.StartTime, 0),
		syncer.CurrentTime(),
		time.Millisecond*time.Duration(nodesConfig.RoundDuration),
		syncer)
	if err != nil {
		return nil, nil, nil, err
	}

	forkDetector, err := processSync.NewBasicForkDetector(rounder)
	if err != nil {
		return nil, nil, nil, err
	}

	if config.Explorer.Enabled {
		serversConfigurationFileName := ctx.GlobalString(serversConfigurationFile.Name)
		dbIndexer, err = CreateElasticIndexer(
			serversConfigurationFileName,
			config.Explorer.IndexerURL,
			shardCoordinator,
			coreComponents.marshalizer,
			coreComponents.hasher,
			log)
		if err != nil {
			return nil, nil, nil, err
		}

		err = setServiceContainer(shardCoordinator, tpsBenchmark)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	shardsGenesisBlocks, err := generateGenesisHeadersForInit(
		nodesConfig,
		genesisConfig,
		shardCoordinator,
		stateComponents.addressConverter,
		coreComponents.hasher,
		coreComponents.marshalizer,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	blockProcessor, blockTracker, err := getBlockProcessorAndTracker(resolversFinder, shardCoordinator,
		dataComponents, coreComponents, stateComponents, forkDetector, shardsGenesisBlocks, nodesConfig)
	if err != nil {
		return nil, nil, nil, err
	}

	nd, err := node.NewNode(
		node.WithMessenger(netMessenger),
		node.WithHasher(coreComponents.hasher),
		node.WithMarshalizer(coreComponents.marshalizer),
		node.WithInitialNodesPubKeys(initialPubKeys),
		node.WithAddressConverter(stateComponents.addressConverter),
		node.WithAccountsAdapter(stateComponents.accountsAdapter),
		node.WithBlockChain(dataComponents.blkc),
		node.WithDataStore(dataComponents.store),
		node.WithRoundDuration(nodesConfig.RoundDuration),
		node.WithConsensusGroupSize(int(nodesConfig.ConsensusGroupSize)),
		node.WithSyncer(syncer),
		node.WithBlockProcessor(blockProcessor),
		node.WithBlockTracker(blockTracker),
		node.WithGenesisTime(time.Unix(nodesConfig.StartTime, 0)),
		node.WithRounder(rounder),
		node.WithShardCoordinator(shardCoordinator),
		node.WithUint64ByteSliceConverter(coreComponents.uint64ByteSliceConverter),
		node.WithSingleSigner(cryptoComponents.singleSigner),
		node.WithMultiSigner(cryptoComponents.multiSigner),
		node.WithKeyGen(keyGen),
		node.WithTxSignPubKey(cryptoComponents.txSignPubKey),
		node.WithTxSignPrivKey(cryptoComponents.txSignPrivKey),
		node.WithPubKey(pubKey),
		node.WithPrivKey(privKey),
		node.WithForkDetector(forkDetector),
		node.WithInterceptorsContainer(interceptorsContainer),
		node.WithResolversFinder(resolversFinder),
		node.WithConsensusType(config.Consensus.Type),
		node.WithTxSingleSigner(cryptoComponents.txSingleSigner),
		node.WithTxStorageSize(config.TxStorage.Cache.Size),
	)
	if err != nil {
		return nil, nil, nil, errors.New("error creating node: " + err.Error())
	}

	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		inBalanceForShard, err := genesisConfig.InitialNodesBalances(shardCoordinator, stateComponents.addressConverter)
		if err != nil {
			return nil, nil, nil, errors.New("initial balances could not be processed " + err.Error())
		}
		err = nd.ApplyOptions(
			node.WithInitialNodesBalances(inBalanceForShard),
			node.WithDataPool(dataComponents.datapool),
			node.WithActiveMetachain(nodesConfig.MetaChainActive))
		if err != nil {
			return nil, nil, nil, errors.New("error creating node: " + err.Error())
		}
		err = nd.CreateShardedStores()
		if err != nil {
			return nil, nil, nil, err
		}
		err = nd.StartHeartbeat(config.Heartbeat)
		if err != nil {
			return nil, nil, nil, err
		}
		err = nd.CreateShardGenesisBlock()
		if err != nil {
			return nil, nil, nil, err
		}
	}
	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		err = nd.ApplyOptions(node.WithMetaDataPool(dataComponents.metaDatapool))
		if err != nil {
			return nil, nil, nil, errors.New("error creating meta-node: " + err.Error())
		}
		err = nd.CreateMetaGenesisBlock()
		if err != nil {
			return nil, nil, nil, err
		}

	}

	externalResolver, err := external.NewExternalResolver(
		shardCoordinator,
		dataComponents.blkc,
		dataComponents.store,
		coreComponents.marshalizer,
		&mockProposerResolver{},
	)
	if err != nil {
		return nil, nil, nil, err
	}

	return nd, externalResolver, tpsBenchmark, nil

}

func getBlockProcessorAndTracker(
	resolversFinder dataRetriever.ResolversFinder,
	shardCoordinator sharding.Coordinator,
	data *Data,
	core *Core,
	state *State,
	forkDetector process.ForkDetector,
	shardsGenesisBlocks map[uint32]data.HeaderHandler,
	nodesConfig *sharding.NodesSetup,
) (process.BlockProcessor, process.BlocksTracker, error) {
	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		argsParser, err := smartContract.NewAtArgumentParser()
		if err != nil {
			return nil, nil, err
		}

		vmAccountsDB, err := hooks.NewVMAccountsDB(state.accountsAdapter, state.addressConverter)
		if err != nil {
			return nil, nil, err
		}

		//TODO: change the mock
		scProcessor, err := smartContract.NewSmartContractProcessor(&mock.VMExecutionHandlerStub{}, argsParser,
			core.hasher, core.marshalizer, state.accountsAdapter, vmAccountsDB, state.addressConverter, shardCoordinator)
		if err != nil {
			return nil, nil, err
		}

		requestHandler, err := requestHandlers.NewShardResolverRequestHandler(resolversFinder, factory.TransactionTopic,
			factory.MiniBlocksTopic, factory.MetachainBlocksTopic, maxTxsToRequest)
		if err != nil {
			return nil, nil, err
		}

		transactionProcessor, err := transaction.NewTxProcessor(state.accountsAdapter, core.hasher,
			state.addressConverter, core.marshalizer, shardCoordinator, scProcessor)
		if err != nil {
			return nil, nil, errors.New("could not create transaction processor: " + err.Error())
		}

		blockTracker, err := track.NewShardBlockTracker(data.datapool, core.marshalizer, shardCoordinator, data.store)
		if err != nil {
			return nil, nil, err
		}

		blockProcessor, err := block.NewShardProcessor(
			coreServiceContainer,
			data.datapool,
			data.store,
			core.hasher,
			core.marshalizer,
			transactionProcessor,
			state.accountsAdapter,
			shardCoordinator,
			forkDetector,
			blockTracker,
			shardsGenesisBlocks,
			nodesConfig.MetaChainActive,
			requestHandler,
		)
		if err != nil {
			return nil, nil, errors.New("could not create block processor: " + err.Error())
		}

		return blockProcessor, blockTracker, nil
	}
	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		requestHandler, err := requestHandlers.NewMetaResolverRequestHandler(resolversFinder, factory.ShardHeadersForMetachainTopic)
		if err != nil {
			return nil, nil, err
		}

		blockTracker, err := track.NewMetaBlockTracker()
		if err != nil {
			return nil, nil, err
		}

		metaProcessor, err := block.NewMetaProcessor(
			coreServiceContainer,
			state.accountsAdapter,
			data.metaDatapool,
			forkDetector,
			shardCoordinator,
			core.hasher,
			core.marshalizer,
			data.store,
			shardsGenesisBlocks,
			requestHandler,
		)
		if err != nil {
			return nil, nil, errors.New("could not create block processor: " + err.Error())
		}
		return metaProcessor, blockTracker, nil
	}
	return nil, nil, errors.New("could not create block processor and tracker")
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

func getInterceptorAndResolverContainerFactory(
	shardCoordinator sharding.Coordinator,
	netMessenger p2p.Messenger,
	data *Data,
	core *Core,
	crypto *Crypto,
	state *State,
) (process.InterceptorsContainerFactory, dataRetriever.ResolversContainerFactory, error) {
	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		//TODO add a real chronology validator and remove null chronology validator
		interceptorContainerFactory, err := shard.NewInterceptorsContainerFactory(
			shardCoordinator,
			netMessenger,
			data.store,
			core.marshalizer,
			core.hasher,
			crypto.txSignKeyGen,
			crypto.txSingleSigner,
			crypto.multiSigner,
			data.datapool,
			state.addressConverter,
			&nullChronologyValidator{},
		)
		if err != nil {
			return nil, nil, err
		}

		dataPacker, err := partitioning.NewSizeDataPacker(core.marshalizer)
		if err != nil {
			return nil, nil, err
		}

		resolversContainerFactory, err := shardfactoryDataRetriever.NewResolversContainerFactory(
			shardCoordinator,
			netMessenger,
			data.store,
			core.marshalizer,
			data.datapool,
			core.uint64ByteSliceConverter,
			dataPacker,
		)
		if err != nil {
			return nil, nil, err
		}

		return interceptorContainerFactory, resolversContainerFactory, nil
	}
	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		//TODO add a real chronology validator and remove null chronology validator
		interceptorContainerFactory, err := metachain.NewInterceptorsContainerFactory(
			shardCoordinator,
			netMessenger,
			data.store,
			core.marshalizer,
			core.hasher,
			crypto.multiSigner,
			data.metaDatapool,
			&nullChronologyValidator{},
		)
		if err != nil {
			return nil, nil, err
		}
		resolversContainerFactory, err := metafactoryDataRetriever.NewResolversContainerFactory(
			shardCoordinator,
			netMessenger,
			data.store,
			core.marshalizer,
			data.metaDatapool,
			core.uint64ByteSliceConverter,
		)
		if err != nil {
			return nil, nil, err
		}
		return interceptorContainerFactory, resolversContainerFactory, nil
	}
	return nil, nil, errors.New("could not create interceptor and resolver container factory")
}

func createNetMessenger(
	p2pConfig *config.P2PConfig,
	log *logger.Logger,
	randReader io.Reader,
) (p2p.Messenger, error) {

	if p2pConfig.Node.Port < 0 {
		return nil, errors.New("cannot start node on port < 0")
	}

	pDiscoveryFactory := factoryP2P.NewPeerDiscovererCreator(*p2pConfig)
	pDiscoverer, err := pDiscoveryFactory.CreatePeerDiscoverer()

	if err != nil {
		return nil, err
	}

	log.Info(fmt.Sprintf("Starting with peer discovery: %s", pDiscoverer.Name()))

	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), randReader)
	sk := (*crypto2.Secp256k1PrivateKey)(prvKey)

	nm, err := libp2p.NewNetworkMessenger(
		context.Background(),
		p2pConfig.Node.Port,
		sk,
		nil,
		loadBalancer.NewOutgoingChannelLoadBalancer(),
		pDiscoverer,
		libp2p.ListenAddrWithIp4AndTcp,
	)

	if err != nil {
		return nil, err
	}
	return nm, nil
}

func getSk(ctx *cli.Context, log *logger.Logger, skName string, skIndexName string, skPemFileName string) ([]byte, error) {
	//if flag is defined, it shall overwrite what was read from pem file
	if ctx.GlobalIsSet(skName) {
		encodedSk := []byte(ctx.GlobalString(skName))
		return decodeAddress(string(encodedSk))
	}

	skIndex := ctx.GlobalInt(skIndexName)
	encodedSk, err := core.LoadSkFromPemFile(skPemFileName, log, skIndex)
	if err != nil {
		return nil, err
	}

	return decodeAddress(string(encodedSk))
}

func getSigningParams(
	ctx *cli.Context,
	log *logger.Logger,
	skName string,
	skIndexName string,
	skPemFileName string,
	suite crypto.Suite,
) (keyGen crypto.KeyGenerator, privKey crypto.PrivateKey, pubKey crypto.PublicKey, err error) {

	sk, err := getSk(ctx, log, skName, skIndexName, skPemFileName)
	if err != nil {
		return nil, nil, nil, err
	}

	keyGen = signing.NewKeyGenerator(suite)

	privKey, err = keyGen.PrivateKeyFromByteArray(sk)
	if err != nil {
		return nil, nil, nil, err
	}

	pubKey = privKey.GeneratePublic()

	return keyGen, privKey, pubKey, err
}

func getPkEncoded(pubKey crypto.PublicKey) string {
	pk, err := pubKey.ToByteArray()
	if err != nil {
		return err.Error()
	}

	return encodeAddress(pk)
}

func getCacherFromConfig(cfg config.CacheConfig) storageUnit.CacheConfig {
	return storageUnit.CacheConfig{
		Size:   cfg.Size,
		Type:   storageUnit.CacheType(cfg.Type),
		Shards: cfg.Shards,
	}
}

func getDBFromConfig(cfg config.DBConfig) storageUnit.DBConfig {
	return storageUnit.DBConfig{
		FilePath:          filepath.Join(config.DefaultPath()+uniqueID, cfg.FilePath),
		Type:              storageUnit.DBType(cfg.Type),
		MaxBatchSize:      cfg.MaxBatchSize,
		BatchDelaySeconds: cfg.BatchDelaySeconds,
	}
}

func getBloomFromConfig(cfg config.BloomFilterConfig) storageUnit.BloomConfig {
	var hashFuncs []storageUnit.HasherType
	if cfg.HashFunc != nil {
		hashFuncs = make([]storageUnit.HasherType, 0)
		for _, hf := range cfg.HashFunc {
			hashFuncs = append(hashFuncs, storageUnit.HasherType(hf))
		}
	}

	return storageUnit.BloomConfig{
		Size:     cfg.Size,
		HashFunc: hashFuncs,
	}
}

func decodeAddress(address string) ([]byte, error) {
	return hex.DecodeString(address)
}

func encodeAddress(address []byte) string {
	return hex.EncodeToString(address)
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

func generateGenesisHeadersForInit(
	nodesSetup *sharding.NodesSetup,
	genesisConfig *sharding.Genesis,
	shardCoordinator sharding.Coordinator,
	addressConverter state.AddressConverter,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
) (map[uint32]data.HeaderHandler, error) {
	//TODO change this rudimentary startup for metachain nodes
	// Talk between Adrian, Robert and Iulian, did not want it to be discarded:
	// --------------------------------------------------------------------
	// Adrian: "This looks like a workaround as the metchain should not deal with individual accounts, but shards data.
	// What I was thinking was that the genesis on metachain (or pre-genesis block) is the nodes allocation to shards,
	// with 0 state root for every shard, as there is no balance yet.
	// Then the shards start operating as they get the initial node allocation, maybe we can do consensus on the
	// genesis as well, I think this would be actually good as then everything is signed and agreed upon.
	// The genesis shard blocks need to be then just the state root, I think we already have that in genesis,
	// so shard nodes can go ahead with individually creating the block, but then run consensus on this.
	// Then this block is sent to metachain who updates the state root of every shard and creates the metablock for
	// the genesis of each of the shards (this is actually the same thing that would happen at new epoch start)."

	shardsGenesisBlocks := make(map[uint32]data.HeaderHandler)

	for shardId := uint32(0); shardId < shardCoordinator.NumberOfShards(); shardId++ {
		newShardCoordinator, err := sharding.NewMultiShardCoordinator(shardCoordinator.NumberOfShards(), shardId)
		if err != nil {
			return nil, err
		}

		accountFactory, err := factoryState.NewAccountFactoryCreator(newShardCoordinator)
		if err != nil {
			return nil, err
		}

		accounts := generateInMemoryAccountsAdapter(accountFactory, hasher, marshalizer)
		initialBalances, err := genesisConfig.InitialNodesBalances(newShardCoordinator, addressConverter)
		if err != nil {
			return nil, err
		}

		genesisBlock, err := genesis.CreateShardGenesisBlockFromInitialBalances(
			accounts,
			newShardCoordinator,
			addressConverter,
			initialBalances,
			uint64(nodesSetup.StartTime),
		)
		if err != nil {
			return nil, err
		}

		shardsGenesisBlocks[shardId] = genesisBlock
	}

	if nodesSetup.IsMetaChainActive() {
		genesisBlock, err := genesis.CreateMetaGenesisBlock(uint64(nodesSetup.StartTime), nodesSetup.InitialNodesPubKeys())
		if err != nil {
			return nil, err
		}

		shardsGenesisBlocks[sharding.MetachainShardId] = genesisBlock
	}

	return shardsGenesisBlocks, nil
}

func generateInMemoryAccountsAdapter(
	accountFactory state.AccountFactory,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
) state.AccountsAdapter {

	dbw, _ := trie.NewDBWriteCache(createMemUnit())
	tr, _ := trie.NewTrie(make([]byte, 32), dbw, hasher)
	adb, _ := state.NewAccountsDB(tr, sha256.Sha256{}, marshalizer, accountFactory)

	return adb
}

func createMemUnit() storage.Storer {
	cache, _ := storageUnit.NewCache(storageUnit.LRUCache, 10, 1)
	persist, _ := memorydb.New()

	unit, _ := storageUnit.NewStorageUnit(cache, persist)
	return unit
}

type Core struct {
	hasher                   hashing.Hasher
	marshalizer              marshal.Marshalizer
	tr                       trie.PatriciaMerkelTree
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
}

type State struct {
	addressConverter state.AddressConverter
	accountsAdapter  state.AccountsAdapter
}

type Data struct {
	blkc         data.ChainHandler
	store        dataRetriever.StorageService
	datapool     dataRetriever.PoolsHolder
	metaDatapool dataRetriever.MetaPoolsHolder
}

type Crypto struct {
	txSingleSigner crypto.SingleSigner
	singleSigner   crypto.SingleSigner
	multiSigner    crypto.MultiSigner
	txSignKeyGen   crypto.KeyGenerator
	txSignPrivKey  crypto.PrivateKey
	txSignPubKey   crypto.PublicKey
}

//TODO: refactor in the next phase
//type Process struct {
//	interceptorsContainer process.InterceptorsContainer
//	resolversFinder       dataRetriever.ResolversFinder
//	rounder               consensus.Rounder
//	forkDetector          process.ForkDetector
//	blockProcessor        process.BlockProcessor
//	blockTracker          process.BlocksTracker
//}

func coreComponentsFactory(config *config.Config) (*Core, error) {
	hasher, err := getHasherFromConfig(config)
	if err != nil {
		return nil, errors.New("could not create hasher: " + err.Error())
	}

	marshalizer, err := getMarshalizerFromConfig(config)
	if err != nil {
		return nil, errors.New("could not create marshalizer: " + err.Error())
	}

	tr, err := getTrie(config.AccountsTrieStorage, hasher)
	if err != nil {
		return nil, errors.New("error creating trie: " + err.Error())
	}
	uint64ByteSliceConverter := uint64ByteSlice.NewBigEndianConverter()

	return &Core{
		hasher:                   hasher,
		marshalizer:              marshalizer,
		tr:                       tr,
		uint64ByteSliceConverter: uint64ByteSliceConverter,
	}, nil
}

func stateComponentsFactory(config *config.Config, shardCoordinator sharding.Coordinator, core *Core) (*State, error) {
	addressConverter, err := addressConverters.NewPlainAddressConverter(config.Address.Length, config.Address.Prefix)
	if err != nil {
		return nil, errors.New("could not create address converter: " + err.Error())
	}

	accountFactory, err := factoryState.NewAccountFactoryCreator(shardCoordinator)
	if err != nil {
		return nil, errors.New("could not create account factory: " + err.Error())
	}

	accountsAdapter, err := state.NewAccountsDB(core.tr, core.hasher, core.marshalizer, accountFactory)
	if err != nil {
		return nil, errors.New("could not create accounts adapter: " + err.Error())
	}

	return &State{
		addressConverter: addressConverter,
		accountsAdapter:  accountsAdapter,
	}, nil
}

func dataComponentsFactory(config *config.Config, shardCoordinator sharding.Coordinator, core *Core) (*Data, error) {
	var datapool dataRetriever.PoolsHolder
	var metaDatapool dataRetriever.MetaPoolsHolder
	blkc, err := createBlockChainFromConfig(config, shardCoordinator)
	if err != nil {
		return nil, errors.New("could not create block chain: " + err.Error())
	}

	store, err := createDataStoreFromConfig(config, shardCoordinator)
	if err != nil {
		return nil, errors.New("could not create local data store: " + err.Error())
	}

	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		datapool, err = createShardDataPoolFromConfig(config, core.uint64ByteSliceConverter)
		if err != nil {
			return nil, errors.New("could not create shard data pools: " + err.Error())
		}
	}
	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		metaDatapool, err = createMetaDataPoolFromConfig(config, core.uint64ByteSliceConverter)
		if err != nil {
			return nil, errors.New("could not create shard data pools: " + err.Error())
		}
	}
	if datapool == nil && metaDatapool == nil {
		return nil, errors.New("could not create data pools: ")
	}

	return &Data{
		blkc:         blkc,
		store:        store,
		datapool:     datapool,
		metaDatapool: metaDatapool,
	}, nil
}

func cryptoComponentsFactory(ctx *cli.Context, config *config.Config, nodesConfig *sharding.NodesSetup, shardCoordinator sharding.Coordinator,
	keyGen crypto.KeyGenerator, privKey crypto.PrivateKey, log *logger.Logger,
) (*Crypto, error) {
	txSingleSigner := &singlesig.SchnorrSigner{}
	singleSigner, err := createSingleSigner(config)
	if err != nil {
		return nil, errors.New("could not create singleSigner: " + err.Error())
	}

	multisigHasher, err := getMultisigHasherFromConfig(config)
	if err != nil {
		return nil, errors.New("could not create multisig hasher: " + err.Error())
	}

	currentShardPubKeys, err := nodesConfig.InitialNodesPubKeysForShard(shardCoordinator.SelfId())
	if err != nil {
		return nil, errors.New("could not start creation of multiSigner: " + err.Error())
	}

	multiSigner, err := createMultiSigner(config, multisigHasher, currentShardPubKeys, privKey, keyGen)
	if err != nil {
		return nil, err
	}

	initialBalancesSkPemFileName := ctx.GlobalString(initialBalancesSkPemFile.Name)
	txSignKeyGen, txSignPrivKey, txSignPubKey, err := getSigningParams(
		ctx,
		log,
		txSignSk.Name,
		txSignSkIndex.Name,
		initialBalancesSkPemFileName,
		kyber.NewBlakeSHA256Ed25519())
	if err != nil {
		return nil, err
	}

	return &Crypto{
		txSingleSigner: txSingleSigner,
		singleSigner:   singleSigner,
		multiSigner:    multiSigner,
		txSignKeyGen:   txSignKeyGen,
		txSignPrivKey:  txSignPrivKey,
		txSignPubKey:   txSignPubKey,
	}, nil
}

func getHasherFromConfig(cfg *config.Config) (hashing.Hasher, error) {
	switch cfg.Hasher.Type {
	case "sha256":
		return sha256.Sha256{}, nil
	case "blake2b":
		return blake2b.Blake2b{}, nil
	}

	return nil, errors.New("no hasher provided in config file")
}

func getMarshalizerFromConfig(cfg *config.Config) (marshal.Marshalizer, error) {
	switch cfg.Marshalizer.Type {
	case "json":
		return marshal.JsonMarshalizer{}, nil
	}

	return nil, errors.New("no marshalizer provided in config file")
}

func getTrie(cfg config.StorageConfig, hasher hashing.Hasher) (*trie.Trie, error) {
	accountsTrieStorage, err := storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(cfg.Cache),
		getDBFromConfig(cfg.DB),
		getBloomFromConfig(cfg.Bloom),
	)
	if err != nil {
		return nil, errors.New("error creating accountsTrieStorage: " + err.Error())
	}

	dbWriteCache, err := trie.NewDBWriteCache(accountsTrieStorage)
	if err != nil {
		return nil, errors.New("error creating dbWriteCache: " + err.Error())
	}

	return trie.NewTrie(make([]byte, 32), dbWriteCache, hasher)
}

func createBlockChainFromConfig(config *config.Config, coordinator sharding.Coordinator) (data.ChainHandler, error) {
	badBlockCache, err := storageUnit.NewCache(
		storageUnit.CacheType(config.BadBlocksCache.Type),
		config.BadBlocksCache.Size,
		config.BadBlocksCache.Shards)
	if err != nil {
		return nil, err
	}

	if coordinator == nil {
		return nil, state.ErrNilShardCoordinator
	}

	if coordinator.SelfId() < coordinator.NumberOfShards() {
		blockChain, err := blockchain.NewBlockChain(badBlockCache)
		if err != nil {
			return nil, err
		}
		return blockChain, nil
	}
	if coordinator.SelfId() == sharding.MetachainShardId {
		blockChain, err := blockchain.NewMetaChain(badBlockCache)
		if err != nil {
			return nil, err
		}
		return blockChain, nil
	}
	return nil, errors.New("can not create blockchain")
}

func createDataStoreFromConfig(config *config.Config, shardCoordinator sharding.Coordinator) (dataRetriever.StorageService, error) {
	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		return createShardDataStoreFromConfig(config, shardCoordinator)
	}
	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		return createMetaChainDataStoreFromConfig(config, shardCoordinator)
	}
	return nil, errors.New("can not create data store")
}

func createShardDataStoreFromConfig(config *config.Config, shardCoordinator sharding.Coordinator) (dataRetriever.StorageService, error) {
	var headerUnit, peerBlockUnit, miniBlockUnit, txUnit, metachainHeaderUnit, metaHdrHashNonceUnit, shardHdrHashNonceUnit *storageUnit.Unit
	var err error

	defer func() {
		// cleanup
		if err != nil {
			if headerUnit != nil {
				_ = headerUnit.DestroyUnit()
			}
			if peerBlockUnit != nil {
				_ = peerBlockUnit.DestroyUnit()
			}
			if miniBlockUnit != nil {
				_ = miniBlockUnit.DestroyUnit()
			}
			if txUnit != nil {
				_ = txUnit.DestroyUnit()
			}
			if metachainHeaderUnit != nil {
				_ = metachainHeaderUnit.DestroyUnit()
			}
			if metaHdrHashNonceUnit != nil {
				_ = metaHdrHashNonceUnit.DestroyUnit()
			}
			if shardHdrHashNonceUnit != nil {
				_ = shardHdrHashNonceUnit.DestroyUnit()
			}
		}
	}()

	txUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.TxStorage.Cache),
		getDBFromConfig(config.TxStorage.DB),
		getBloomFromConfig(config.TxStorage.Bloom))
	if err != nil {
		return nil, err
	}

	miniBlockUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.MiniBlocksStorage.Cache),
		getDBFromConfig(config.MiniBlocksStorage.DB),
		getBloomFromConfig(config.MiniBlocksStorage.Bloom))
	if err != nil {
		return nil, err
	}

	peerBlockUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.PeerBlockBodyStorage.Cache),
		getDBFromConfig(config.PeerBlockBodyStorage.DB),
		getBloomFromConfig(config.PeerBlockBodyStorage.Bloom))
	if err != nil {
		return nil, err
	}

	headerUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.BlockHeaderStorage.Cache),
		getDBFromConfig(config.BlockHeaderStorage.DB),
		getBloomFromConfig(config.BlockHeaderStorage.Bloom))
	if err != nil {
		return nil, err
	}

	metachainHeaderUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.MetaBlockStorage.Cache),
		getDBFromConfig(config.MetaBlockStorage.DB),
		getBloomFromConfig(config.MetaBlockStorage.Bloom))
	if err != nil {
		return nil, err
	}

	metaHdrHashNonceUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.MetaHdrNonceHashStorage.Cache),
		getDBFromConfig(config.MetaHdrNonceHashStorage.DB),
		getBloomFromConfig(config.MetaHdrNonceHashStorage.Bloom),
	)
	if err != nil {
		return nil, err
	}

	shardHdrHashNonceUnit, err = storageUnit.NewShardedStorageUnitFromConf(
		getCacherFromConfig(config.ShardHdrNonceHashStorage.Cache),
		getDBFromConfig(config.ShardHdrNonceHashStorage.DB),
		getBloomFromConfig(config.ShardHdrNonceHashStorage.Bloom),
		shardCoordinator.SelfId(),
	)
	if err != nil {
		return nil, err
	}

	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, txUnit)
	store.AddStorer(dataRetriever.MiniBlockUnit, miniBlockUnit)
	store.AddStorer(dataRetriever.PeerChangesUnit, peerBlockUnit)
	store.AddStorer(dataRetriever.BlockHeaderUnit, headerUnit)
	store.AddStorer(dataRetriever.MetaBlockUnit, metachainHeaderUnit)
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, metaHdrHashNonceUnit)
	store.AddStorer(dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()), shardHdrHashNonceUnit)

	return store, err
}

func createMetaChainDataStoreFromConfig(config *config.Config, shardCoordinator sharding.Coordinator) (dataRetriever.StorageService, error) {
	var peerDataUnit, shardDataUnit, metaBlockUnit, headerUnit, metaHdrHashNonceUnit  *storageUnit.Unit
	var shardHdrHashNonceUnits []*storageUnit.Unit
	var err error


	defer func() {
		// cleanup
		if err != nil {
			if peerDataUnit != nil {
				_ = peerDataUnit.DestroyUnit()
			}
			if shardDataUnit != nil {
				_ = shardDataUnit.DestroyUnit()
			}
			if metaBlockUnit != nil {
				_ = metaBlockUnit.DestroyUnit()
			}
			if headerUnit != nil {
				_ = headerUnit.DestroyUnit()
			}
			if metaHdrHashNonceUnit != nil {
				_ = metaHdrHashNonceUnit.DestroyUnit()
			}
			if shardHdrHashNonceUnits != nil {
				for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
					_ = shardHdrHashNonceUnits[i].DestroyUnit()
				}
			}
		}
	}()

	metaBlockUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.MetaBlockStorage.Cache),
		getDBFromConfig(config.MetaBlockStorage.DB),
		getBloomFromConfig(config.MetaBlockStorage.Bloom))
	if err != nil {
		return nil, err
	}

	shardDataUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.ShardDataStorage.Cache),
		getDBFromConfig(config.ShardDataStorage.DB),
		getBloomFromConfig(config.ShardDataStorage.Bloom))
	if err != nil {
		return nil, err
	}

	peerDataUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.PeerDataStorage.Cache),
		getDBFromConfig(config.PeerDataStorage.DB),
		getBloomFromConfig(config.PeerDataStorage.Bloom))
	if err != nil {
		return nil, err
	}

	headerUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.BlockHeaderStorage.Cache),
		getDBFromConfig(config.BlockHeaderStorage.DB),
		getBloomFromConfig(config.BlockHeaderStorage.Bloom))
	if err != nil {
		return nil, err
	}

	metaHdrHashNonceUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.MetaHdrNonceHashStorage.Cache),
		getDBFromConfig(config.MetaHdrNonceHashStorage.DB),
		getBloomFromConfig(config.MetaHdrNonceHashStorage.Bloom),
	)
	if err != nil {
		return nil, err
	}

	shardHdrHashNonceUnits = make([]*storageUnit.Unit, shardCoordinator.NumberOfShards())
	for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
		shardHdrHashNonceUnits[i], err = storageUnit.NewShardedStorageUnitFromConf(
			getCacherFromConfig(config.ShardHdrNonceHashStorage.Cache),
			getDBFromConfig(config.ShardHdrNonceHashStorage.DB),
			getBloomFromConfig(config.ShardHdrNonceHashStorage.Bloom),
			i,
		)
		if err != nil {
			return nil, err
		}
	}

	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.MetaBlockUnit, metaBlockUnit)
	store.AddStorer(dataRetriever.MetaShardDataUnit, shardDataUnit)
	store.AddStorer(dataRetriever.MetaPeerDataUnit, peerDataUnit)
	store.AddStorer(dataRetriever.BlockHeaderUnit, headerUnit)
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, metaHdrHashNonceUnit)
	for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
		store.AddStorer(dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(i), shardHdrHashNonceUnits[i])
	}

	return store, err
}

func createShardDataPoolFromConfig(
	config *config.Config,
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter,
) (dataRetriever.PoolsHolder, error) {

	fmt.Println("creatingShardDataPool from config")

	txPool, err := shardedData.NewShardedData(getCacherFromConfig(config.TxDataPool))
	if err != nil {
		fmt.Println("error creating txpool")
		return nil, err
	}

	cacherCfg := getCacherFromConfig(config.BlockHeaderDataPool)
	hdrPool, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		fmt.Println("error creating hdrpool")
		return nil, err
	}

	cacherCfg = getCacherFromConfig(config.MetaBlockBodyDataPool)
	metaBlockBody, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		fmt.Println("error creating metaBlockBody")
		return nil, err
	}

	cacherCfg = getCacherFromConfig(config.BlockHeaderNoncesDataPool)
	hdrNoncesCacher, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		fmt.Println("error creating hdrNoncesCacher")
		return nil, err
	}
	hdrNonces, err := dataPool.NewNonceToHashCacher(hdrNoncesCacher, uint64ByteSliceConverter)
	if err != nil {
		fmt.Println("error creating hdrNonces")
		return nil, err
	}

	cacherCfg = getCacherFromConfig(config.TxBlockBodyDataPool)
	txBlockBody, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		fmt.Println("error creating txBlockBody")
		return nil, err
	}

	cacherCfg = getCacherFromConfig(config.PeerBlockBodyDataPool)
	peerChangeBlockBody, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		fmt.Println("error creating peerChangeBlockBody")
		return nil, err
	}

	cacherCfg = getCacherFromConfig(config.MetaHeaderNoncesDataPool)
	metaBlockNoncesCacher, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		fmt.Println("error creating metaBlockNoncesCacher")
		return nil, err
	}
	metaBlockNonces, err := dataPool.NewNonceToHashCacher(metaBlockNoncesCacher, uint64ByteSliceConverter)
	if err != nil {
		fmt.Println("error creating metaBlockNonces")
		return nil, err
	}

	return dataPool.NewShardedDataPool(
		txPool,
		hdrPool,
		hdrNonces,
		txBlockBody,
		peerChangeBlockBody,
		metaBlockBody,
		metaBlockNonces,
	)
}

func createMetaDataPoolFromConfig(
	config *config.Config,
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter,
) (dataRetriever.MetaPoolsHolder, error) {
	cacherCfg := getCacherFromConfig(config.MetaBlockBodyDataPool)
	metaBlockBody, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		fmt.Println("error creating metaBlockBody")
		return nil, err
	}

	miniBlockHashes, err := shardedData.NewShardedData(getCacherFromConfig(config.MiniBlockHeaderHashesDataPool))
	if err != nil {
		fmt.Println("error creating miniBlockHashes")
		return nil, err
	}

	cacherCfg = getCacherFromConfig(config.ShardHeadersDataPool)
	shardHeaders, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		fmt.Println("error creating shardHeaders")
		return nil, err
	}

	shardHeadersNoncesCacher, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		fmt.Println("error creating shard headers nonces pool")
		return nil, err
	}
	shardHeadersNonces, err := dataPool.NewNonceToHashCacher(shardHeadersNoncesCacher, uint64ByteSliceConverter)
	if err != nil {
		fmt.Println("error creating shard headers nonces pool")
		return nil, err
	}

	cacherCfg = getCacherFromConfig(config.MetaHeaderNoncesDataPool)
	metaBlockNoncesCacher, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		fmt.Println("error creating metaBlockNoncesCacher")
		return nil, err
	}
	metaBlockNonces, err := dataPool.NewNonceToHashCacher(metaBlockNoncesCacher, uint64ByteSliceConverter)
	if err != nil {
		fmt.Println("error creating metaBlockNonces")
		return nil, err
	}

	return dataPool.NewMetaDataPool(metaBlockBody, miniBlockHashes, shardHeaders, metaBlockNonces, shardHeadersNonces)
}

func createSingleSigner(config *config.Config) (crypto.SingleSigner, error) {
	switch config.Consensus.Type {
	case blsConsensusType:
		return &singlesig.BlsSingleSigner{}, nil
	case bnConsensusType:
		return &singlesig.SchnorrSigner{}, nil
	}

	return nil, errors.New("no consensus type provided in config file")
}

func getMultisigHasherFromConfig(cfg *config.Config) (hashing.Hasher, error) {
	if cfg.Consensus.Type == blsConsensusType && cfg.MultisigHasher.Type != "blake2b" {
		return nil, errors.New("wrong multisig hasher provided for bls consensus type")
	}

	switch cfg.MultisigHasher.Type {
	case "sha256":
		return sha256.Sha256{}, nil
	case "blake2b":
		if cfg.Consensus.Type == blsConsensusType {
			return blake2b.Blake2b{HashSize: blsHashSize}, nil
		}
		return blake2b.Blake2b{}, nil
	}

	return nil, errors.New("no multisig hasher provided in config file")
}

func createMultiSigner(
	config *config.Config,
	hasher hashing.Hasher,
	pubKeys []string,
	privateKey crypto.PrivateKey,
	keyGen crypto.KeyGenerator,
) (crypto.MultiSigner, error) {

	switch config.Consensus.Type {
	case blsConsensusType:
		blsSigner := &blsMultiSig.KyberMultiSignerBLS{}
		return multisig.NewBLSMultisig(blsSigner, hasher, pubKeys, privateKey, keyGen, uint16(0))
	case bnConsensusType:
		return multisig.NewBelNevMultisig(hasher, pubKeys, privateKey, keyGen, uint16(0))
	}

	return nil, errors.New("no consensus type provided in config file")
}
