package bls

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/block"
)

type sovereignSubRoundEndV2Creator struct {
	outGoingOperationsPool block.OutGoingOperationsPool
	bridgeOpHandler        BridgeOperationsHandler
}

func NewSovereignSubRoundEndV2Creator(
	outGoingOperationsPool block.OutGoingOperationsPool,
	bridgeOpHandler BridgeOperationsHandler,
) (*sovereignSubRoundEndV2Creator, error) {
	if check.IfNil(outGoingOperationsPool) {
		return nil, errors.ErrNilOutGoingOperationsPool
	}

	return &sovereignSubRoundEndV2Creator{
		outGoingOperationsPool: outGoingOperationsPool,
		bridgeOpHandler:        bridgeOpHandler,
	}, nil
}

func (c *sovereignSubRoundEndV2Creator) CreateAndAddSubRoundEnd(
	subroundEndRoundInstance *subroundEndRound,
	worker spos.WorkerHandler,
	consensusCore spos.ConsensusCoreHandler,
) error {
	subroundSignatureV2Instance, err := NewSubroundEndRoundV2(subroundEndRoundInstance)
	if err != nil {
		return err
	}
	sovEndRound, err := NewSovereignSubRoundEndRound(
		subroundSignatureV2Instance,
		c.outGoingOperationsPool,
		c.bridgeOpHandler,
	)
	if err != nil {
		return err
	}

	worker.AddReceivedMessageCall(MtBlockHeaderFinalInfo, sovEndRound.receivedBlockHeaderFinalInfo)
	worker.AddReceivedMessageCall(MtInvalidSigners, sovEndRound.receivedInvalidSignersInfo)
	worker.AddReceivedHeaderHandler(sovEndRound.receivedHeader)
	consensusCore.Chronology().AddSubround(sovEndRound)

	return nil
}

func (c *sovereignSubRoundEndV2Creator) IsInterfaceNil() bool {
	return c == nil
}
