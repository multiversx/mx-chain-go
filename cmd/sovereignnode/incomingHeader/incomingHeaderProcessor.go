package incomingHeader

import (
	"encoding/hex"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("headerSubscriber")

// ArgsIncomingHeaderProcessor is a struct placeholder for args needed to create a new incoming header processor
type ArgsIncomingHeaderProcessor struct {
	HeadersPool HeadersPool
	TxPool      TransactionPool
	Marshaller  marshal.Marshalizer
	Hasher      hashing.Hasher
	StartRound  uint64
}

type incomingHeaderProcessor struct {
	scrProc            *scrProcessor
	extendedHeaderProc *extendedHeaderProcessor
	startRound         uint64
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

	scrProc := &scrProcessor{
		txPool:     args.TxPool,
		marshaller: args.Marshaller,
		hasher:     args.Hasher,
	}

	extendedHearProc := &extendedHeaderProcessor{
		headersPool: args.HeadersPool,
		marshaller:  args.Marshaller,
		hasher:      args.Hasher,
	}

	log.Debug("NewIncomingHeaderProcessor", "starting round to notarize main chain headers", args.StartRound)

	return &incomingHeaderProcessor{
		scrProc:            scrProc,
		extendedHeaderProc: extendedHearProc,
		startRound:         args.StartRound,
	}, nil
}

// AddHeader will receive the incoming header, validate it, create incoming mbs and transactions and add them to pool
func (ihp *incomingHeaderProcessor) AddHeader(headerHash []byte, header sovereign.IncomingHeaderHandler) error {
	log.Info("received incoming header", "hash", hex.EncodeToString(headerHash))
	round := header.GetHeaderHandler().GetRound()
	if round < ihp.startRound {
		log.Debug("do not notarize incoming header, round lower than start round",
			"round", round,
			"start round", ihp.startRound)
		return nil
	}

	headerV2, castOk := header.GetHeaderHandler().(*block.HeaderV2)
	if !castOk {
		return errInvalidHeaderType
	}

	incomingSCRs, err := ihp.scrProc.createIncomingSCRs(header.GetIncomingEventHandlers())
	if err != nil {
		return err
	}

	extendedHeader := createExtendedHeader(headerV2, incomingSCRs)
	err = ihp.extendedHeaderProc.addExtendedHeaderToPool(extendedHeader)
	if err != nil {
		return err
	}

	ihp.scrProc.addSCRsToPool(incomingSCRs)
	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ihp *incomingHeaderProcessor) IsInterfaceNil() bool {
	return ihp == nil
}
