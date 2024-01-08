package bls

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/outport"
)

// factory defines the data needed by this factory to create all the subrounds and give them their specific
// functionality
type factory struct {
	consensusCore  spos.ConsensusCoreHandler
	consensusState *spos.ConsensusState
	worker         spos.WorkerHandler

	appStatusHandler      core.AppStatusHandler
	outportHandler        outport.OutportHandler
	sentSignaturesTracker spos.SentSignaturesTracker
	chainID               []byte
	currentPid            core.PeerID
}

// NewSubroundsFactory creates a new consensusState object
func NewSubroundsFactory(
	consensusDataContainer spos.ConsensusCoreHandler,
	consensusState *spos.ConsensusState,
	worker spos.WorkerHandler,
	chainID []byte,
	currentPid core.PeerID,
	appStatusHandler core.AppStatusHandler,
	sentSignaturesTracker spos.SentSignaturesTracker,
) (*factory, error) {
	err := checkNewFactoryParams(
		consensusDataContainer,
		consensusState,
		worker,
		chainID,
		appStatusHandler,
		sentSignaturesTracker,
	)
	if err != nil {
		return nil, err
	}

	fct := factory{
		consensusCore:         consensusDataContainer,
		consensusState:        consensusState,
		worker:                worker,
		appStatusHandler:      appStatusHandler,
		chainID:               chainID,
		currentPid:            currentPid,
		sentSignaturesTracker: sentSignaturesTracker,
	}

	return &fct, nil
}

func checkNewFactoryParams(
	container spos.ConsensusCoreHandler,
	state *spos.ConsensusState,
	worker spos.WorkerHandler,
	chainID []byte,
	appStatusHandler core.AppStatusHandler,
	sentSignaturesTracker spos.SentSignaturesTracker,
) error {
	err := spos.ValidateConsensusCore(container)
	if err != nil {
		return err
	}
	if state == nil {
		return spos.ErrNilConsensusState
	}
	if check.IfNil(worker) {
		return spos.ErrNilWorker
	}
	if check.IfNil(appStatusHandler) {
		return spos.ErrNilAppStatusHandler
	}
	if check.IfNil(sentSignaturesTracker) {
		return spos.ErrNilSentSignatureTracker
	}
	if len(chainID) == 0 {
		return spos.ErrInvalidChainID
	}

	return nil
}

// SetOutportHandler method will update the value of the factory's outport
func (fct *factory) SetOutportHandler(driver outport.OutportHandler) {
	fct.outportHandler = driver
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
	return fct.consensusCore.RoundHandler().TimeDuration()
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
		fct.chainID,
		fct.currentPid,
		fct.appStatusHandler,
	)
	if err != nil {
		return err
	}

	subroundStartRoundInstance, err := NewSubroundStartRound(
		subround,
		fct.worker.Extend,
		processingThresholdPercent,
		fct.worker.ExecuteStoredMessages,
		fct.worker.ResetConsensusMessages,
		fct.sentSignaturesTracker,
	)
	if err != nil {
		return err
	}

	err = subroundStartRoundInstance.SetOutportHandler(fct.outportHandler)
	if err != nil {
		return err
	}

	fct.consensusCore.Chronology().AddSubround(subroundStartRoundInstance)

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
		fct.chainID,
		fct.currentPid,
		fct.appStatusHandler,
	)
	if err != nil {
		return err
	}

	subroundBlockInstance, err := NewSubroundBlock(
		subround,
		processingThresholdPercent,
		fct.worker,
	)
	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtBlockBodyAndHeader, subroundBlockInstance.receivedBlockBodyAndHeader)
	fct.worker.AddReceivedMessageCall(MtBlockBody, subroundBlockInstance.receivedBlockBody)
	fct.worker.AddReceivedMessageCall(MtBlockHeader, subroundBlockInstance.receivedBlockHeader)
	fct.consensusCore.Chronology().AddSubround(subroundBlockInstance)

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
		fct.chainID,
		fct.currentPid,
		fct.appStatusHandler,
	)
	if err != nil {
		return err
	}

	subroundSignatureObject, err := NewSubroundSignature(
		subround,
		fct.appStatusHandler,
		fct.sentSignaturesTracker,
		fct.worker,
	)
	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtSignature, subroundSignatureObject.receivedSignature)
	fct.consensusCore.Chronology().AddSubround(subroundSignatureObject)

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
		fct.chainID,
		fct.currentPid,
		fct.appStatusHandler,
	)
	if err != nil {
		return err
	}

	subroundEndRoundObject, err := NewSubroundEndRound(
		subround,
		fct.worker.Extend,
		spos.MaxThresholdPercent,
		fct.worker.DisplayStatistics,
		fct.appStatusHandler,
		fct.sentSignaturesTracker,
	)
	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtBlockHeaderFinalInfo, subroundEndRoundObject.receivedBlockHeaderFinalInfo)
	fct.worker.AddReceivedMessageCall(MtInvalidSigners, subroundEndRoundObject.receivedInvalidSignersInfo)
	fct.worker.AddReceivedHeaderHandler(subroundEndRoundObject.receivedHeader)
	fct.consensusCore.Chronology().AddSubround(subroundEndRoundObject)

	return nil
}

func (fct *factory) initConsensusThreshold() {
	pBFTThreshold := core.GetPBFTThreshold(fct.consensusState.ConsensusGroupSize())
	pBFTFallbackThreshold := core.GetPBFTFallbackThreshold(fct.consensusState.ConsensusGroupSize())
	fct.consensusState.SetThreshold(SrBlock, 1)
	fct.consensusState.SetThreshold(SrSignature, pBFTThreshold)
	fct.consensusState.SetFallbackThreshold(SrBlock, 1)
	fct.consensusState.SetFallbackThreshold(SrSignature, pBFTFallbackThreshold)
}

// IsInterfaceNil returns true if there is no value under the interface
func (fct *factory) IsInterfaceNil() bool {
	return fct == nil
}
