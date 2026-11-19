package scrCommon

import (
	"errors"

	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// GetDataTrieLoadError surfaces storage failures even when a nested call handles the VM failure code.
func GetDataTrieLoadError(vm vmcommon.VMExecutionHandler, executionError error) error {
	if executionError != nil {
		if errors.Is(executionError, state.ErrDataTrieLoadFailed) {
			return executionError
		}
		// Input validation can return before the VM clears runtime errors from the previous execution.
		return nil
	}
	provider, ok := vm.(interface{ GetRuntimeErrors() error })
	if !ok {
		return nil
	}
	return findDataTrieLoadError(provider.GetRuntimeErrors())
}

func findDataTrieLoadError(err error) error {
	if !errors.Is(err, state.ErrDataTrieLoadFailed) {
		return nil
	}
	if loadError, ok := err.(*state.DataTrieLoadError); ok {
		return loadError
	}

	// Check the aggregate because Unwrap can discard later runtime failures.
	if aggregate, ok := err.(interface{ GetAllErrors() []error }); ok {
		for _, inner := range aggregate.GetAllErrors() {
			if loadError := findDataTrieLoadError(inner); loadError != nil {
				return loadError
			}
		}
	}
	return findDataTrieLoadError(errors.Unwrap(err))
}
