package proxy

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	v1 "github.com/multiversx/mx-chain-go/consensus/spos/bls/v1"
	v2 "github.com/multiversx/mx-chain-go/consensus/spos/bls/v2"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/outport"
)

var log = logger.GetOrCreate("consensus/spos/bls/proxy")

// SubroundsHandlerArgs struct contains the needed data for the SubroundsHandler
type SubroundsHandlerArgs struct {
	Chronology           consensus.ChronologyHandler
	ConsensusCoreHandler spos.ConsensusCoreHandler
	ConsensusState       spos.ConsensusStateHandler
	Worker               factory.ConsensusWorker
	SignatureThrottler   core.Throttler
	AppStatusHandler     core.AppStatusHandler
	OutportHandler       outport.OutportHandler
	SentSignatureTracker spos.SentSignaturesTracker
	EnableEpochsHandler  core.EnableEpochsHandler
	CommonConfigsHandler common.CommonConfigsHandler
	ChainID              []byte
	CurrentPid           core.PeerID
}

// subroundsFactory defines the methods needed to generate the subrounds
type subroundsFactory interface {
	GenerateSubrounds(epoch uint32) error
	SetOutportHandler(driver outport.OutportHandler)
	IsInterfaceNil() bool
}

type consensusStateMachineType int

// SubroundsHandler struct contains the needed data for the SubroundsHandler
type SubroundsHandler struct {
	chronology           consensus.ChronologyHandler
	consensusCoreHandler spos.ConsensusCoreHandler
	consensusState       spos.ConsensusStateHandler
	worker               factory.ConsensusWorker
	signatureThrottler   core.Throttler
	appStatusHandler     core.AppStatusHandler
	outportHandler       outport.OutportHandler
	sentSignatureTracker spos.SentSignaturesTracker
	enableEpochsHandler  core.EnableEpochsHandler
	commonConfigsHandler common.CommonConfigsHandler
	chainID              []byte
	currentPid           core.PeerID
	currentConsensusType consensusStateMachineType
}

// EpochConfirmed is called when the epoch is confirmed (this is registered as callback)
func (s *SubroundsHandler) EpochConfirmed(epoch uint32, _ uint64) {
	err := s.initSubroundsForEpoch(epoch)
	if err != nil {
		log.Error("SubroundsHandler.EpochConfirmed: cannot initialize subrounds", "error", err)
	}
}

const (
	consensusNone consensusStateMachineType = iota
	consensusV1
	consensusV2
)

// NewSubroundsHandler creates a new SubroundsHandler object
func NewSubroundsHandler(args *SubroundsHandlerArgs) (*SubroundsHandler, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	subroundHandler := &SubroundsHandler{
		chronology:           args.Chronology,
		consensusCoreHandler: args.ConsensusCoreHandler,
		consensusState:       args.ConsensusState,
		worker:               args.Worker,
		signatureThrottler:   args.SignatureThrottler,
		appStatusHandler:     args.AppStatusHandler,
		outportHandler:       args.OutportHandler,
		sentSignatureTracker: args.SentSignatureTracker,
		enableEpochsHandler:  args.EnableEpochsHandler,
		commonConfigsHandler: args.CommonConfigsHandler,
		chainID:              args.ChainID,
		currentPid:           args.CurrentPid,
		currentConsensusType: consensusNone,
	}

	subroundHandler.consensusCoreHandler.EpochNotifier().RegisterNotifyHandler(subroundHandler)

	return subroundHandler, nil
}

func checkArgs(args *SubroundsHandlerArgs) error {
	if check.IfNil(args.Chronology) {
		return ErrNilChronologyHandler
	}
	if check.IfNil(args.ConsensusCoreHandler) {
		return ErrNilConsensusCoreHandler
	}
	if check.IfNil(args.ConsensusState) {
		return ErrNilConsensusState
	}
	if check.IfNil(args.Worker) {
		return ErrNilWorker
	}
	if check.IfNil(args.SignatureThrottler) {
		return ErrNilSignatureThrottler
	}
	if check.IfNil(args.AppStatusHandler) {
		return ErrNilAppStatusHandler
	}
	if check.IfNil(args.OutportHandler) {
		return ErrNilOutportHandler
	}
	if check.IfNil(args.SentSignatureTracker) {
		return ErrNilSentSignatureTracker
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return ErrNilEnableEpochsHandler
	}
	if check.IfNil(args.CommonConfigsHandler) {
		return common.ErrNilCommonConfigsHandler
	}
	if args.ChainID == nil {
		return ErrNilChainID
	}
	if len(args.CurrentPid) == 0 {
		return ErrNilCurrentPid
	}
	// outport handler can be nil if not configured so no need to check it

	return nil
}

// Start starts the sub-rounds handler
func (s *SubroundsHandler) Start(epoch uint32) error {
	return s.initSubroundsForEpoch(epoch)
}

// initSubroundsForEpoch generates the subrounds for the target consensus type of the given epoch, if it
// differs from the currently active consensus type
func (s *SubroundsHandler) initSubroundsForEpoch(epoch uint32) error {
	targetConsensusType := s.getTargetConsensusType(epoch)

	if s.currentConsensusType == targetConsensusType {
		return nil
	}

	s.currentConsensusType = targetConsensusType
	return s.generateSubroundsForCurrentType(epoch)
}

func (s *SubroundsHandler) getTargetConsensusType(epoch uint32) consensusStateMachineType {
	if s.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, epoch) {
		return consensusV2
	}

	return consensusV1
}

// generateSubroundsForCurrentType generates the subrounds matching the currently set consensus type
func (s *SubroundsHandler) generateSubroundsForCurrentType(epoch uint32) error {
	if s.currentConsensusType == consensusNone {
		return nil
	}

	var fct subroundsFactory
	var err error

	if s.currentConsensusType == consensusV2 {
		fct, err = v2.NewSubroundsFactory(
			s.consensusCoreHandler,
			s.consensusState,
			s.worker,
			s.chainID,
			s.currentPid,
			s.appStatusHandler,
			s.sentSignatureTracker,
			s.signatureThrottler,
			s.outportHandler,
			s.commonConfigsHandler,
		)
	} else {
		fct, err = v1.NewSubroundsFactory(
			s.consensusCoreHandler,
			s.consensusState,
			s.worker,
			s.chainID,
			s.currentPid,
			s.appStatusHandler,
			s.sentSignatureTracker,
			s.outportHandler,
			s.commonConfigsHandler,
		)
	}
	if err != nil {
		return err
	}

	err = s.chronology.Close()
	if err != nil {
		log.Warn("SubroundsHandler.generateSubroundsForCurrentType: cannot close the chronology", "error", err)
	}

	err = fct.GenerateSubrounds(epoch)
	if err != nil {
		return err
	}

	log.Debug("SubroundsHandler.generateSubroundsForCurrentType: reset consensus round state")
	s.worker.ResetConsensusRoundState()
	s.chronology.StartRounds()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *SubroundsHandler) IsInterfaceNil() bool {
	return s == nil
}
