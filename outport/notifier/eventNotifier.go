package notifier

import (
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	nodeData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	outportSenderData "github.com/multiversx/mx-chain-core-go/websocketOutportDriver/data"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("outport/eventNotifier")

const (
	pushEventEndpoint       = "/events/push"
	revertEventsEndpoint    = "/events/revert"
	finalizedEventsEndpoint = "/events/finalized"
)

// RevertBlock holds revert event data
type RevertBlock struct {
	Hash  string `json:"hash"`
	Nonce uint64 `json:"nonce"`
	Round uint64 `json:"round"`
	Epoch uint32 `json:"epoch"`
}

// FinalizedBlock holds finalized block data
type FinalizedBlock struct {
	Hash string `json:"hash"`
}

type eventNotifier struct {
	httpClient      httpClientHandler
	marshalizer     marshal.Marshalizer
	hasher          hashing.Hasher
	pubKeyConverter core.PubkeyConverter
}

// ArgsEventNotifier defines the arguments needed for event notifier creation
type ArgsEventNotifier struct {
	HttpClient      httpClientHandler
	Marshaller      marshal.Marshalizer
	Hasher          hashing.Hasher
	PubKeyConverter core.PubkeyConverter
}

// NewEventNotifier creates a new instance of the eventNotifier
// It implements all methods of process.Indexer
func NewEventNotifier(args ArgsEventNotifier) (*eventNotifier, error) {
	err := checkEventNotifierArgs(args)
	if err != nil {
		return nil, err
	}

	return &eventNotifier{
		httpClient:      args.HttpClient,
		marshalizer:     args.Marshaller,
		hasher:          args.Hasher,
		pubKeyConverter: args.PubKeyConverter,
	}, nil
}

func checkEventNotifierArgs(args ArgsEventNotifier) error {
	if check.IfNil(args.HttpClient) {
		return ErrNilHTTPClientWrapper
	}
	if check.IfNil(args.Marshaller) {
		return ErrNilMarshaller
	}
	if check.IfNil(args.Hasher) {
		return ErrNilHasher
	}
	if check.IfNil(args.PubKeyConverter) {
		return ErrNilPubKeyConverter
	}

	return nil
}

// SaveBlock converts block data in order to be pushed to subscribers
func (en *eventNotifier) SaveBlock(args *outport.ArgsSaveBlockData) error {
	log.Debug("eventNotifier: SaveBlock called at block", "block hash", args.HeaderHash)
	if args.TransactionsPool == nil {
		return ErrNilTransactionsPool
	}

	argsSaveBlock := outportSenderData.ArgsSaveBlock{
		HeaderType:        core.GetHeaderType(args.Header),
		ArgsSaveBlockData: *args,
	}

	err := en.httpClient.Post(pushEventEndpoint, argsSaveBlock)
	if err != nil {
		return fmt.Errorf("%w in eventNotifier.SaveBlock while posting block data", err)
	}

	return nil
}

// RevertIndexedBlock converts revert data in order to be pushed to subscribers
func (en *eventNotifier) RevertIndexedBlock(header nodeData.HeaderHandler, _ nodeData.BodyHandler) error {
	blockHash, err := core.CalculateHash(en.marshalizer, en.hasher, header)
	if err != nil {
		return fmt.Errorf("%w in eventNotifier.RevertIndexedBlock while computing the block hash", err)
	}

	revertBlock := RevertBlock{
		Hash:  hex.EncodeToString(blockHash),
		Nonce: header.GetNonce(),
		Round: header.GetRound(),
		Epoch: header.GetEpoch(),
	}

	err = en.httpClient.Post(revertEventsEndpoint, revertBlock)
	if err != nil {
		return fmt.Errorf("%w in eventNotifier.RevertIndexedBlock while posting event data", err)
	}

	return nil
}

// FinalizedBlock converts finalized block data in order to push it to subscribers
func (en *eventNotifier) FinalizedBlock(headerHash []byte) error {
	finalizedBlock := FinalizedBlock{
		Hash: hex.EncodeToString(headerHash),
	}

	err := en.httpClient.Post(finalizedEventsEndpoint, finalizedBlock)
	if err != nil {
		return fmt.Errorf("%w in eventNotifier.FinalizedBlock while posting event data", err)
	}

	return nil
}

// SaveRoundsInfo returns nil
func (en *eventNotifier) SaveRoundsInfo(_ []*outport.RoundInfo) error {
	return nil
}

// SaveValidatorsRating returns nil
func (en *eventNotifier) SaveValidatorsRating(_ string, _ []*outport.ValidatorRatingInfo) error {
	return nil
}

// SaveValidatorsPubKeys returns nil
func (en *eventNotifier) SaveValidatorsPubKeys(_ map[uint32][][]byte, _ uint32) error {
	return nil
}

// SaveAccounts does nothing
func (en *eventNotifier) SaveAccounts(_ uint64, _ map[string]*outport.AlteredAccount, _ uint32) error {
	return nil
}

// IsInterfaceNil returns whether the interface is nil
func (en *eventNotifier) IsInterfaceNil() bool {
	return en == nil
}

// Close returns nil
func (en *eventNotifier) Close() error {
	return nil
}
