package enableEpochsHandlerMock

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
)

// EnableEpochsHandlerStub -
type EnableEpochsHandlerStub struct {
	sync.RWMutex
	activeFlags                map[core.EnableEpochFlag]struct{}
	GetCurrentEpochCalled      func() uint32
	IsFlagDefinedCalled        func(flag core.EnableEpochFlag) bool
	IsFlagEnabledCalled        func(flag core.EnableEpochFlag) bool
	IsFlagEnabledInEpochCalled func(flag core.EnableEpochFlag, epoch uint32) bool
	GetActivationEpochCalled   func(flag core.EnableEpochFlag) uint32
}

// NewEnableEpochsHandlerStubWithNoFlagsDefined -
func NewEnableEpochsHandlerStubWithNoFlagsDefined() *EnableEpochsHandlerStub {
	return &EnableEpochsHandlerStub{
		activeFlags: make(map[core.EnableEpochFlag]struct{}),
		IsFlagDefinedCalled: func(flag core.EnableEpochFlag) bool {
			return false
		},
	}
}

// NewEnableEpochsHandlerStub -
func NewEnableEpochsHandlerStub(flags ...core.EnableEpochFlag) *EnableEpochsHandlerStub {
	stub := &EnableEpochsHandlerStub{
		activeFlags: make(map[core.EnableEpochFlag]struct{}),
	}
	for _, flag := range flags {
		stub.activeFlags[flag] = struct{}{}
	}

	return stub
}

// AddActiveFlags -
func (stub *EnableEpochsHandlerStub) AddActiveFlags(flags ...core.EnableEpochFlag) {
	stub.Lock()
	defer stub.Unlock()

	for _, flag := range flags {
		stub.activeFlags[flag] = struct{}{}
	}
}

// RemoveActiveFlags -
func (stub *EnableEpochsHandlerStub) RemoveActiveFlags(flags ...core.EnableEpochFlag) {
	stub.Lock()
	defer stub.Unlock()

	for _, flag := range flags {
		delete(stub.activeFlags, flag)
	}
}

// GetActivationEpoch -
func (stub *EnableEpochsHandlerStub) GetActivationEpoch(flag core.EnableEpochFlag) uint32 {
	if stub.GetActivationEpochCalled != nil {
		return stub.GetActivationEpochCalled(flag)
	}
	return 0
}

// IsFlagDefined -
func (stub *EnableEpochsHandlerStub) IsFlagDefined(flag core.EnableEpochFlag) bool {
	if stub.IsFlagDefinedCalled != nil {
		return stub.IsFlagDefinedCalled(flag)
	}
	return true
}

// IsFlagEnabled -
func (stub *EnableEpochsHandlerStub) IsFlagEnabled(flag core.EnableEpochFlag) bool {
	if stub.IsFlagEnabledCalled != nil {
		return stub.IsFlagEnabledCalled(flag)
	}

	stub.RLock()
	defer stub.RUnlock()
	_, found := stub.activeFlags[flag]
	return found
}

// IsFlagEnabledInEpoch -
func (stub *EnableEpochsHandlerStub) IsFlagEnabledInEpoch(flag core.EnableEpochFlag, epoch uint32) bool {
	if stub.IsFlagEnabledInEpochCalled != nil {
		return stub.IsFlagEnabledInEpochCalled(flag, epoch)
	}
	return false
}

// GetCurrentEpoch -
func (stub *EnableEpochsHandlerStub) GetCurrentEpoch() uint32 {
	if stub.GetCurrentEpochCalled != nil {
		return stub.GetCurrentEpochCalled()
	}
	return 0
}

// IsConsensusModelV2Enabled -
func (stub *EnableEpochsHandlerStub) IsConsensusModelV2Enabled() bool {
	stub.RLock()
	defer stub.RUnlock()

	return stub.IsConsensusModelV2EnabledField
}

// IsInterfaceNil -
func (stub *EnableEpochsHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
