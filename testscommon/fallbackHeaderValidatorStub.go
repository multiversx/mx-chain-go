package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// FallBackHeaderValidatorStub -
type FallBackHeaderValidatorStub struct {
	ShouldApplyFallbackValidationCalled              func(headerHandler data.HeaderHandler) bool
	ShouldApplyFallbackValidationForHeaderWithCalled func(shardID uint32, startOfEpochBlock bool, round uint64, prevHeaderHash []byte) bool
}

// ShouldApplyFallbackValidationForHeaderWith -
func (fhvs *FallBackHeaderValidatorStub) ShouldApplyFallbackValidationForHeaderWith(shardID uint32, startOfEpochBlock bool, round uint64, prevHeaderHash []byte) bool {
	if fhvs.ShouldApplyFallbackValidationForHeaderWithCalled != nil {
		return fhvs.ShouldApplyFallbackValidationForHeaderWithCalled(shardID, startOfEpochBlock, round, prevHeaderHash)
	}
	return false
}

// ShouldApplyFallbackValidation -
func (fhvs *FallBackHeaderValidatorStub) ShouldApplyFallbackValidation(headerHandler data.HeaderHandler) bool {
	if fhvs.ShouldApplyFallbackValidationCalled != nil {
		return fhvs.ShouldApplyFallbackValidationCalled(headerHandler)
	}
	return false
}

// IsInterfaceNil -
func (fhvs *FallBackHeaderValidatorStub) IsInterfaceNil() bool {
	return fhvs == nil
}
