package sposFactory

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bls"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
)

// GetSubroundsFactory returns a subrounds factory depending of the given parameter
func GetSubroundsFactory(
	consensusDataContainer spos.ConsensusCoreHandler,
	consensusState *spos.ConsensusState,
	worker spos.WorkerHandler,
	consensusType string,
) (spos.SubroundsFactory, error) {

	switch consensusType {
	case blsConsensusType:
		return bls.NewSubroundsFactory(consensusDataContainer, consensusState, worker)
	case bnConsensusType:
		return bn.NewSubroundsFactory(consensusDataContainer, consensusState, worker)
	}

	return nil, ErrInvalidConsensusType
}

// GetConsensusCoreFactory returns a consensus service depending of the given parameter
func GetConsensusCoreFactory(consensusType string) (spos.ConsensusService, error) {
	switch consensusType {
	case blsConsensusType:
		return bls.NewConsensusService()
	case bnConsensusType:
		return bn.NewConsensusService()
	}

	return nil, ErrInvalidConsensusType
}
