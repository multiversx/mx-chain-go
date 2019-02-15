package main

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/cmd/facade"
	"github.com/ElrondNetwork/elrond-go-sandbox/cmd/flags"
	"github.com/ElrondNetwork/elrond-go-sandbox/config"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kv2/multisig"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/interceptor"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/resolver"
	sync2 "github.com/ElrondNetwork/elrond-go-sandbox/process/sync"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	beevikntp "github.com/beevik/ntp"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kv2/singlesig"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kv2"
)

var bootNodeHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}
   {{if len .Authors}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}{{if .Copyright }}
VERSION:
   {{.Version}}
   {{end}}
`
var configurationFile = "./config/config.testnet.json"
var uniqueID = ""

type initialNode struct {
	Address string `json:"address"`
	PubKey  string `json:"pubkey"`
	Balance string `json:"balance"`
}

type genesis struct {
	StartTime          int64         `json:"startTime"`
	RoundDuration      uint64        `json:"roundDuration"`
	ConsensusGroupSize int           `json:"consensusGroupSize"`
	ElasticSubrounds   bool          `json:"elasticSubrounds"`
	InitialNodes       []initialNode `json:"initialNodes"`
}

type netMessengerConfig struct {
	ctx             context.Context
	port            int
	maxAllowedPeers int
	marshalizer     marshal.Marshalizer
	hasher          hashing.Hasher
	pubSubStrategy  p2p.PubSubStrategy
}

func main() {
	log := logger.NewDefaultLogger()
	log.SetLevel(logger.LogInfo)

	app := cli.NewApp()
	cli.AppHelpTemplate = bootNodeHelpTemplate
	app.Name = "BootNode CLI App"
	app.Usage = "This is the entry point for starting a new bootstrap node - the app will start after the genesis timestamp"
	app.Flags = []cli.Flag{flags.GenesisFile, flags.Port, flags.MaxAllowedPeers, flags.PrivateKey}
	app.Action = func(c *cli.Context) error {
		return startNode(c, log)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startNode(ctx *cli.Context, log *logger.Logger) error {
	log.Info("Starting node...")

	stop := make(chan bool, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	generalConfig, err := loadMainConfig(configurationFile, log)
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Initialized with config from: %s", configurationFile))

	genesisConfig, err := loadGenesisConfiguration(ctx.GlobalString(flags.GenesisFile.Name), log)
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Initialized with genesis config from: %s", ctx.GlobalString(flags.GenesisFile.Name)))

	syncer := ntp.NewSyncTime(time.Hour, beevikntp.Query)
	go syncer.StartSync()

	// TODO: The next 5 lines should be deleted when we are done testing from a precalculated (not hard coded)
	//  timestamp
	if genesisConfig.StartTime == 0 {
		time.Sleep(1000 * time.Millisecond)
		ntpTime := syncer.CurrentTime(syncer.ClockOffset())
		genesisConfig.StartTime = (ntpTime.Unix()/60 + 1) * 60
	}

	startTime := time.Unix(genesisConfig.StartTime, 0)
	log.Info(fmt.Sprintf("Start time in seconds: %d", startTime.Unix()))

	uniqueID = fmt.Sprintf("%d", ctx.GlobalInt(flags.Port.Name))

	currentNode, err := createNode(ctx, generalConfig, genesisConfig, syncer, log)

	if err != nil {
		return err
	}

	ef := facade.NewElrondNodeFacade(currentNode)

	ef.SetLogger(log)
	ef.SetSyncer(syncer)

	wg := sync.WaitGroup{}
	go ef.StartBackgroundServices(&wg)
	wg.Wait()

	if !ctx.Bool(flags.WithUI.Name) {
		log.Info("Bootstrapping node....")
		err = ef.StartNode()
		if err != nil {
			log.Error("starting node failed", err.Error())
		}
	}

	go func() {
		<-sigs
		log.Info("terminating at user's signal...")
		stop <- true
	}()

	log.Info("Application is now running...")
	<-stop

	return nil
}

func loadFile(dest interface{}, relativePath string, log *logger.Logger) error {
	path, err := filepath.Abs(relativePath)
	fmt.Println(path)
	if err != nil {
		log.Error("cannot create absolute path for the provided file", err.Error())
		return err
	}
	f, err := os.Open(path)
	defer func() {
		err = f.Close()
		if err != nil {
			log.Error("cannot close file: ", err.Error())
		}
	}()
	if err != nil {
		return err
	}

	jsonParser := json.NewDecoder(f)
	err = jsonParser.Decode(dest)
	if err != nil {
		return err
	}
	return nil
}

func loadMainConfig(filepath string, log *logger.Logger) (*config.Config, error) {
	cfg := &config.Config{}
	err := loadFile(cfg, filepath, log)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func loadGenesisConfiguration(genesisFilePath string, log *logger.Logger) (*genesis, error) {
	cfg := &genesis{}
	err := loadFile(cfg, genesisFilePath, log)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (g *genesis) initialNodesPubkeys(log *logger.Logger) []string {
	var pubKeys []string
	for _, in := range g.InitialNodes {
		pubKey, err := decodeAddress(in.PubKey)

		if err != nil {
			log.Error(fmt.Sprintf("%s is not a valid public key. Ignored", in))
			continue
		}

		pubKeys = append(pubKeys, string(pubKey))
	}
	return pubKeys
}

func (g *genesis) initialNodesBalances(log *logger.Logger) map[string]*big.Int {
	var pubKeys = make(map[string]*big.Int)
	for _, in := range g.InitialNodes {
		balance, ok := new(big.Int).SetString(in.Balance, 10)
		if ok {
			pubKey, err := decodeAddress(in.PubKey)
			if err != nil {
				log.Error(fmt.Sprintf("%s is not a valid public key. Ignored", in.PubKey))
				continue
			}
			pubKeys[string(pubKey)] = balance
		} else {
			log.Warn(fmt.Sprintf("error decoding balance %s for public key %s - setting to 0", in.Balance, in.PubKey))
			pubKeys[in.PubKey] = big.NewInt(0)
		}

	}
	return pubKeys
}

func createNode(ctx *cli.Context, cfg *config.Config, genesisConfig *genesis, syncer ntp.SyncTimer, log *logger.Logger) (*node.Node, error) {
	appContext := context.Background()

	hasher, err := getHasherFromConfig(cfg)
	if err != nil {
		return nil, errors.New("could not create hasher: " + err.Error())
	}

	marshalizer, err := getMarshalizerFromConfig(cfg)
	if err != nil {
		return nil, errors.New("could not create marshalizer: " + err.Error())
	}

	tr, err := getTrie(cfg.AccountsTrieStorage, hasher)
	if err != nil {
		return nil, errors.New("error creating node: " + err.Error())
	}

	addressConverter, err := state.NewPlainAddressConverter(cfg.Address.Length, cfg.Address.Prefix)
	if err != nil {
		return nil, errors.New("could not create address converter: " + err.Error())
	}

	accountsAdapter, err := state.NewAccountsDB(tr, hasher, marshalizer)
	if err != nil {
		return nil, errors.New("could not create accounts adapter: " + err.Error())
	}

	blkc, err := createBlockChainFromConfig(cfg)
	if err != nil {
		return nil, errors.New("could not create block chain: " + err.Error())
	}

	transactionProcessor, err := transaction.NewTxProcessor(accountsAdapter, hasher, addressConverter, marshalizer)
	if err != nil {
		return nil, errors.New("could not create transaction processor: " + err.Error())
	}

	uint64ByteSliceConverter := uint64ByteSlice.NewBigEndianConverter()

	datapool, err := createDataPoolFromConfig(cfg, uint64ByteSliceConverter)
	if err != nil {
		return nil, errors.New("could not create transient data pool: " + err.Error())
	}

	shardCoordinator := &sharding.OneShardCoordinator{}

	initialPubKeys := genesisConfig.initialNodesPubkeys(log)

	keyGen, privKey, pubKey, err := getSigningParams(ctx, log)

	if err != nil {
		return nil, err
	}

	singlesigner := &singlesig.SchnorrSigner{}

	multisigner, err := multisig.NewBelNevMultisig(hasher, initialPubKeys, privKey, keyGen, uint16(0))

	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	netMessenger, err := createNetMessenger(netMessengerConfig{
		ctx:             appContext,
		port:            ctx.GlobalInt(flags.Port.Name),
		maxAllowedPeers: ctx.GlobalInt(flags.MaxAllowedPeers.Name),
		marshalizer:     marshalizer,
		hasher:          hasher,
		pubSubStrategy:  p2p.GossipSub,
	})
	if err != nil {
		return nil, err
	}

	interceptorsContainer := interceptor.NewContainer()
	resolversContainer := resolver.NewContainer()

	processorFactory, err := factory.NewProcessorsCreator(factory.ProcessorsCreatorConfig{
		InterceptorContainer:     interceptorsContainer,
		ResolverContainer:        resolversContainer,
		Messenger:                netMessenger,
		Blockchain:               blkc,
		DataPool:                 datapool,
		ShardCoordinator:         shardCoordinator,
		AddrConverter:            addressConverter,
		Hasher:                   hasher,
		Marshalizer:              marshalizer,
		MultiSigner:              multisigner,
		SingleSigner:             singlesigner,
		KeyGen:                   keyGen,
		Uint64ByteSliceConverter: uint64ByteSliceConverter,
	})
	if err != nil {
		return nil, err
	}

	err = processorFactory.CreateInterceptors()
	if err != nil {
		return nil, err
	}

	err = processorFactory.CreateResolvers()
	if err != nil {
		return nil, err
	}

	forkDetector := sync2.NewBasicForkDetector()

	res, err := processorFactory.ResolverContainer().Get(string(factory.TransactionTopic))
	if err != nil {
		return nil, err
	}
	txResolver, ok := res.(*transaction.TxResolver)
	if !ok {
		return nil, errors.New("tx resolver is not of type transaction.TxResolver")
	}

	blockProcessor, err := block.NewBlockProcessor(
		datapool,
		hasher,
		marshalizer,
		transactionProcessor,
		accountsAdapter,
		shardCoordinator,
		forkDetector,
		createRequestTransactionHandler(txResolver, log),
	)

	if err != nil {
		return nil, errors.New("could not create block processor: " + err.Error())
	}

	nd, err := node.NewNode(
		node.WithMessenger(netMessenger),
		node.WithHasher(hasher),
		node.WithContext(appContext),
		node.WithMarshalizer(marshalizer),
		node.WithInitialNodesPubKeys(initialPubKeys),
		node.WithInitialNodesBalances(genesisConfig.initialNodesBalances(log)),
		node.WithAddressConverter(addressConverter),
		node.WithAccountsAdapter(accountsAdapter),
		node.WithBlockChain(blkc),
		node.WithRoundDuration(genesisConfig.RoundDuration),
		node.WithConsensusGroupSize(genesisConfig.ConsensusGroupSize),
		node.WithSyncer(syncer),
		node.WithBlockProcessor(blockProcessor),
		node.WithGenesisTime(time.Unix(genesisConfig.StartTime, 0)),
		node.WithElasticSubrounds(genesisConfig.ElasticSubrounds),
		node.WithDataPool(datapool),
		node.WithShardCoordinator(shardCoordinator),
		node.WithUint64ByteSliceConverter(uint64ByteSliceConverter),
		node.WithSinglesig(singlesigner),
		node.WithMultisig(multisigner),
		node.WithKeyGenerator(keyGen),
		node.WithPublicKey(pubKey),
		node.WithPrivateKey(privKey),
		node.WithForkDetector(forkDetector),
		node.WithProcessorCreator(processorFactory),
	)

	if err != nil {
		return nil, errors.New("error creating node: " + err.Error())
	}

	err = nd.CreateShardedStores()
	if err != nil {
		return nil, err
	}

	return nd, nil
}

func createRequestTransactionHandler(txResolver *transaction.TxResolver, log *logger.Logger) func(destShardID uint32, txHash []byte) {
	return func(destShardID uint32, txHash []byte) {
		_ = txResolver.RequestTransactionFromHash(txHash)
		log.Debug(fmt.Sprintf("Requested tx for shard %d with hash %s from network\n", destShardID, toB64(txHash)))
	}
}

func createNetMessenger(config netMessengerConfig) (p2p.Messenger, error) {
	if config.port == 0 {
		return nil, errors.New("cannot start node on port 0")
	}

	if config.maxAllowedPeers == 0 {
		return nil, errors.New("cannot start node without providing maxAllowedPeers")
	}

	//TODO check if libp2p provides a better random source
	cp := &p2p.ConnectParams{}
	cp.Port = config.port
	cp.GeneratePrivPubKeys(time.Now().UnixNano())
	cp.GenerateIDFromPubKey()

	nm, err := p2p.NewNetMessenger(config.ctx, config.marshalizer, config.hasher, cp, config.maxAllowedPeers, config.pubSubStrategy)
	if err != nil {
		return nil, err
	}
	return nm, nil
}

func getSk(ctx *cli.Context) ([]byte, error) {
	if !ctx.GlobalIsSet(flags.PrivateKey.Name) {
		if ctx.GlobalString(flags.PrivateKey.Name) == "" {
			return nil, errors.New("no private key file provided")
		}
	}

	encodedSk, err := ioutil.ReadFile(ctx.GlobalString(flags.PrivateKey.Name))
	if err != nil {
		encodedSk = []byte(ctx.GlobalString(flags.PrivateKey.Name))
	}
	return decodeAddress(string(encodedSk))
}

func getSigningParams(ctx *cli.Context, log *logger.Logger) (
	keyGen crypto.KeyGenerator,
	privKey crypto.PrivateKey,
	pubKey crypto.PublicKey,
	err error,
) {
	sk, err := getSk(ctx)

	if err != nil {
		return nil, nil, nil, err
	}

	suite := kv2.NewBlakeSHA256Ed25519()
	keyGen = signing.NewKeyGenerator(suite)
	privKey, err = keyGen.PrivateKeyFromByteArray(sk)

	if err != nil {
		return nil, nil, nil, err
	}

	pubKey = privKey.GeneratePublic()

	pk, _ := pubKey.ToByteArray()

	skEncoded := encodeAddress(sk)
	pkEncoded := encodeAddress(pk)

	log.Info("starting with private key: " + skEncoded)
	log.Info("starting with public key: " + pkEncoded)

	return keyGen, privKey, pubKey, err
}

func getTrie(cfg config.StorageConfig, hasher hashing.Hasher) (*trie.Trie, error) {
	accountsTrieStorage, err := storage.NewStorageUnitFromConf(
		getCacherFromConfig(cfg.Cache),
		getDBFromConfig(cfg.DB),
		getBloomFromConfig(cfg.Bloom),
	)
	if err != nil {
		return nil, errors.New("error creating node: " + err.Error())
	}

	dbWriteCache, err := trie.NewDBWriteCache(accountsTrieStorage)
	if err != nil {
		return nil, errors.New("error creating node: " + err.Error())
	}

	return trie.NewTrie(make([]byte, 32), dbWriteCache, hasher)
}

func getHasherFromConfig(cfg *config.Config) (hashing.Hasher, error) {
	switch cfg.Hasher.Type {
	case "sha256":
		return sha256.Sha256{}, nil
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

func getCacherFromConfig(cfg config.CacheConfig) storage.CacheConfig {
	return storage.CacheConfig{
		Size: cfg.Size,
		Type: storage.CacheType(cfg.Type),
	}
}

func getDBFromConfig(cfg config.DBConfig) storage.DBConfig {
	return storage.DBConfig{
		FilePath: filepath.Join(config.DefaultPath()+uniqueID, cfg.FilePath),
		Type:     storage.DBType(cfg.Type),
	}
}

func getBloomFromConfig(cfg config.BloomFilterConfig) storage.BloomConfig {
	var hashFuncs []storage.HasherType
	if cfg.HashFunc != nil {
		hashFuncs = make([]storage.HasherType, 0)
		for _, hf := range cfg.HashFunc {
			hashFuncs = append(hashFuncs, storage.HasherType(hf))
		}
	}

	return storage.BloomConfig{
		Size:     cfg.Size,
		HashFunc: hashFuncs,
	}
}

func createDataPoolFromConfig(config *config.Config, uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter) (data.TransientDataHolder, error) {
	txPool, err := shardedData.NewShardedData(getCacherFromConfig(config.TxDataPool))
	if err != nil {
		return nil, err
	}

	hdrPool, err := shardedData.NewShardedData(getCacherFromConfig(config.BlockHeaderDataPool))
	if err != nil {
		return nil, err
	}

	cacherCfg := getCacherFromConfig(config.BlockHeaderNoncesDataPool)
	hdrNoncesCacher, err := storage.NewCache(cacherCfg.Type, cacherCfg.Size)
	if err != nil {
		return nil, err
	}
	hdrNonces, err := dataPool.NewNonceToHashCacher(hdrNoncesCacher, uint64ByteSliceConverter)
	if err != nil {
		return nil, err
	}

	cacherCfg = getCacherFromConfig(config.TxBlockBodyDataPool)
	txBlockBody, err := storage.NewCache(cacherCfg.Type, cacherCfg.Size)
	if err != nil {
		return nil, err
	}

	cacherCfg = getCacherFromConfig(config.PeerBlockBodyDataPool)
	peerChangeBlockBody, err := storage.NewCache(cacherCfg.Type, cacherCfg.Size)
	if err != nil {
		return nil, err
	}

	cacherCfg = getCacherFromConfig(config.StateBlockBodyDataPool)
	stateBlockBody, err := storage.NewCache(cacherCfg.Type, cacherCfg.Size)
	if err != nil {
		return nil, err
	}

	return dataPool.NewDataPool(
		txPool,
		hdrPool,
		hdrNonces,
		txBlockBody,
		peerChangeBlockBody,
		stateBlockBody,
	)
}

func createBlockChainFromConfig(config *config.Config) (*blockchain.BlockChain, error) {
	var headerUnit, peerBlockUnit, stateBlockUnit, txBlockUnit, txUnit *storage.Unit
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
			if stateBlockUnit != nil {
				_ = stateBlockUnit.DestroyUnit()
			}
			if txBlockUnit != nil {
				_ = txBlockUnit.DestroyUnit()
			}
			if txUnit != nil {
				_ = txUnit.DestroyUnit()
			}
		}
	}()

	badBlockCache, err := storage.NewCache(
		storage.CacheType(config.BadBlocksCache.Type),
		config.BadBlocksCache.Size)

	if err != nil {
		return nil, err
	}

	txUnit, err = storage.NewStorageUnitFromConf(
		getCacherFromConfig(config.TxStorage.Cache),
		getDBFromConfig(config.TxStorage.DB),
		getBloomFromConfig(config.TxStorage.Bloom))

	if err != nil {
		return nil, err
	}

	txBlockUnit, err = storage.NewStorageUnitFromConf(
		getCacherFromConfig(config.TxBlockBodyStorage.Cache),
		getDBFromConfig(config.TxBlockBodyStorage.DB),
		getBloomFromConfig(config.TxBlockBodyStorage.Bloom))

	if err != nil {
		return nil, err
	}

	stateBlockUnit, err = storage.NewStorageUnitFromConf(
		getCacherFromConfig(config.StateBlockBodyStorage.Cache),
		getDBFromConfig(config.StateBlockBodyStorage.DB),
		getBloomFromConfig(config.StateBlockBodyStorage.Bloom))

	if err != nil {
		return nil, err
	}

	peerBlockUnit, err = storage.NewStorageUnitFromConf(
		getCacherFromConfig(config.PeerBlockBodyStorage.Cache),
		getDBFromConfig(config.PeerBlockBodyStorage.DB),
		getBloomFromConfig(config.PeerBlockBodyStorage.Bloom))

	if err != nil {
		return nil, err
	}

	headerUnit, err = storage.NewStorageUnitFromConf(
		getCacherFromConfig(config.BlockHeaderStorage.Cache),
		getDBFromConfig(config.BlockHeaderStorage.DB),
		getBloomFromConfig(config.BlockHeaderStorage.Bloom))

	if err != nil {
		return nil, err
	}

	blockChain, err := blockchain.NewBlockChain(
		badBlockCache,
		txUnit,
		txBlockUnit,
		stateBlockUnit,
		peerBlockUnit,
		headerUnit)

	if err != nil {
		return nil, err
	}

	return blockChain, err
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
