package main

import (
	"context"
	"encoding/base64"
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
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/multisig"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/schnorr"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	beevikntp "github.com/beevik/ntp"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
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
	RoundDuration      int64         `json:"roundDuration"`
	ConsensusGroupSize int           `json:"consensusGroupSize"`
	ElasticSubrounds   bool          `json:"elasticSubrounds"`
	InitialNodes       []initialNode `json:"initialNodes"`
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

	syncer := ntp.NewSyncTime(time.Millisecond*time.Duration(genesisConfig.RoundDuration), beevikntp.Query)
	go syncer.StartSync()

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
			log.Error("Starting node failed", err.Error())
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
		log.Error("Cannot create absolute path for the provided file", err.Error())
		return err
	}
	f, err := os.Open(path)
	defer func() {
		err = f.Close()
		if err != nil {
			log.Error("Cannot close file: ", err.Error())
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

func (g *genesis) initialNodesPubkeys() []string {
	var pubKeys []string
	for _, in := range g.InitialNodes {
		pubKeys = append(pubKeys, in.PubKey)
	}
	return pubKeys
}

func (g *genesis) initialNodesBalances(log *logger.Logger) map[string]big.Int {
	var pubKeys = make(map[string]big.Int)
	for _, in := range g.InitialNodes {
		balance, ok := new(big.Int).SetString(in.Balance, 10)
		if ok {
			pubKeys[in.PubKey] = *balance
		} else {
			log.Warn(fmt.Sprintf("Error decoding balance %s for public key %s - setting to 0", in.Balance, in.PubKey))
			pubKeys[in.PubKey] = *big.NewInt(0)
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

	addressConverter, err := state.NewHashAddressConverter(hasher, cfg.Address.Length, cfg.Address.Prefix)
	if err != nil {
		return nil, errors.New("could not create address converter: " + err.Error())
	}

	accountsAdapter, err := state.NewAccountsDB(tr, hasher, marshalizer)
	if err != nil {
		return nil, errors.New("could not create accounts adapter: " + err.Error())
	}

	blkc, err := createBlockChainFromConfig(blockChainConfig())
	if err != nil {
		return nil, errors.New("could not create block chain: " + err.Error())
	}

	transactionProcessor, err := transaction.NewTxProcessor(accountsAdapter, hasher, addressConverter, marshalizer)
	if err != nil {
		return nil, errors.New("could not create transaction processor: " + err.Error())
	}

	txPoolCacher := getCacherFromConfig(cfg.TxPoolStorage)
	shrdData, err := shardedData.NewShardedData(txPoolCacher)
	if err != nil {
		return nil, errors.New("could not create sharded data: " + err.Error())
	}

	blockProcessor := block.NewBlockProcessor(shrdData, hasher, marshalizer, transactionProcessor, accountsAdapter, &sharding.OneShardCoordinator{})

	if err != nil {
		return nil, errors.New("error creating node: " + err.Error())
	}

	sk, err := getSk(ctx)

	if err != nil {
		return nil, err
	}

	initialPubKeys := genesisConfig.initialNodesPubkeys()

	keyGen, privKey, pubKey, err := loadSkPk(sk, log)

	multisigner, err := multisig.NewBelNevMultisig(hasher, initialPubKeys, privKey, uint16(0))

	if err != nil {
		return nil, err
	}

	nd, err := node.NewNode(
		node.WithHasher(hasher),
		node.WithContext(appContext),
		node.WithMarshalizer(marshalizer),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithMaxAllowedPeers(ctx.GlobalInt(flags.MaxAllowedPeers.Name)),
		node.WithPort(ctx.GlobalInt(flags.Port.Name)),
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
		node.WithMultisig(multisigner),
		node.WithKeyGenerator(keyGen),
		node.WithPublicKey(pubKey),
		node.WithPrivateKey(privKey),
	)

	return nd, nil
}

func getSk(ctx *cli.Context) ([]byte, error) {
	if !ctx.GlobalIsSet(flags.PrivateKey.Name) {
		if ctx.GlobalString(flags.PrivateKey.Name) == "" {
			return nil, errors.New("no private key file provided")
		}
	}

	b64sk, err := ioutil.ReadFile(ctx.GlobalString(flags.PrivateKey.Name))
	if err != nil {
		b64sk = []byte(ctx.GlobalString(flags.PrivateKey.Name))
	}
	decodedSk := make([]byte, base64.StdEncoding.DecodedLen(len(b64sk)))
	l, err := base64.StdEncoding.Decode(decodedSk, b64sk)

	if err != nil {
		return nil, errors.New("could not decode private key: " + err.Error())
	}

	return decodedSk[:l], nil
}

func loadSkPk(sk []byte, log *logger.Logger) (
	keyGen crypto.KeyGenerator,
	privKey crypto.PrivateKey,
	pubKey crypto.PublicKey,
	err error,
) {
	keyGen = schnorr.NewKeyGenerator()
	privKey, err = keyGen.PrivateKeyFromByteArray(sk)

	if err == nil {
		pubKey = privKey.GeneratePublic()

		base64sk := make([]byte, base64.StdEncoding.EncodedLen(len(sk)))
		base64.StdEncoding.Encode(base64sk, sk)
		log.Info("starting with private key: " + string(base64sk))

		pk, _ := pubKey.ToByteArray()
		base64pk := make([]byte, base64.StdEncoding.EncodedLen(len(pk)))
		base64.StdEncoding.Encode(base64pk, pk)
		log.Info("starting with public key: " + string(base64pk))
	}

	return
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
	hashFuncs := make([]storage.HasherType, 0)
	for _, hf := range cfg.HashFunc {
		hashFuncs = append(hashFuncs, storage.HasherType(hf))
	}

	return storage.BloomConfig{
		Size:     cfg.Size,
		HashFunc: hashFuncs,
	}
}

func createBlockChainFromConfig(blConfig *blockchain.Config) (*blockchain.BlockChain, error) {
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

	txBadBlockCache, err := storage.NewCache(
		blConfig.TxBadBlockBodyCache.Type,
		blConfig.TxBadBlockBodyCache.Size)

	if err != nil {
		return nil, err
	}

	txUnit, err = storage.NewStorageUnitFromConf(
		blConfig.TxStorage.CacheConf,
		blConfig.TxStorage.DBConf,
		blConfig.TxStorage.BloomConf)

	if err != nil {
		return nil, err
	}

	txBlockUnit, err = storage.NewStorageUnitFromConf(
		blConfig.TxBlockBodyStorage.CacheConf,
		blConfig.TxBlockBodyStorage.DBConf,
		blConfig.TxBlockBodyStorage.BloomConf)

	if err != nil {
		return nil, err
	}

	stateBlockUnit, err = storage.NewStorageUnitFromConf(
		blConfig.StateBlockBodyStorage.CacheConf,
		blConfig.StateBlockBodyStorage.DBConf,
		blConfig.StateBlockBodyStorage.BloomConf)

	if err != nil {
		return nil, err
	}

	peerBlockUnit, err = storage.NewStorageUnitFromConf(
		blConfig.PeerBlockBodyStorage.CacheConf,
		blConfig.PeerBlockBodyStorage.DBConf,
		blConfig.PeerBlockBodyStorage.BloomConf)

	if err != nil {
		return nil, err
	}

	headerUnit, err = storage.NewStorageUnitFromConf(
		blConfig.BlockHeaderStorage.CacheConf,
		blConfig.BlockHeaderStorage.DBConf,
		blConfig.BlockHeaderStorage.BloomConf)

	if err != nil {
		return nil, err
	}

	blockChain, err := blockchain.NewBlockChain(
		txBadBlockCache,
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

func blockChainConfig() *blockchain.Config {
	cacher := storage.CacheConfig{Type: storage.LRUCache, Size: 100}
	bloom := storage.BloomConfig{Size: 2048, HashFunc: []storage.HasherType{storage.Keccak, storage.Blake2b, storage.Fnv}}
	persisterTxBlockBodyStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: filepath.Join(config.DefaultPath()+uniqueID, "TxBlockBodyStorage")}
	persisterStateBlockBodyStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: filepath.Join(config.DefaultPath()+uniqueID, "StateBlockBodyStorage")}
	persisterPeerBlockBodyStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: filepath.Join(config.DefaultPath()+uniqueID, "PeerBlockBodyStorage")}
	persisterBlockHeaderStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: filepath.Join(config.DefaultPath()+uniqueID, "BlockHeaderStorage")}
	persisterTxStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: filepath.Join(config.DefaultPath()+uniqueID, "TxStorage")}
	return &blockchain.Config{
		TxBlockBodyStorage:    storage.UnitConfig{CacheConf: cacher, DBConf: persisterTxBlockBodyStorage, BloomConf: bloom},
		StateBlockBodyStorage: storage.UnitConfig{CacheConf: cacher, DBConf: persisterStateBlockBodyStorage, BloomConf: bloom},
		PeerBlockBodyStorage:  storage.UnitConfig{CacheConf: cacher, DBConf: persisterPeerBlockBodyStorage, BloomConf: bloom},
		BlockHeaderStorage:    storage.UnitConfig{CacheConf: cacher, DBConf: persisterBlockHeaderStorage, BloomConf: bloom},
		TxStorage:             storage.UnitConfig{CacheConf: cacher, DBConf: persisterTxStorage, BloomConf: bloom},
		TxPoolStorage:         cacher,
		TxBadBlockBodyCache:   cacher,
	}
}
