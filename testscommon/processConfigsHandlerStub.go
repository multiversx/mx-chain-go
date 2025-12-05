package testscommon

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/configs"
	"github.com/multiversx/mx-chain-go/config"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// GetDefaultProcessConfigsHandler -
func GetDefaultProcessConfigsHandler() common.ProcessConfigsHandler {
	processConfigsHandler, _ := configs.NewProcessConfigsHandler([]config.ProcessConfigByEpoch{{
		EnableEpoch:                       0,
		MaxMetaNoncesBehind:               15,
		MaxMetaNoncesBehindForGlobalStuck: 30,
		MaxShardNoncesBehind:              15,
	}},
		[]config.ProcessConfigByRound{
			{
				EnableRound:                        0,
				MaxRoundsWithoutNewBlockReceived:   10,
				MaxRoundsWithoutCommittedBlock:     10,
				RoundModulusTriggerWhenSyncIsStuck: 20,
			},
		},
	)

	return processConfigsHandler
}

// ProcessConfigsHandlerStub -
type ProcessConfigsHandlerStub struct {
	GetMaxMetaNoncesBehindByEpochCalled               func(epoch uint32) uint32
	GetMaxMetaNoncesBehindForGlobalStuckByEpochCalled func(epoch uint32) uint32
	GetMaxShardNoncesBehindByEpochCalled              func(epoch uint32) uint32
	GetMaxRoundsWithoutNewBlockReceivedByRoundCalled  func(round uint64) uint32
	GetMaxRoundsWithoutCommittedBlockCalled           func(round uint64) uint32
	GetRoundModulusTriggerWhenSyncIsStuckCalled       func(round uint64) uint32
}

// GetMaxMetaNoncesBehindByEpoch -
func (p *ProcessConfigsHandlerStub) GetMaxMetaNoncesBehindByEpoch(epoch uint32) uint32 {
	if p.GetMaxMetaNoncesBehindByEpochCalled != nil {
		return p.GetMaxMetaNoncesBehindByEpochCalled(epoch)
	}

	return 0
}

// GetMaxMetaNoncesBehindForGlobalStuckByEpoch -
func (p *ProcessConfigsHandlerStub) GetMaxMetaNoncesBehindForGlobalStuckByEpoch(epoch uint32) uint32 {
	if p.GetMaxMetaNoncesBehindForGlobalStuckByEpochCalled != nil {
		return p.GetMaxMetaNoncesBehindForGlobalStuckByEpochCalled(epoch)
	}

	return 0
}

// GetMaxShardNoncesBehindByEpoch -
func (p *ProcessConfigsHandlerStub) GetMaxShardNoncesBehindByEpoch(epoch uint32) uint32 {
	if p.GetMaxShardNoncesBehindByEpochCalled != nil {
		return p.GetMaxShardNoncesBehindByEpochCalled(epoch)
	}

	return 0
}

// GetMaxRoundsWithoutNewBlockReceivedByRound -
func (p *ProcessConfigsHandlerStub) GetMaxRoundsWithoutNewBlockReceivedByRound(round uint64) uint32 {
	if p.GetMaxRoundsWithoutNewBlockReceivedByRoundCalled != nil {
		return p.GetMaxRoundsWithoutNewBlockReceivedByRoundCalled(round)
	}

	return 0
}

// GetMaxRoundsWithoutCommittedBlock -
func (p *ProcessConfigsHandlerStub) GetMaxRoundsWithoutCommittedBlock(round uint64) uint32 {
	if p.GetMaxRoundsWithoutCommittedBlockCalled != nil {
		return p.GetMaxRoundsWithoutCommittedBlockCalled(round)
	}

	return 0
}

// GetRoundModulusTriggerWhenSyncIsStuck -
func (p *ProcessConfigsHandlerStub) GetRoundModulusTriggerWhenSyncIsStuck(round uint64) uint32 {
	if p.GetRoundModulusTriggerWhenSyncIsStuckCalled != nil {
		return p.GetRoundModulusTriggerWhenSyncIsStuckCalled(round)
	}

	return 0
}

// IsInterfaceNil -
func (p *ProcessConfigsHandlerStub) IsInterfaceNil() bool {
	return p == nil
}

// SetActivationRound -
func (p *ProcessConfigsHandlerStub) SetActivationRound(round uint64, log logger.Logger) {

}
