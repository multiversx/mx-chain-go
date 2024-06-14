package bls

import (
	"context"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
)

type subroundBlockV2 struct {
	*subroundBlock
}

// NewSubroundBlockV2 creates a subroundBlockV2 object
func NewSubroundBlockV2(subroundBlock *subroundBlock) (*subroundBlockV2, error) {
	if subroundBlock == nil {
		return nil, spos.ErrNilSubround
	}

	sr := &subroundBlockV2{
		subroundBlock,
	}

	sr.Job = sr.doBlockJob

	return sr, nil
}

// doBlockJob method does the job of the subround Block
func (sr *subroundBlockV2) doBlockJob(ctx context.Context) bool {
	args, deferFunc := sr.doBlockComputation()
	defer deferFunc()

	if args == nil {
		return false
	}

	cnsDta := &consensus.Message{
		PubKey:     []byte(args.leader),
		RoundIndex: sr.RoundHandler().Index(),
	}

	return sr.processReceivedBlock(ctx, cnsDta)
}
