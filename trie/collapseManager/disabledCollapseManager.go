package collapseManager

import (
	"github.com/multiversx/mx-chain-go/common"
)

type disabledCollapseManager struct{}

// NewDisabledCollapseManager creates a new disabled collapse manager
func NewDisabledCollapseManager() *disabledCollapseManager {
	return &disabledCollapseManager{}
}

func (d *disabledCollapseManager) MarkKeyAsAccessed(key []byte, sizeLoadedInMemory int) {

}

func (d *disabledCollapseManager) RemoveKey(key []byte, sizeLoadedInMemory int) {

}

func (d *disabledCollapseManager) ShouldCollapseTrie() bool {
	return false
}

func (d *disabledCollapseManager) GetCollapsibleLeaves() ([][]byte, error) {
	return nil, nil
}

func (d *disabledCollapseManager) AddSizeInMemory(size int) {

}

func (d *disabledCollapseManager) GetSizeInMemory() int {
	return 0
}

func (d *disabledCollapseManager) CloneWithoutState() common.TrieCollapseManager {
	return NewDisabledCollapseManager()
}

func (d *disabledCollapseManager) IsCollapseEnabled() bool {
	return false
}

func (d *disabledCollapseManager) IsInterfaceNil() bool {
	return d == nil
}
