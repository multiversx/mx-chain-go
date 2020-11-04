package factory

import (
	"bytes"
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
)

// PrintStack -
func PrintStack() {

	buffer := new(bytes.Buffer)
	err := pprof.Lookup("goroutine").WriteTo(buffer, 2)
	if err != nil {
		fmt.Println("could not dump goroutines")
	}

	fmt.Printf("\n%s", buffer.String())
}

// CleanupWorkingDir -
func CleanupWorkingDir() {
	workingDir := WorkingDir
	if _, err := os.Stat(workingDir); !os.IsNotExist(err) {
		err = os.RemoveAll(workingDir)
		if err != nil {
			fmt.Println("CleanupWorkingDir failed:" + err.Error())
		}
	}
}

func CreateDefaultConfig() *config.Configs {
	generalConfig, _ := core.LoadMainConfig(ConfigPath)
	ratingsConfig, _ := core.LoadRatingsConfig(RatingsPath)
	economicsConfig, _ := core.LoadEconomicsConfig(EconomicsPath)
	prefsConfig, _ := core.LoadPreferencesConfig(PrefsPath)
	p2pConfig, _ := core.LoadP2PConfig(P2pPath)
	externalConfig, _ := core.LoadExternalConfig(ExternalPath)
	systemSCConfig, _ := core.LoadSystemSmartContractsConfig(SystemSCConfigPath)

	configs := &config.Configs{}
	configs.GeneralConfig = generalConfig
	configs.RatingsConfig = ratingsConfig
	configs.EconomicsConfig = economicsConfig
	configs.SystemSCConfig = systemSCConfig
	configs.PreferencesConfig = prefsConfig
	configs.P2pConfig = p2pConfig
	configs.ExternalConfig = externalConfig
	configs.FlagsConfig = &config.ContextFlagsConfig{
		WorkingDir:                       "workingDir",
		UseLogView:                       true,
		ValidatorKeyPemFileName:          ValidatorKeyPemPath,
		GasScheduleConfigurationFileName: GasSchedule,
		Version:                          Version,
		GenesisFileName:                  GenesisPath,
		SmartContractsFileName:           GenesisSmartContracts,
		NodesFileName:                    NodesSetupPath,
	}

	return configs
}
