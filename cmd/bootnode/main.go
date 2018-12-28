package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/config"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/schnorr"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"github.com/ElrondNetwork/elrond-go-sandbox/cmd/facade"
	"github.com/ElrondNetwork/elrond-go-sandbox/cmd/flags"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
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

type initialNode struct {
	PubKey  string `json:"pubkey"`
	Balance uint64 `json:"balance"`
}

type genesis struct {
	StartTime       int64         `json:"startTime"`
	ClockSyncPeriod int           `json:"clockSyncPeriod"`
	InitialNodes    []initialNode `json:"initialNodes"`
}

func main() {
	log := logger.NewDefaultLogger()
	log.SetLevel(logger.LogInfo)

	app := cli.NewApp()
	cli.AppHelpTemplate = bootNodeHelpTemplate
	app.Name = "BootNode CLI App"
	app.Usage = "This is the entry point for starting a new bootstrap node - the app will start after the genesis timestamp"
	app.Flags = []cli.Flag{flags.GenesisFile, flags.Port, flags.WithUI, flags.MaxAllowedPeers, flags.PrivateKey}
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

	currentNode, err := createNode(ctx, generalConfig, genesisConfig, log)
	if err != nil {
		return err
	}
	ef := facade.NewElrondNodeFacade(currentNode)
	ef.SetLogger(log)
	ef.StartNTP(genesisConfig.ClockSyncPeriod)

	wg := sync.WaitGroup{}
	go ef.StartBackgroundServices(&wg)

	ef.WaitForStartTime(time.Unix(genesisConfig.StartTime, 0))
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

func createNode(ctx *cli.Context, cfg *config.Config, genesisConfig *genesis, log *logger.Logger) (*node.Node, error) {
	appContext := context.Background()
	hasher := sha256.Sha256{}
	marshalizer := marshal.JsonMarshalizer{}

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

	nd, err := node.NewNode(
		node.WithHasher(hasher),
		node.WithContext(appContext),
		node.WithMarshalizer(marshalizer),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithMaxAllowedPeers(ctx.GlobalInt(flags.MaxAllowedPeers.Name)),
		node.WithPort(ctx.GlobalInt(flags.Port.Name)),
		node.WithInitialNodeAddresses(genesisConfig.initialNodesPubkeys()),
		node.WithAddressConverter(addressConverter),
		node.WithAccountsAdapter(accountsAdapter),
	)

	if err != nil {
		return nil, errors.New("error creating node: " + err.Error())
	}

	sk, err := getSk(ctx)
	if err != nil {
		log.Error("node is starting without a private key...")
	} else {
		loadSk(nd, sk, log)
	}

	return nd, nil
}

func getSk(ctx *cli.Context) ([]byte, error) {
	if !ctx.GlobalIsSet(flags.PrivateKey.Name) {
		return nil, errors.New("no private key file provided")
	}
	b64sk, err := ioutil.ReadFile(ctx.GlobalString(flags.PrivateKey.Name))
	if err != nil {
		return nil, errors.New("could not read private key file")
	}
	decodedSk := make([]byte, base64.StdEncoding.DecodedLen(len(b64sk)))
	l, _ := base64.StdEncoding.Decode(decodedSk, b64sk)

	return decodedSk[:l], nil
}

func loadSk(nd *node.Node, sk []byte, log *logger.Logger) {
	generator := schnorr.NewKeyGenerator()
	secretKey, err := generator.PrivateKeyFromByteArray(sk)
	if err == nil {
		err = nd.ApplyOptions(node.WithPrivateKey(secretKey))
		if err != nil {
			log.Error(err.Error())
		}
	} else {
		log.Error("error unpacking private key")
	}
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

func getCacherFromConfig(cfg config.CacheConfig) storage.CacheConfig {
	return storage.CacheConfig{
		Size: cfg.Size,
		Type: storage.CacheType(cfg.Type),
	}
}

func getDBFromConfig(cfg config.DBConfig) storage.DBConfig {
	return storage.DBConfig{
		FilePath: filepath.Join(config.DefaultPath(), cfg.FilePath),
		Type: storage.DBType(cfg.Type),
	}
}

func getBloomFromConfig(cfg config.BloomFilterConfig) storage.BloomConfig {
	hashFuncs := make([]storage.HasherType, 0)
	for _, hf := range cfg.HashFunc {
		hashFuncs = append(hashFuncs, storage.HasherType(hf))
	}

	return storage.BloomConfig{
		Size: cfg.Size,
		HashFunc: hashFuncs,
	}
}
