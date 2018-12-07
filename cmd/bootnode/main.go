package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/cmd/flags"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	beevikntp "github.com/beevik/ntp"
	"github.com/urfave/cli"
	"os"
	"sync"
	"time"
)

var (
	defaultGenesisJSON 	= "genesis.json"
	consensusGroupSize 	= 21
	maxAllowedPeers 	= 4
	roundDuration 		= time.Duration(1000 * time.Millisecond)
	pbftThreshold 		= consensusGroupSize*2/3 + 1
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
	Address string `json:"address"`
}

type Account struct {
	Nounce int `json:"nounce"`
	Balance int	`json:"balance"`
	CodeHash string `json:"codeHash"`
	Root string `json:"root"`
}

type Genesis struct {
	StartTime int64 `json:"startTime"`
	ClockSyncPeriod int8 `json:"clockSyncPeriod"`
	InitialNodes []InitialNode `json:"initialNodes"`
	Accounts []Account `json:"accounts"`
}

func main() {
	log := logger.NewDefaultLogger()
	log.SetLevel(logger.LogInfo)

	app := cli.NewApp()
	cli.AppHelpTemplate = bootNodeHelpTemplate
	app.Name = "BootNode CLI App"
	app.Usage = "This is the entrypoint for starting a new bootstrap node - the app will start after the genessis timestamp"
	app.Flags = []cli.Flag{ flags.GenesisFile, flags.Port, flags.WithUI }
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

	wg := sync.WaitGroup{}
	appContext := context.Background()
	hasher := sha256.Sha256{}
	marshalizer := marshal.JsonMarshalizer{}

	if ctx.GlobalIsSet(flags.GenesisFile.Name) {
		defaultGenesisJSON = ctx.GlobalString(flags.GenesisFile.Name)
	}

	initialConfig, err := loadInitialConfiguration(defaultGenesisJSON, log)
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Initialized with config from: %s", defaultGenesisJSON))

	// NTP Start
	syncedTime := ntp.NewSyncTime(time.Second * time.Duration(initialConfig.ClockSyncPeriod), func(host string) (response *beevikntp.Response, e error) {
		return nil, errors.New("this should be implemented")
	})

	// 1. Start with an empty node
	localNode := node.NewNode(
		node.WithHasher(hasher),
		node.WithContext(appContext),
		node.WithMarshalizer(marshalizer),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithMaxAllowedPeers(maxAllowedPeers),
		node.WithPort(ctx.GlobalInt(flags.Port.Name)),
	)
	go wakeup(ctx, &wg, localNode)

	// 2. Wait until we reach the config genesis time
	if !syncedTime.CurrentTime(syncedTime.ClockOffset()).After(time.Unix(initialConfig.StartTime, 0)) {
		log.Info("Elrond protocol not started yet, waiting ...")
	}
	for {
		if syncedTime.CurrentTime(syncedTime.ClockOffset()).After(time.Unix(initialConfig.StartTime, 0)) {

			break
		}
		time.Sleep(time.Duration(5 * time.Millisecond))
	}
	// 3. Start all services
	wg.Wait()

	// If not in UI mode we should automatically boot a node
	if !ctx.Bool(flags.WithUI.Name) {
		fmt.Println("Bootstraping node....")
		err = localNode.Start()
		if err != nil {
			log.Error("Could not start node: ", err.Error())
		}
		err = localNode.ConnectToAddresses(initialConfig.InitialNodesAddresses())
		if err != nil {
			log.Error("Could not connect to addresses", err.Error())
		}
	}

	return nil
}

func wakeup(ctx *cli.Context, group *sync.WaitGroup, localNode *node.Node) {
	group.Add(1)
	go wakeREST(group, localNode)
}

func wakeREST(group *sync.WaitGroup, localNode *node.Node) {
	// TODO: Next task will refactor the api. We are not able to boot it here right now,
	//  but it will be possible after the changes are implemented. We should do that then
	group.Done()
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

func (g *Genesis) InitialNodesAddresses() []string {
	var addresses []string
	for  _, in := range g.InitialNodes {
		addresses = append(addresses, in.Address)
	}
	return addresses
}