package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/logging"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/urfave/cli"
)

const (
	defaultLogsPath = "logs"
	logFilePrefix   = "elrond-go"
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
	machineID := core.GetAnonymizedMachineID(app.Name)

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
		return startNodeRunner(c, log, app.Version)
	}

	err := app.Run(os.Args)
	if err != nil {
		os.Exit(1)
	}
}

func startNodeRunner(c *cli.Context, log logger.Logger, version string) error {
	flagsConfig := getFlagsConfig(c, log)

	fileLogging, errLogger := attachFileLogger(log, flagsConfig)
	if errLogger != nil {
		return errLogger
	}

	cfgs, errCfg := readConfigs(c, log)
	if errCfg != nil {
		return errCfg
	}

	if !check.IfNil(fileLogging) {
		err := fileLogging.ChangeFileLifeSpan(time.Second * time.Duration(cfgs.GeneralConfig.Logs.LogFileLifeSpanInSec))
		if err != nil {
			return err
		}
	}

	err := applyFlags(c, cfgs, flagsConfig, log)
	if err != nil {
		return err
	}

	cfgs.FlagsConfig.Version = version

	nodeRunner, errRunner := node.NewNodeRunner(cfgs)
	if errRunner != nil {
		return errRunner
	}

	err = nodeRunner.Start()
	if err != nil {
		log.Error(err.Error())
	}

	if !check.IfNil(fileLogging) {
		err = fileLogging.Close()
		log.LogIfError(err)
	}

	return err
}

func readConfigs(ctx *cli.Context, log logger.Logger) (*config.Configs, error) {
	log.Trace("reading Configs")

	configurationPaths := &config.ConfigurationPathsHolder{}

	configurationPaths.MainConfig = ctx.GlobalString(configurationFile.Name)
	generalConfig, err := core.LoadMainConfig(configurationPaths.MainConfig)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationPaths.MainConfig)

	configurationPaths.ApiRoutes = ctx.GlobalString(configurationApiFile.Name)
	apiRoutesConfig, err := core.LoadApiConfig(configurationPaths.ApiRoutes)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationPaths.ApiRoutes)

	configurationPaths.Economics = ctx.GlobalString(configurationEconomicsFile.Name)
	economicsConfig, err := core.LoadEconomicsConfig(configurationPaths.Economics)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationPaths.Economics)

	configurationPaths.SystemSC = ctx.GlobalString(configurationSystemSCFile.Name)
	systemSCConfig, err := core.LoadSystemSmartContractsConfig(configurationPaths.SystemSC)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationPaths.SystemSC)

	configurationPaths.Ratings = ctx.GlobalString(configurationRatingsFile.Name)
	ratingsConfig, err := core.LoadRatingsConfig(configurationPaths.Ratings)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationPaths.Ratings)

	configurationPaths.Preferences = ctx.GlobalString(configurationPreferencesFile.Name)
	preferencesConfig, err := core.LoadPreferencesConfig(configurationPaths.Preferences)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationPaths.Preferences)

	configurationPaths.External = ctx.GlobalString(externalConfigFile.Name)
	externalConfig, err := core.LoadExternalConfig(configurationPaths.External)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationPaths.External)

	configurationPaths.P2p = ctx.GlobalString(p2pConfigurationFile.Name)
	p2pConfig, err := core.LoadP2PConfig(configurationPaths.P2p)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationPaths.P2p)

	configurationPaths.Epoch = ctx.GlobalString(epochConfigurationFile.Name)
	epochConfig, err := core.LoadEpochConfig(configurationPaths.Epoch)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationPaths.Epoch)

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
		GeneralConfig:            generalConfig,
		ApiRoutesConfig:          apiRoutesConfig,
		EconomicsConfig:          economicsConfig,
		SystemSCConfig:           systemSCConfig,
		RatingsConfig:            ratingsConfig,
		PreferencesConfig:        preferencesConfig,
		ExternalConfig:           externalConfig,
		P2pConfig:                p2pConfig,
		ConfigurationPathsHolder: configurationPaths,
		EpochConfig:              epochConfig,
	}, nil
}

func attachFileLogger(log logger.Logger, flagsConfig *config.ContextFlagsConfig) (factory.FileLoggingHandler, error) {
	var fileLogging factory.FileLoggingHandler
	var err error
	if flagsConfig.SaveLogFile {
		fileLogging, err = logging.NewFileLogging(flagsConfig.WorkingDir, defaultLogsPath, logFilePrefix)
		if err != nil {
			return nil, fmt.Errorf("%w creating a log file", err)
		}
	}

	err = logger.SetDisplayByteSlice(logger.ToHex)
	log.LogIfError(err)
	logger.ToggleCorrelation(flagsConfig.EnableLogCorrelation)
	logger.ToggleLoggerName(flagsConfig.EnableLogName)
	logLevelFlagValue := flagsConfig.LogLevel
	err = logger.SetLogLevel(logLevelFlagValue)
	if err != nil {
		return nil, err
	}

	if flagsConfig.DisableAnsiColor {
		err = logger.RemoveLogObserver(os.Stdout)
		if err != nil {
			return nil, err
		}

		err = logger.AddLogObserver(os.Stdout, &logger.PlainFormatter{})
		if err != nil {
			return nil, err
		}
	}
	log.Trace("logger updated", "level", logLevelFlagValue, "disable ANSI color", flagsConfig.DisableAnsiColor)

	return fileLogging, nil
}
