package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go-logger/file"
	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/reflectcommon"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/urfave/cli"
	// test point 1 for custom profiler
)

const (
	defaultLogsPath = "logs"
	logFilePrefix   = "elrond-go"
)

var (
	memoryBallastObject []byte
	nodeHelpTemplate    = `NAME:
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
var appVersion = common.UnVersionedAppString

func main() {
	_ = logger.SetDisplayByteSlice(logger.ToHexShort)
	log := logger.GetOrCreate("main")

	// test point 2 for custom profiler

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
		log.Error(err.Error())
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
		timeLogLifeSpan := time.Second * time.Duration(cfgs.GeneralConfig.Logs.LogFileLifeSpanInSec)
		sizeLogLifeSpanInMB := uint64(cfgs.GeneralConfig.Logs.LogFileLifeSpanInMB)

		err := fileLogging.ChangeFileLifeSpan(timeLogLifeSpan, sizeLogLifeSpanInMB)
		if err != nil {
			return err
		}
	}

	err := applyFlags(c, cfgs, flagsConfig, log)
	if err != nil {
		return err
	}

	memBallastValue := c.GlobalUint64(memBallast.Name)
	if memBallastValue > 0 {
		// memory ballast is an optimization for golang's garbage collector. If set to a high value, it can decrease
		// the number of times when GC performs STW processes, that results is a better performance over high load
		memoryBallastObject = make([]byte, memBallastValue*core.MegabyteSize)
		log.Debug("initialized memory ballast object", "size", core.ConvertBytes(uint64(len(memoryBallastObject))))
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
	generalConfig, err := common.LoadMainConfig(configurationPaths.MainConfig)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationPaths.MainConfig)

	configurationPaths.ApiRoutes = ctx.GlobalString(configurationApiFile.Name)
	apiRoutesConfig, err := common.LoadApiConfig(configurationPaths.ApiRoutes)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationPaths.ApiRoutes)

	configurationPaths.Economics = ctx.GlobalString(configurationEconomicsFile.Name)
	economicsConfig, err := common.LoadEconomicsConfig(configurationPaths.Economics)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationPaths.Economics)

	configurationPaths.SystemSC = ctx.GlobalString(configurationSystemSCFile.Name)
	systemSCConfig, err := common.LoadSystemSmartContractsConfig(configurationPaths.SystemSC)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationPaths.SystemSC)

	configurationPaths.Ratings = ctx.GlobalString(configurationRatingsFile.Name)
	ratingsConfig, err := common.LoadRatingsConfig(configurationPaths.Ratings)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationPaths.Ratings)

	configurationPaths.Preferences = ctx.GlobalString(configurationPreferencesFile.Name)
	preferencesConfig, err := common.LoadPreferencesConfig(configurationPaths.Preferences)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationPaths.Preferences)

	for _, overridable := range preferencesConfig.Preferences.OverridableConfigTomlValues {
		err = reflectcommon.AdaptStructureValueBasedOnPath(generalConfig, overridable.Path, overridable.Value)
		if err != nil {
			return nil, err
		}

		log.Info("updated config toml value", "path", overridable.Path, "new value", overridable.Value)
	}

	configurationPaths.External = ctx.GlobalString(externalConfigFile.Name)
	externalConfig, err := common.LoadExternalConfig(configurationPaths.External)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationPaths.External)

	configurationPaths.P2p = ctx.GlobalString(p2pConfigurationFile.Name)
	p2pConfig, err := common.LoadP2PConfig(configurationPaths.P2p)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationPaths.P2p)

	configurationPaths.Epoch = ctx.GlobalString(epochConfigurationFile.Name)
	epochConfig, err := common.LoadEpochConfig(configurationPaths.Epoch)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationPaths.Epoch)

	configurationPaths.RoundActivation = ctx.GlobalString(roundConfigurationFile.Name)
	roundConfig, err := common.LoadRoundConfig(configurationPaths.RoundActivation)
	if err != nil {
		return nil, err
	}
	log.Debug("config", "file", configurationPaths.RoundActivation)

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
		RoundConfig:              roundConfig,
	}, nil
}

func attachFileLogger(log logger.Logger, flagsConfig *config.ContextFlagsConfig) (factory.FileLoggingHandler, error) {
	var fileLogging factory.FileLoggingHandler
	var err error
	if flagsConfig.SaveLogFile {
		args := file.ArgsFileLogging{
			WorkingDir:      flagsConfig.WorkingDir,
			DefaultLogsPath: defaultLogsPath,
			LogFilePrefix:   logFilePrefix,
		}
		fileLogging, err = file.NewFileLogging(args)
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
