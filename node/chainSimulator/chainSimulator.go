package chainSimulator

import (
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/testdata"
)

type simulator struct {
	chanStopNodeProcess    chan endProcess.ArgEndProcess
	syncedBroadcastNetwork components.SyncedBroadcastNetworkHandler
	nodes                  []ChainHandler
	numOfShards            uint32
}

// NewChainSimulator will create a new instance of simulator
func NewChainSimulator(
	tempDir string,
	numOfShards uint32,
	pathToInitialConfig string,
	genesisTimestamp int64,
	roundDurationInMillis uint64,
) (*simulator, error) {
	syncedBroadcastNetwork := components.NewSyncedBroadcastNetwork()

	instance := &simulator{
		syncedBroadcastNetwork: syncedBroadcastNetwork,
		nodes:                  make([]ChainHandler, 0),
		numOfShards:            numOfShards,
		chanStopNodeProcess:    make(chan endProcess.ArgEndProcess),
	}

	err := instance.createChainHandlers(tempDir, numOfShards, pathToInitialConfig, genesisTimestamp, roundDurationInMillis)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (s *simulator) createChainHandlers(
	tempDir string,
	numOfShards uint32,
	originalConfigPath string,
	genesisTimestamp int64,
	roundDurationInMillis uint64,
) error {
	outputConfigs, err := configs.CreateChainSimulatorConfigs(configs.ArgsChainSimulatorConfigs{
		NumOfShards:               numOfShards,
		OriginalConfigsPath:       originalConfigPath,
		GenesisAddressWithStake:   testdata.GenesisAddressWithStake,
		GenesisAddressWithBalance: testdata.GenesisAddressWithBalance,
		GenesisTimeStamp:          genesisTimestamp,
		RoundDurationInMillis:     roundDurationInMillis,
		TempDir:                   tempDir,
	})
	if err != nil {
		return err
	}

	for idx := range outputConfigs.ValidatorsPrivateKeys {
		chainHandler, errCreate := s.createChainHandler(outputConfigs.Configs, idx, outputConfigs.GasScheduleFilename)
		if errCreate != nil {
			return errCreate
		}

		s.nodes = append(s.nodes, chainHandler)
	}

	return nil
}

func (s *simulator) createChainHandler(
	configs *config.Configs,
	skIndex int,
	gasScheduleFilename string,
) (ChainHandler, error) {
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
		SkIndex:                  skIndex,
	}

	testNode, err := components.NewTestOnlyProcessingNode(args)
	if err != nil {
		return nil, err
	}

	return process.NewBlocksCreator(testNode)
}

// GenerateBlocks will generate the provided number of blocks
func (s *simulator) GenerateBlocks(numOfBlocks int) error {
	for idx := 0; idx < numOfBlocks; idx++ {
		s.incrementRoundOnAllValidators()
		err := s.allNodesCreateBlocks()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *simulator) incrementRoundOnAllValidators() {
	for _, node := range s.nodes {
		node.IncrementRound()
	}
}

func (s *simulator) allNodesCreateBlocks() error {
	for _, node := range s.nodes {
		err := node.CreateNewBlock()
		if err != nil {
			return err
		}
	}

	return nil
}

// Stop will stop the simulator
func (s *simulator) Stop() {
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *simulator) IsInterfaceNil() bool {
	return s == nil
}
