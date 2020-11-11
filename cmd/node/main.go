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
	filePathPlaceholder = "[path]"
	// genesisFile defines a flag for the path of the bootstrapping file.
	genesisFile = cli.StringFlag{
		Name: "genesis-file",
		Usage: "The `" + filePathPlaceholder + "` for the genesis file. This JSON file contains initial data to " +
			"bootstrap from, such as initial balances for accounts.",
		Value: "./config/genesis.json",
	}
	// smartContractsFile defines a flag for the path of the file containing initial smart contracts.
	smartContractsFile = cli.StringFlag{
		Name: "smart-contracts-file",
		Usage: "The `" + filePathPlaceholder + "` for the initial smart contracts file. This JSON file contains data used " +
			"to deploy initial smart contracts such as delegation smart contracts",
		Value: "./config/genesisSmartContracts.json",
	}
	// nodesFile defines a flag for the path of the initial nodes file.
	nodesFile = cli.StringFlag{
		Name: "nodes-setup-file",
		Usage: "The `" + filePathPlaceholder + "` for the nodes setup. This JSON file contains initial nodes info, " +
			"such as consensus group size, round duration, validators public keys and so on.",
		Value: "./config/nodesSetup.json",
	}
	// configurationFile defines a flag for the path to the main toml configuration file
	configurationFile = cli.StringFlag{
		Name: "config",
		Usage: "The `" + filePathPlaceholder + "` for the main configuration file. This TOML file contain the main " +
			"configurations such as storage setups, epoch duration and so on.",
		Value: "./config/config.toml",
	}
	// configurationEconomicsFile defines a flag for the path to the economics toml configuration file
	configurationEconomicsFile = cli.StringFlag{
		Name: "config-economics",
		Usage: "The `" + filePathPlaceholder + "` for the economics configuration file. This TOML file contains " +
			"economics configurations such as minimum gas price for a transactions and so on.",
		Value: "./config/economics.toml",
	}
	// configurationApiFile defines a flag for the path to the api routes toml configuration file
	configurationApiFile = cli.StringFlag{
		Name: "config-api",
		Usage: "The `" + filePathPlaceholder + "` for the api configuration file. This TOML file contains " +
			"all available routes for Rest API and options to enable or disable them.",
		Value: "./config/api.toml",
	}
	// configurationSystemSCFile defines a flag for the path to the system sc toml configuration file
	configurationSystemSCFile = cli.StringFlag{
		Name:  "config-systemSmartContracts",
		Usage: "The `" + filePathPlaceholder + "` for the system smart contracts configuration file.",
		Value: "./config/systemSmartContractsConfig.toml",
	}
	// configurationRatingsFile defines a flag for the path to the ratings toml configuration file
	configurationRatingsFile = cli.StringFlag{
		Name:  "config-ratings",
		Usage: "The ratings configuration file to load",
		Value: "./config/ratings.toml",
	}
	// configurationPreferencesFile defines a flag for the path to the preferences toml configuration file
	configurationPreferencesFile = cli.StringFlag{
		Name: "config-preferences",
		Usage: "The `" + filePathPlaceholder + "` for the preferences configuration file. This TOML file contains " +
			"preferences configurations, such as the node display name or the shard to start in when starting as observer",
		Value: "./config/prefs.toml",
	}
	// externalConfigFile defines a flag for the path to the external toml configuration file
	externalConfigFile = cli.StringFlag{
		Name: "config-external",
		Usage: "The `" + filePathPlaceholder + "` for the external configuration file. This TOML file contains" +
			" external configurations such as ElasticSearch's URL and login information",
		Value: "./config/external.toml",
	}
	// p2pConfigurationFile defines a flag for the path to the toml file containing P2P configuration
	p2pConfigurationFile = cli.StringFlag{
		Name: "p2p-config",
		Usage: "The `" + filePathPlaceholder + "` for the p2p configuration file. This TOML file contains peer-to-peer " +
			"configurations such as port, target peer count or KadDHT settings",
		Value: "./config/p2p.toml",
	}
	// gasScheduleConfigurationFile defines a flag for the path to the toml file containing the gas costs used in SmartContract execution
	gasScheduleConfigurationFile = cli.StringFlag{
		Name: "gas-costs-config",
		Usage: "The `" + filePathPlaceholder + "` for the gas costs configuration file. This TOML file contains " +
			"gas costs used in SmartContract execution",
		Value: "./config/gasSchedule.toml",
	}
	// port defines a flag for setting the port on which the node will listen for connections
	port = cli.StringFlag{
		Name: "port",
		Usage: "The `[p2p port]` number on which the application will start. Can use single values such as " +
			"`0, 10230, 15670` or range of ports such as `5000-10000`",
		Value: "0",
	}
	// profileMode defines a flag for profiling the binary
	// If enabled, it will open the pprof routes over the default gin rest webserver.
	// There are several routes that will be available for profiling (profiling can be analyzed with: go tool pprof):
	//  /debug/pprof/ (can be accessed in the browser, will list the available options)
	//  /debug/pprof/goroutine
	//  /debug/pprof/heap
	//  /debug/pprof/threadcreate
	//  /debug/pprof/block
	//  /debug/pprof/mutex
	//  /debug/pprof/profile (CPU profile)
	//  /debug/pprof/trace?seconds=5 (CPU trace) -> being a trace, can be analyzed with: go tool trace
	// Usage: go tool pprof http(s)://ip.of.the.server/debug/pprof/xxxxx
	profileMode = cli.BoolFlag{
		Name: "profile-mode",
		Usage: "Boolean option for enabling the profiling mode. If set, the /debug/pprof routes will be available " +
			"on the node for profiling the application.",
	}
	// useHealthService is used to enable the health service
	useHealthService = cli.BoolFlag{
		Name:  "use-health-service",
		Usage: "Boolean option for enabling the health service.",
	}
	// validatorKeyIndex defines a flag that specifies the 0-th based index of the private key to be used from validatorKey.pem file
	validatorKeyIndex = cli.IntFlag{
		Name:  "sk-index",
		Usage: "The index in the PEM file of the private key to be used by the node.",
		Value: 0,
	}
	// gopsEn used to enable diagnosis of running go processes
	gopsEn = cli.BoolFlag{
		Name:  "gops-enable",
		Usage: "Boolean option for enabling gops over the process. If set, stack can be viewed by calling 'gops stack <pid>'.",
	}
	// storageCleanup defines a flag for choosing the option of starting the node from scratch. If it is not set (false)
	// it starts from the last state stored on disk
	storageCleanup = cli.BoolFlag{
		Name: "storage-cleanup",
		Usage: "Boolean option for starting the node with clean storage. If set, the Node will empty its storage " +
			"before starting, otherwise it will start from the last state stored on disk..",
	}

	// restApiInterface defines a flag for the interface on which the rest API will try to bind with
	restApiInterface = cli.StringFlag{
		Name: "rest-api-interface",
		Usage: "The interface `address and port` to which the REST API will attempt to bind. " +
			"To bind to all available interfaces, set this flag to :8080",
		Value: facade.DefaultRestInterface,
	}

	// restApiDebug defines a flag for starting the rest API engine in debug mode
	restApiDebug = cli.BoolFlag{
		Name:  "rest-api-debug",
		Usage: "Boolean option for starting the Rest API in debug mode.",
	}

	// nodeDisplayName defines the friendly name used by a node in the public monitoring tools. If set, will override
	// the NodeDisplayName from prefs.toml
	nodeDisplayName = cli.StringFlag{
		Name: "display-name",
		Usage: "The user-friendly name for the node, appearing in the public monitoring tools. Will override the " +
			"name set in the preferences TOML file.",
		Value: "",
	}

	// identityFlagName defines the keybase's identity. If set, will override the identity from prefs.toml
	identityFlagName = cli.StringFlag{
		Name:  "keybase-identity",
		Usage: "The keybase's identity. If set, will override the one set in the preferences TOML file.",
		Value: "",
	}

	//useLogView is used when termui interface is not needed.
	useLogView = cli.BoolFlag{
		Name: "use-log-view",
		Usage: "Boolean option for enabling the simple node's interface. If set, the node will not enable the " +
			"user-friendly terminal view of the node.",
	}

	// validatorKeyPemFile defines a flag for the path to the validator key used in block signing
	validatorKeyPemFile = cli.StringFlag{
		Name:  "validator-key-pem-file",
		Usage: "The `filepath` for the PEM file which contains the secret keys for the validator key.",
		Value: "./config/validatorKey.pem",
	}
	// logLevel defines the logger level
	logLevel = cli.StringFlag{
		Name: "log-level",
		Usage: "This flag specifies the logger `level(s)`. It can contain multiple comma-separated value. For example" +
			", if set to *:INFO the logs for all packages will have the INFO level. However, if set to *:INFO,api:DEBUG" +
			" the logs for all packages will have the INFO level, excepting the api package which will receive a DEBUG" +
			" log level.",
		Value: "*:" + logger.LogInfo.String(),
	}
	//logFile is used when the log output needs to be logged in a file
	logSaveFile = cli.BoolFlag{
		Name:  "log-save",
		Usage: "Boolean option for enabling log saving. If set, it will automatically save all the logs into a file.",
	}
	//logWithCorrelation is used to enable log correlation elements
	logWithCorrelation = cli.BoolFlag{
		Name:  "log-correlation",
		Usage: "Boolean option for enabling log correlation elements.",
	}
	//logWithLoggerName is used to enable log correlation elements
	logWithLoggerName = cli.BoolFlag{
		Name:  "log-logger-name",
		Usage: "Boolean option for logger name in the logs.",
	}
	// disableAnsiColor defines if the logger subsystem should prevent displaying ANSI colors
	disableAnsiColor = cli.BoolFlag{
		Name:  "disable-ansi-color",
		Usage: "Boolean option for disabling ANSI colors in the logging system.",
	}
	// bootstrapRoundIndex defines a flag that specifies the round index from which node should bootstrap from storage
	bootstrapRoundIndex = cli.Uint64Flag{
		Name:  "bootstrap-round-index",
		Usage: "This flag specifies the round `index` from which node should bootstrap from storage.",
		Value: math.MaxUint64,
	}
	// enableTxIndexing enables transaction indexing. There can be cases when it's too expensive to index all transactions
	//  so we provide the command line option to disable this behaviour
	enableTxIndexing = cli.BoolTFlag{
		Name: "tx-indexing",
		Usage: "Boolean option for enabling transactions indexing. There can be cases when it's too expensive to " +
			"index all transactions so this flag will disable this.",
	}

	// workingDirectory defines a flag for the path for the working directory.
	workingDirectory = cli.StringFlag{
		Name:  "working-directory",
		Usage: "This flag specifies the `directory` where the node will store databases, logs and statistics.",
		Value: "",
	}

	// destinationShardAsObserver defines a flag for the prefered shard to be assigned to as an observer.
	destinationShardAsObserver = cli.StringFlag{
		Name: "destination-shard-as-observer",
		Usage: "This flag specifies the shard to start in when running as an observer. It will override the configuration " +
			"set in the preferences TOML config file.",
		Value: "",
	}

	keepOldEpochsData = cli.BoolFlag{
		Name: "keep-old-epochs-data",
		Usage: "Boolean option for enabling a node to keep old epochs data. If set, the node won't remove any database " +
			"and will have a full history over epochs.",
	}

	numEpochsToSave = cli.Uint64Flag{
		Name: "num-epochs-to-keep",
		Usage: "This flag represents the number of epochs which will kept in the databases. It is relevant only if " +
			"the full archive flag is not set.",
		Value: uint64(2),
	}

	numActivePersisters = cli.Uint64Flag{
		Name: "num-active-persisters",
		Usage: "This flag represents the number of databases (1 database = 1 epoch) which are kept open at a moment. " +
			"It is relevant even if the node is full archive or not.",
		Value: uint64(2),
	}

	fullArchive = cli.BoolFlag{
		Name:  "full-archive",
		Usage: "Boolean option for settings an observer as full archive, which will sync the entire database of its shard",
	}

	startInEpoch = cli.BoolFlag{
		Name: "start-in-epoch",
		Usage: "Boolean option for enabling a node the fast bootstrap mechanism from the network." +
			"Should be enabled if data is not available in local disk.",
	}

	// importDbDirectory defines a flag for the optional import DB directory on which the node will re-check the blockchain against
	importDbDirectory = cli.StringFlag{
		Name: "import-db",
		Usage: "This flag, if set, will make the node start the import process using the provided data path. Will re-check" +
			"and re-process everything",
		Value: "",
	}
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
		var cfgs *config.Configs
		cfgs, err = readConfigs(c, log)
		if err != nil {
			return err
		}

		applyFlags(c, cfgs, log)
		cfgs.FlagsConfig.Version = app.Version

		nodeRunner, err := node.NewNodeRunner(cfgs, log)
		if err != nil {
			return err
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

	importDbDirectoryValue := ctx.GlobalString(importDbDirectory.Name)
	if len(importDbDirectoryValue) > 0 {
		importCheckpointRoundsModulus := uint(generalConfig.EpochStartConfig.RoundsPerEpoch)
		log.Info("import DB directory is set, altering config values!",
			"GeneralSettings.StartInEpochEnabled", "false",
			"StateTriesConfig.CheckpointRoundsModulus", importCheckpointRoundsModulus,
			"import DB path", importDbDirectoryValue,
		)
		generalConfig.GeneralSettings.StartInEpochEnabled = false
		generalConfig.StateTriesConfig.CheckpointRoundsModulus = importCheckpointRoundsModulus
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
		ConfigurationEconomicsFileName:   configurationEconomicsFileName,
		ConfigurationRatingsFileName:     configurationRatingsFileName,
		ConfigurationPreferencesFileName: configurationPreferencesFileName,
		P2pConfigurationFileName:         p2pConfigurationFileName,
	}, nil
}
