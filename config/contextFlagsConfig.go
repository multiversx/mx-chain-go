package config

// ContextFlagsConfig will keep the values for the cli.Context flags
type ContextFlagsConfig struct {
	WorkingDir                   string
	DbDir                        string
	LogsDir                      string
	EnableGops                   bool
	SaveLogFile                  bool
	EnableLogCorrelation         bool
	EnableLogName                bool
	LogLevel                     string
	DisableAnsiColor             bool
	CleanupStorage               bool
	UseHealthService             bool
	SessionInfoFileOutput        string
	BootstrapRoundIndex          uint64
	RestApiInterface             string
	EnablePprof                  bool
	UseLogView                   bool
	ValidatorKeyIndex            int
	EnableRestAPIServerDebugMode bool
	BaseVersion                  string
	Version                      string
	ForceStartFromNetwork        bool
	DisableConsensusWatchdog     bool
	SerializeSnapshots           bool
	OperationMode                string
	RepopulateTokensSupplies     bool
	P2PPrometheusMetricsEnabled  bool
}

// ImportDbConfig will hold the import-db parameters
type ImportDbConfig struct {
	IsImportDBMode                bool
	ImportDBTargetShardID         uint32
	ImportDBWorkingDir            string
	ImportDbNoSigCheckFlag        bool
	ImportDbSaveTrieEpochRootHash bool
}
