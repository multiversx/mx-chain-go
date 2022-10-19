package bls

import "github.com/ElrondNetwork/elrond-go/consensus/spos"

// factoryV2 defines the data needed by this factory to create all the subrounds and give them their specific functionality
type factoryV2 struct {
	*factory
}

// NewSubroundsFactoryV2 creates a new factoryV2 object
func NewSubroundsFactoryV2(factory *factory) (*factoryV2, error) {
	if factory == nil {
		return nil, spos.ErrNilFactory
	}

	fct := &factoryV2{
		factory,
	}

	fct.generateBlockSubroundMethod = fct.generateBlockSubround

	return fct, nil
}

func (fct *factoryV2) generateBlockSubround() error {
	subround, err := spos.NewSubround(
		SrStartRound,
		SrBlock,
		SrSignature,
		int64(float64(fct.getTimeDuration())*srBlockStartTime),
		int64(float64(fct.getTimeDuration())*srBlockEndTime),
		getSubroundName(SrBlock),
		fct.consensusState,
		fct.worker.GetConsensusStateChangedChannel(),
		fct.worker.ExecuteStoredMessages,
		fct.consensusCore,
		fct.chainID,
		fct.currentPid,
		fct.appStatusHandler,
	)
	if err != nil {
		return err
	}

	subroundBlock, err := NewSubroundBlock(
		subround,
		fct.worker.Extend,
		processingThresholdPercent,
	)
	if err != nil {
		return err
	}

	subroundBlockV2, err := NewSubroundBlockV2(subroundBlock)
	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtBlockBodyAndHeader, subroundBlockV2.receivedBlockBodyAndHeader)
	fct.worker.AddReceivedMessageCall(MtBlockBody, subroundBlockV2.receivedBlockBody)
	fct.worker.AddReceivedMessageCall(MtBlockHeader, subroundBlockV2.receivedBlockHeader)
	fct.consensusCore.Chronology().AddSubround(subroundBlockV2)

	return nil
}
