package bls

import (
	"github.com/multiversx/mx-chain-go/consensus/spos"
)

type subRoundEndV2Creator struct {
}

func NewSubRoundEndV2Creator() *subRoundEndV2Creator {
	return &subRoundEndV2Creator{}
}

func (c *subRoundEndV2Creator) CreateAndAddSubRoundEnd(
	subroundEndRoundInstance *subroundEndRound,
	worker spos.WorkerHandler,
	consensusCore spos.ConsensusCoreHandler,
) error {
	subroundSignatureV2Instance, errV2 := NewSubroundEndRoundV2(subroundEndRoundInstance)
	if errV2 != nil {
		return errV2
	}

	worker.AddReceivedMessageCall(MtBlockHeaderFinalInfo, subroundSignatureV2Instance.receivedBlockHeaderFinalInfo)
	worker.AddReceivedMessageCall(MtInvalidSigners, subroundSignatureV2Instance.receivedInvalidSignersInfo)
	worker.AddReceivedHeaderHandler(subroundSignatureV2Instance.receivedHeader)
	consensusCore.Chronology().AddSubround(subroundSignatureV2Instance)

	return nil
}

func (c *subRoundEndV2Creator) IsInterfaceNil() bool {
	return c == nil
}
