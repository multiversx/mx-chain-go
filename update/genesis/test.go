package genesis

import "github.com/ElrondNetwork/elrond-go/data/state"

// ExportNodesSetupJson -
func (se *stateExport) ExportNodesSetupJson(validators map[uint32][]*state.ValidatorInfo) error {
	return se.exportNodesSetupJson(validators)
}
