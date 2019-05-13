package bn

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/commonSubround"
)

// factory defines the data needed by this factory to create all the subrounds and give them their specific
// functionality
type factory struct {
	consensusCore  spos.ConsensusCoreHandler
	consensusState *spos.ConsensusState
	worker         spos.WorkerHandler
}

// NewSubroundsFactory creates a new factory for BN subrounds
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
		consensusCore:  consensusDataContainer,
		consensusState: consensusState,
		worker:         worker,
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

	if worker == nil {
		return spos.ErrNilWorker
	}

	return nil
}

// GenerateSubrounds will generate the subrounds used in Belare & Naveen Cns
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

	err = fct.generateCommitmentHashSubround()
	if err != nil {
		return err
	}

	err = fct.generateBitmapSubround()
	if err != nil {
		return err
	}

	err = fct.generateCommitmentSubround()
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
	)

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
		SrCommitmentHash,
		int64(float64(fct.getTimeDuration())*srBlockStartTime),
		int64(float64(fct.getTimeDuration())*srBlockEndTime),
		getSubroundName(SrBlock),
		fct.consensusState,
		fct.worker.GetConsensusStateChangedChannel(),
		fct.consensusCore,
	)
	if err != nil {
		return err
	}

	subroundBlock, err := commonSubround.NewSubroundBlock(
		subround,
		fct.worker.SendConsensusMessage,
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

func (fct *factory) generateCommitmentHashSubround() error {
	subround, err := spos.NewSubround(
		SrBlock,
		SrCommitmentHash,
		SrBitmap,
		int64(float64(fct.getTimeDuration())*srCommitmentHashStartTime),
		int64(float64(fct.getTimeDuration())*srCommitmentHashEndTime),
		getSubroundName(SrCommitmentHash),
		fct.consensusState,
		fct.worker.GetConsensusStateChangedChannel(),
		fct.consensusCore,
	)
	if err != nil {
		return err
	}

	subroundCommitmentHash, err := NewSubroundCommitmentHash(
		subround,
		fct.worker.SendConsensusMessage,
		fct.worker.Extend,
	)
	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtCommitmentHash, subroundCommitmentHash.receivedCommitmentHash)
	fct.consensusCore.Chronology().AddSubround(subroundCommitmentHash)

	return nil
}

func (fct *factory) generateBitmapSubround() error {
	subround, err := spos.NewSubround(
		SrCommitmentHash,
		SrBitmap,
		SrCommitment,
		int64(float64(fct.getTimeDuration())*srBitmapStartTime),
		int64(float64(fct.getTimeDuration())*srBitmapEndTime),
		getSubroundName(SrBitmap),
		fct.consensusState,
		fct.worker.GetConsensusStateChangedChannel(),
		fct.consensusCore,
	)
	if err != nil {
		return err
	}

	subroundBitmap, err := NewSubroundBitmap(
		subround,
		fct.worker.SendConsensusMessage,
		fct.worker.Extend,
	)
	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtBitmap, subroundBitmap.receivedBitmap)
	fct.consensusCore.Chronology().AddSubround(subroundBitmap)

	return nil
}

func (fct *factory) generateCommitmentSubround() error {
	subround, err := spos.NewSubround(
		SrBitmap,
		SrCommitment,
		SrSignature,
		int64(float64(fct.getTimeDuration())*srCommitmentStartTime),
		int64(float64(fct.getTimeDuration())*srCommitmentEndTime),
		getSubroundName(SrCommitment),
		fct.consensusState,
		fct.worker.GetConsensusStateChangedChannel(),
		fct.consensusCore,
	)
	if err != nil {
		return err
	}

	subroundCommitment, err := NewSubroundCommitment(
		subround,
		fct.worker.SendConsensusMessage,
		fct.worker.Extend,
	)
	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtCommitment, subroundCommitment.receivedCommitment)
	fct.consensusCore.Chronology().AddSubround(subroundCommitment)

	return nil
}

func (fct *factory) generateSignatureSubround() error {
	subround, err := spos.NewSubround(
		SrCommitment,
		SrSignature,
		SrEndRound,
		int64(float64(fct.getTimeDuration())*srSignatureStartTime),
		int64(float64(fct.getTimeDuration())*srSignatureEndTime),
		getSubroundName(SrSignature),
		fct.consensusState,
		fct.worker.GetConsensusStateChangedChannel(),
		fct.consensusCore,
	)
	if err != nil {
		return err
	}

	subroundSignature, err := NewSubroundSignature(
		subround,
		fct.worker.SendConsensusMessage,
		fct.worker.Extend,
	)
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
		fct.consensusCore,
	)
	if err != nil {
		return err
	}

	subroundEndRound, err := NewSubroundEndRound(
		subround,
		fct.worker.BroadcastBlock,
		fct.worker.Extend,
	)
	if err != nil {
		return err
	}

	fct.consensusCore.Chronology().AddSubround(subroundEndRound)

	return nil
}

func (fct *factory) initConsensusThreshold() {
	pbftThreshold := fct.consensusState.ConsensusGroupSize()*2/3 + 1
	fct.consensusState.SetThreshold(SrBlock, 1)
	fct.consensusState.SetThreshold(SrCommitmentHash, pbftThreshold)
	fct.consensusState.SetThreshold(SrBitmap, pbftThreshold)
	fct.consensusState.SetThreshold(SrCommitment, pbftThreshold)
	fct.consensusState.SetThreshold(SrSignature, pbftThreshold)
}
