package incomingHeader

import (
	"encoding/hex"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"

	sovereignBlock "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	"github.com/multiversx/mx-chain-go/errors"
)

var log = logger.GetOrCreate("headerSubscriber")

// ArgsIncomingHeaderProcessor is a struct placeholder for args needed to create a new incoming header processor
type ArgsIncomingHeaderProcessor struct {
	HeadersPool                     HeadersPool
	OutGoingOperationsPool          sovereignBlock.OutGoingOperationsPool
	TxPool                          TransactionPool
	Marshaller                      marshal.Marshalizer
	Hasher                          hashing.Hasher
	MainChainNotarizationStartRound uint64
	DataCodec                       SovereignDataCodec
	TopicsChecker                   TopicsChecker
}

type incomingHeaderProcessor struct {
	eventsProc         *incomingEventsProcessor
	extendedHeaderProc *extendedHeaderProcessor

	outGoingPool                    sovereignBlock.OutGoingOperationsPool
	mainChainNotarizationStartRound uint64
}

// NewIncomingHeaderProcessor creates an incoming header processor which should be able to receive incoming headers and events
// from a chain to local sovereign chain. This handler will validate the events(using proofs in the future) and create
// incoming miniblocks and transaction(which will be added in pool) to be executed in sovereign shard.
func NewIncomingHeaderProcessor(args ArgsIncomingHeaderProcessor) (*incomingHeaderProcessor, error) {
	if check.IfNil(args.HeadersPool) {
		return nil, errNilHeadersPool
	}
	if check.IfNil(args.TxPool) {
		return nil, errNilTxPool
	}
	if check.IfNil(args.Marshaller) {
		return nil, core.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, core.ErrNilHasher
	}
	if check.IfNil(args.OutGoingOperationsPool) {
		return nil, errors.ErrNilOutGoingOperationsPool
	}
	if check.IfNil(args.DataCodec) {
		return nil, errors.ErrNilDataCodec
	}
	if check.IfNil(args.TopicsChecker) {
		return nil, errors.ErrNilTopicsChecker
	}

	depositProc := &depositEventProc{
		marshaller:    args.Marshaller,
		hasher:        args.Hasher,
		dataCodec:     args.DataCodec,
		topicsChecker: args.TopicsChecker,
	}

	executedOpProc := &executedBridgeOpEventProc{
		depositEventProc: depositProc,
	}

	eventsProc := &incomingEventsProcessor{
		handlers: make(map[string]IncomingEventHandler),
	}
	err := eventsProc.registerProcessor(eventIDDepositIncomingTransfer, depositProc)
	if err != nil {
		return nil, nil
	}
	err = eventsProc.registerProcessor(eventIDExecutedOutGoingBridgeOp, executedOpProc)
	if err != nil {
		return nil, nil
	}

	extendedHearProc := &extendedHeaderProcessor{
		headersPool: args.HeadersPool,
		txPool:      args.TxPool,
		marshaller:  args.Marshaller,
		hasher:      args.Hasher,
	}

	log.Debug("NewIncomingHeaderProcessor", "starting round to notarize main chain headers", args.MainChainNotarizationStartRound)

	return &incomingHeaderProcessor{
		eventsProc:                      eventsProc,
		extendedHeaderProc:              extendedHearProc,
		outGoingPool:                    args.OutGoingOperationsPool,
		mainChainNotarizationStartRound: args.MainChainNotarizationStartRound,
	}, nil
}

// AddHeader will receive the incoming header, validate it, create incoming mbs and transactions and add them to pool
func (ihp *incomingHeaderProcessor) AddHeader(headerHash []byte, header sovereign.IncomingHeaderHandler) error {
	if check.IfNil(header) || check.IfNil(header.GetHeaderHandler()) {
		return data.ErrNilHeader
	}

	log.Info("received incoming header",
		"hash", hex.EncodeToString(headerHash),
		"nonce", header.GetHeaderHandler().GetNonce(),
		"round", header.GetHeaderHandler().GetRound(),
	)
	round := header.GetHeaderHandler().GetRound()

	// pre-genesis header, needed to track/link genesis header on top of this one. Every node with an enabled notifier
	// will validate that the next genesis header with round == mainChainNotarizationStartRound is on top of pre-genesis header.
	// just save internal header to tracker, no need to process anything from it
	if round == ihp.mainChainNotarizationStartRound-1 {
		log.Debug("received pre-genesis header", "round", header.GetHeaderHandler().GetRound())
		return ihp.extendedHeaderProc.addPreGenesisExtendedHeaderToPool(header)
	}

	if round < ihp.mainChainNotarizationStartRound {
		log.Debug("do not notarize incoming header, round lower than main chain notarization start round",
			"round", round,
			"start round", ihp.mainChainNotarizationStartRound)
		return nil
	}

	res, err := ihp.eventsProc.processIncomingEvents(header.GetIncomingEventHandlers())
	if err != nil {
		return err
	}

	extendedHeader, err := createExtendedHeader(header, res.scrs)
	if err != nil {
		return err
	}

	err = ihp.extendedHeaderProc.addExtendedHeaderAndSCRsToPool(extendedHeader, res.scrs)
	if err != nil {
		return err
	}

	ihp.addConfirmedBridgeOpsToPool(res.confirmedBridgeOps)
	return nil
}

func (ihp *incomingHeaderProcessor) addConfirmedBridgeOpsToPool(ops []*ConfirmedBridgeOp) {
	for _, op := range ops {
		// This is not a critical error. This might just happen when a leader tries to re-send unconfirmed confirmation
		// that have been already executed, but the confirmation from notifier comes too late, and we receive a double
		// confirmation.
		err := ihp.outGoingPool.ConfirmOperation(op.HashOfHashes, op.Hash)
		if err != nil {
			log.Debug("incomingHeaderProcessor.AddHeader.addConfirmedBridgeOpsToPool",
				"error", err,
				"hashOfHashes", hex.EncodeToString(op.HashOfHashes),
				"hash", hex.EncodeToString(op.Hash),
			)
		}
	}
}

// CreateExtendedHeader will create an extended shard header with incoming scrs and mbs from the events of the received header
func (ihp *incomingHeaderProcessor) CreateExtendedHeader(header sovereign.IncomingHeaderHandler) (data.ShardHeaderExtendedHandler, error) {
	res, err := ihp.eventsProc.processIncomingEvents(header.GetIncomingEventHandlers())
	if err != nil {
		return nil, err
	}

	return createExtendedHeader(header, res.scrs)
}

// RegisterEventHandler will register an extra incoming event processor. For the registered processor, a subscription
// should be added to NotifierConfig.SubscribedEvents from sovereignConfig.toml
func (ihp *incomingHeaderProcessor) RegisterEventHandler(event string, proc IncomingEventHandler) error {
	return ihp.eventsProc.registerProcessor(event, proc)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ihp *incomingHeaderProcessor) IsInterfaceNil() bool {
	return ihp == nil
}
