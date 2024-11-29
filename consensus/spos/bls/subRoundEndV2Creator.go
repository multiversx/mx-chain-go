package bls

import (
	"github.com/multiversx/mx-chain-go/consensus/spos"
)

type subRoundEndV2Creator struct {
}

// NewSubRoundEndV2Creator creates a new subround end v2 factory
func NewSubRoundEndV2Creator() *subRoundEndV2Creator {
	return &subRoundEndV2Creator{}
}

// CreateAndAddSubRoundEnd creates a new subround end v2 and adds it to the consensus
func (c *subRoundEndV2Creator) CreateAndAddSubRoundEnd(
	subroundEndRoundInstance *subroundEndRound,
	worker spos.WorkerHandler,
	consensusCore spos.ConsensusCoreHandler,
) error {
	subroundEndV2Instance, err := NewSubroundEndRoundV2(subroundEndRoundInstance)
	if err != nil {
		return err
	}

	worker.AddReceivedMessageCall(MtBlockHeaderFinalInfo, subroundEndV2Instance.receivedBlockHeaderFinalInfo)
	worker.AddReceivedMessageCall(MtInvalidSigners, subroundEndV2Instance.receivedInvalidSignersInfo)
	worker.AddReceivedHeaderHandler(subroundEndV2Instance.receivedHeader)
	consensusCore.Chronology().AddSubround(subroundEndV2Instance)

	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (c *subRoundEndV2Creator) IsInterfaceNil() bool {
	return c == nil
}
