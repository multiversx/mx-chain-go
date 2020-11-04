package config

//ContextFlagsConfig will keep the values for the cli.Context flags
type ContextFlagsConfig struct {
	WorkingDir                       string
	NodesFileName                    string
	EnableGops                       bool
	SaveLogFile                      bool
	EnableLogCorrelation             bool
	EnableLogName                    bool
	LogLevel                         string
	DisableAnsiColor                 bool
	GenesisFileName                  string
	CleanupStorage                   bool
	UseHealthService                 bool
	SessionInfoFileOutput            string
	GasScheduleConfigurationFileName string
	EnableTxIndexing                 bool
	SmartContractsFileName           string
	BootstrapRoundIndex              uint64
	RestApiInterface                 string
	EnablePprof                      bool
	UseLogView                       bool
	ValidatorKeyPemFileName          string
	ValidatorKeyIndex                int
	EnableRestAPIServerDebugMode     bool
	Version                          string
	IsInImportMode                   bool
	ImportDbNoSigCheckFlag           bool
	ElasticSearchTemplatesPath       string
}
