package config

//ContextFlagsConfig will keep the values for the cli.Context flags
type ContextFlagsConfig struct {
	WorkingDir                   string
	EnableGops                   bool
	SaveLogFile                  bool
	EnableLogCorrelation         bool
	EnableLogName                bool
	LogLevel                     string
	DisableAnsiColor             bool
	CleanupStorage               bool
	UseHealthService             bool
	SessionInfoFileOutput        string
	EnableTxIndexing             bool
	BootstrapRoundIndex          uint64
	RestApiInterface             string
	EnablePprof                  bool
	UseLogView                   bool
	ValidatorKeyIndex            int
	EnableRestAPIServerDebugMode bool
	Version                      string
}

// ImportDbConfig will hold the import-db parameters
type ImportDbConfig struct {
	IsImportDBMode                bool
	ImportDBStartInEpoch          uint32
	ImportDBTargetShardID         uint32
	ImportDBWorkingDir            string
	ImportDbNoSigCheckFlag        bool
	ImportDbSaveTrieEpochRootHash bool
}
