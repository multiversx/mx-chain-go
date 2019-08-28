package bls

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/commonSubround"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
)

// factory defines the data needed by this factory to create all the subrounds and give them their specific
// functionality
type factory struct {
	consensusCore  spos.ConsensusCoreHandler
	consensusState *spos.ConsensusState
	worker         spos.WorkerHandler

	appStatusHandler core.AppStatusHandler
}

// NewSubroundsFactory creates a new consensusState object
func NewSubroundsFactory(
	consensusDataContainer spos.ConsensusCoreHandler,
	consensusState *spos.ConsensusState,
	worker spos.WorkerHandler,
) (*factory, error) {
	err := checkNewFactoryParams(
		consensusDataContainer,
		consensusState,
		worker,
	)
	if err != nil {
		return nil, err
	}

	fct := factory{
		consensusCore:    consensusDataContainer,
		consensusState:   consensusState,
		worker:           worker,
		appStatusHandler: statusHandler.NewNilStatusHandler(),
	}

	return &fct, nil
}

func checkNewFactoryParams(
	container spos.ConsensusCoreHandler,
	state *spos.ConsensusState,
	worker spos.WorkerHandler,
) error {
	err := spos.ValidateConsensusCore(container)
	if err != nil {
		return err
	}
	if state == nil {
		return spos.ErrNilConsensusState
	}
	if worker == nil || worker.IsInterfaceNil() {
		return spos.ErrNilWorker
	}

	return nil
}

// SetAppStatusHandler method set appStatusHandler
func (fct *factory) SetAppStatusHandler(ash core.AppStatusHandler) error {
	if ash == nil || ash.IsInterfaceNil() {
		return spos.ErrNilAppStatusHandler
	}

	fct.appStatusHandler = ash
	return nil
}

// GenerateSubrounds will generate the subrounds used in BLS Cns
func (fct *factory) GenerateSubrounds() error {
	fct.initConsensusThreshold()
	fct.consensusCore.Chronology().RemoveAllSubrounds()
	fct.worker.RemoveAllReceivedMessagesCalls()

	err := fct.generateStartRoundSubround()
	if err != nil {
		return err
	}

	err = fct.generateBlockSubround()
	if err != nil {
		return err
	}

	err = fct.generateSignatureSubround()
	if err != nil {
		return err
	}

	err = fct.generateEndRoundSubround()
	if err != nil {
		return err
	}

	return nil
}

func (fct *factory) getTimeDuration() time.Duration {
	return fct.consensusCore.Rounder().TimeDuration()
}

func (fct *factory) generateStartRoundSubround() error {
	subround, err := spos.NewSubround(
		-1,
		SrStartRound,
		SrBlock,
		int64(float64(fct.getTimeDuration())*srStartStartTime),
		int64(float64(fct.getTimeDuration())*srStartEndTime),
		getSubroundName(SrStartRound),
		fct.consensusState,
		fct.worker.GetConsensusStateChangedChannel(),
		fct.worker.ExecuteStoredMessages,
		fct.consensusCore,
	)
	if err != nil {
		return err
	}

	subroundStartRound, err := commonSubround.NewSubroundStartRound(
		subround,
		fct.worker.Extend,
		processingThresholdPercent,
		getSubroundName,
		fct.worker.ExecuteStoredMessages,
		fct.worker.BroadcastUnnotarisedBlocks,
	)
	if err != nil {
		return err
	}

	err = subroundStartRound.SetAppStatusHandler(fct.appStatusHandler)
	if err != nil {
		return err
	}

	fct.consensusCore.Chronology().AddSubround(subroundStartRound)

	return nil
}

func (fct *factory) generateBlockSubround() error {
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
	)
	if err != nil {
		return err
	}

	subroundBlock, err := commonSubround.NewSubroundBlock(
		subround,
		fct.worker.Extend,
		int(MtBlockBody),
		int(MtBlockHeader),
		processingThresholdPercent,
		getSubroundName,
	)
	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtBlockBody, subroundBlock.ReceivedBlockBody)
	fct.worker.AddReceivedMessageCall(MtBlockHeader, subroundBlock.ReceivedBlockHeader)
	fct.consensusCore.Chronology().AddSubround(subroundBlock)

	return nil
}

func (fct *factory) generateSignatureSubround() error {
	subround, err := spos.NewSubround(
		SrBlock,
		SrSignature,
		SrEndRound,
		int64(float64(fct.getTimeDuration())*srSignatureStartTime),
		int64(float64(fct.getTimeDuration())*srSignatureEndTime),
		getSubroundName(SrSignature),
		fct.consensusState,
		fct.worker.GetConsensusStateChangedChannel(),
		fct.worker.ExecuteStoredMessages,
		fct.consensusCore,
	)
	if err != nil {
		return err
	}

	subroundSignature, err := NewSubroundSignature(
		subround,
		fct.worker.Extend,
	)
	if err != nil {
		return err
	}

	err = subroundSignature.SetAppStatusHandler(fct.appStatusHandler)
	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtSignature, subroundSignature.receivedSignature)
	fct.consensusCore.Chronology().AddSubround(subroundSignature)

	return nil
}

func (fct *factory) generateEndRoundSubround() error {
	subround, err := spos.NewSubround(
		SrSignature,
		SrEndRound,
		-1,
		int64(float64(fct.getTimeDuration())*srEndStartTime),
		int64(float64(fct.getTimeDuration())*srEndEndTime),
		getSubroundName(SrEndRound),
		fct.consensusState,
		fct.worker.GetConsensusStateChangedChannel(),
		fct.worker.ExecuteStoredMessages,
		fct.consensusCore,
	)
	if err != nil {
		return err
	}

	subroundEndRound, err := NewSubroundEndRound(
		subround,
		fct.worker.Extend,
	)
	if err != nil {
		return err
	}

	err = subroundEndRound.SetAppStatusHandler(fct.appStatusHandler)
	if err != nil {
		return err
	}

	fct.consensusCore.Chronology().AddSubround(subroundEndRound)

	return nil
}

func (fct *factory) initConsensusThreshold() {
	pbftThreshold := fct.consensusState.ConsensusGroupSize()*2/3 + 1
	fct.consensusState.SetThreshold(SrBlock, 1)
	fct.consensusState.SetThreshold(SrSignature, pbftThreshold)
}

// IsInterfaceNil returns true if there is no value under the interface
func (fct *factory) IsInterfaceNil() bool {
	if fct == nil {
		return true
	}
	return false
}
