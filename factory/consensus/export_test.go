package consensus

import "github.com/ElrondNetwork/elrond-go/consensus/spos"

// GetSubroundsFactory -
func (ccf *consensusComponentsFactoryV2) GetSubroundsFactory(
	consensusDataContainer *spos.ConsensusCore,
	consensusState *spos.ConsensusState,
	cc *consensusComponents,
) (spos.SubroundsFactory, error) {
	return ccf.getSubroundsFactory(consensusDataContainer, consensusState, cc)
}
