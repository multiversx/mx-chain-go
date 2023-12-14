package mock

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
)

// EnableEpochsHandlerMock -
type EnableEpochsHandlerMock struct {
	WaitingListFixEnableEpochField            uint32
	RefactorPeersMiniBlocksEnableEpochField   uint32
	IsRefactorPeersMiniBlocksFlagEnabledField bool
	CurrentEpoch                              uint32
}

// GetActivationEpoch -
func (mock *EnableEpochsHandlerMock) GetActivationEpoch(flag core.EnableEpochFlag) uint32 {
	switch flag {
	case common.RefactorPeersMiniBlocksFlag:
		return mock.RefactorPeersMiniBlocksEnableEpochField
	case common.WaitingListFixFlag:
		return mock.WaitingListFixEnableEpochField

	default:
		return 0
	}
}

// IsFlagDefined returns true
func (mock *EnableEpochsHandlerMock) IsFlagDefined(_ core.EnableEpochFlag) bool {
	return true
}

// IsFlagEnabled returns true
func (mock *EnableEpochsHandlerMock) IsFlagEnabled(_ core.EnableEpochFlag) bool {
	return true
}

// IsFlagEnabledInEpoch returns true
func (mock *EnableEpochsHandlerMock) IsFlagEnabledInEpoch(_ core.EnableEpochFlag, _ uint32) bool {
	return true
}

// GetCurrentEpoch -
func (mock *EnableEpochsHandlerMock) GetCurrentEpoch() uint32 {
	return mock.CurrentEpoch
}

// FixGasRemainingForSaveKeyValueBuiltinFunctionEnabled -
func (mock *EnableEpochsHandlerMock) FixGasRemainingForSaveKeyValueBuiltinFunctionEnabled() bool {
	return false
}

// IsEquivalentMessagesFlagEnabled -
func (mock *EnableEpochsHandlerMock) IsEquivalentMessagesFlagEnabled() bool {
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (mock *EnableEpochsHandlerMock) IsInterfaceNil() bool {
	return mock == nil
}
