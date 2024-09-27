package proxy

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
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
	ChainID              []byte
	CurrentPid           core.PeerID
}

// SubroundsFactory defines the methods needed to generate the subrounds
type SubroundsFactory interface {
	GenerateSubrounds() error
	SetOutportHandler(driver outport.OutportHandler)
	IsInterfaceNil() bool
}

type ConsensusStateMachineType int

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
	chainID              []byte
	currentPid           core.PeerID
	currentConsensusType ConsensusStateMachineType
}

const (
	ConsensusNone ConsensusStateMachineType = iota
	ConsensusV1
	ConsensusV2
)

func NewSubroundsHandler(args *SubroundsHandlerArgs) (*SubroundsHandler, error) {
	if check.IfNil(args.Chronology) {
		return nil, bls.ErrNilChronologyHandler
	}
	if check.IfNil(args.ConsensusCoreHandler) {
		return nil, bls.ErrNilConsensusCoreHandler
	}
	if check.IfNil(args.ConsensusState) {
		return nil, bls.ErrNilConsensusState
	}
	if check.IfNil(args.Worker) {
		return nil, bls.ErrNilWorker
	}
	if check.IfNil(args.SignatureThrottler) {
		return nil, bls.ErrNilSignatureThrottler
	}
	if check.IfNil(args.AppStatusHandler) {
		return nil, bls.ErrNilAppStatusHandler
	}
	if check.IfNil(args.OutportHandler) {
		return nil, bls.ErrNilOutportHandler
	}
	if check.IfNil(args.SentSignatureTracker) {
		return nil, bls.ErrNilSentSignatureTracker
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, bls.ErrNilEnableEpochsHandler
	}
	if args.ChainID == nil {
		return nil, bls.ErrNilChainID
	}
	if len(args.CurrentPid) == 0 {
		return nil, bls.ErrNilCurrentPid
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
		chainID:              args.ChainID,
		currentPid:           args.CurrentPid,
		currentConsensusType: ConsensusNone,
	}

	subroundHandler.consensusCoreHandler.EpochStartRegistrationHandler().RegisterHandler(subroundHandler)

	return subroundHandler, nil
}

// Start starts the sub-rounds handler
func (s *SubroundsHandler) Start(epoch uint32) error {
	return s.initSubroundsForEpoch(epoch)
}

func (s *SubroundsHandler) initSubroundsForEpoch(epoch uint32) error {
	var err error
	var fct SubroundsFactory
	if s.enableEpochsHandler.IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, epoch) {
		if s.currentConsensusType == ConsensusV2 {
			return nil
		}

		s.currentConsensusType = ConsensusV2
		fct, err = v2.NewSubroundsFactory(
			s.consensusCoreHandler,
			s.consensusState,
			s.worker,
			s.chainID,
			s.currentPid,
			s.appStatusHandler,
			s.sentSignatureTracker,
			s.signatureThrottler,
		)
	} else {
		if s.currentConsensusType == ConsensusV1 {
			return nil
		}

		s.currentConsensusType = ConsensusV1
		fct, err = v1.NewSubroundsFactory(
			s.consensusCoreHandler,
			s.consensusState,
			s.worker,
			s.chainID,
			s.currentPid,
			s.appStatusHandler,
			s.sentSignatureTracker,
		)
	}
	if err != nil {
		return err
	}

	err = s.chronology.Close()
	if err != nil {
		log.Warn("SubroundsHandler.initSubroundsForEpoch: cannot close the chronology", "error", err)
	}

	fct.SetOutportHandler(s.outportHandler)
	err = fct.GenerateSubrounds()
	if err != nil {
		return err
	}

	s.chronology.StartRounds()
	return nil
}

// EpochStartAction is called when the epoch starts
func (s *SubroundsHandler) EpochStartAction(hdr data.HeaderHandler) {
	err := s.initSubroundsForEpoch(hdr.GetEpoch())
	if err != nil {
		log.Error("SubroundsHandler.EpochStartAction: cannot initialize subrounds", "error", err)
	}
}

// EpochStartPrepare prepares the subrounds handler for the epoch start
func (s *SubroundsHandler) EpochStartPrepare(_ data.HeaderHandler, _ data.BodyHandler) {
}

// NotifyOrder returns the order of the subrounds handler
func (s *SubroundsHandler) NotifyOrder() uint32 {
	return common.ConsensusHandlerOrder
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *SubroundsHandler) IsInterfaceNil() bool {
	return s == nil
}
