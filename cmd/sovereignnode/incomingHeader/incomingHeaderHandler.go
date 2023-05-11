package incomingHeader

import (
	"encoding/hex"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("headerSubscriber")

// ArgsIncomingHeaderHandler is a struct placeholder for args needed to create a new incoming header handler
type ArgsIncomingHeaderHandler struct {
	HeadersPool HeadersPool
	TxPool      TransactionPool
	Marshaller  marshal.Marshalizer
	Hasher      hashing.Hasher
}

type incomingHeaderHandler struct {
	headersPool HeadersPool
	txPool      TransactionPool
	marshaller  marshal.Marshalizer
	hasher      hashing.Hasher
}

// NewIncomingHeaderHandler creates an incoming header handler which should be able to receive incoming headers and events
// from a chain to local sovereign chain. This handler will validate the events(using proofs in the future) and create
// incoming miniblocks and transaction(which will be added in pool) to be executed in sovereign shard.
func NewIncomingHeaderHandler(args ArgsIncomingHeaderHandler) (*incomingHeaderHandler, error) {
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

	return &incomingHeaderHandler{
		headersPool: args.HeadersPool,
		txPool:      args.TxPool,
		marshaller:  args.Marshaller,
		hasher:      args.Hasher,
	}, nil
}

// AddHeader will receive the incoming header, validate it, create incoming mbs and transactions and add them to pool
func (ihs *incomingHeaderHandler) AddHeader(headerHash []byte, header sovereign.IncomingHeaderHandler) error {
	log.Info("received incoming header", "hash", hex.EncodeToString(headerHash))

	headerV2, castOk := header.GetHeaderHandler().(*block.HeaderV2)
	if !castOk {
		return errInvalidHeaderType
	}

	incomingSCRs, err := ihs.createIncomingSCRs(header.GetIncomingEventHandlers())
	if err != nil {
		return err
	}

	extendedHeader := &block.ShardHeaderExtended{
		Header:             headerV2,
		IncomingMiniBlocks: createIncomingMb(incomingSCRs),
	}

	err = ihs.addExtendedHeaderToPool(extendedHeader)
	if err != nil {
		return err
	}

	ihs.addSCRsToPool(incomingSCRs)
	return nil
}

func (ihs *incomingHeaderHandler) addExtendedHeaderToPool(extendedHeader data.ShardHeaderExtendedHandler) error {
	extendedHeaderHash, err := core.CalculateHash(ihs.marshaller, ihs.hasher, extendedHeader)
	if err != nil {
		return err
	}

	ihs.headersPool.AddHeader(extendedHeaderHash, extendedHeader)
	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ihs *incomingHeaderHandler) IsInterfaceNil() bool {
	return ihs == nil
}
