package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
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
	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/partitioning"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/statistics"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber"
	blsMultiSig "github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber/multisig"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber/singlesig"
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
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/resolvers"
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
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory/shard"
	processSync "github.com/ElrondNetwork/elrond-go-sandbox/process/sync"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/track"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/storageUnit"
	beevikntp "github.com/beevik/ntp"
	"github.com/btcsuite/btcd/btcec"
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

	configurationFile        = "./config/config.toml"
	p2pConfigurationFile     = "./config/p2p.toml"
	initialBalancesSkPemFile = "./config/initialBalancesSk.pem"
	initialNodesSkPemFile    = "./config/initialNodesSk.pem"

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

func main() {
	log := logger.DefaultLogger()
	log.SetLevel(logger.LogInfo)

	app := cli.NewApp()
	cli.AppHelpTemplate = nodeHelpTemplate
	app.Name = "Elrond Node CLI App"
	app.Version = "v0.0.1"
	app.Usage = "This is the entry point for starting a new Elrond node - the app will start after the genesis timestamp"
	app.Flags = []cli.Flag{genesisFile, nodesFile, port, txSignSk, sk, profileMode, txSignSkIndex, skIndex, numOfNodes, storageCleanup}
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

	log.Info("Starting node...")

	stop := make(chan bool, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	generalConfig, err := loadMainConfig(configurationFile, log)
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Initialized with config from: %s", configurationFile))

	p2pConfig, err := core.LoadP2PConfig(p2pConfigurationFile)
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Initialized with p2p config from: %s", p2pConfigurationFile))
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

	syncer := ntp.NewSyncTime(time.Hour, beevikntp.Query)
	go syncer.StartSync()

	//TODO: The next 5 lines should be deleted when we are done testing from a precalculated (not hard coded) timestamp
	if nodesConfig.StartTime == 0 {
		time.Sleep(1000 * time.Millisecond)
		ntpTime := syncer.CurrentTime()
		nodesConfig.StartTime = (ntpTime.Unix()/60 + 1) * 60
	}

	startTime := time.Unix(nodesConfig.StartTime, 0)
	log.Info(fmt.Sprintf("Start time in seconds: %d", startTime.Unix()))

	suite, err := getSuite(generalConfig)
	if err != nil {
		return err
	}

	keyGen, privKey, pubKey, err := getSigningParams(
		ctx,
		log,
		sk.Name,
		skIndex.Name,
		initialNodesSkPemFile,
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

	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		currentNode, externalResolver, tpsBenchmark, err = createShardNode(
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
	}

	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		currentNode, externalResolver, tpsBenchmark, err = createMetaNode(
			ctx,
			generalConfig,
			nodesConfig,
			genesisConfig,
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
	}

	if currentNode == nil {
		return errors.New("node was not created")
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

func createSingleSigner(config *config.Config) (crypto.SingleSigner, error) {
	switch config.Consensus.Type {
	case blsConsensusType:
		return &singlesig.BlsSingleSigner{}, nil
	case bnConsensusType:
		return &singlesig.SchnorrSigner{}, nil
	}

	return nil, errors.New("no consensus type provided in config file")
}

func createShardNode(
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

	hasher, err := getHasherFromConfig(config)
	if err != nil {
		return nil, nil, nil, errors.New("could not create hasher: " + err.Error())
	}

	marshalizer, err := getMarshalizerFromConfig(config)
	if err != nil {
		return nil, nil, nil, errors.New("could not create marshalizer: " + err.Error())
	}

	tr, err := getTrie(config.AccountsTrieStorage, hasher)
	if err != nil {
		return nil, nil, nil, errors.New("error creating trie: " + err.Error())
	}

	addressConverter, err := addressConverters.NewPlainAddressConverter(config.Address.Length, config.Address.Prefix)
	if err != nil {
		return nil, nil, nil, errors.New("could not create address converter: " + err.Error())
	}

	accountFactory, err := factoryState.NewAccountFactoryCreator(shardCoordinator)
	if err != nil {
		return nil, nil, nil, errors.New("could not create account factory: " + err.Error())
	}

	accountsAdapter, err := state.NewAccountsDB(tr, hasher, marshalizer, accountFactory)
	if err != nil {
		return nil, nil, nil, errors.New("could not create accounts adapter: " + err.Error())
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

	transactionProcessor, err := transaction.NewTxProcessor(accountsAdapter, hasher, addressConverter, marshalizer, shardCoordinator)
	if err != nil {
		return nil, nil, nil, errors.New("could not create transaction processor: " + err.Error())
	}

	blkc, err := createBlockChainFromConfig(config)
	if err != nil {
		return nil, nil, nil, errors.New("could not create block chain: " + err.Error())
	}

	store, err := createShardDataStoreFromConfig(config)
	if err != nil {
		return nil, nil, nil, errors.New("could not create local data store: " + err.Error())
	}

	uint64ByteSliceConverter := uint64ByteSlice.NewBigEndianConverter()
	datapool, err := createShardDataPoolFromConfig(config, uint64ByteSliceConverter)
	if err != nil {
		return nil, nil, nil, errors.New("could not create shard data pools: " + err.Error())
	}

	inBalanceForShard, err := genesisConfig.InitialNodesBalances(shardCoordinator, addressConverter)
	if err != nil {
		return nil, nil, nil, errors.New("initial balances could not be processed " + err.Error())
	}

	txSingleSigner := &singlesig.SchnorrSigner{}
	singleSigner, err := createSingleSigner(config)
	if err != nil {
		return nil, nil, nil, errors.New("could not create singleSigner: " + err.Error())
	}

	multisigHasher, err := getMultisigHasherFromConfig(config)
	if err != nil {
		return nil, nil, nil, errors.New("could not create multisig hasher: " + err.Error())
	}

	currentShardPubKeys, err := nodesConfig.InitialNodesPubKeysForShard(shardCoordinator.SelfId())
	if err != nil {
		return nil, nil, nil, errors.New("could not start creation of multiSigner: " + err.Error())
	}

	multiSigner, err := createMultiSigner(config, multisigHasher, currentShardPubKeys, privKey, keyGen)
	if err != nil {
		return nil, nil, nil, err
	}

	var randReader io.Reader
	if p2pConfig.Node.Seed != "" {
		randReader = NewSeedRandReader(hasher.Compute(p2pConfig.Node.Seed))
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

	dataPacker, err := partitioning.NewSizeDataPacker(marshalizer)
	if err != nil {
		return nil, nil, nil, err
	}

	txSignKeyGen, txSignPrivKey, txSignPubKey, err := getSigningParams(
		ctx,
		log,
		txSignSk.Name,
		txSignSkIndex.Name,
		initialBalancesSkPemFile,
		kyber.NewBlakeSHA256Ed25519())

	if err != nil {
		return nil, nil, nil, err
	}

	log.Info("Starting with tx sign public key: " + getPkEncoded(txSignPubKey))

	//TODO add a real chronology validator and remove null chronology validator
	interceptorContainerFactory, err := shard.NewInterceptorsContainerFactory(
		shardCoordinator,
		netMessenger,
		store,
		marshalizer,
		hasher,
		txSignKeyGen,
		txSingleSigner,
		multiSigner,
		datapool,
		addressConverter,
		&nullChronologyValidator{},
		tpsBenchmark,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	//TODO refactor all these factory calls
	interceptorsContainer, err := interceptorContainerFactory.Create()
	if err != nil {
		return nil, nil, nil, err
	}

	resolversContainerFactory, err := shardfactoryDataRetriever.NewResolversContainerFactory(
		shardCoordinator,
		netMessenger,
		store,
		marshalizer,
		datapool,
		uint64ByteSliceConverter,
		dataPacker,
	)
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

	blockTracker, err := track.NewShardBlockTracker(datapool, marshalizer, shardCoordinator, store)
	if err != nil {
		return nil, nil, nil, err
	}

	shardsGenesisBlocks, err := generateGenesisHeadersForInit(
		nodesConfig,
		genesisConfig,
		shardCoordinator,
		addressConverter,
		hasher,
		marshalizer,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	blockProcessor, err := block.NewShardProcessor(
		datapool,
		store,
		hasher,
		marshalizer,
		transactionProcessor,
		accountsAdapter,
		shardCoordinator,
		forkDetector,
		blockTracker,
		createTxRequestHandler(resolversFinder, factory.TransactionTopic, log),
		createRequestHandler(resolversFinder, factory.MiniBlocksTopic, log),
	)
	if err != nil {
		return nil, nil, nil, errors.New("could not create block processor: " + err.Error())
	}

	err = blockProcessor.SetLastNotarizedHeadersSlice(shardsGenesisBlocks, nodesConfig.MetaChainActive)
	if err != nil {
		return nil, nil, nil, err
	}

	err = blockProcessor.SetOnRequestHeaderHandlerByNonce(createHeaderRequestHandlerByNonce(resolversFinder, factory.MetachainBlocksTopic, log))
	if err != nil {
		return nil, nil, nil, err
	}

	nd, err := node.NewNode(
		node.WithMessenger(netMessenger),
		node.WithHasher(hasher),
		node.WithMarshalizer(marshalizer),
		node.WithInitialNodesPubKeys(initialPubKeys),
		node.WithInitialNodesBalances(inBalanceForShard),
		node.WithAddressConverter(addressConverter),
		node.WithAccountsAdapter(accountsAdapter),
		node.WithBlockChain(blkc),
		node.WithDataStore(store),
		node.WithRoundDuration(nodesConfig.RoundDuration),
		node.WithConsensusGroupSize(int(nodesConfig.ConsensusGroupSize)),
		node.WithSyncer(syncer),
		node.WithBlockProcessor(blockProcessor),
		node.WithBlockTracker(blockTracker),
		node.WithGenesisTime(time.Unix(nodesConfig.StartTime, 0)),
		node.WithRounder(rounder),
		node.WithDataPool(datapool),
		node.WithShardCoordinator(shardCoordinator),
		node.WithUint64ByteSliceConverter(uint64ByteSliceConverter),
		node.WithSingleSigner(singleSigner),
		node.WithMultiSigner(multiSigner),
		node.WithKeyGen(keyGen),
		node.WithTxSignPubKey(txSignPubKey),
		node.WithTxSignPrivKey(txSignPrivKey),
		node.WithPubKey(pubKey),
		node.WithPrivKey(privKey),
		node.WithForkDetector(forkDetector),
		node.WithInterceptorsContainer(interceptorsContainer),
		node.WithResolversFinder(resolversFinder),
		node.WithConsensusType(config.Consensus.Type),
		node.WithTxSingleSigner(txSingleSigner),
		node.WithActiveMetachain(nodesConfig.MetaChainActive),
		node.WithTxStorageSize(config.TxStorage.Cache.Size),
	)
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

	externalResolver, err := external.NewExternalResolver(
		shardCoordinator,
		blkc,
		store,
		marshalizer,
		&mockProposerResolver{},
	)
	if err != nil {
		return nil, nil, nil, err
	}

	err = nd.CreateShardGenesisBlock()
	if err != nil {
		return nil, nil, nil, err
	}

	return nd, externalResolver, tpsBenchmark, nil
}

func createMetaNode(
	ctx *cli.Context,
	config *config.Config,
	nodesConfig *sharding.NodesSetup,
	genesisConfig *sharding.Genesis,
	p2pConfig *config.P2PConfig,
	syncer ntp.SyncTimer,
	keyGen crypto.KeyGenerator,
	privKey crypto.PrivateKey,
	pubKey crypto.PublicKey,
	shardCoordinator sharding.Coordinator,
	log *logger.Logger,
) (*node.Node, *external.ExternalResolver, *statistics.TpsBenchmark, error) {

	hasher, err := getHasherFromConfig(config)
	if err != nil {
		return nil, nil, nil, errors.New("could not create hasher: " + err.Error())
	}

	marshalizer, err := getMarshalizerFromConfig(config)
	if err != nil {
		return nil, nil, nil, errors.New("could not create marshalizer: " + err.Error())
	}

	tr, err := getTrie(config.AccountsTrieStorage, hasher)
	if err != nil {
		return nil, nil, nil, errors.New("error creating trie: " + err.Error())
	}

	addressConverter, err := addressConverters.NewPlainAddressConverter(config.Address.Length, config.Address.Prefix)
	if err != nil {
		return nil, nil, nil, errors.New("could not create address converter: " + err.Error())
	}

	accountFactory, err := factoryState.NewAccountFactoryCreator(shardCoordinator)
	if err != nil {
		return nil, nil, nil, errors.New("could not create account factory: " + err.Error())
	}

	accountsAdapter, err := state.NewAccountsDB(tr, hasher, marshalizer, accountFactory)
	if err != nil {
		return nil, nil, nil, errors.New("could not create accounts adapter: " + err.Error())
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

	metaChain, err := createMetaChainFromConfig(config)
	if err != nil {
		return nil, nil, nil, errors.New("could not create block chain: " + err.Error())
	}

	metaStore, err := createMetaChainDataStoreFromConfig(config)
	if err != nil {
		return nil, nil, nil, errors.New("could not create local data store: " + err.Error())
	}

	uint64ByteSliceConverter := uint64ByteSlice.NewBigEndianConverter()
	metaDatapool, err := createMetaDataPoolFromConfig(config, uint64ByteSliceConverter)
	if err != nil {
		return nil, nil, nil, errors.New("could not create shard data pools: " + err.Error())
	}

	txSingleSigner := &singlesig.SchnorrSigner{}
	singleSigner, err := createSingleSigner(config)
	if err != nil {
		return nil, nil, nil, errors.New("could not create singleSigner: " + err.Error())
	}

	multisigHasher, err := getMultisigHasherFromConfig(config)
	if err != nil {
		return nil, nil, nil, errors.New("could not create multisig hasher: " + err.Error())
	}

	currentShardPubKeys, err := nodesConfig.InitialNodesPubKeysForShard(shardCoordinator.SelfId())
	if err != nil {
		return nil, nil, nil, errors.New("could not start creation of multiSigner: " + err.Error())
	}

	multiSigner, err := createMultiSigner(config, multisigHasher, currentShardPubKeys, privKey, keyGen)
	if err != nil {
		return nil, nil, nil, err
	}

	var randReader io.Reader
	if p2pConfig.Node.Seed != "" {
		randReader = NewSeedRandReader(hasher.Compute(p2pConfig.Node.Seed))
	} else {
		randReader = rand.Reader
	}

	netMessenger, err := createNetMessenger(p2pConfig, log, randReader)
	if err != nil {
		return nil, nil, nil, err
	}

	_, txSignPrivKey, txSignPubKey, err := getSigningParams(
		ctx,
		log,
		txSignSk.Name,
		txSignSkIndex.Name,
		initialBalancesSkPemFile,
		kyber.NewBlakeSHA256Ed25519())

	if err != nil {
		return nil, nil, nil, err
	}

	tpsBenchmark, err := statistics.NewTPSBenchmark(shardCoordinator.NumberOfShards(), nodesConfig.RoundDuration/1000)
	if err != nil {
		return nil, nil, nil, err
	}

	log.Info("Starting with tx sign public key: " + getPkEncoded(txSignPubKey))

	//TODO add a real chronology validator and remove null chronology validator
	interceptorContainerFactory, err := metachain.NewInterceptorsContainerFactory(
		shardCoordinator,
		netMessenger,
		metaStore,
		marshalizer,
		hasher,
		multiSigner,
		metaDatapool,
		&nullChronologyValidator{},
		tpsBenchmark,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	//TODO refactor all these factory calls
	interceptorsContainer, err := interceptorContainerFactory.Create()
	if err != nil {
		return nil, nil, nil, err
	}

	resolversContainerFactory, err := metafactoryDataRetriever.NewResolversContainerFactory(
		shardCoordinator,
		netMessenger,
		metaStore,
		marshalizer,
		metaDatapool,
		uint64ByteSliceConverter,
	)
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

	blockTracker, err := track.NewMetaBlockTracker()
	if err != nil {
		return nil, nil, nil, err
	}

	shardsGenesisBlocks, err := generateGenesisHeadersForInit(
		nodesConfig,
		genesisConfig,
		shardCoordinator,
		addressConverter,
		hasher,
		marshalizer,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	metaProcessor, err := block.NewMetaProcessor(
		accountsAdapter,
		metaDatapool,
		forkDetector,
		shardCoordinator,
		hasher,
		marshalizer,
		metaStore,
		createRequestHandler(resolversFinder, factory.ShardHeadersForMetachainTopic, log),
	)
	if err != nil {
		return nil, nil, nil, errors.New("could not create block processor: " + err.Error())
	}

	err = metaProcessor.SetLastNotarizedHeadersSlice(shardsGenesisBlocks, nodesConfig.MetaChainActive)
	if err != nil {
		return nil, nil, nil, err
	}

	err = metaProcessor.SetOnRequestHeaderHandlerByNonce(createHeaderRequestHandlerByNonce(resolversFinder, factory.ShardHeadersForMetachainTopic, log))
	if err != nil {
		return nil, nil, nil, err
	}

	nd, err := node.NewNode(
		node.WithMessenger(netMessenger),
		node.WithHasher(hasher),
		node.WithMarshalizer(marshalizer),
		node.WithInitialNodesPubKeys(initialPubKeys),
		node.WithAddressConverter(addressConverter),
		node.WithAccountsAdapter(accountsAdapter),
		node.WithBlockChain(metaChain),
		node.WithDataStore(metaStore),
		node.WithRoundDuration(nodesConfig.RoundDuration),
		node.WithConsensusGroupSize(int(nodesConfig.MetaChainConsensusGroupSize)),
		node.WithSyncer(syncer),
		node.WithBlockProcessor(metaProcessor),
		node.WithBlockTracker(blockTracker),
		node.WithGenesisTime(time.Unix(nodesConfig.StartTime, 0)),
		node.WithRounder(rounder),
		node.WithMetaDataPool(metaDatapool),
		node.WithShardCoordinator(shardCoordinator),
		node.WithUint64ByteSliceConverter(uint64ByteSliceConverter),
		node.WithSingleSigner(singleSigner),
		node.WithMultiSigner(multiSigner),
		node.WithKeyGen(keyGen),
		node.WithTxSignPubKey(txSignPubKey),
		node.WithTxSignPrivKey(txSignPrivKey),
		node.WithPubKey(pubKey),
		node.WithPrivKey(privKey),
		node.WithForkDetector(forkDetector),
		node.WithInterceptorsContainer(interceptorsContainer),
		node.WithResolversFinder(resolversFinder),
		node.WithConsensusType(config.Consensus.Type),
		node.WithTxSingleSigner(txSingleSigner),
		node.WithTxStorageSize(config.TxStorage.Cache.Size),
	)
	if err != nil {
		return nil, nil, nil, errors.New("error creating meta-node: " + err.Error())
	}

	externalResolver, err := external.NewExternalResolver(
		shardCoordinator,
		metaChain,
		metaStore,
		marshalizer,
		&mockProposerResolver{},
	)
	if err != nil {
		return nil, nil, nil, err
	}

	err = nd.CreateMetaGenesisBlock()
	if err != nil {
		return nil, nil, nil, err
	}

	return nd, externalResolver, tpsBenchmark, nil
}

func createTxRequestHandler(resolversFinder dataRetriever.ResolversFinder, baseTopic string, log *logger.Logger) func(destShardID uint32, txHashes [][]byte) {
	return func(destShardID uint32, txHashes [][]byte) {
		log.Debug(fmt.Sprintf("Requesting %d transactions from shard %d from network...\n", len(txHashes), destShardID))
		resolver, err := resolversFinder.CrossShardResolver(baseTopic, destShardID)
		if err != nil {
			log.Error(fmt.Sprintf("missing resolver to %s topic to shard %d", baseTopic, destShardID))
			return
		}

		txResolver, ok := resolver.(*resolvers.TxResolver)
		if !ok {
			log.Error("wrong assertion type when creating transaction resolver")
			return
		}

		go func() {
			dataSplit := &partitioning.DataSplit{}
			sliceBatches, err := dataSplit.SplitDataInChunks(txHashes, maxTxsToRequest)
			if err != nil {
				log.Error("error requesting transactions: " + err.Error())
				return
			}

			for _, batch := range sliceBatches {
				err = txResolver.RequestDataFromHashArray(batch)
				if err != nil {
					log.Debug("error requesting tx batch: " + err.Error())
				}
			}
		}()
	}
}

func createRequestHandler(resolversFinder dataRetriever.ResolversFinder, baseTopic string, log *logger.Logger) func(destShardID uint32, txHash []byte) {
	return func(destShardID uint32, txHash []byte) {
		log.Debug(fmt.Sprintf("Requesting %s from shard %d with hash %s from network\n", baseTopic, destShardID, toB64(txHash)))
		resolver, err := resolversFinder.CrossShardResolver(baseTopic, destShardID)
		if err != nil {
			log.Error(fmt.Sprintf("missing resolver to %s topic to shard %d", baseTopic, destShardID))
			return
		}

		err = resolver.RequestDataFromHash(txHash)
		if err != nil {
			log.Debug(err.Error())
		}
	}
}

func createHeaderRequestHandlerByNonce(resolversFinder dataRetriever.ResolversFinder, baseTopic string, log *logger.Logger) func(destShardID uint32, nonce uint64) {
	return func(destShardID uint32, nonce uint64) {
		log.Debug(fmt.Sprintf("Requesting %s from shard %d with nonce %d from network\n", baseTopic, destShardID, nonce))
		resolver, err := resolversFinder.CrossShardResolver(baseTopic, destShardID)
		if err != nil {
			log.Error(fmt.Sprintf("missing resolver to %s topic to shard %d", baseTopic, destShardID))
			return
		}

		headerResolver, ok := resolver.(*resolvers.HeaderResolver)
		if !ok {
			log.Error(fmt.Sprintf("resolver is not a header resolverto %s topic to shard %d", baseTopic, destShardID))
		}

		headerResolver.RequestDataFromNonce(nonce)
		if err != nil {
			log.Debug(err.Error())
		}
	}
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

func getHasherFromConfig(cfg *config.Config) (hashing.Hasher, error) {
	switch cfg.Hasher.Type {
	case "sha256":
		return sha256.Sha256{}, nil
	case "blake2b":
		return blake2b.Blake2b{}, nil
	}

	return nil, errors.New("no hasher provided in config file")
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

func getMarshalizerFromConfig(cfg *config.Config) (marshal.Marshalizer, error) {
	switch cfg.Marshalizer.Type {
	case "json":
		return marshal.JsonMarshalizer{}, nil
	}

	return nil, errors.New("no marshalizer provided in config file")
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

func createBlockChainFromConfig(config *config.Config) (data.ChainHandler, error) {
	badBlockCache, err := storageUnit.NewCache(
		storageUnit.CacheType(config.BadBlocksCache.Type),
		config.BadBlocksCache.Size,
		config.BadBlocksCache.Shards)
	if err != nil {
		return nil, err
	}

	blockChain, err := blockchain.NewBlockChain(
		badBlockCache,
	)
	if err != nil {
		return nil, err
	}

	return blockChain, err
}

func createShardDataStoreFromConfig(config *config.Config) (dataRetriever.StorageService, error) {
	var headerUnit, peerBlockUnit, miniBlockUnit, txUnit, metachainHeaderUnit *storageUnit.Unit
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

	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, txUnit)
	store.AddStorer(dataRetriever.MiniBlockUnit, miniBlockUnit)
	store.AddStorer(dataRetriever.PeerChangesUnit, peerBlockUnit)
	store.AddStorer(dataRetriever.BlockHeaderUnit, headerUnit)
	store.AddStorer(dataRetriever.MetaBlockUnit, metachainHeaderUnit)

	return store, err
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

	return dataPool.NewMetaDataPool(metaBlockBody, miniBlockHashes, shardHeaders, metaBlockNonces)
}

func createMetaChainFromConfig(config *config.Config) (*blockchain.MetaChain, error) {
	badBlockCache, err := storageUnit.NewCache(
		storageUnit.CacheType(config.BadBlocksCache.Type),
		config.BadBlocksCache.Size,
		config.BadBlocksCache.Shards)
	if err != nil {
		return nil, err
	}

	metaChain, err := blockchain.NewMetaChain(
		badBlockCache,
	)
	if err != nil {
		return nil, err
	}

	return metaChain, err
}

func createMetaChainDataStoreFromConfig(config *config.Config) (dataRetriever.StorageService, error) {
	var peerDataUnit, shardDataUnit, metaBlockUnit, headerUnit *storageUnit.Unit
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

	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.MetaBlockUnit, metaBlockUnit)
	store.AddStorer(dataRetriever.MetaShardDataUnit, shardDataUnit)
	store.AddStorer(dataRetriever.MetaPeerDataUnit, peerDataUnit)
	store.AddStorer(dataRetriever.BlockHeaderUnit, headerUnit)

	return store, err
}

func decodeAddress(address string) ([]byte, error) {
	return hex.DecodeString(address)
}

func encodeAddress(address []byte) string {
	return hex.EncodeToString(address)
}

func toB64(buff []byte) string {
	if buff == nil {
		return "<NIL>"
	}
	return base64.StdEncoding.EncodeToString(buff)
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
