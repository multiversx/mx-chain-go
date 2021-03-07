package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	elasticIndexer "github.com/ElrondNetwork/elastic-indexer-go"
	indexerFactory "github.com/ElrondNetwork/elastic-indexer-go/factory"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/cmd/node/metrics"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/round"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/accumulator"
	"github.com/ElrondNetwork/elrond-go/core/alarm"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/closing"
	"github.com/ElrondNetwork/elrond-go/core/dblookupext"
	dbLookupFactory "github.com/ElrondNetwork/elrond-go/core/dblookupext/factory"
	"github.com/ElrondNetwork/elrond-go/core/forking"
	"github.com/ElrondNetwork/elrond-go/core/logging"
	"github.com/ElrondNetwork/elrond-go/core/parsers"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/core/versioning"
	"github.com/ElrondNetwork/elrond-go/core/watchdog"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/data/state"
	stateFactory "github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/facade"
	mainFactory "github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/fallback"
	"github.com/ElrondNetwork/elrond-go/genesis/parsing"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/health"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/node/nodeDebugFactory"
	"github.com/ElrondNetwork/elrond-go/node/stakeValuesProcessor"
	"github.com/ElrondNetwork/elrond-go/node/txsimulator"
	"github.com/ElrondNetwork/elrond-go/ntp"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/headerCheck"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	"github.com/ElrondNetwork/elrond-go/process/rating"
	"github.com/ElrondNetwork/elrond-go/process/rating/peerHonesty"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/blackList"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/redundancy"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go/storage/pathmanager"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	"github.com/ElrondNetwork/elrond-go/update"
	exportFactory "github.com/ElrondNetwork/elrond-go/update/factory"
	"github.com/ElrondNetwork/elrond-go/update/trigger"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/denisbrodbeck/machineid"
	"github.com/google/gops/agent"
	"github.com/urfave/cli"
)

const (
	defaultStatsPath             = "stats"
	defaultLogsPath              = "logs"
	logFilePrefix                = "elrond-go"
	notSetDestinationShardID     = "disabled"
	metachainShardName           = "metachain"
	secondsToWaitForP2PBootstrap = 20
	maxTimeToClose               = 10 * time.Second
	maxMachineIDLen              = 10
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
	// gasScheduleConfigurationDirectory defines a flag for the path to the directory containing the gas costs used in execution
	gasScheduleConfigurationDirectory = cli.StringFlag{
		Name:  "gas-costs-config",
		Usage: "The `" + filePathPlaceholder + "` for the gas costs configuration directory.",
		Value: "./config/gasSchedules",
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
	// elasticSearchTemplates defines a flag for the path to the elasticsearch templates
	elasticSearchTemplates = cli.StringFlag{
		Name:  "elasticsearch-templates-path",
		Usage: "The `path` to the elasticsearch templates directory containing the templates in .json format",
		Value: "./config/elasticIndexTemplates",
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
	// importDbNoSigCheck defines a flag for the optional import DB no signature check option
	importDbNoSigCheck = cli.BoolFlag{
		Name:  "import-db-no-sig-check",
		Usage: "This flag, if set, will cause the signature checks on headers to be skipped. Can be used only if the import-db was previously set",
	}

	// redundancyLevel defines a flag that specifies the level of redundancy used by the current instance for the node (-1 = disabled, 0 = main instance (default), 1 = first backup, 2 = second backup, etc.)
	redundancyLevel = cli.Int64Flag{
		Name:  "redundancy-level",
		Usage: "This flag specifies the level of redundancy used by the current instance for the node (-1 = disabled, 0 = main instance (default), 1 = first backup, 2 = second backup, etc.)",
		Value: 0,
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
	app.Flags = []cli.Flag{
		genesisFile,
		smartContractsFile,
		nodesFile,
		configurationFile,
		configurationApiFile,
		configurationEconomicsFile,
		configurationSystemSCFile,
		configurationRatingsFile,
		configurationPreferencesFile,
		externalConfigFile,
		p2pConfigurationFile,
		gasScheduleConfigurationDirectory,
		validatorKeyIndex,
		validatorKeyPemFile,
		port,
		profileMode,
		useHealthService,
		storageCleanup,
		gopsEn,
		nodeDisplayName,
		identityFlagName,
		restApiInterface,
		restApiDebug,
		disableAnsiColor,
		elasticSearchTemplates,
		logLevel,
		logSaveFile,
		logWithCorrelation,
		logWithLoggerName,
		useLogView,
		bootstrapRoundIndex,
		workingDirectory,
		destinationShardAsObserver,
		keepOldEpochsData,
		startInEpoch,
		importDbDirectory,
		importDbNoSigCheck,
		redundancyLevel,
	}
	app.Authors = []cli.Author{
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
		},
	}

	app.Action = func(c *cli.Context) error {
		return startNode(c, log, app.Version)
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func getSuite(config *config.Config) (crypto.Suite, error) {
	switch config.Consensus.Type {
	case consensus.BlsConsensusType:
		return mcl.NewSuiteBLS12(), nil
	default:
		return nil, errors.New("no consensus provided in config file")
	}
}

func startNode(ctx *cli.Context, log logger.Logger, version string) error {
	log.Trace("startNode called")
	chanStopNodeProcess := make(chan endProcess.ArgEndProcess, 1)
	workingDir := getWorkingDir(ctx, log)

	var fileLogging factory.FileLoggingHandler
	var err error
	withLogFile := ctx.GlobalBool(logSaveFile.Name)
	if withLogFile {
		fileLogging, err = logging.NewFileLogging(workingDir, defaultLogsPath, logFilePrefix)
		if err != nil {
			return fmt.Errorf("%w creating a log file", err)
		}
	}

	err = logger.SetDisplayByteSlice(logger.ToHex)
	log.LogIfError(err)
	logger.ToggleCorrelation(ctx.GlobalBool(logWithCorrelation.Name))
	logger.ToggleLoggerName(ctx.GlobalBool(logWithLoggerName.Name))
	logLevelFlagValue := ctx.GlobalString(logLevel.Name)
	err = logger.SetLogLevel(logLevelFlagValue)
	if err != nil {
		return err
	}
	noAnsiColor := ctx.GlobalBool(disableAnsiColor.Name)
	if noAnsiColor {
		err = logger.RemoveLogObserver(os.Stdout)
		if err != nil {
			//we need to print this manually as we do not have console log observer
			fmt.Println("error removing log observer: " + err.Error())
			return err
		}

		err = logger.AddLogObserver(os.Stdout, &logger.PlainFormatter{})
		if err != nil {
			//we need to print this manually as we do not have console log observer
			fmt.Println("error setting log observer: " + err.Error())
			return err
		}
	}
	log.Trace("logger updated", "level", logLevelFlagValue, "disable ANSI color", noAnsiColor)

	enableGopsIfNeeded(ctx, log)

	log.Info("starting node", "version", version, "pid", os.Getpid())
	log.Trace("reading configs")

	configurationFileName := ctx.GlobalString(configurationFile.Name)
	generalConfig, err := loadMainConfig(configurationFileName)
	if err != nil {
		return err
	}
	log.Debug("config", "file", configurationFileName)

	p2pConfigurationFileName := ctx.GlobalString(p2pConfigurationFile.Name)
	p2pConfig, err := core.LoadP2PConfig(p2pConfigurationFileName)
	if err != nil {
		return err
	}
	log.Debug("config", "file", p2pConfigurationFileName)

	importDbDirectoryValue := ctx.GlobalString(importDbDirectory.Name)
	isInImportMode := len(importDbDirectoryValue) > 0
	importDbNoSigCheckFlag := ctx.GlobalBool(importDbNoSigCheck.Name) && isInImportMode
	applyCompatibleConfigs(isInImportMode, importDbNoSigCheckFlag, log, generalConfig, p2pConfig)

	configurationApiFileName := ctx.GlobalString(configurationApiFile.Name)
	apiRoutesConfig, err := loadApiConfig(configurationApiFileName)
	if err != nil {
		return err
	}
	log.Debug("config", "file", configurationApiFileName)

	configurationEconomicsFileName := ctx.GlobalString(configurationEconomicsFile.Name)
	economicsConfig, err := loadEconomicsConfig(configurationEconomicsFileName)
	if err != nil {
		return err
	}
	log.Debug("config", "file", configurationEconomicsFileName)

	configurationSystemSCConfigFileName := ctx.GlobalString(configurationSystemSCFile.Name)
	systemSCConfig, err := loadSystemSmartContractsConfig(configurationSystemSCConfigFileName)
	if err != nil {
		return err
	}
	log.Debug("config", "file", configurationSystemSCConfigFileName)

	configurationRatingsFileName := ctx.GlobalString(configurationRatingsFile.Name)
	ratingsConfig, err := loadRatingsConfig(configurationRatingsFileName)
	if err != nil {
		return err
	}
	log.Debug("config", "file", configurationRatingsFileName)

	configurationPreferencesFileName := ctx.GlobalString(configurationPreferencesFile.Name)
	preferencesConfig, err := loadPreferencesConfig(configurationPreferencesFileName)
	if err != nil {
		return err
	}
	log.Debug("config", "file", configurationPreferencesFileName)

	externalConfigurationFileName := ctx.GlobalString(externalConfigFile.Name)
	externalConfig, err := loadExternalConfig(externalConfigurationFileName)
	if err != nil {
		return err
	}
	log.Debug("config", "file", externalConfigurationFileName)

	if ctx.IsSet(port.Name) {
		p2pConfig.Node.Port = ctx.GlobalString(port.Name)
	}

	if !check.IfNil(fileLogging) {
		err = fileLogging.ChangeFileLifeSpan(time.Second * time.Duration(generalConfig.Logs.LogFileLifeSpanInSec))
		if err != nil {
			return err
		}
	}

	epochNotifier := forking.NewGenericEpochNotifier()

	addressPubkeyConverter, err := stateFactory.NewPubkeyConverter(generalConfig.AddressPubkeyConverter)
	if err != nil {
		return fmt.Errorf("%w for AddressPubKeyConverter", err)
	}
	validatorPubkeyConverter, err := stateFactory.NewPubkeyConverter(generalConfig.ValidatorPubkeyConverter)
	if err != nil {
		return fmt.Errorf("%w for ValidatorPubkeyConverter", err)
	}

	//TODO when refactoring main, maybe initialize economics data before this line
	totalSupply, ok := big.NewInt(0).SetString(economicsConfig.GlobalSettings.GenesisTotalSupply, 10)
	if !ok {
		return fmt.Errorf("can not parse total suply from economics.toml, %s is not a valid value",
			economicsConfig.GlobalSettings.GenesisTotalSupply)
	}

	log.Debug("config", "file", ctx.GlobalString(genesisFile.Name))

	exportFolder := filepath.Join(workingDir, generalConfig.Hardfork.ImportFolder)
	nodesSetupPath := ctx.GlobalString(nodesFile.Name)
	if generalConfig.Hardfork.AfterHardFork {
		exportFolderNodesSetupPath := filepath.Join(exportFolder, core.NodesSetupJsonFileName)
		if !core.DoesFileExist(exportFolderNodesSetupPath) {
			return fmt.Errorf("cannot find %s in the export folder", core.NodesSetupJsonFileName)
		}

		nodesSetupPath = exportFolderNodesSetupPath
	}
	genesisNodesConfig, err := sharding.NewNodesSetup(
		nodesSetupPath,
		addressPubkeyConverter,
		validatorPubkeyConverter,
		generalConfig.GeneralSettings.GenesisMaxNumberOfShards,
	)
	if err != nil {
		return err
	}
	log.Debug("config", "file", nodesSetupPath)

	if generalConfig.Hardfork.AfterHardFork {
		log.Debug("changed genesis time after hardfork",
			"old genesis time", genesisNodesConfig.StartTime,
			"new genesis time", generalConfig.Hardfork.GenesisTime)
		genesisNodesConfig.StartTime = generalConfig.Hardfork.GenesisTime
	}

	syncer := ntp.NewSyncTime(generalConfig.NTPConfig, nil)
	syncer.StartSyncingTime()

	log.Debug("NTP average clock offset", "value", syncer.ClockOffset())

	if ctx.IsSet(startInEpoch.Name) {
		log.Debug("start in epoch is enabled")
		generalConfig.GeneralSettings.StartInEpochEnabled = ctx.GlobalBool(startInEpoch.Name)
		if generalConfig.GeneralSettings.StartInEpochEnabled {
			delayedStartInterval := 2 * time.Second
			time.Sleep(delayedStartInterval)
		}
	}

	//TODO: The next 5 lines should be deleted when we are done testing from a precalculated (not hard coded) timestamp
	if genesisNodesConfig.StartTime == 0 {
		time.Sleep(1000 * time.Millisecond)
		ntpTime := syncer.CurrentTime()
		genesisNodesConfig.StartTime = (ntpTime.Unix()/60 + 1) * 60
	}

	startTime := time.Unix(genesisNodesConfig.StartTime, 0)

	log.Info("start time",
		"formatted", startTime.Format("Mon Jan 2 15:04:05 MST 2006"),
		"seconds", startTime.Unix())

	log.Trace("getting suite")
	suite, err := getSuite(generalConfig)
	if err != nil {
		return err
	}

	validatorKeyPemFileName := ctx.GlobalString(validatorKeyPemFile.Name)
	cryptoParamsLoader, err := mainFactory.NewCryptoSigningParamsLoader(
		validatorPubkeyConverter,
		ctx.GlobalInt(validatorKeyIndex.Name),
		validatorKeyPemFileName,
		suite,
		isInImportMode,
	)
	if err != nil {
		return err
	}

	cryptoParams, err := cryptoParamsLoader.Get()
	if err != nil {
		return fmt.Errorf("%w: consider regenerating your keys", err)
	}

	log.Debug("block sign pubkey", "value", cryptoParams.PublicKeyString)

	if ctx.IsSet(destinationShardAsObserver.Name) {
		preferencesConfig.Preferences.DestinationShardAsObserver = ctx.GlobalString(destinationShardAsObserver.Name)
	}

	if ctx.IsSet(nodeDisplayName.Name) {
		preferencesConfig.Preferences.NodeDisplayName = ctx.GlobalString(nodeDisplayName.Name)
	}

	if ctx.IsSet(identityFlagName.Name) {
		preferencesConfig.Preferences.Identity = ctx.GlobalString(identityFlagName.Name)
	}

	if ctx.IsSet(redundancyLevel.Name) {
		preferencesConfig.Preferences.RedundancyLevel = ctx.GlobalInt64(redundancyLevel.Name)
	}

	err = cleanupStorageIfNecessary(workingDir, ctx, log)
	if err != nil {
		return err
	}

	pathTemplateForPruningStorer := filepath.Join(
		workingDir,
		factory.DefaultDBPath,
		genesisNodesConfig.ChainID,
		fmt.Sprintf("%s_%s", factory.DefaultEpochString, core.PathEpochPlaceholder),
		fmt.Sprintf("%s_%s", factory.DefaultShardString, core.PathShardPlaceholder),
		core.PathIdentifierPlaceholder)

	pathTemplateForStaticStorer := filepath.Join(
		workingDir,
		factory.DefaultDBPath,
		genesisNodesConfig.ChainID,
		factory.DefaultStaticDbString,
		fmt.Sprintf("%s_%s", factory.DefaultShardString, core.PathShardPlaceholder),
		core.PathIdentifierPlaceholder)

	var pathManager *pathmanager.PathManager
	pathManager, err = pathmanager.NewPathManager(pathTemplateForPruningStorer, pathTemplateForStaticStorer)
	if err != nil {
		return err
	}

	genesisShardCoordinator, nodeType, err := createShardCoordinator(genesisNodesConfig, cryptoParams.PublicKey, preferencesConfig.Preferences, log)
	if err != nil {
		return err
	}
	var shardId = core.GetShardIDString(genesisShardCoordinator.SelfId())

	log.Trace("creating crypto components")
	cryptoArgs := mainFactory.CryptoComponentsFactoryArgs{
		Config:                               *generalConfig,
		NodesConfig:                          genesisNodesConfig,
		ShardCoordinator:                     genesisShardCoordinator,
		KeyGen:                               cryptoParams.KeyGenerator,
		PrivKey:                              cryptoParams.PrivateKey,
		ActivateBLSPubKeyMessageVerification: systemSCConfig.StakingSystemSCConfig.ActivateBLSPubKeyMessageVerification,
	}
	cryptoComponentsFactory, err := mainFactory.NewCryptoComponentsFactory(cryptoArgs, importDbNoSigCheckFlag)
	if err != nil {
		return err
	}
	cryptoComponents, err := cryptoComponentsFactory.Create()
	if err != nil {
		return err
	}

	accountsParser, err := parsing.NewAccountsParser(
		ctx.GlobalString(genesisFile.Name),
		totalSupply,
		addressPubkeyConverter,
		cryptoComponents.TxSignKeyGen,
	)
	if err != nil {
		return err
	}

	smartContractParser, err := parsing.NewSmartContractsParser(
		ctx.GlobalString(smartContractsFile.Name),
		addressPubkeyConverter,
		cryptoComponents.TxSignKeyGen,
	)
	if err != nil {
		return err
	}

	log.Trace("creating core components")

	healthService := health.NewHealthService(generalConfig.Health, workingDir)
	if ctx.IsSet(useHealthService.Name) {
		healthService.Start()
	}

	coreArgs := mainFactory.CoreComponentsFactoryArgs{
		Config:                *generalConfig,
		ShardId:               shardId,
		ChainID:               []byte(genesisNodesConfig.ChainID),
		MinTransactionVersion: genesisNodesConfig.MinTransactionVersion,
	}
	coreComponentsFactory := mainFactory.NewCoreComponentsFactory(coreArgs)
	coreComponents, err := coreComponentsFactory.Create()
	if err != nil {
		return err
	}

	chanCreateViews := make(chan struct{}, 1)
	chanLogRewrite := make(chan struct{}, 1)
	handlersArgs, err := factory.NewStatusHandlersFactoryArgs(
		useLogView.Name,
		ctx,
		coreComponents.InternalMarshalizer,
		coreComponents.Uint64ByteSliceConverter,
		chanCreateViews,
		chanLogRewrite,
	)
	if err != nil {
		return err
	}

	statusHandlersInfo, err := factory.CreateStatusHandlers(handlersArgs)
	if err != nil {
		return err
	}

	coreComponents.StatusHandler = statusHandlersInfo.StatusHandler

	log.Trace("creating network components")
	networkComponentFactory, err := mainFactory.NewNetworkComponentsFactory(
		*p2pConfig,
		*generalConfig,
		coreComponents.StatusHandler,
		coreComponents.InternalMarshalizer,
		syncer,
	)
	if err != nil {
		return err
	}
	networkComponents, err := networkComponentFactory.Create()
	if err != nil {
		return err
	}
	err = networkComponents.NetMessenger.Bootstrap()
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("waiting %d seconds for network discovery...", secondsToWaitForP2PBootstrap))
	time.Sleep(secondsToWaitForP2PBootstrap * time.Second)

	log.Trace("creating economics data components")
	argsNewEconomicsData := economics.ArgsNewEconomicsData{
		Economics:                      economicsConfig,
		PenalizedTooMuchGasEnableEpoch: generalConfig.GeneralSettings.PenalizedTooMuchGasEnableEpoch,
		GasPriceModifierEnableEpoch:    generalConfig.GeneralSettings.GasPriceModifierEnableEpoch,
		EpochNotifier:                  epochNotifier,
	}
	economicsData, err := economics.NewEconomicsData(argsNewEconomicsData)
	if err != nil {
		return err
	}

	err = economicsData.SetStatusHandler(coreComponents.StatusHandler)
	if err != nil {
		log.Debug("cannot set status handler to economicsData", "error", err)
	}

	log.Trace("creating ratings data components")

	ratingDataArgs := rating.RatingsDataArg{
		Config:                   ratingsConfig,
		ShardConsensusSize:       genesisNodesConfig.ConsensusGroupSize,
		MetaConsensusSize:        genesisNodesConfig.MetaChainConsensusGroupSize,
		ShardMinNodes:            genesisNodesConfig.MinNodesPerShard,
		MetaMinNodes:             genesisNodesConfig.MetaChainMinNodes,
		RoundDurationMiliseconds: genesisNodesConfig.RoundDuration,
	}
	ratingsData, err := rating.NewRatingsData(ratingDataArgs)
	if err != nil {
		return err
	}

	rater, err := rating.NewBlockSigningRater(ratingsData)
	if err != nil {
		return err
	}

	argsNodesShuffler := &sharding.NodesShufflerArgs{
		NodesShard:           genesisNodesConfig.MinNodesPerShard,
		NodesMeta:            genesisNodesConfig.MetaChainMinNodes,
		Hysteresis:           genesisNodesConfig.Hysteresis,
		Adaptivity:           genesisNodesConfig.Adaptivity,
		ShuffleBetweenShards: true,
		MaxNodesEnableConfig: generalConfig.GeneralSettings.MaxNodesChangeEnableEpoch,
	}

	nodesShuffler, err := sharding.NewHashValidatorsShuffler(argsNodesShuffler)
	if err != nil {
		return err
	}

	destShardIdAsObserver, err := processDestinationShardAsObserver(preferencesConfig.Preferences)
	if err != nil {
		return err
	}

	startRound := int64(0)
	if generalConfig.Hardfork.AfterHardFork {
		startRound = int64(generalConfig.Hardfork.StartRound)
	}
	rounder, err := round.NewRound(
		time.Unix(genesisNodesConfig.StartTime, 0),
		syncer.CurrentTime(),
		time.Millisecond*time.Duration(genesisNodesConfig.RoundDuration),
		syncer,
		startRound,
	)
	if err != nil {
		return err
	}

	importStartHandler, err := trigger.NewImportStartHandler(filepath.Join(workingDir, factory.DefaultDBPath), appVersion)
	if err != nil {
		return err
	}

	bootstrapDataProvider, err := storageFactory.NewBootstrapDataProvider(coreComponents.InternalMarshalizer)
	if err != nil {
		return err
	}

	latestStorageDataProvider, err := factory.CreateLatestStorageDataProvider(
		bootstrapDataProvider,
		coreComponents.InternalMarshalizer,
		coreComponents.Hasher,
		*generalConfig,
		genesisNodesConfig.ChainID,
		workingDir,
		factory.DefaultDBPath,
		factory.DefaultEpochString,
		factory.DefaultShardString,
	)
	if err != nil {
		return err
	}

	unitOpener, err := factory.CreateUnitOpener(
		bootstrapDataProvider,
		latestStorageDataProvider,
		coreComponents.InternalMarshalizer,
		*generalConfig,
		genesisNodesConfig.ChainID,
		workingDir,
		factory.DefaultDBPath,
		factory.DefaultEpochString,
		factory.DefaultShardString,
	)
	if err != nil {
		return err
	}

	versionsCache, err := storageUnit.NewCache(storageFactory.GetCacherFromConfig(generalConfig.Versions.Cache))
	if err != nil {
		return err
	}

	headerIntegrityVerifier, err := headerCheck.NewHeaderIntegrityVerifier(
		[]byte(genesisNodesConfig.ChainID),
		generalConfig.Versions.VersionsByEpochs,
		generalConfig.Versions.DefaultVersion,
		versionsCache,
	)
	if err != nil {
		return err
	}

	epochStartBootstrapArgs := bootstrap.ArgsEpochStartBootstrap{
		PublicKey:                  cryptoParams.PublicKey,
		Marshalizer:                coreComponents.InternalMarshalizer,
		TxSignMarshalizer:          coreComponents.TxSignMarshalizer,
		Hasher:                     coreComponents.Hasher,
		Messenger:                  networkComponents.NetMessenger,
		GeneralConfig:              *generalConfig,
		EconomicsData:              economicsData,
		SingleSigner:               cryptoComponents.TxSingleSigner,
		BlockSingleSigner:          cryptoComponents.SingleSigner,
		KeyGen:                     cryptoComponents.TxSignKeyGen,
		BlockKeyGen:                cryptoComponents.BlockSignKeyGen,
		GenesisNodesConfig:         genesisNodesConfig,
		GenesisShardCoordinator:    genesisShardCoordinator,
		PathManager:                pathManager,
		StorageUnitOpener:          unitOpener,
		WorkingDir:                 workingDir,
		DefaultDBPath:              factory.DefaultDBPath,
		DefaultEpochString:         factory.DefaultEpochString,
		DefaultShardString:         factory.DefaultShardString,
		Rater:                      rater,
		DestinationShardAsObserver: destShardIdAsObserver,
		Uint64Converter:            coreComponents.Uint64ByteSliceConverter,
		NodeShuffler:               nodesShuffler,
		Rounder:                    rounder,
		AddressPubkeyConverter:     addressPubkeyConverter,
		LatestStorageDataProvider:  latestStorageDataProvider,
		ArgumentsParser:            smartContract.NewArgumentParser(),
		StatusHandler:              coreComponents.StatusHandler,
		HeaderIntegrityVerifier:    headerIntegrityVerifier,
		TxSignHasher:               coreComponents.TxSignHasher,
		EpochNotifier:              epochNotifier,
	}
	bootstrapper, err := bootstrap.NewEpochStartBootstrap(epochStartBootstrapArgs)
	if err != nil {
		log.Error("could not create bootstrap", "err", err)
		return err
	}

	bootstrapParameters, err := bootstrapper.Bootstrap()
	if err != nil {
		log.Error("bootstrap return error", "error", err)
		return err
	}

	trieContainer, trieStorageManager := bootstrapper.GetTriesComponents()
	triesComponents := &mainFactory.TriesComponents{
		TriesContainer:      trieContainer,
		TrieStorageManagers: trieStorageManager,
	}

	log.Info("bootstrap parameters", "shardId", bootstrapParameters.SelfShardId, "epoch", bootstrapParameters.Epoch, "numShards", bootstrapParameters.NumOfShards)

	shardCoordinator, err := sharding.NewMultiShardCoordinator(bootstrapParameters.NumOfShards, bootstrapParameters.SelfShardId)
	if err != nil {
		return err
	}

	currentEpoch := bootstrapParameters.Epoch
	storerEpoch := currentEpoch
	if !generalConfig.StoragePruning.Enabled {
		// TODO: refactor this as when the pruning storer is disabled, the default directory path is Epoch_0
		// and it should be Epoch_ALL or something similar
		storerEpoch = 0
	}

	var shardIdString = core.GetShardIDString(shardCoordinator.SelfId())
	logger.SetCorrelationShard(shardIdString)

	log.Trace("initializing stats file")
	err = initStatsFileMonitor(generalConfig, pathManager, shardId)
	if err != nil {
		return err
	}

	log.Trace("creating data components")
	epochStartNotifier := notifier.NewEpochStartSubscriptionHandler()

	dataArgs := mainFactory.DataComponentsFactoryArgs{
		Config:             *generalConfig,
		EconomicsData:      economicsData,
		ShardCoordinator:   shardCoordinator,
		Core:               coreComponents,
		PathManager:        pathManager,
		EpochStartNotifier: epochStartNotifier,
		CurrentEpoch:       storerEpoch,
	}
	dataComponentsFactory, err := mainFactory.NewDataComponentsFactory(dataArgs)
	if err != nil {
		return err
	}
	dataComponents, err := dataComponentsFactory.Create()
	if err != nil {
		return err
	}

	healthService.RegisterComponent(dataComponents.Datapool.Transactions())
	healthService.RegisterComponent(dataComponents.Datapool.UnsignedTransactions())
	healthService.RegisterComponent(dataComponents.Datapool.RewardTransactions())

	log.Trace("initializing metrics")
	err = metrics.InitMetrics(
		coreComponents.StatusHandler,
		cryptoParams.PublicKeyString,
		nodeType,
		shardCoordinator,
		genesisNodesConfig,
		version,
		economicsConfig,
		generalConfig.EpochStartConfig.RoundsPerEpoch,
	)
	if err != nil {
		return err
	}

	chanLogRewrite <- struct{}{}
	chanCreateViews <- struct{}{}

	err = statusHandlersInfo.UpdateStorerAndMetricsForPersistentHandler(dataComponents.Store.GetStorer(dataRetriever.StatusMetricsUnit))
	if err != nil {
		return err
	}

	log.Trace("creating nodes coordinator")
	if ctx.IsSet(keepOldEpochsData.Name) {
		generalConfig.StoragePruning.CleanOldEpochsData = !ctx.GlobalBool(keepOldEpochsData.Name)
	}
	log.Info("Bootstrap", "epoch", bootstrapParameters.Epoch)
	if bootstrapParameters.NodesConfig != nil {
		log.Info("the epoch from nodesConfig is", "epoch", bootstrapParameters.NodesConfig.CurrentEpoch)
	}

	nodesCoordinator, nodeShufflerOut, err := createNodesCoordinator(
		log,
		genesisNodesConfig,
		preferencesConfig.Preferences,
		epochStartNotifier,
		cryptoParams.PublicKey,
		coreComponents.InternalMarshalizer,
		coreComponents.Hasher,
		rater,
		dataComponents.Store.GetStorer(dataRetriever.BootstrapUnit),
		nodesShuffler,
		generalConfig.EpochStartConfig,
		shardCoordinator.SelfId(),
		chanStopNodeProcess,
		bootstrapParameters,
		currentEpoch,
	)
	if err != nil {
		return err
	}

	log.Trace("creating state components")
	stateArgs := mainFactory.StateComponentsFactoryArgs{
		Config:           *generalConfig,
		ShardCoordinator: shardCoordinator,
		Core:             coreComponents,
		PathManager:      pathManager,
		Tries:            triesComponents,
	}
	stateComponentsFactory, err := mainFactory.NewStateComponentsFactory(stateArgs)
	if err != nil {
		return err
	}
	stateComponents, err := stateComponentsFactory.Create()
	if err != nil {
		return err
	}

	metrics.SaveStringMetric(coreComponents.StatusHandler, core.MetricNodeDisplayName, preferencesConfig.Preferences.NodeDisplayName)
	metrics.SaveStringMetric(coreComponents.StatusHandler, core.MetricChainId, genesisNodesConfig.ChainID)
	metrics.SaveUint64Metric(coreComponents.StatusHandler, core.MetricGasPerDataByte, economicsData.GasPerDataByte())
	metrics.SaveUint64Metric(coreComponents.StatusHandler, core.MetricMinGasPrice, economicsData.MinGasPrice())
	metrics.SaveUint64Metric(coreComponents.StatusHandler, core.MetricMinGasLimit, economicsData.MinGasLimit())
	metrics.SaveStringMetric(coreComponents.StatusHandler, core.MetricRewardsTopUpGradientPoint, economicsData.RewardsTopUpGradientPoint().String())
	metrics.SaveStringMetric(coreComponents.StatusHandler, core.MetricTopUpFactor, fmt.Sprintf("%g", economicsData.RewardsTopUpFactor()))
	metrics.SaveStringMetric(coreComponents.StatusHandler, core.MetricGasPriceModifier, fmt.Sprintf("%g", economicsData.GasPriceModifier()))

	sessionInfoFileOutput := fmt.Sprintf("%s:%s\n%s:%s\n%s:%v\n%s:%s\n%s:%v\n",
		"PkBlockSign", cryptoParams.PublicKeyString,
		"ShardId", shardId,
		"TotalShards", shardCoordinator.NumberOfShards(),
		"AppVersion", version,
		"GenesisTimeStamp", startTime.Unix(),
	)

	sessionInfoFileOutput += "\nStarted with parameters:\n"
	for _, flag := range ctx.App.Flags {
		flagValue := fmt.Sprintf("%v", ctx.GlobalGeneric(flag.GetName()))
		if flagValue != "" {
			sessionInfoFileOutput += fmt.Sprintf("%s = %v\n", flag.GetName(), flagValue)
		}
	}

	statsFolder := filepath.Join(workingDir, defaultStatsPath)
	copyConfigToStatsFolder(
		log,
		statsFolder,
		[]string{
			configurationFileName,
			configurationEconomicsFileName,
			configurationRatingsFileName,
			configurationPreferencesFileName,
			p2pConfigurationFileName,
			configurationFileName,
			ctx.GlobalString(genesisFile.Name),
			ctx.GlobalString(nodesFile.Name),
		})

	statsFile := filepath.Join(statsFolder, "session.info")
	err = ioutil.WriteFile(statsFile, []byte(sessionInfoFileOutput), os.ModePerm)
	log.LogIfError(err)

	//TODO: remove this in the future and add just a log debug
	computedRatingsData := filepath.Join(statsFolder, "ratings.info")
	computedRatingsDataStr := createStringFromRatingsData(ratingsData)
	err = ioutil.WriteFile(computedRatingsData, []byte(computedRatingsDataStr), os.ModePerm)
	log.LogIfError(err)

	log.Trace("creating tps benchmark components")
	initialTpsBenchmark := statusHandlersInfo.LoadTpsBenchmarkFromStorage(
		dataComponents.Store.GetStorer(dataRetriever.StatusMetricsUnit),
		coreComponents.InternalMarshalizer,
	)
	tpsBenchmark, err := statistics.NewTPSBenchmarkWithInitialData(
		statusHandlersInfo.StatusHandler,
		initialTpsBenchmark,
		shardCoordinator.NumberOfShards(),
		genesisNodesConfig.RoundDuration/1000,
	)
	if err != nil {
		return err
	}

	esIndexer, err := createElasticIndexer(
		externalConfig.ElasticSearchConnector,
		coreComponents.InternalMarshalizer,
		coreComponents.Hasher,
		nodesCoordinator,
		epochStartNotifier,
		addressPubkeyConverter,
		validatorPubkeyConverter,
		stateComponents.AccountsAdapter,
		economicsConfig.GlobalSettings.Denomination,
		shardCoordinator,
		economicsData,
		isInImportMode,
	)
	if err != nil {
		return err
	}

	gasScheduleConfigurationFolderName := ctx.GlobalString(gasScheduleConfigurationDirectory.Name)
	argsGasScheduleNotifier := forking.ArgsNewGasScheduleNotifier{
		GasScheduleConfig: generalConfig.GasSchedule,
		ConfigDir:         gasScheduleConfigurationFolderName,
		EpochNotifier:     epochNotifier,
	}
	gasScheduleNotifier, err := forking.NewGasScheduleNotifier(argsGasScheduleNotifier)
	if err != nil {
		return err
	}

	log.Trace("creating time cache for requested items components")
	requestedItemsHandler := timecache.NewTimeCache(time.Duration(uint64(time.Millisecond) * genesisNodesConfig.RoundDuration))

	whiteListCache, err := storageUnit.NewCache(storageFactory.GetCacherFromConfig(generalConfig.WhiteListPool))
	if err != nil {
		return err
	}
	whiteListRequest, err := interceptors.NewWhiteListDataVerifier(whiteListCache)
	if err != nil {
		return err
	}

	whiteListerVerifiedTxs, err := createWhiteListerVerifiedTxs(generalConfig)
	if err != nil {
		return err
	}

	historyRepoFactoryArgs := &dbLookupFactory.ArgsHistoryRepositoryFactory{
		SelfShardID: shardCoordinator.SelfId(),
		Config:      generalConfig.DbLookupExtensions,
		Hasher:      coreComponents.Hasher,
		Marshalizer: coreComponents.InternalMarshalizer,
		Store:       dataComponents.Store,
	}
	historyRepositoryFactory, err := dbLookupFactory.NewHistoryRepositoryFactory(historyRepoFactoryArgs)
	if err != nil {
		return err
	}

	historyRepository, err := historyRepositoryFactory.Create()
	if err != nil {
		return err
	}

	txSimulatorProcessorArgs := &txsimulator.ArgsTxSimulator{
		AddressPubKeyConverter: addressPubkeyConverter,
		ShardCoordinator:       shardCoordinator,
	}

	fallbackHeaderValidator, err := fallback.NewFallbackHeaderValidator(
		dataComponents.Datapool.Headers(),
		coreComponents.InternalMarshalizer,
		dataComponents.Store,
	)
	if err != nil {
		return err
	}

	log.Trace("creating process components")
	processArgs := factory.NewProcessComponentsFactoryArgs(
		&coreArgs,
		accountsParser,
		smartContractParser,
		economicsData,
		genesisNodesConfig,
		gasScheduleNotifier,
		rounder,
		shardCoordinator,
		nodesCoordinator,
		dataComponents,
		coreComponents,
		cryptoComponents,
		stateComponents,
		networkComponents,
		triesComponents,
		requestedItemsHandler,
		whiteListRequest,
		whiteListerVerifiedTxs,
		epochStartNotifier,
		*generalConfig,
		currentEpoch,
		rater,
		generalConfig.Marshalizer.SizeCheckDelta,
		generalConfig.StateTriesConfig.CheckpointRoundsModulus,
		generalConfig.GeneralSettings.MaxComputableRounds,
		generalConfig.Antiflood.NumConcurrentResolverJobs,
		generalConfig.BlockSizeThrottleConfig.MinSizeInBytes,
		generalConfig.BlockSizeThrottleConfig.MaxSizeInBytes,
		ratingsConfig.General.MaxRating,
		validatorPubkeyConverter,
		ratingsData,
		systemSCConfig,
		version,
		importStartHandler,
		coreComponents.Uint64ByteSliceConverter,
		workingDir,
		esIndexer,
		tpsBenchmark,
		historyRepository,
		epochNotifier,
		txSimulatorProcessorArgs,
		ctx.GlobalString(importDbDirectory.Name),
		chanStopNodeProcess,
		fallbackHeaderValidator,
	)
	processComponents, err := factory.ProcessComponentsFactory(processArgs)
	if err != nil {
		return err
	}

	transactionSimulator, err := txsimulator.NewTransactionSimulator(*txSimulatorProcessorArgs)
	if err != nil {
		return err
	}

	hardForkTrigger, err := createHardForkTrigger(
		generalConfig,
		cryptoParams.KeyGenerator,
		cryptoParams.PublicKey,
		shardCoordinator,
		nodesCoordinator,
		coreComponents,
		stateComponents,
		dataComponents,
		cryptoComponents,
		processComponents,
		networkComponents,
		whiteListRequest,
		whiteListerVerifiedTxs,
		chanStopNodeProcess,
		epochStartNotifier,
		importStartHandler,
		genesisNodesConfig,
		workingDir,
		epochNotifier,
	)
	if err != nil {
		return err
	}

	err = hardForkTrigger.AddCloser(nodeShufflerOut)
	if err != nil {
		return fmt.Errorf("%w when adding nodeShufflerOut in hardForkTrigger", err)
	}

	if !esIndexer.IsNilIndexer() {
		esIndexer.SetTxLogsProcessor(processComponents.TxLogsProcessor)
		processComponents.TxLogsProcessor.EnableLogToBeSavedInCache()
	}

	observerBLSPrivateKey, observerBLSPublicKey := cryptoComponents.BlockSignKeyGen.GeneratePair()
	observerBLSPublicKeyBuff, err := observerBLSPublicKey.ToByteArray()
	if err != nil {
		log.Error("error generating observerBLSPublicKeyBuff", "error", err)
	} else {
		log.Debug("generated BLS private key for redundancy handler. This key will be used on heartbeat messages "+
			"if the node is in backup mode and the main node is active", "hex public key", observerBLSPublicKeyBuff)
	}
	arg := redundancy.ArgNodeRedundancy{
		RedundancyLevel:    preferencesConfig.Preferences.RedundancyLevel,
		Messenger:          networkComponents.NetMessenger,
		ObserverPrivateKey: observerBLSPrivateKey,
	}

	nodeRedundancy, err := redundancy.NewNodeRedundancy(arg)
	if err != nil {
		return err
	}

	log.Trace("creating node structure")
	currentNode, err := createNode(
		generalConfig,
		ratingsConfig,
		preferencesConfig,
		genesisNodesConfig,
		economicsData,
		syncer,
		cryptoParams.KeyGenerator,
		cryptoParams.PrivateKey,
		cryptoParams.PublicKey,
		shardCoordinator,
		nodesCoordinator,
		coreComponents,
		stateComponents,
		dataComponents,
		cryptoComponents,
		processComponents,
		networkComponents,
		ctx.GlobalUint64(bootstrapRoundIndex.Name),
		version,
		esIndexer,
		requestedItemsHandler,
		epochStartNotifier,
		whiteListRequest,
		whiteListerVerifiedTxs,
		chanStopNodeProcess,
		hardForkTrigger,
		historyRepository,
		fallbackHeaderValidator,
		isInImportMode,
		nodeRedundancy,
	)
	if err != nil {
		return err
	}

	log.Trace("creating software checker structure")
	softwareVersionChecker, err := factory.CreateSoftwareVersionChecker(coreComponents.StatusHandler, generalConfig.SoftwareVersionConfig)
	if err != nil {
		log.Debug("nil software version checker", "error", err.Error())
	} else {
		softwareVersionChecker.StartCheckSoftwareVersion()
	}

	if shardCoordinator.SelfId() == core.MetachainShardId {
		log.Trace("activating nodesCoordinator's validators indexing")
		indexValidatorsListIfNeeded(esIndexer, nodesCoordinator, processComponents.EpochStartTrigger.Epoch(), log)
	}

	log.Trace("creating api resolver structure")
	apiWorkingDir := filepath.Join(workingDir, factory.TemporaryPath)
	apiResolver, err := createApiResolver(
		generalConfig,
		stateComponents.AccountsAdapter,
		stateComponents.PeerAccounts,
		stateComponents.AddressPubkeyConverter,
		dataComponents.Store,
		dataComponents.Datapool,
		dataComponents.Blkc,
		coreComponents.InternalMarshalizer,
		coreComponents.Hasher,
		coreComponents.Uint64ByteSliceConverter,
		shardCoordinator,
		statusHandlersInfo.StatusMetrics,
		gasScheduleNotifier,
		economicsData,
		cryptoComponents.MessageSignVerifier,
		genesisNodesConfig,
		systemSCConfig,
		rater,
		epochNotifier,
		apiWorkingDir,
	)
	if err != nil {
		return err
	}

	log.Trace("starting status pooling components")
	statusPollingInterval := time.Duration(generalConfig.GeneralSettings.StatusPollingIntervalSec) * time.Second
	err = metrics.StartStatusPolling(
		currentNode.GetAppStatusHandler(),
		statusPollingInterval,
		networkComponents,
		processComponents,
		shardCoordinator,
	)
	if err != nil {
		return err
	}

	updateMachineStatisticsDuration := time.Second
	err = metrics.StartMachineStatisticsPolling(coreComponents.StatusHandler, epochStartNotifier, updateMachineStatisticsDuration)
	if err != nil {
		return err
	}

	log.Trace("creating elrond node facade")
	restAPIServerDebugMode := ctx.GlobalBool(restApiDebug.Name)

	argNodeFacade := facade.ArgNodeFacade{
		Node:                   currentNode,
		ApiResolver:            apiResolver,
		TxSimulatorProcessor:   transactionSimulator,
		RestAPIServerDebugMode: restAPIServerDebugMode,
		WsAntifloodConfig:      generalConfig.Antiflood.WebServer,
		FacadeConfig: config.FacadeConfig{
			RestApiInterface: ctx.GlobalString(restApiInterface.Name),
			PprofEnabled:     ctx.GlobalBool(profileMode.Name),
		},
		ApiRoutesConfig: *apiRoutesConfig,
		AccountsState:   stateComponents.AccountsAdapter,
		PeerState:       stateComponents.PeerAccounts,
	}

	ef, err := facade.NewNodeFacade(argNodeFacade)
	if err != nil {
		return fmt.Errorf("%w while creating NodeFacade", err)
	}

	ef.SetSyncer(syncer)
	ef.SetTpsBenchmark(tpsBenchmark)

	log.Trace("starting background services")
	ef.StartBackgroundServices()

	log.Debug("starting node...")
	err = ef.StartNode()
	if err != nil {
		log.Error("starting node failed", "epoch", currentEpoch, "error", err.Error())
		return err
	}

	log.Info("application is now running")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	var sig endProcess.ArgEndProcess
	select {
	case <-sigs:
		log.Info("terminating at user's signal...")
	case sig = <-chanStopNodeProcess:
		log.Info("terminating at internal stop signal", "reason", sig.Reason, "description", sig.Description)
	}

	chanCloseComponents := make(chan struct{})
	go func() {
		closeAllComponents(log, healthService, dataComponents, triesComponents, networkComponents, chanCloseComponents)
	}()

	select {
	case <-chanCloseComponents:
	case <-time.After(maxTimeToClose):
		log.Warn("force closing the node", "error", "closeAllComponents did not finished on time")
	}

	log.Debug("closing node")
	if !check.IfNil(fileLogging) {
		err = fileLogging.Close()
		log.LogIfError(err)
	}

	return nil
}

func applyCompatibleConfigs(isInImportMode bool, importDbNoSigCheckFlag bool, log logger.Logger, config *config.Config, p2pConfig *config.P2PConfig) {
	if isInImportMode {
		importCheckpointRoundsModulus := uint(config.EpochStartConfig.RoundsPerEpoch)
		log.Warn("the node is in import mode! Will auto-set some config values, including storage config values",
			"GeneralSettings.StartInEpochEnabled", "false",
			"StateTriesConfig.CheckpointRoundsModulus", importCheckpointRoundsModulus,
			"StoragePruning.NumActivePersisters", config.StoragePruning.NumEpochsToKeep,
			"TrieStorageManagerConfig.MaxSnapshots", math.MaxUint32,
			"p2p.ThresholdMinConnectedPeers", 0,
			"no sig check", importDbNoSigCheckFlag,
			"heartbeat sender", "off",
		)
		config.GeneralSettings.StartInEpochEnabled = false
		config.StateTriesConfig.CheckpointRoundsModulus = importCheckpointRoundsModulus
		config.StoragePruning.NumActivePersisters = config.StoragePruning.NumEpochsToKeep
		config.TrieStorageManagerConfig.MaxSnapshots = math.MaxUint32
		p2pConfig.Node.ThresholdMinConnectedPeers = 0
		config.Heartbeat.DurationToConsiderUnresponsiveInSec = math.MaxInt32
		config.Heartbeat.MinTimeToWaitBetweenBroadcastsInSec = math.MaxInt32 - 2
		config.Heartbeat.MaxTimeToWaitBetweenBroadcastsInSec = math.MaxInt32 - 1

		alterStorageConfigsForDBImport(config)
	}
}

func alterStorageConfigsForDBImport(config *config.Config) {
	changeStorageConfigForDBImport(&config.MiniBlocksStorage)
	changeStorageConfigForDBImport(&config.BlockHeaderStorage)
	changeStorageConfigForDBImport(&config.MetaBlockStorage)
	changeStorageConfigForDBImport(&config.ShardHdrNonceHashStorage)
	changeStorageConfigForDBImport(&config.MetaHdrNonceHashStorage)
	changeStorageConfigForDBImport(&config.PeerAccountsTrieStorage)
}

func changeStorageConfigForDBImport(storageConfig *config.StorageConfig) {
	alterCoefficient := uint32(10)

	storageConfig.Cache.Capacity = storageConfig.Cache.Capacity * alterCoefficient
	storageConfig.DB.MaxBatchSize = storageConfig.DB.MaxBatchSize * int(alterCoefficient)
}

func closeAllComponents(
	log logger.Logger,
	healthService io.Closer,
	dataComponents *mainFactory.DataComponents,
	triesComponents *mainFactory.TriesComponents,
	networkComponents *mainFactory.NetworkComponents,
	chanCloseComponents chan struct{},
) {
	log.Debug("closing health service...")
	err := healthService.Close()
	log.LogIfError(err)

	log.Debug("closing all store units....")
	err = dataComponents.Store.CloseAll()
	log.LogIfError(err)

	dataTries := triesComponents.TriesContainer.GetAll()
	for _, trie := range dataTries {
		err = trie.ClosePersister()
		log.LogIfError(err)
	}

	log.Debug("calling close on the network messenger instance...")
	err = networkComponents.NetMessenger.Close()
	log.LogIfError(err)

	chanCloseComponents <- struct{}{}
}

func createStringFromRatingsData(ratingsData *rating.RatingsData) string {
	metaChainStepHandler := ratingsData.MetaChainRatingsStepHandler()
	shardChainHandler := ratingsData.ShardChainRatingsStepHandler()
	computedRatingsDataStr := fmt.Sprintf(
		"meta:\n"+
			"ProposerIncrease=%v\n"+
			"ProposerDecrease=%v\n"+
			"ValidatorIncrease=%v\n"+
			"ValidatorDecrease=%v\n\n"+
			"shard:\n"+
			"ProposerIncrease=%v\n"+
			"ProposerDecrease=%v\n"+
			"ValidatorIncrease=%v\n"+
			"ValidatorDecrease=%v",
		metaChainStepHandler.ProposerIncreaseRatingStep(),
		metaChainStepHandler.ProposerDecreaseRatingStep(),
		metaChainStepHandler.ValidatorIncreaseRatingStep(),
		metaChainStepHandler.ValidatorDecreaseRatingStep(),
		shardChainHandler.ProposerIncreaseRatingStep(),
		shardChainHandler.ProposerDecreaseRatingStep(),
		shardChainHandler.ValidatorIncreaseRatingStep(),
		shardChainHandler.ValidatorDecreaseRatingStep(),
	)
	return computedRatingsDataStr
}

func cleanupStorageIfNecessary(workingDir string, ctx *cli.Context, log logger.Logger) error {
	storageCleanupFlagValue := ctx.GlobalBool(storageCleanup.Name)
	if storageCleanupFlagValue {
		dbPath := filepath.Join(
			workingDir,
			factory.DefaultDBPath)
		log.Trace("cleaning storage", "path", dbPath)
		err := os.RemoveAll(dbPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func copyConfigToStatsFolder(log logger.Logger, statsFolder string, configs []string) {
	err := os.MkdirAll(statsFolder, os.ModePerm)
	log.LogIfError(err)

	for _, configFile := range configs {
		copySingleFile(statsFolder, configFile)
	}
}

func copySingleFile(folder string, configFile string) {
	fileName := filepath.Base(configFile)

	source, err := core.OpenFile(configFile)
	if err != nil {
		return
	}
	defer func() {
		err = source.Close()
		if err != nil {
			fmt.Printf("Could not close %s\n", source.Name())
		}
	}()

	destPath := filepath.Join(folder, fileName)
	destination, err := os.Create(destPath)
	if err != nil {
		return
	}
	defer func() {
		err = destination.Close()
		if err != nil {
			fmt.Printf("Could not close %s\n", source.Name())
		}
	}()

	_, err = io.Copy(destination, source)
	if err != nil {
		fmt.Printf("Could not copy %s\n", source.Name())
	}
}

func getWorkingDir(ctx *cli.Context, log logger.Logger) string {
	var workingDir string
	var err error
	if ctx.IsSet(workingDirectory.Name) {
		workingDir = ctx.GlobalString(workingDirectory.Name)
	} else {
		workingDir, err = os.Getwd()
		if err != nil {
			log.LogIfError(err)
			workingDir = ""
		}
	}
	log.Trace("working directory", "path", workingDir)

	return workingDir
}

func indexValidatorsListIfNeeded(
	elasticIndexer process.Indexer,
	coordinator sharding.NodesCoordinator,
	epoch uint32,
	log logger.Logger,

) {
	if check.IfNil(elasticIndexer) {
		return
	}

	validatorsPubKeys, err := coordinator.GetAllEligibleValidatorsPublicKeys(epoch)
	if err != nil {
		log.Warn("GetAllEligibleValidatorPublicKeys for epoch 0 failed", "error", err)
	}

	if len(validatorsPubKeys) > 0 {
		elasticIndexer.SaveValidatorsPubKeys(validatorsPubKeys, epoch)
	}
}

func enableGopsIfNeeded(ctx *cli.Context, log logger.Logger) {
	var gopsEnabled bool
	if ctx.IsSet(gopsEn.Name) {
		gopsEnabled = ctx.GlobalBool(gopsEn.Name)
	}

	if gopsEnabled {
		if err := agent.Listen(agent.Options{}); err != nil {
			log.Error("failure to init gops", "error", err.Error())
		}
	}

	log.Trace("gops", "enabled", gopsEnabled)
}

func loadMainConfig(filepath string) (*config.Config, error) {
	cfg := &config.Config{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadApiConfig(filepath string) (*config.ApiRoutesConfig, error) {
	cfg := &config.ApiRoutesConfig{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadEconomicsConfig(filepath string) (*config.EconomicsConfig, error) {
	cfg := &config.EconomicsConfig{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadSystemSmartContractsConfig(filepath string) (*config.SystemSmartContractsConfig, error) {
	cfg := &config.SystemSmartContractsConfig{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadRatingsConfig(filepath string) (config.RatingsConfig, error) {
	cfg := &config.RatingsConfig{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return config.RatingsConfig{}, err
	}

	return *cfg, nil
}

func loadPreferencesConfig(filepath string) (*config.Preferences, error) {
	cfg := &config.Preferences{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadExternalConfig(filepath string) (*config.ExternalConfig, error) {
	cfg := &config.ExternalConfig{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, fmt.Errorf("cannot load external config: %w", err)
	}

	return cfg, nil
}

func getShardIdFromNodePubKey(pubKey crypto.PublicKey, nodesConfig *sharding.NodesSetup) (uint32, error) {
	if pubKey == nil {
		return 0, errors.New("nil public key")
	}

	publicKey, err := pubKey.ToByteArray()
	if err != nil {
		return 0, err
	}

	selfShardId, err := nodesConfig.GetShardIDForPubKey(publicKey)
	if err != nil {
		return 0, err
	}

	return selfShardId, err
}

func createShardCoordinator(
	nodesConfig *sharding.NodesSetup,
	pubKey crypto.PublicKey,
	prefsConfig config.PreferencesConfig,
	log logger.Logger,
) (sharding.Coordinator, core.NodeType, error) {

	selfShardId, err := getShardIdFromNodePubKey(pubKey, nodesConfig)
	nodeType := core.NodeTypeValidator
	if err == sharding.ErrPublicKeyNotFoundInGenesis {
		nodeType = core.NodeTypeObserver
		log.Info("starting as observer node")

		selfShardId, err = processDestinationShardAsObserver(prefsConfig)
		if err != nil {
			return nil, "", err
		}
		if selfShardId == core.DisabledShardIDAsObserver {
			selfShardId = uint32(0)
		}
	}
	if err != nil {
		return nil, "", err
	}

	var shardName string
	if selfShardId == core.MetachainShardId {
		shardName = metachainShardName
	} else {
		shardName = fmt.Sprintf("%d", selfShardId)
	}
	log.Info("shard info", "started in shard", shardName)

	shardCoordinator, err := sharding.NewMultiShardCoordinator(nodesConfig.NumberOfShards(), selfShardId)
	if err != nil {
		return nil, "", err
	}

	return shardCoordinator, nodeType, nil
}

func createNodesCoordinator(
	log logger.Logger,
	nodesConfig *sharding.NodesSetup,
	prefsConfig config.PreferencesConfig,
	epochStartNotifier epochStart.RegistrationHandler,
	pubKey crypto.PublicKey,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	ratingAndListIndexHandler sharding.PeerAccountListAndRatingHandler,
	bootStorer storage.Storer,
	nodeShuffler sharding.NodesShuffler,
	epochConfig config.EpochStartConfig,
	currentShardID uint32,
	chanStopNodeProcess chan endProcess.ArgEndProcess,
	bootstrapParameters bootstrap.Parameters,
	startEpoch uint32,
) (sharding.NodesCoordinator, update.Closer, error) {
	shardIDAsObserver, err := processDestinationShardAsObserver(prefsConfig)
	if err != nil {
		return nil, nil, err
	}
	if shardIDAsObserver == core.DisabledShardIDAsObserver {
		shardIDAsObserver = uint32(0)
	}

	nbShards := nodesConfig.NumberOfShards()
	shardConsensusGroupSize := int(nodesConfig.ConsensusGroupSize)
	metaConsensusGroupSize := int(nodesConfig.MetaChainConsensusGroupSize)
	eligibleNodesInfo, waitingNodesInfo := nodesConfig.InitialNodesInfo()

	eligibleValidators, errEligibleValidators := sharding.NodesInfoToValidators(eligibleNodesInfo)
	if errEligibleValidators != nil {
		return nil, nil, errEligibleValidators
	}

	waitingValidators, errWaitingValidators := sharding.NodesInfoToValidators(waitingNodesInfo)
	if errWaitingValidators != nil {
		return nil, nil, errWaitingValidators
	}

	currentEpoch := startEpoch
	if bootstrapParameters.NodesConfig != nil {
		nodeRegistry := bootstrapParameters.NodesConfig
		currentEpoch = bootstrapParameters.Epoch
		eligibles := nodeRegistry.EpochsConfig[fmt.Sprintf("%d", currentEpoch)].EligibleValidators
		eligibleValidators, err = sharding.SerializableValidatorsToValidators(eligibles)
		if err != nil {
			return nil, nil, err
		}

		waitings := nodeRegistry.EpochsConfig[fmt.Sprintf("%d", currentEpoch)].WaitingValidators
		waitingValidators, err = sharding.SerializableValidatorsToValidators(waitings)
		if err != nil {
			return nil, nil, err
		}
	}

	pubKeyBytes, err := pubKey.ToByteArray()
	if err != nil {
		return nil, nil, err
	}

	consensusGroupCache, err := lrucache.NewCache(25000)
	if err != nil {
		return nil, nil, err
	}

	maxThresholdEpochDuration := epochConfig.MaxShuffledOutRestartThreshold
	if !(maxThresholdEpochDuration >= 0.0 && maxThresholdEpochDuration <= 1.0) {
		return nil, nil, fmt.Errorf("invalid max threshold for shuffled out handler")
	}
	minThresholdEpochDuration := epochConfig.MinShuffledOutRestartThreshold
	if !(minThresholdEpochDuration >= 0.0 && minThresholdEpochDuration <= 1.0) {
		return nil, nil, fmt.Errorf("invalid min threshold for shuffled out handler")
	}

	epochDuration := int64(nodesConfig.RoundDuration) * epochConfig.RoundsPerEpoch
	minDurationBeforeStopProcess := int64(minThresholdEpochDuration * float64(epochDuration))
	maxDurationBeforeStopProcess := int64(maxThresholdEpochDuration * float64(epochDuration))

	minDurationInterval := time.Millisecond * time.Duration(minDurationBeforeStopProcess)
	maxDurationInterval := time.Millisecond * time.Duration(maxDurationBeforeStopProcess)

	log.Debug("closing.NewShuffleOutCloser",
		"minDurationInterval", minDurationInterval,
		"maxDurationInterval", maxDurationInterval,
	)

	nodeShufflerOut, err := closing.NewShuffleOutCloser(
		minDurationInterval,
		maxDurationInterval,
		chanStopNodeProcess,
	)
	if err != nil {
		return nil, nil, err
	}
	shuffledOutHandler, err := sharding.NewShuffledOutTrigger(pubKeyBytes, currentShardID, nodeShufflerOut.EndOfProcessingHandler)
	if err != nil {
		return nil, nil, err
	}

	argumentsNodesCoordinator := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: shardConsensusGroupSize,
		MetaConsensusGroupSize:  metaConsensusGroupSize,
		Marshalizer:             marshalizer,
		Hasher:                  hasher,
		Shuffler:                nodeShuffler,
		EpochStartNotifier:      epochStartNotifier,
		BootStorer:              bootStorer,
		ShardIDAsObserver:       shardIDAsObserver,
		NbShards:                nbShards,
		EligibleNodes:           eligibleValidators,
		WaitingNodes:            waitingValidators,
		SelfPublicKey:           pubKeyBytes,
		ConsensusGroupCache:     consensusGroupCache,
		ShuffledOutHandler:      shuffledOutHandler,
		Epoch:                   currentEpoch,
		StartEpoch:              startEpoch,
	}

	baseNodesCoordinator, err := sharding.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)
	if err != nil {
		return nil, nil, err
	}

	nodesCoordinator, err := sharding.NewIndexHashedNodesCoordinatorWithRater(baseNodesCoordinator, ratingAndListIndexHandler)
	if err != nil {
		return nil, nil, err
	}

	return nodesCoordinator, nodeShufflerOut, nil
}

func processDestinationShardAsObserver(prefsConfig config.PreferencesConfig) (uint32, error) {
	destShard := strings.ToLower(prefsConfig.DestinationShardAsObserver)
	if len(destShard) == 0 {
		return 0, errors.New("option DestinationShardAsObserver is not set in prefs.toml")
	}

	if destShard == notSetDestinationShardID {
		return core.DisabledShardIDAsObserver, nil
	}

	if destShard == metachainShardName {
		return core.MetachainShardId, nil
	}

	val, err := strconv.ParseUint(destShard, 10, 32)
	if err != nil {
		return 0, errors.New("error parsing DestinationShardAsObserver option: " + err.Error())
	}

	return uint32(val), err
}

// createElasticIndexer creates a new elasticIndexer where the server listens on the url,
// authentication for the server is using the username and password
func createElasticIndexer(
	elasticSearchConfig config.ElasticSearchConfig,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	nodesCoordinator sharding.NodesCoordinator,
	startNotifier notifier.EpochStartNotifier,
	addressPubkeyConverter core.PubkeyConverter,
	validatorPubkeyConverter core.PubkeyConverter,
	accountsDB state.AccountsAdapter,
	denomination int,
	shardCoordinator sharding.Coordinator,
	economicsHandler process.TransactionFeeCalculator,
	isInImportDBMode bool,
) (process.Indexer, error) {

	indexerFactoryArgs := &indexerFactory.ArgsIndexerFactory{
		Enabled:                  elasticSearchConfig.Enabled,
		IndexerCacheSize:         elasticSearchConfig.IndexerCacheSize,
		ShardCoordinator:         shardCoordinator,
		Url:                      elasticSearchConfig.URL,
		UserName:                 elasticSearchConfig.Username,
		Password:                 elasticSearchConfig.Password,
		Marshalizer:              marshalizer,
		Hasher:                   hasher,
		EpochStartNotifier:       startNotifier,
		NodesCoordinator:         nodesCoordinator,
		AddressPubkeyConverter:   addressPubkeyConverter,
		ValidatorPubkeyConverter: validatorPubkeyConverter,
		EnabledIndexes:           elasticSearchConfig.EnabledIndexes,
		AccountsDB:               accountsDB,
		Denomination:             denomination,
		TransactionFeeCalculator: economicsHandler,
		Options: &elasticIndexer.Options{
			UseKibana: elasticSearchConfig.UseKibana,
		},
		IsInImportDBMode: isInImportDBMode,
	}

	return indexerFactory.NewIndexer(indexerFactoryArgs)
}
func getConsensusGroupSize(nodesConfig *sharding.NodesSetup, shardCoordinator sharding.Coordinator) (uint32, error) {
	if shardCoordinator.SelfId() == core.MetachainShardId {
		return nodesConfig.MetaChainConsensusGroupSize, nil
	}
	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		return nodesConfig.ConsensusGroupSize, nil
	}

	return 0, state.ErrUnknownShardId
}

func createHardForkTrigger(
	config *config.Config,
	keyGen crypto.KeyGenerator,
	pubKey crypto.PublicKey,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	coreData *mainFactory.CoreComponents,
	stateComponents *mainFactory.StateComponents,
	data *mainFactory.DataComponents,
	crypto *mainFactory.CryptoComponents,
	process *factory.Process,
	network *mainFactory.NetworkComponents,
	whiteListRequest process.WhiteListHandler,
	whiteListerVerifiedTxs process.WhiteListHandler,
	chanStopNodeProcess chan endProcess.ArgEndProcess,
	epochStartNotifier factory.EpochStartNotifier,
	importStartHandler update.ImportStartHandler,
	nodesSetup update.GenesisNodesSetupHandler,
	workingDir string,
	epochNotifier process.EpochNotifier,
) (node.HardforkTrigger, error) {

	selfPubKeyBytes, err := pubKey.ToByteArray()
	if err != nil {
		return nil, err
	}
	triggerPubKeyBytes, err := stateComponents.ValidatorPubkeyConverter.Decode(config.Hardfork.PublicKeyToListenFrom)
	if err != nil {
		return nil, fmt.Errorf("%w while decoding HardforkConfig.PublicKeyToListenFrom", err)
	}

	accountsDBs := make(map[state.AccountsDbIdentifier]state.AccountsAdapter)
	accountsDBs[state.UserAccountsState] = stateComponents.AccountsAdapter
	accountsDBs[state.PeerAccountsState] = stateComponents.PeerAccounts
	hardForkConfig := config.Hardfork
	exportFolder := filepath.Join(workingDir, hardForkConfig.ImportFolder)
	argsExporter := exportFactory.ArgsExporter{
		TxSignMarshalizer:         coreData.TxSignMarshalizer,
		Marshalizer:               coreData.InternalMarshalizer,
		Hasher:                    coreData.Hasher,
		HeaderValidator:           process.HeaderValidator,
		Uint64Converter:           coreData.Uint64ByteSliceConverter,
		DataPool:                  data.Datapool,
		StorageService:            data.Store,
		RequestHandler:            process.RequestHandler,
		ShardCoordinator:          shardCoordinator,
		Messenger:                 network.NetMessenger,
		ActiveAccountsDBs:         accountsDBs,
		ExistingResolvers:         process.ResolversFinder,
		ExportFolder:              exportFolder,
		ExportTriesStorageConfig:  hardForkConfig.ExportTriesStorageConfig,
		ExportStateStorageConfig:  hardForkConfig.ExportStateStorageConfig,
		ExportStateKeysConfig:     hardForkConfig.ExportKeysStorageConfig,
		WhiteListHandler:          whiteListRequest,
		WhiteListerVerifiedTxs:    whiteListerVerifiedTxs,
		InterceptorsContainer:     process.InterceptorsContainer,
		MultiSigner:               crypto.MultiSigner,
		NodesCoordinator:          nodesCoordinator,
		SingleSigner:              crypto.TxSingleSigner,
		AddressPubKeyConverter:    stateComponents.AddressPubkeyConverter,
		ValidatorPubKeyConverter:  stateComponents.ValidatorPubkeyConverter,
		BlockKeyGen:               keyGen,
		KeyGen:                    crypto.TxSignKeyGen,
		BlockSigner:               crypto.SingleSigner,
		HeaderSigVerifier:         process.HeaderSigVerifier,
		HeaderIntegrityVerifier:   process.HeaderIntegrityVerifier,
		MaxTrieLevelInMemory:      config.StateTriesConfig.MaxStateTrieLevelInMemory,
		InputAntifloodHandler:     network.InputAntifloodHandler,
		OutputAntifloodHandler:    network.OutputAntifloodHandler,
		ValidityAttester:          process.BlockTracker,
		ChainID:                   coreData.ChainID,
		RoundHandler:              process.Rounder,
		GenesisNodesSetupHandler:  nodesSetup,
		InterceptorDebugConfig:    config.Debug.InterceptorResolver,
		MinTxVersion:              coreData.MinTransactionVersion,
		EnableSignTxWithHashEpoch: config.GeneralSettings.TransactionSignedWithTxHashEnableEpoch,
		TxSignHasher:              coreData.TxSignHasher,
		EpochNotifier:             epochNotifier,
		NumConcurrentTrieSyncers:  config.TrieSync.NumConcurrentTrieSyncers,
		MaxHardCapForMissingNodes: config.TrieSync.MaxHardCapForMissingNodes,
		TrieSyncerVersion:         config.TrieSync.TrieSyncerVersion,
	}
	hardForkExportFactory, err := exportFactory.NewExportHandlerFactory(argsExporter)
	if err != nil {
		return nil, err
	}

	atArgumentParser := smartContract.NewArgumentParser()
	argTrigger := trigger.ArgHardforkTrigger{
		TriggerPubKeyBytes:        triggerPubKeyBytes,
		SelfPubKeyBytes:           selfPubKeyBytes,
		Enabled:                   config.Hardfork.EnableTrigger,
		EnabledAuthenticated:      config.Hardfork.EnableTriggerFromP2P,
		ArgumentParser:            atArgumentParser,
		EpochProvider:             process.EpochStartTrigger,
		ExportFactoryHandler:      hardForkExportFactory,
		ChanStopNodeProcess:       chanStopNodeProcess,
		EpochConfirmedNotifier:    epochStartNotifier,
		CloseAfterExportInMinutes: config.Hardfork.CloseAfterExportInMinutes,
		ImportStartHandler:        importStartHandler,
		RoundHandler:              process.Rounder,
	}
	hardforkTrigger, err := trigger.NewTrigger(argTrigger)
	if err != nil {
		return nil, err
	}

	return hardforkTrigger, nil
}

func createNode(
	config *config.Config,
	ratingConfig config.RatingsConfig,
	preferencesConfig *config.Preferences,
	nodesConfig *sharding.NodesSetup,
	economicsData process.FeeHandler,
	syncer ntp.SyncTimer,
	keyGen crypto.KeyGenerator,
	privKey crypto.PrivateKey,
	pubKey crypto.PublicKey,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	coreData *mainFactory.CoreComponents,
	stateComponents *mainFactory.StateComponents,
	data *mainFactory.DataComponents,
	crypto *mainFactory.CryptoComponents,
	process *factory.Process,
	network *mainFactory.NetworkComponents,
	bootstrapRoundIndex uint64,
	version string,
	esIndexer process.Indexer,
	requestedItemsHandler dataRetriever.RequestedItemsHandler,
	epochStartRegistrationHandler epochStart.RegistrationHandler,
	whiteListRequest process.WhiteListHandler,
	whiteListerVerifiedTxs process.WhiteListHandler,
	chanStopNodeProcess chan endProcess.ArgEndProcess,
	hardForkTrigger node.HardforkTrigger,
	historyRepository dblookupext.HistoryRepository,
	fallbackHeaderValidator consensus.FallbackHeaderValidator,
	isInImportDbMode bool,
	nodeRedundancyHandler consensus.NodeRedundancyHandler,
) (*node.Node, error) {
	var err error
	var consensusGroupSize uint32
	consensusGroupSize, err = getConsensusGroupSize(nodesConfig, shardCoordinator)
	if err != nil {
		return nil, err
	}

	var txAccumulator node.Accumulator
	txAccumulatorConfig := config.Antiflood.TxAccumulator
	txAccumulator, err = accumulator.NewTimeAccumulator(
		time.Duration(txAccumulatorConfig.MaxAllowedTimeInMilliseconds)*time.Millisecond,
		time.Duration(txAccumulatorConfig.MaxDeviationTimeInMilliseconds)*time.Millisecond,
	)
	if err != nil {
		return nil, err
	}

	networkShardingCollector, err := factory.PrepareNetworkShardingCollector(
		network,
		config,
		nodesCoordinator,
		shardCoordinator,
		epochStartRegistrationHandler,
		process.EpochStartTrigger.MetaEpoch(),
	)
	if err != nil {
		return nil, err
	}

	factory.PrepareOpenTopics(network.InputAntifloodHandler, shardCoordinator)

	alarmScheduler := alarm.NewAlarmScheduler()
	watchdogTimer, err := watchdog.NewWatchdog(alarmScheduler, chanStopNodeProcess)
	if err != nil {
		return nil, err
	}

	peerDenialEvaluator, err := blackList.NewPeerDenialEvaluator(
		network.PeerBlackListHandler,
		network.PkTimeCache,
		networkShardingCollector,
	)
	if err != nil {
		return nil, err
	}

	err = network.NetMessenger.SetPeerDenialEvaluator(peerDenialEvaluator)
	if err != nil {
		return nil, err
	}

	peerHonestyHandler, err := createPeerHonestyHandler(config, ratingConfig, network.PkTimeCache)
	if err != nil {
		return nil, err
	}

	txVersionCheckerHandler := versioning.NewTxVersionChecker(coreData.MinTransactionVersion)

	var nd *node.Node
	nd, err = node.NewNode(
		node.WithMessenger(network.NetMessenger),
		node.WithHasher(coreData.Hasher),
		node.WithInternalMarshalizer(coreData.InternalMarshalizer, config.Marshalizer.SizeCheckDelta),
		node.WithVmMarshalizer(coreData.VmMarshalizer),
		node.WithTxSignMarshalizer(coreData.TxSignMarshalizer),
		node.WithTxFeeHandler(economicsData),
		node.WithInitialNodesPubKeys(crypto.InitialPubKeys),
		node.WithAddressPubkeyConverter(stateComponents.AddressPubkeyConverter),
		node.WithValidatorPubkeyConverter(stateComponents.ValidatorPubkeyConverter),
		node.WithAccountsAdapter(stateComponents.AccountsAdapter),
		node.WithBlockChain(data.Blkc),
		node.WithDataStore(data.Store),
		node.WithRoundDuration(nodesConfig.RoundDuration),
		node.WithConsensusGroupSize(int(consensusGroupSize)),
		node.WithSyncer(syncer),
		node.WithBlockProcessor(process.BlockProcessor),
		node.WithGenesisTime(time.Unix(nodesConfig.StartTime, 0)),
		node.WithRounder(process.Rounder),
		node.WithShardCoordinator(shardCoordinator),
		node.WithNodesCoordinator(nodesCoordinator),
		node.WithUint64ByteSliceConverter(coreData.Uint64ByteSliceConverter),
		node.WithSingleSigner(crypto.SingleSigner),
		node.WithMultiSigner(crypto.MultiSigner),
		node.WithKeyGen(keyGen),
		node.WithKeyGenForAccounts(crypto.TxSignKeyGen),
		node.WithPubKey(pubKey),
		node.WithPrivKey(privKey),
		node.WithForkDetector(process.ForkDetector),
		node.WithInterceptorsContainer(process.InterceptorsContainer),
		node.WithResolversFinder(process.ResolversFinder),
		node.WithConsensusType(config.Consensus.Type),
		node.WithTxSingleSigner(crypto.TxSingleSigner),
		node.WithBootstrapRoundIndex(bootstrapRoundIndex),
		node.WithAppStatusHandler(coreData.StatusHandler),
		node.WithIndexer(esIndexer),
		node.WithEpochStartTrigger(process.EpochStartTrigger),
		node.WithEpochStartEventNotifier(epochStartRegistrationHandler),
		node.WithBlockBlackListHandler(process.BlackListHandler),
		node.WithPeerDenialEvaluator(peerDenialEvaluator),
		node.WithNetworkShardingCollector(networkShardingCollector),
		node.WithBootStorer(process.BootStorer),
		node.WithRequestedItemsHandler(requestedItemsHandler),
		node.WithHeaderSigVerifier(process.HeaderSigVerifier),
		node.WithHeaderIntegrityVerifier(process.HeaderIntegrityVerifier),
		node.WithValidatorStatistics(process.ValidatorsStatistics),
		node.WithValidatorsProvider(process.ValidatorsProvider),
		node.WithChainID(coreData.ChainID),
		node.WithMinTransactionVersion(nodesConfig.MinTransactionVersion),
		node.WithBlockTracker(process.BlockTracker),
		node.WithRequestHandler(process.RequestHandler),
		node.WithInputAntifloodHandler(network.InputAntifloodHandler),
		node.WithTxAccumulator(txAccumulator),
		node.WithHardforkTrigger(hardForkTrigger),
		node.WithWhiteListHandler(whiteListRequest),
		node.WithWhiteListHandlerVerified(whiteListerVerifiedTxs),
		node.WithAddressSignatureSize(config.AddressPubkeyConverter.SignatureLength),
		node.WithValidatorSignatureSize(config.ValidatorPubkeyConverter.SignatureLength),
		node.WithPublicKeySize(config.ValidatorPubkeyConverter.Length),
		node.WithNodeStopChannel(chanStopNodeProcess),
		node.WithPeerHonestyHandler(peerHonestyHandler),
		node.WithFallbackHeaderValidator(fallbackHeaderValidator),
		node.WithWatchdogTimer(watchdogTimer),
		node.WithPeerSignatureHandler(crypto.PeerSignatureHandler),
		node.WithHistoryRepository(historyRepository),
		node.WithEnableSignTxWithHashEpoch(config.GeneralSettings.TransactionSignedWithTxHashEnableEpoch),
		node.WithTxSignHasher(coreData.TxSignHasher),
		node.WithTxVersionChecker(txVersionCheckerHandler),
		node.WithImportMode(isInImportDbMode),
		node.WithNodeRedundancyHandler(nodeRedundancyHandler),
	)
	if err != nil {
		return nil, errors.New("error creating node: " + err.Error())
	}

	err = nd.StartHeartbeat(config.Heartbeat, version, preferencesConfig.Preferences)
	if err != nil {
		return nil, err
	}

	err = nd.ApplyOptions(node.WithDataPool(data.Datapool))
	if err != nil {
		return nil, errors.New("error creating node: " + err.Error())
	}

	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		err = nd.CreateShardedStores()
		if err != nil {
			return nil, err
		}
	}
	if shardCoordinator.SelfId() == core.MetachainShardId {
		err = nd.ApplyOptions(node.WithPendingMiniBlocksHandler(process.PendingMiniBlocksHandler))
		if err != nil {
			return nil, errors.New("error creating meta-node: " + err.Error())
		}
	}

	err = nodeDebugFactory.CreateInterceptedDebugHandler(
		nd,
		process.InterceptorsContainer,
		process.ResolversFinder,
		config.Debug.InterceptorResolver,
	)
	if err != nil {
		return nil, err
	}

	return nd, nil
}

func createPeerHonestyHandler(
	config *config.Config,
	ratingConfig config.RatingsConfig,
	pkTimeCache process.TimeCacher,
) (consensus.PeerHonestyHandler, error) {

	cache, err := storageUnit.NewCache(storageFactory.GetCacherFromConfig(config.PeerHonesty))
	if err != nil {
		return nil, err
	}

	return peerHonesty.NewP2pPeerHonesty(ratingConfig.PeerHonesty, pkTimeCache, cache)
}

func initStatsFileMonitor(
	config *config.Config,
	pathManager storage.PathManagerHandler,
	shardId string,
) error {
	err := startStatisticsMonitor(config, pathManager, shardId)
	if err != nil {
		return err
	}

	return nil
}

func startStatisticsMonitor(
	generalConfig *config.Config,
	pathManager storage.PathManagerHandler,
	shardId string,
) error {
	if !generalConfig.ResourceStats.Enabled {
		return nil
	}

	if generalConfig.ResourceStats.RefreshIntervalInSec < 1 {
		return errors.New("invalid RefreshIntervalInSec in section [ResourceStats]. Should be an integer higher than 1")
	}

	resMon := statistics.NewResourceMonitor()

	go func() {
		for {
			resMon.SaveStatistics(generalConfig, pathManager, shardId)
			time.Sleep(time.Second * time.Duration(generalConfig.ResourceStats.RefreshIntervalInSec))
		}
	}()

	return nil
}

func createApiResolver(
	generalConfig *config.Config,
	accnts state.AccountsAdapter,
	validatorAccounts state.AccountsAdapter,
	pubkeyConv core.PubkeyConverter,
	storageService dataRetriever.StorageService,
	dataPool dataRetriever.PoolsHolder,
	blockChain data.ChainHandler,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
	shardCoordinator sharding.Coordinator,
	statusMetrics external.StatusMetricsHandler,
	gasScheduleNotifier core.GasScheduleNotifier,
	economics process.EconomicsDataHandler,
	messageSigVerifier vm.MessageSignVerifier,
	nodesSetup sharding.GenesisNodesSetupHandler,
	systemSCConfig *config.SystemSmartContractsConfig,
	rater sharding.PeerAccountListAndRatingHandler,
	epochNotifier process.EpochNotifier,
	workingDir string,
) (facade.ApiResolver, error) {
	scQueryService, err := createScQueryService(
		generalConfig,
		accnts,
		validatorAccounts,
		pubkeyConv,
		storageService,
		dataPool,
		blockChain,
		marshalizer,
		hasher,
		uint64Converter,
		shardCoordinator,
		gasScheduleNotifier,
		economics,
		messageSigVerifier,
		nodesSetup,
		systemSCConfig,
		rater,
		epochNotifier,
		workingDir,
	)
	if err != nil {
		return nil, err
	}

	builtInFuncs, err := createBuiltinFuncs(
		gasScheduleNotifier,
		marshalizer,
		accnts,
	)
	if err != nil {
		return nil, err
	}

	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:  pubkeyConv,
		ShardCoordinator: shardCoordinator,
		BuiltInFuncNames: builtInFuncs.Keys(),
		ArgumentParser:   parsers.NewCallArgsParser(),
	}
	txTypeHandler, err := coordinator.NewTxTypeHandler(argsTxTypeHandler)
	if err != nil {
		return nil, err
	}

	txCostHandler, err := transaction.NewTransactionCostEstimator(txTypeHandler, economics, scQueryService, gasScheduleNotifier)
	if err != nil {
		return nil, err
	}

	args := &stakeValuesProcessor.ArgsTotalStakedValueHandler{
		ShardID:             shardCoordinator.SelfId(),
		InternalMarshalizer: marshalizer,
		Accounts:            accnts,
		NodePrice:           systemSCConfig.StakingSystemSCConfig.GenesisNodePrice,
	}
	totalStakedValueHandler, err := stakeValuesProcessor.CreateTotalStakedValueHandler(args)
	if err != nil {
		return nil, err
	}

	return external.NewNodeApiResolver(scQueryService, statusMetrics, txCostHandler, totalStakedValueHandler)
}

//TODO refactor this code when moving into feat/soft-restart. Maybe use arguments instead of endless parameter lists
func createScQueryService(
	generalConfig *config.Config,
	accnts state.AccountsAdapter,
	validatorAccounts state.AccountsAdapter,
	pubkeyConv core.PubkeyConverter,
	storageService dataRetriever.StorageService,
	dataPool dataRetriever.PoolsHolder,
	blockChain data.ChainHandler,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
	shardCoordinator sharding.Coordinator,
	gasScheduleNotifier core.GasScheduleNotifier,
	economics process.EconomicsDataHandler,
	messageSigVerifier vm.MessageSignVerifier,
	nodesSetup sharding.GenesisNodesSetupHandler,
	systemSCConfig *config.SystemSmartContractsConfig,
	rater sharding.PeerAccountListAndRatingHandler,
	epochNotifier process.EpochNotifier,
	workingDir string,
) (process.SCQueryService, error) {
	numConcurrentVms := generalConfig.VirtualMachine.Querying.NumConcurrentVMs
	if numConcurrentVms < 1 {
		return nil, fmt.Errorf("VirtualMachine.Querying.NumConcurrentVms should be a positive number more than 1")
	}

	list := make([]process.SCQueryService, 0, numConcurrentVms)
	for i := 0; i < numConcurrentVms; i++ {
		scQueryService, err := createScQueryElement(
			generalConfig,
			accnts,
			validatorAccounts,
			pubkeyConv,
			storageService,
			dataPool,
			blockChain,
			marshalizer,
			hasher,
			uint64Converter,
			shardCoordinator,
			gasScheduleNotifier,
			economics,
			messageSigVerifier,
			nodesSetup,
			systemSCConfig,
			rater,
			epochNotifier,
			workingDir,
			i,
		)

		if err != nil {
			return nil, err
		}

		list = append(list, scQueryService)
	}

	sqQueryDispatcher, err := smartContract.NewScQueryServiceDispatcher(list)
	if err != nil {
		return nil, err
	}

	return sqQueryDispatcher, nil
}

func createScQueryElement(
	generalConfig *config.Config,
	accnts state.AccountsAdapter,
	validatorAccounts state.AccountsAdapter,
	pubkeyConv core.PubkeyConverter,
	storageService dataRetriever.StorageService,
	dataPool dataRetriever.PoolsHolder,
	blockChain data.ChainHandler,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
	shardCoordinator sharding.Coordinator,
	gasScheduleNotifier core.GasScheduleNotifier,
	economics process.EconomicsDataHandler,
	messageSigVerifier vm.MessageSignVerifier,
	nodesSetup sharding.GenesisNodesSetupHandler,
	systemSCConfig *config.SystemSmartContractsConfig,
	rater sharding.PeerAccountListAndRatingHandler,
	epochNotifier process.EpochNotifier,
	workingDir string,
	index int,
) (process.SCQueryService, error) {
	var vmFactory process.VirtualMachinesContainerFactory
	var err error

	builtInFuncs, err := createBuiltinFuncs(
		gasScheduleNotifier,
		marshalizer,
		accnts,
	)
	if err != nil {
		return nil, err
	}

	cacherCfg := storageFactory.GetCacherFromConfig(generalConfig.SmartContractDataPool)
	smartContractsCache, err := storageUnit.NewCache(cacherCfg)
	if err != nil {
		return nil, err
	}

	scStorage := generalConfig.SmartContractsStorageForSCQuery
	scStorage.DB.FilePath += fmt.Sprintf("%d", index)
	argsHook := hooks.ArgBlockChainHook{
		Accounts:           accnts,
		PubkeyConv:         pubkeyConv,
		StorageService:     storageService,
		BlockChain:         blockChain,
		ShardCoordinator:   shardCoordinator,
		Marshalizer:        marshalizer,
		Uint64Converter:    uint64Converter,
		BuiltInFunctions:   builtInFuncs,
		DataPool:           dataPool,
		ConfigSCStorage:    scStorage,
		CompiledSCPool:     smartContractsCache,
		WorkingDir:         workingDir,
		NilCompiledSCStore: true,
	}

	if shardCoordinator.SelfId() == core.MetachainShardId {
		argsNewVmFactory := metachain.ArgsNewVMContainerFactory{
			ArgBlockChainHook:   argsHook,
			Economics:           economics,
			MessageSignVerifier: messageSigVerifier,
			GasSchedule:         gasScheduleNotifier,
			NodesConfigProvider: nodesSetup,
			Hasher:              hasher,
			Marshalizer:         marshalizer,
			SystemSCConfig:      systemSCConfig,
			ValidatorAccountsDB: validatorAccounts,
			ChanceComputer:      rater,
			EpochNotifier:       epochNotifier,
		}
		vmFactory, err = metachain.NewVMContainerFactory(argsNewVmFactory)
		if err != nil {
			return nil, err
		}
	} else {
		queryVirtualMachineConfig := generalConfig.VirtualMachine.Querying.VirtualMachineConfig
		queryVirtualMachineConfig.OutOfProcessEnabled = true
		argsNewVMFactory := shard.ArgVMContainerFactory{
			Config:                         queryVirtualMachineConfig,
			BlockGasLimit:                  economics.MaxGasLimitPerBlock(shardCoordinator.SelfId()),
			GasSchedule:                    gasScheduleNotifier,
			ArgBlockChainHook:              argsHook,
			DeployEnableEpoch:              generalConfig.GeneralSettings.SCDeployEnableEpoch,
			AheadOfTimeGasUsageEnableEpoch: generalConfig.GeneralSettings.AheadOfTimeGasUsageEnableEpoch,
			ArwenV3EnableEpoch:             generalConfig.GeneralSettings.RepairCallbackEnableEpoch,
		}

		vmFactory, err = shard.NewVMContainerFactory(argsNewVMFactory)
		if err != nil {
			return nil, err
		}
	}

	vmContainer, err := vmFactory.Create()
	if err != nil {
		return nil, err
	}

	err = builtInFunctions.SetPayableHandler(builtInFuncs, vmFactory.BlockChainHookImpl())
	if err != nil {
		return nil, err
	}

	return smartContract.NewSCQueryService(vmContainer, economics, vmFactory.BlockChainHookImpl(), blockChain)
}

func createBuiltinFuncs(
	gasScheduleNotifier core.GasScheduleNotifier,
	marshalizer marshal.Marshalizer,
	accnts state.AccountsAdapter,
) (process.BuiltInFunctionContainer, error) {
	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasSchedule:     gasScheduleNotifier,
		MapDNSAddresses: make(map[string]struct{}),
		Marshalizer:     marshalizer,
		Accounts:        accnts,
	}
	builtInFuncFactory, err := builtInFunctions.NewBuiltInFunctionsFactory(argsBuiltIn)
	if err != nil {
		return nil, err
	}

	return builtInFuncFactory.CreateBuiltInFunctionContainer()
}

func createWhiteListerVerifiedTxs(generalConfig *config.Config) (process.WhiteListHandler, error) {
	whiteListCacheVerified, err := storageUnit.NewCache(storageFactory.GetCacherFromConfig(generalConfig.WhiteListerVerifiedTxs))
	if err != nil {
		return nil, err
	}
	return interceptors.NewWhiteListDataVerifier(whiteListCacheVerified)
}
