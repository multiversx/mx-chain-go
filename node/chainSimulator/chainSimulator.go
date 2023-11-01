package chainSimulator

import (
	"errors"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
)

const (
	genesisAddressWithStake = "erd10z6sdhwfy8jtuf87j5gnq7lt7fd2wfmhkg8zfzf79lrapzq265yqlnmtm7"
	//genesisAddressWithStakeSK   = "ZWRlZDAyNDczZTE4NjQ2MTY5NzNhZTIwY2IzYjg3NWFhM2ZmZWU1NWE2MGQ5NDgy\nMjhmMzk4ZTQ4OTk1NjA3NTc4YjUwNmRkYzkyMWU0YmUyNGZlOTUxMTMwN2JlYmYy\nNWFhNzI3NzdiMjBlMjQ4OTNlMmZjN2QwODgwYWQ1MDg="
	genesisAddressWithBalance = "erd1rhrm20mmf2pugzxc3twlu3fa264hxeefnglsy4ads4dpccs9s3jsg6qdrz"
	//genesisAddressWithBalanceSK = "YWQxMTM2YTEyNWZkM2YxY2ZiMTU0MTU5NDQyZTRiYzZhM2I1YzMwOTU5NDMwMjk5\nNThhYzQ2NGRhN2NlMTNlYjFkYzdiNTNmN2I0YTgzYzQwOGQ4OGFkZGZlNDUzZDU2\nYWI3MzY3Mjk5YTNmMDI1N2FkODU1YTFjNjIwNTg0NjU="
)

type simulator struct {
	chanStopNodeProcess    chan endProcess.ArgEndProcess
	syncedBroadcastNetwork components.SyncedBroadcastNetworkHandler
	nodes                  []ChainHandler
	numOfShards            uint32
}

func NewChainSimulator(numOfShards uint32, pathToInitialConfig string) (*simulator, error) {
	if pathToInitialConfig == "" {
		return nil, errors.New("invalid provided path to the initial config")
	}

	syncedBroadcastNetwork := components.NewSyncedBroadcastNetwork()

	instance := &simulator{
		syncedBroadcastNetwork: syncedBroadcastNetwork,
		nodes:                  make([]ChainHandler, 0),
		numOfShards:            numOfShards,
		chanStopNodeProcess:    make(chan endProcess.ArgEndProcess),
	}

	err := instance.createChainHandlers(numOfShards, pathToInitialConfig)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (s *simulator) createChainHandlers(numOfShards uint32, originalConfigPath string) error {
	outputConfigs, err := configs.CreateChainSimulatorConfigs(configs.ArgsChainSimulatorConfigs{
		NumOfShards:               numOfShards,
		OriginalConfigsPath:       originalConfigPath,
		GenesisAddressWithStake:   genesisAddressWithStake,
		GenesisAddressWithBalance: genesisAddressWithBalance,
	})
	if err != nil {
		return err
	}

	metaChainHandler, err := s.createChainHandler(core.MetachainShardId, outputConfigs.Configs, 0, outputConfigs.GasScheduleFilename)
	if err != nil {
		return err
	}

	s.nodes = append(s.nodes, metaChainHandler)

	for idx := uint32(0); idx < numOfShards; idx++ {
		shardChainHandler, errS := s.createChainHandler(idx, outputConfigs.Configs, int(idx)+1, outputConfigs.GasScheduleFilename)
		if errS != nil {
			return errS
		}

		s.nodes = append(s.nodes, shardChainHandler)
	}

	return nil
}

func (s *simulator) createChainHandler(shardID uint32, configs *config.Configs, skIndex int, gasScheduleFilename string) (ChainHandler, error) {
	args := components.ArgsTestOnlyProcessingNode{
		Config:                   *configs.GeneralConfig,
		EpochConfig:              *configs.EpochConfig,
		EconomicsConfig:          *configs.EconomicsConfig,
		RoundsConfig:             *configs.RoundConfig,
		PreferencesConfig:        *configs.PreferencesConfig,
		ImportDBConfig:           *configs.ImportDbConfig,
		ContextFlagsConfig:       *configs.FlagsConfig,
		SystemSCConfig:           *configs.SystemSCConfig,
		ConfigurationPathsHolder: *configs.ConfigurationPathsHolder,
		ChanStopNodeProcess:      s.chanStopNodeProcess,
		SyncedBroadcastNetwork:   s.syncedBroadcastNetwork,
		NumShards:                s.numOfShards,
		GasScheduleFilename:      gasScheduleFilename,
		ShardID:                  shardID,
		SkKeyIndex:               skIndex,
	}

	return components.NewTestOnlyProcessingNode(args)
}

func (s *simulator) GenerateBlocks(numOfBlocks int) error {
	for idx := 0; idx < numOfBlocks; idx++ {
		for _, node := range s.nodes {
			err := node.ProcessBlock()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *simulator) Stop() {
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *simulator) IsInterfaceNil() bool {
	return s == nil
}
