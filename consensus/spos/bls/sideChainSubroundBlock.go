package bls

import (
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"golang.org/x/net/context"
	"time"
)

type sideChainSubroundBlock struct {
	*subroundBlock
}

// NewSideChainSubroundBlock creates a sideChainSubroundBlock object
func NewSideChainSubroundBlock(subroundBlock *subroundBlock) (*sideChainSubroundBlock, error) {
	if subroundBlock == nil {
		return nil, spos.ErrNilSubround
	}

	scsr := &sideChainSubroundBlock{
		subroundBlock,
	}

	scsr.Job = scsr.doBlockJob

	return scsr, nil
}

// doBlockJob method does the job of the subround Block
func (scsr *sideChainSubroundBlock) doBlockJob(ctx context.Context) bool {
	if !scsr.IsSelfLeaderInCurrentRound() { // is NOT self leader in this round?
		return false
	}

	if scsr.RoundHandler().Index() <= scsr.getRoundInLastCommittedBlock() {
		return false
	}

	if scsr.IsSelfJobDone(scsr.Current()) {
		return false
	}

	if scsr.IsSubroundFinished(scsr.Current()) {
		return false
	}

	metricStatTime := time.Now()
	defer scsr.computeSubroundProcessingMetric(metricStatTime, common.MetricCreatedProposedBlock)

	header, err := scsr.createHeader()
	if err != nil {
		printLogMessage(ctx, "doBlockJob.createHeader", err)
		return false
	}

	header, body, err := scsr.createBlock(header)
	if err != nil {
		printLogMessage(ctx, "doBlockJob.createBlock", err)
		return false
	}

	sentWithSuccess := scsr.sendBlock(header, body)
	if !sentWithSuccess {
		return false
	}

	cnsDta := &consensus.Message{
		PubKey:     []byte(scsr.SelfPubKey()),
		RoundIndex: scsr.RoundHandler().Index(),
	}

	return scsr.processReceivedBlock(ctx, cnsDta)
}
