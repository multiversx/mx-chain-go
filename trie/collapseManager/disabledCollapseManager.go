package collapseManager

import (
	"github.com/multiversx/mx-chain-go/common"
)

type disabledCollapseManager struct{}

// NewDisabledCollapseManager creates a new disabled collapse manager
func NewDisabledCollapseManager() *disabledCollapseManager {
	return &disabledCollapseManager{}
}

// MarkKeyAsAccessed does nothing for this implementation
func (d *disabledCollapseManager) MarkKeyAsAccessed(_ []byte, _ int) {
}

// RemoveKey does nothing for this implementation
func (d *disabledCollapseManager) RemoveKey(_ []byte, _ int) {
}

// ShouldCollapseTrie always returns false for this implementation
func (d *disabledCollapseManager) ShouldCollapseTrie() bool {
	return false
}

// GetCollapsibleLeaves always returns nil for this implementation
func (d *disabledCollapseManager) GetCollapsibleLeaves() ([][]byte, error) {
	return nil, nil
}

// AddSizeInMemory does nothing for this implementation
func (d *disabledCollapseManager) AddSizeInMemory(_ int) {
}

// GetSizeInMemory always returns 0 for this implementation
func (d *disabledCollapseManager) GetSizeInMemory() int {
	return 0
}

// CloneWithoutState returns a new disabled collapse manager
func (d *disabledCollapseManager) CloneWithoutState() common.TrieCollapseManager {
	return NewDisabledCollapseManager()
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledCollapseManager) IsInterfaceNil() bool {
	return d == nil
}
