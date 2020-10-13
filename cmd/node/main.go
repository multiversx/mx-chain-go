package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/denisbrodbeck/machineid"
	"github.com/urfave/cli"
)

const (
	defaultLogsPath = "logs"
	maxTimeToClose  = 10 * time.Second
	maxMachineIDLen = 10
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
)

// appVersion should be populated at build time using ldflags
// Usage examples:
// linux/mac:
//            go build -i -v -ldflags="-X main.appVersion=$(git describe --tags --long --dirty)"
// windows:
//            for /f %i in ('git describe --tags --long --dirty') do set VERS=%i
//            go build -i -v -ldflags="-X main.appVersion=%VERS%"
var appVersion = core.UnVersionedAppString

type configs struct {
	generalConfig                    *config.Config
	apiRoutesConfig                  *config.ApiRoutesConfig
	economicsConfig                  *config.EconomicsConfig
	systemSCConfig                   *config.SystemSmartContractsConfig
	ratingsConfig                    *config.RatingsConfig
	preferencesConfig                *config.Preferences
	externalConfig                   *config.ExternalConfig
	p2pConfig                        *config.P2PConfig
	configurationFileName            string
	configurationEconomicsFileName   string
	configurationRatingsFileName     string
	configurationPreferencesFileName string
	p2pConfigurationFileName         string
}

func main() {
	_ = logger.SetDisplayByteSlice(logger.ToHexShort)
	log := logger.GetOrCreate("main")

	app := cli.NewApp()
	cli.AppHelpTemplate = nodeHelpTemplate
	app.Name = "Elrond Node CLI App"
	machineID, err := machineid.ProtectedID(app.Name)
	if err != nil {
		log.Warn("error fetching machine id", "error", err)
		machineID = "unknown"
	}
	if len(machineID) > maxMachineIDLen {
		machineID = machineID[:maxMachineIDLen]
	}

	app.Version = fmt.Sprintf("%s/%s/%s-%s/%s", appVersion, runtime.Version(), runtime.GOOS, runtime.GOARCH, machineID)
	app.Usage = "This is the entry point for starting a new Elrond node - the app will start after the genesis timestamp"
	app.Flags = getFlags()
	app.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}

	app.Action = func(c *cli.Context) error {
		nodeStarter, err := NewNodeRunner(c, log, app.Version)
		if err != nil {
			return err
		}

		return nodeStarter.StartNode()
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
