package main

import (
	"encoding/json"
	"errors"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/cmd/flags"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
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
	Address 		string				`json:"address"`
}

type Account struct {
	Nounce 			int					`json:"nounce"`
	Balance 		int					`json:"balance"`
	CodeHash 		string				`json:"codeHash"`
	Root 			string				`json:"root"`
}

type Genesis struct {
	StartTime 		int64		 		`json:"startTime"`
	ClockSyncPeriod int8		`json:"clockSyncPeriod"`
	InitialNodes 	[]InitialNode		`json:"initialNodes"`
	Accounts 		[]Account			`json:"accounts"`
}

func main() {

	log := logger.NewDefaultLogger()
	log.SetLevel(logger.LogInfo)

	app := cli.NewApp()
	cli.AppHelpTemplate = bootNodeHelpTemplate
	app.Name = "BootNode CLI App"
	app.Usage = "This is the entrypoint for starting a new bootstrap node - the app will start after the genessis timestamp"
	app.Flags = []cli.Flag{ flags.GenesisFile }
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

	if ctx.GlobalIsSet(flags.GenesisFile.Name) {
		defaultGenesisJSON = ctx.GlobalString(flags.GenesisFile.Name)
	}

	initialConfig, err := loadInitialConfiguration(defaultGenesisJSON, log)
	if err != nil {
		return err
	}

	// NTP Start
	syncedTime := ntp.NewSyncTime(time.Second * time.Duration(initialConfig.ClockSyncPeriod), func(host string) (response *beevikntp.Response, e error) {
		return nil, errors.New("this should be implemented")
	})

	// 1. Boot everything that we can before we reach the genesis start time
	wg := sync.WaitGroup{}
	localNode := node.NewNode(node.WithPort(4000))
	if ctx.Bool(flags.WithUI.Name) {
		wg.Add(1)
		go wakeREST(&wg, localNode)
	}

	// 2. Wait until we reach the config genesis time
	if syncedTime.CurrentTime(syncedTime.ClockOffset()).After(time.Unix(initialConfig.StartTime, 0)) {
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
		localNode.Start()
	}

	return nil
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