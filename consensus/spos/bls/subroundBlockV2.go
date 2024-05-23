package bls

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-go/common"
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
	//TODO: check what can be reuese from the normal subrounds so that we do not have duplicate code
	isSelfLeader := sr.IsSelfLeaderInCurrentRound() && sr.ShouldConsiderSelfKeyInConsensus()
	if !isSelfLeader && !sr.IsMultiKeyLeaderInCurrentRound() { // is NOT self leader in this round?
		return false
	}

	if sr.RoundHandler().Index() <= sr.getRoundInLastCommittedBlock() {
		return false
	}

	if sr.IsLeaderJobDone(sr.Current()) {
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

	leader, errGetLeader := sr.GetLeader()
	if errGetLeader != nil {
		log.Debug("doBlockJob.GetLeader", "error", errGetLeader)
		return false
	}

	cnsDta := &consensus.Message{
		PubKey:     []byte(leader),
		RoundIndex: sr.RoundHandler().Index(),
	}

	return sr.processReceivedBlock(ctx, cnsDta)
}
