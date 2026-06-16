package v1

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/outport"
)

// factory defines the data needed by this factory to create all the subrounds and give them their specific
// functionality
type factory struct {
	consensusCore  spos.ConsensusCoreHandler
	consensusState spos.ConsensusStateHandler
	worker         spos.WorkerHandler

	appStatusHandler      core.AppStatusHandler
	outportHandler        outport.OutportHandler
	sentSignaturesTracker spos.SentSignaturesTracker
	chainID               []byte
	currentPid            core.PeerID
	commonConfigsHandler  common.CommonConfigsHandler
}

// NewSubroundsFactory creates a new consensusState object
func NewSubroundsFactory(
	consensusDataContainer spos.ConsensusCoreHandler,
	consensusState spos.ConsensusStateHandler,
	worker spos.WorkerHandler,
	chainID []byte,
	currentPid core.PeerID,
	appStatusHandler core.AppStatusHandler,
	sentSignaturesTracker spos.SentSignaturesTracker,
	outportHandler outport.OutportHandler,
	commonConfigsHandler common.CommonConfigsHandler,
) (*factory, error) {
	// no need to check the outportHandler, it can be nil
	err := checkNewFactoryParams(
		consensusDataContainer,
		consensusState,
		worker,
		chainID,
		appStatusHandler,
		sentSignaturesTracker,
		commonConfigsHandler,
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
		outportHandler:        outportHandler,
		commonConfigsHandler:  commonConfigsHandler,
	}

	return &fct, nil
}

func checkNewFactoryParams(
	container spos.ConsensusCoreHandler,
	state spos.ConsensusStateHandler,
	worker spos.WorkerHandler,
	chainID []byte,
	appStatusHandler core.AppStatusHandler,
	sentSignaturesTracker spos.SentSignaturesTracker,
	commonConfigsHandler common.CommonConfigsHandler,
) error {
	err := spos.ValidateConsensusCore(container)
	if err != nil {
		return err
	}
	if check.IfNil(state) {
		return spos.ErrNilConsensusState
	}
	if check.IfNil(worker) {
		return spos.ErrNilWorker
	}
	if check.IfNil(appStatusHandler) {
		return spos.ErrNilAppStatusHandler
	}
	if check.IfNil(sentSignaturesTracker) {
		return ErrNilSentSignatureTracker
	}
	if check.IfNil(commonConfigsHandler) {
		return common.ErrNilCommonConfigsHandler
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
func (fct *factory) GenerateSubrounds(epoch uint32) error {
	fct.initConsensusThreshold()
	fct.consensusCore.Chronology().RemoveAllSubrounds()
	fct.worker.RemoveAllReceivedMessagesCalls()
	fct.worker.RemoveAllReceivedHeaderHandlers()

	timing := fct.commonConfigsHandler.GetSubroundsTimingByEpoch(epoch)

	err := fct.generateStartRoundSubround(timing)
	if err != nil {
		return err
	}

	err = fct.generateBlockSubround(timing)
	if err != nil {
		return err
	}

	err = fct.generateSignatureSubround(timing)
	if err != nil {
		return err
	}

	err = fct.generateEndRoundSubround(timing)
	if err != nil {
		return err
	}

	return nil
}

func (fct *factory) getTimeDuration() time.Duration {
	return fct.consensusCore.RoundHandler().TimeDuration()
}

func (fct *factory) generateStartRoundSubround(timing config.SubroundsTimingConfig) error {
	subround, err := spos.NewSubround(
		-1,
		bls.SrStartRound,
		bls.SrBlock,
		fct.getTimeDuration(),
		timing.SubroundStartStartTime,
		timing.SubroundStartEndTime,
		bls.GetSubroundName(bls.SrStartRound),
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
		int(timing.ProcessingThresholdPercent),
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

func (fct *factory) generateBlockSubround(timing config.SubroundsTimingConfig) error {
	subround, err := spos.NewSubround(
		bls.SrStartRound,
		bls.SrBlock,
		bls.SrSignature,
		fct.getTimeDuration(),
		timing.SubroundBlockStartTime,
		timing.SubroundBlockEndTime,
		bls.GetSubroundName(bls.SrBlock),
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
		fct.worker.Extend,
		int(timing.ProcessingThresholdPercent),
	)
	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(bls.MtBlockBodyAndHeader, subroundBlockInstance.receivedBlockBodyAndHeader)
	fct.worker.AddReceivedMessageCall(bls.MtBlockBody, subroundBlockInstance.receivedBlockBody)
	fct.worker.AddReceivedMessageCall(bls.MtBlockHeader, subroundBlockInstance.receivedBlockHeader)
	fct.worker.AddReceivedHeaderHandler(subroundBlockInstance.receivedFullHeader)
	fct.consensusCore.Chronology().AddSubround(subroundBlockInstance)

	return nil
}

func (fct *factory) generateSignatureSubround(timing config.SubroundsTimingConfig) error {
	subround, err := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		fct.getTimeDuration(),
		timing.SubroundSignatureStartTime,
		timing.SubroundSignatureEndTime,
		bls.GetSubroundName(bls.SrSignature),
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
		fct.worker.Extend,
		fct.appStatusHandler,
		fct.sentSignaturesTracker,
	)
	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(bls.MtSignature, subroundSignatureObject.receivedSignature)
	fct.consensusCore.Chronology().AddSubround(subroundSignatureObject)

	return nil
}

func (fct *factory) generateEndRoundSubround(timing config.SubroundsTimingConfig) error {
	subround, err := spos.NewSubround(
		bls.SrSignature,
		bls.SrEndRound,
		-1,
		fct.getTimeDuration(),
		timing.SubroundEndStartTime,
		timing.SubroundEndEndTime,
		bls.GetSubroundName(bls.SrEndRound),
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

	fct.worker.AddReceivedMessageCall(bls.MtBlockHeaderFinalInfo, subroundEndRoundObject.receivedBlockHeaderFinalInfo)
	fct.worker.AddReceivedMessageCall(bls.MtInvalidSigners, subroundEndRoundObject.receivedInvalidSignersInfo)
	fct.worker.AddReceivedHeaderHandler(subroundEndRoundObject.receivedHeader)
	fct.consensusCore.Chronology().AddSubround(subroundEndRoundObject)

	return nil
}

func (fct *factory) initConsensusThreshold() {
	pBFTThreshold := core.GetPBFTThreshold(fct.consensusState.ConsensusGroupSize())
	pBFTFallbackThreshold := core.GetPBFTFallbackThreshold(fct.consensusState.ConsensusGroupSize())
	fct.consensusState.SetThreshold(bls.SrBlock, 1)
	fct.consensusState.SetThreshold(bls.SrSignature, pBFTThreshold)
	fct.consensusState.SetFallbackThreshold(bls.SrBlock, 1)
	fct.consensusState.SetFallbackThreshold(bls.SrSignature, pBFTFallbackThreshold)
}

// IsInterfaceNil returns true if there is no value under the interface
func (fct *factory) IsInterfaceNil() bool {
	return fct == nil
}
