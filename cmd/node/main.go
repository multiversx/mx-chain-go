package main

import (
	"fmt"
	"os"
	"runtime"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/denisbrodbeck/machineid"
	"github.com/urfave/cli"
)

const (
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
		cfgs, errCfg := readConfigs(c, log)
		if errCfg != nil {
			return errCfg
		}

		err = applyFlags(c, cfgs, log)
		if err != nil {
			return err
		}

		cfgs.FlagsConfig.Version = app.Version

		nodeRunner, errRunner := node.NewNodeRunner(cfgs)
		if errRunner != nil {
			return errRunner
		}

		return nodeRunner.Start()
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func readConfigs(ctx *cli.Context, log logger.Logger) (*config.Configs, error) {
	log.Trace("reading Configs")

	configurationFileName := ctx.GlobalString(configurationFile.Name)
	generalConfig, err := core.LoadMainConfig(configurationFileName)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationFileName)

	configurationApiFileName := ctx.GlobalString(configurationApiFile.Name)
	apiRoutesConfig, err := core.LoadApiConfig(configurationApiFileName)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationApiFileName)

	configurationEconomicsFileName := ctx.GlobalString(configurationEconomicsFile.Name)
	economicsConfig, err := core.LoadEconomicsConfig(configurationEconomicsFileName)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationEconomicsFileName)

	configurationSystemSCConfigFileName := ctx.GlobalString(configurationSystemSCFile.Name)
	systemSCConfig, err := core.LoadSystemSmartContractsConfig(configurationSystemSCConfigFileName)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationSystemSCConfigFileName)

	configurationRatingsFileName := ctx.GlobalString(configurationRatingsFile.Name)
	ratingsConfig, err := core.LoadRatingsConfig(configurationRatingsFileName)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationRatingsFileName)

	configurationPreferencesFileName := ctx.GlobalString(configurationPreferencesFile.Name)
	preferencesConfig, err := core.LoadPreferencesConfig(configurationPreferencesFileName)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationPreferencesFileName)

	externalConfigurationFileName := ctx.GlobalString(externalConfigFile.Name)
	externalConfig, err := core.LoadExternalConfig(externalConfigurationFileName)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", externalConfigurationFileName)

	p2pConfigurationFileName := ctx.GlobalString(p2pConfigurationFile.Name)
	p2pConfig, err := core.LoadP2PConfig(p2pConfigurationFileName)
	if err != nil {
		return nil, err
	}

	log.Debug("config", "file", p2pConfigurationFileName)
	if ctx.IsSet(port.Name) {
		p2pConfig.Node.Port = ctx.GlobalString(port.Name)
	}
	if ctx.IsSet(destinationShardAsObserver.Name) {
		preferencesConfig.Preferences.DestinationShardAsObserver = ctx.GlobalString(destinationShardAsObserver.Name)
	}
	if ctx.IsSet(nodeDisplayName.Name) {
		preferencesConfig.Preferences.NodeDisplayName = ctx.GlobalString(nodeDisplayName.Name)
	}
	if ctx.IsSet(identityFlagName.Name) {
		preferencesConfig.Preferences.Identity = ctx.GlobalString(identityFlagName.Name)
	}

	return &config.Configs{
		GeneralConfig:                    generalConfig,
		ApiRoutesConfig:                  apiRoutesConfig,
		EconomicsConfig:                  economicsConfig,
		SystemSCConfig:                   systemSCConfig,
		RatingsConfig:                    ratingsConfig,
		PreferencesConfig:                preferencesConfig,
		ExternalConfig:                   externalConfig,
		P2pConfig:                        p2pConfig,
		ConfigurationFileName:            configurationFileName,
		ConfigurationApiRoutesFileName:   configurationApiFileName,
		ConfigurationEconomicsFileName:   configurationEconomicsFileName,
		ConfigurationSystemSCFilename:    configurationSystemSCConfigFileName,
		ConfigurationRatingsFileName:     configurationRatingsFileName,
		ConfigurationPreferencesFileName: configurationPreferencesFileName,
		ConfigurationExternalFileName:    externalConfigurationFileName,
		P2pConfigurationFileName:         p2pConfigurationFileName,
		ConfigurationGasScheduleFileName: gasScheduleConfigurationFile.Name,
	}, nil
}
