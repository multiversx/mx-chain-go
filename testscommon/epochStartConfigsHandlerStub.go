package testscommon

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/configs"
	"github.com/multiversx/mx-chain-go/config"
)

// GetDefaultEpochStartConfigsHandler -
func GetDefaultEpochStartConfigsHandler() common.EpochStartConfigsHandler {
	epochStartConfigsHandler, _ := configs.NewEpochStartConfigsHandler(
		[]config.EpochStartConfigByEpoch{
			{EnableEpoch: 0, GracePeriodRounds: 25, ExtraDelayForRequestBlockInfoInMilliseconds: 3000},
		},
	)

	return epochStartConfigsHandler
}

// EpochStartConfigsHandlerStub -
type EpochStartConfigsHandlerStub struct {
	GetGracePeriodRoundsByEpochCalled          func(epoch uint32) uint32
	GetExtraDelayForRequestBlockInfoInMsCalled func(epoch uint32) uint32
}

// GetGracePeriodRoundsByEpoch -
func (e *EpochStartConfigsHandlerStub) GetGracePeriodRoundsByEpoch(epoch uint32) uint32 {
	if e.GetGracePeriodRoundsByEpochCalled != nil {
		return e.GetGracePeriodRoundsByEpochCalled(epoch)
	}

	return 0
}

// GetExtraDelayForRequestBlockInfoInMs -
func (e *EpochStartConfigsHandlerStub) GetExtraDelayForRequestBlockInfoInMs(epoch uint32) uint32 {
	if e.GetExtraDelayForRequestBlockInfoInMsCalled != nil {
		return e.GetExtraDelayForRequestBlockInfoInMsCalled(epoch)
	}

	return 0
}

// IsInterfaceNil -
func (e *EpochStartConfigsHandlerStub) IsInterfaceNil() bool {
	return e == nil
}
