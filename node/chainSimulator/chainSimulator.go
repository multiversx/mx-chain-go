package chainSimulator

import (
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components"
	"github.com/multiversx/mx-chain-go/testscommon"
)

const (
	NumOfShards = 3
)

type simulator struct {
	chanStopNodeProcess    chan endProcess.ArgEndProcess
	syncedBroadcastNetwork components.SyncedBroadcastNetworkHandler
	nodes                  []ChainHandler
}

func NewChainSimulator() (*simulator, error) {
	syncedBroadcastNetwork := components.NewSyncedBroadcastNetwork()

	return &simulator{
		syncedBroadcastNetwork: syncedBroadcastNetwork,
	}, nil
}

func (s *simulator) createChanHandler(shardID uint32) (ChainHandler, error) {
	generalConfig := testscommon.GetGeneralConfig()

	args := components.ArgsTestOnlyProcessingNode{
		Config:                   generalConfig,
		EpochConfig:              config.EpochConfig{},
		EconomicsConfig:          config.EconomicsConfig{},
		RoundsConfig:             config.RoundConfig{},
		PreferencesConfig:        config.Preferences{},
		ImportDBConfig:           config.ImportDbConfig{},
		ContextFlagsConfig:       config.ContextFlagsConfig{},
		SystemSCConfig:           config.SystemSmartContractsConfig{},
		ConfigurationPathsHolder: config.ConfigurationPathsHolder{},
		ChanStopNodeProcess:      nil,
		SyncedBroadcastNetwork:   s.syncedBroadcastNetwork,
		GasScheduleFilename:      "",
		ValidatorPemFile:         "",
		WorkingDir:               "",
		NodesSetupPath:           "",
		NumShards:                NumOfShards,
		ShardID:                  shardID,
	}

	return components.NewTestOnlyProcessingNode(args)
}

func (s *simulator) GenerateBlocks(numOfBlock int) error {
	return nil
}

func (s *simulator) Stop() {
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *simulator) IsInterfaceNil() bool {
	return s == nil
}
