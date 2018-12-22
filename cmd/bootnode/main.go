package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/cmd/facade"
	"github.com/ElrondNetwork/elrond-go-sandbox/cmd/flags"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transactionPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/fnv"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/bloom"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/leveldb"
	beevikntp "github.com/beevik/ntp"
	"sync"
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

type InitialNode struct {
	PubKey  string `json:"pubkey"`
	Balance uint64 `json:"balance"`
}

type Genesis struct {
	StartTime          int64         `json:"startTime"`
	RoundDuration      int64         `json:"roundDuration"`
	ConsensusGroupSize int           `json:"consensusGroupSize"`
	InitialNodes       []InitialNode `json:"initialNodes"`
}

func main() {
	log := logger.NewDefaultLogger()
	log.SetLevel(logger.LogInfo)

	app := cli.NewApp()
	cli.AppHelpTemplate = bootNodeHelpTemplate
	app.Name = "BootNode CLI App"
	app.Usage = "This is the entrypoint for starting a new bootstrap node - the app will start after the genessis timestamp"
	//app.Flags = []cli.Flag{flags.GenesisFile, flags.Port, flags.WithUI, flags.MaxAllowedPeers}
	app.Flags = []cli.Flag{flags.GenesisFile, flags.Port, flags.MaxAllowedPeers, flags.SelfPubKey}
	app.Action = func(c *cli.Context) error {
		err := startNode(c, log)
		if err != nil {
			log.Error("Could not start node", err.Error())
			return err
		}
		return nil
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

	initialConfig, err := loadInitialConfiguration(ctx.GlobalString(flags.GenesisFile.Name), log)
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Initialized with config from: %s", ctx.GlobalString(flags.GenesisFile.Name)))

	hasher := sha256.Sha256{}
	marshalizer := marshal.JsonMarshalizer{}
	syncer := ntp.NewSyncTime(time.Millisecond*time.Duration(initialConfig.RoundDuration), beevikntp.Query)
	cacher, err := storage.NewCache(storage.LRUCache, 100)
	persister, err := leveldb.NewDB("leveldb_temp")
	bloomFilter, err := bloom.NewFilter(2048, []hashing.Hasher{keccak.Keccak{}, blake2b.Blake2b{}, fnv.Fnv{}})
	storer, err := storage.NewStorageUnit(cacher, persister, bloomFilter)
	dBWriteCacher, err := trie.NewDBWriteCache(storer)
	patriciaMerkelTree, err := trie.NewTrie(nil, dBWriteCacher, hasher)
	accountAdapter, err := state.NewAccountsDB(patriciaMerkelTree, hasher, marshalizer)
	addressConverter, err := state.NewHashAddressConverter(hasher, 32, "")
	transactionProcessor, err := transaction.NewTxProcessor(accountAdapter, hasher, addressConverter, marshalizer)
	txPoolAccesser := transactionPool.NewTransactionPool(nil)
	blockProcessor := block.NewBlockProcessor(txPoolAccesser, hasher, marshalizer, transactionProcessor, accountAdapter, 1)

	log.Info(fmt.Sprintf("Current time in seconds: %d", syncer.CurrentTime(syncer.ClockOffset()).Unix()))

	// 1. Start with an empty node
	currentNode := CreateNode(
		ctx.GlobalInt(flags.MaxAllowedPeers.Name),
		ctx.GlobalInt(flags.Port.Name),
		initialConfig.InitialNodesPubKeys(),
		ctx.GlobalString(flags.SelfPubKey.Name),
		initialConfig.RoundDuration,
		initialConfig.ConsensusGroupSize,
		hasher,
		marshalizer,
		syncer,
		blockProcessor,
		time.Unix(initialConfig.StartTime, 0))

	ef := facade.NewElrondNodeFacade(currentNode)

	ef.SetLogger(log)
	ef.SetSyncer(syncer)

	wg := sync.WaitGroup{}
	go ef.StartBackgroundServices(&wg)

	// 2. Wait until we reach the config genesis time
	//ef.WaitForStartTime(time.Unix(initialConfig.StartTime, 0))
	//ef.WaitForStartTime(genesisTime)

	wg.Wait()

	// If not in UI mode we should automatically boot a node
	if !ctx.Bool(flags.WithUI.Name) {
		log.Info("Bootstraping node....")
		err = ef.StartNode()
		if err != nil {
			log.Error("Starting node failed", err.Error())
		}
	}

	// Hold the program until stopped by user
	go func() {
		<-sigs
		log.Info("terminating at user's signal...")
		stop <- true
	}()

	log.Info("Application is now running...")
	<-stop

	return nil
}

func loadInitialConfiguration(genesisFilePath string, log *logger.Logger) (*Genesis, error) {
	f, err := os.Open(genesisFilePath)
	defer func() {
		err := f.Close()
		if err != nil {
			log.Error("Cannot close configuration file: ", err.Error())
		}
	}()
	if err != nil {
		log.Error("Cannot open configuration file", err.Error())
		return nil, err
	}
	genesis := &Genesis{}
	jsonParser := json.NewDecoder(f)
	err = jsonParser.Decode(genesis)
	if err != nil {
		log.Error("Cannot decode configuration file", err.Error())
		return nil, err
	}
	return genesis, nil
}

func (g *Genesis) InitialNodesPubKeys() []string {
	var pubKeys []string
	for _, in := range g.InitialNodes {
		pubKeys = append(pubKeys, in.PubKey)
	}
	return pubKeys
}

func CreateNode(
	maxAllowedPeers int,
	port int,
	initialNodesPubKeys []string,
	selfPubKey string,
	roundDuration int64,
	consensusGroupSize int,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	syncer ntp.SyncTimer,
	blockProcessor process.BlockProcessor,
	genesisTime time.Time,
) *node.Node {
	appContext := context.Background()
	nd := node.NewNode(
		node.WithHasher(hasher),
		node.WithContext(appContext),
		node.WithMarshalizer(marshalizer),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithMaxAllowedPeers(maxAllowedPeers),
		node.WithPort(port),
		node.WithInitialNodesPubKeys(initialNodesPubKeys),
		node.WithSelfPubKey(selfPubKey),
		node.WithRoundDuration(roundDuration),
		node.WithConsensusGroupSize(consensusGroupSize),
		node.WithSyncer(syncer),
		node.WithBlockProcessor(blockProcessor),
		node.WithGenesisTime(genesisTime),
	)

	return nd
}
