package disabled

import (
	"context"

	"github.com/multiversx/mx-chain-go/common"
)

type disabledDataTrieHandler struct {
}

// NewDisabledDataTrieHandler returns a new instance of disabledDataTrieHandler
func NewDisabledDataTrieHandler() *disabledDataTrieHandler {
	return &disabledDataTrieHandler{}
}

// RootHash returns an empty byte array
func (ddth *disabledDataTrieHandler) RootHash() ([]byte, error) {
	return []byte{}, nil
}

// GetAllLeavesOnChannel does nothing for this implementation
func (ddth *disabledDataTrieHandler) GetAllLeavesOnChannel(
	leavesChannels *common.TrieIteratorChannels,
	_ context.Context,
	_ []byte,
	_ common.KeyBuilder,
	_ common.TrieLeafParser,
) error {
	if leavesChannels.LeavesChan != nil {
		close(leavesChannels.LeavesChan)
	}
	if leavesChannels.ErrChan != nil {
		leavesChannels.ErrChan.Close()
	}

	return nil
}

// IsMigratedToLatestVersion returns true
func (ddth *disabledDataTrieHandler) IsMigratedToLatestVersion() (bool, error) {
	return true, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ddth *disabledDataTrieHandler) IsInterfaceNil() bool {
	return ddth == nil
}
