package sposFactory

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/bls"
	"github.com/ElrondNetwork/elrond-go/outport"
)

// GetSubroundsFactoryV2 returns a subrounds factory depending of the given parameter
func GetSubroundsFactoryV2(
	consensusDataContainer spos.ConsensusCoreHandler,
	consensusState *spos.ConsensusState,
	worker spos.WorkerHandler,
	consensusType string,
	appStatusHandler core.AppStatusHandler,
	outportHandler outport.OutportHandler,
	chainID []byte,
	currentPid core.PeerID,
) (spos.SubroundsFactory, error) {
	switch consensusType {
	case blsConsensusType:
		subRoundFactoryBls, err := bls.NewSubroundsFactory(
			consensusDataContainer,
			consensusState,
			worker,
			chainID,
			currentPid,
			appStatusHandler,
		)
		if err != nil {
			return nil, err
		}

		subRoundFactoryBlsV2, err := bls.NewSubroundsFactoryV2(subRoundFactoryBls)
		if err != nil {
			return nil, err
		}

		subRoundFactoryBlsV2.SetOutportHandler(outportHandler)

		return subRoundFactoryBlsV2, nil
	default:
		return nil, ErrInvalidConsensusType
	}
}
