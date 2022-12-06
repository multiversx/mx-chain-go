package bls

import (
	"context"
	"time"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
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
	if !sr.IsSelfLeaderInCurrentRound() { // is NOT self leader in this round?
		return false
	}

	if sr.RoundHandler().Index() <= sr.getRoundInLastCommittedBlock() {
		return false
	}

	if sr.IsSelfJobDone(sr.Current()) {
		return false
	}

	if sr.IsSubroundFinished(sr.Current()) {
		return false
	}

	metricStatTime := time.Now()
	defer sr.computeSubroundProcessingMetric(metricStatTime, common.MetricCreatedProposedBlock)

	header, err := sr.createHeader()
	if err != nil {
		printLogMessage(ctx, "doBlockJob.createHeader", err)
		return false
	}

	header, body, err := sr.createBlock(header)
	if err != nil {
		printLogMessage(ctx, "doBlockJob.createBlock", err)
		return false
	}

	sentWithSuccess := sr.sendBlock(header, body)
	if !sentWithSuccess {
		return false
	}

	cnsDta := &consensus.Message{
		PubKey:     []byte(sr.SelfPubKey()),
		RoundIndex: sr.RoundHandler().Index(),
	}

	return sr.processReceivedBlock(ctx, cnsDta)
}
