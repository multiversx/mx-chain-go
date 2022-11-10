package bls

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/outport"
)

// factory defines the data needed by this factory to create all the subrounds and give them their specific
// functionality
type factory struct {
	consensusCore  spos.ConsensusCoreHandler
	consensusState *spos.ConsensusState
	worker         spos.WorkerHandler

	appStatusHandler  core.AppStatusHandler
	outportHandler    outport.OutportHandler
	chainID           []byte
	currentPid        core.PeerID
	subroundBlockType consensus.SubroundBlockType
}

// NewSubroundsFactory creates a new factory object
func NewSubroundsFactory(
	consensusDataContainer spos.ConsensusCoreHandler,
	consensusState *spos.ConsensusState,
	worker spos.WorkerHandler,
	chainID []byte,
	currentPid core.PeerID,
	appStatusHandler core.AppStatusHandler,
	subroundBlockType consensus.SubroundBlockType,
) (*factory, error) {
	err := checkNewFactoryParams(
		consensusDataContainer,
		consensusState,
		worker,
		chainID,
		appStatusHandler,
	)
	if err != nil {
		return nil, err
	}

	fct := factory{
		consensusCore:     consensusDataContainer,
		consensusState:    consensusState,
		worker:            worker,
		appStatusHandler:  appStatusHandler,
		chainID:           chainID,
		currentPid:        currentPid,
		subroundBlockType: subroundBlockType,
	}

	return &fct, nil
}

func checkNewFactoryParams(
	container spos.ConsensusCoreHandler,
	state *spos.ConsensusState,
	worker spos.WorkerHandler,
	chainID []byte,
	appStatusHandler core.AppStatusHandler,
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
	if len(chainID) == 0 {
		return spos.ErrInvalidChainID
	}

	return nil
}

// SetOutportHandler method will update the value of the factory's outport
func (fct *factory) SetOutportHandler(driver outport.OutportHandler) {
	fct.outportHandler = driver
}

// GenerateSubrounds will generate the subrounds used in BLS consensus
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

	subroundStartRound, err := NewSubroundStartRound(
		subround,
		fct.worker.Extend,
		processingThresholdPercent,
		fct.worker.ExecuteStoredMessages,
		fct.worker.ResetConsensusMessages,
	)
	if err != nil {
		return err
	}

	err = subroundStartRound.SetOutportHandler(fct.outportHandler)
	if err != nil {
		return err
	}

	fct.consensusCore.Chronology().AddSubround(subroundStartRound)

	return nil
}

func (fct *factory) generateBlockSubround() error {
	subroundBlock, err := fct.generateBlockSubroundV1()
	if err != nil {
		return err
	}

	switch fct.subroundBlockType {
	case consensus.SubroundBlockTypeV1:
		fct.worker.AddReceivedMessageCall(MtBlockBodyAndHeader, subroundBlock.receivedBlockBodyAndHeader)
		fct.worker.AddReceivedMessageCall(MtBlockBody, subroundBlock.receivedBlockBody)
		fct.worker.AddReceivedMessageCall(MtBlockHeader, subroundBlock.receivedBlockHeader)
		fct.consensusCore.Chronology().AddSubround(subroundBlock)

		return nil
	case consensus.SubroundBlockTypeV2:
		subroundBlockV2, errV2 := NewSubroundBlockV2(subroundBlock)
		if errV2 != nil {
			return errV2
		}

		fct.worker.AddReceivedMessageCall(MtBlockBodyAndHeader, subroundBlockV2.receivedBlockBodyAndHeader)
		fct.worker.AddReceivedMessageCall(MtBlockBody, subroundBlockV2.receivedBlockBody)
		fct.worker.AddReceivedMessageCall(MtBlockHeader, subroundBlockV2.receivedBlockHeader)
		fct.consensusCore.Chronology().AddSubround(subroundBlockV2)

		return nil
	default:
		return fmt.Errorf("%w type %v", ErrUnimplementedSubroundType, fct.subroundBlockType)
	}
}

func (fct *factory) generateBlockSubroundV1() (*subroundBlock, error) {
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
		return nil, err
	}

	subroundBlock, err := NewSubroundBlock(
		subround,
		fct.worker.Extend,
		processingThresholdPercent,
	)
	if err != nil {
		return nil, err
	}

	return subroundBlock, nil
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
		fct.worker.Extend,
		fct.appStatusHandler,
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
	)
	if err != nil {
		return err
	}

	fct.worker.AddReceivedMessageCall(MtBlockHeaderFinalInfo, subroundEndRoundObject.receivedBlockHeaderFinalInfo)
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
