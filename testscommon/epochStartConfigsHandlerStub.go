package testscommon

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/configs"
	"github.com/multiversx/mx-chain-go/config"
)

// GetDefaultCommonConfigsHandler -
func GetDefaultCommonConfigsHandler() common.CommonConfigsHandler {
	commonConfigsHandler, _ := configs.NewCommonConfigsHandler(
		[]config.EpochStartConfigByEpoch{
			{EnableEpoch: 0, GracePeriodRounds: 25, ExtraDelayForRequestBlockInfoInMilliseconds: 3000},
		},
		[]config.EpochStartConfigByRound{
			{EnableRound: 0, MaxRoundsWithoutCommittedStartInEpochBlock: 50},
		},
		[]config.ConsensusConfigByEpoch{
			{EnableEpoch: 0, NumRoundsToWaitBeforeSignalingChronologyStuck: 10},
		},
	)

	return commonConfigsHandler
}

// CommonConfigsHandlerStub -
type CommonConfigsHandlerStub struct {
	GetGracePeriodRoundsByEpochCalled                          func(epoch uint32) uint32
	GetExtraDelayForRequestBlockInfoInMsCalled                 func(epoch uint32) uint32
	GetMaxRoundsWithoutCommittedStartInEpochBlockInRoundCalled func(round uint64) uint32
}

// GetGracePeriodRoundsByEpoch -
func (e *CommonConfigsHandlerStub) GetGracePeriodRoundsByEpoch(epoch uint32) uint32 {
	if e.GetGracePeriodRoundsByEpochCalled != nil {
		return e.GetGracePeriodRoundsByEpochCalled(epoch)
	}

	return 0
}

// GetExtraDelayForRequestBlockInfoInMs -
func (e *CommonConfigsHandlerStub) GetExtraDelayForRequestBlockInfoInMs(epoch uint32) uint32 {
	if e.GetExtraDelayForRequestBlockInfoInMsCalled != nil {
		return e.GetExtraDelayForRequestBlockInfoInMsCalled(epoch)
	}

	return 0
}

// GetMaxRoundsWithoutCommittedStartInEpochBlockInRound -
func (e *CommonConfigsHandlerStub) GetMaxRoundsWithoutCommittedStartInEpochBlockInRound(round uint64) uint32 {
	if e.GetMaxRoundsWithoutCommittedStartInEpochBlockInRoundCalled != nil {
		return e.GetMaxRoundsWithoutCommittedStartInEpochBlockInRoundCalled(round)
	}

	return 0
}

// IsInterfaceNil -
func (e *CommonConfigsHandlerStub) IsInterfaceNil() bool {
	return e == nil
}
