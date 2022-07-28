package notifier

import (
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	nodeData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("outport/eventNotifier")

const (
	pushEventEndpoint       = "/events/push"
	revertEventsEndpoint    = "/events/revert"
	finalizedEventsEndpoint = "/events/finalized"
)

// SaveBlockData holds the data that will be sent to notifier instance
type SaveBlockData struct {
	Hash      string                                                  `json:"hash"`
	Txs       map[string]nodeData.TransactionHandlerWithGasUsedAndFee `json:"txs"`
	Scrs      map[string]nodeData.TransactionHandlerWithGasUsedAndFee `json:"scrs"`
	LogEvents []Event                                                 `json:"events"`
}

// Event holds event data
type Event struct {
	Address    string   `json:"address"`
	Identifier string   `json:"identifier"`
	Topics     [][]byte `json:"topics"`
	Data       []byte   `json:"data"`
}

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
	Marshalizer     marshal.Marshalizer
	Hasher          hashing.Hasher
	PubKeyConverter core.PubkeyConverter
}

// NewEventNotifier creates a new instance of the eventNotifier
// It implements all methods of process.Indexer
func NewEventNotifier(args ArgsEventNotifier) (*eventNotifier, error) {
	return &eventNotifier{
		httpClient:      args.HttpClient,
		marshalizer:     args.Marshalizer,
		hasher:          args.Hasher,
		pubKeyConverter: args.PubKeyConverter,
	}, nil
}

// SaveBlock converts block data in order to be pushed to subscribers
func (en *eventNotifier) SaveBlock(args *indexer.ArgsSaveBlockData) error {
	log.Debug("eventNotifier: SaveBlock called at block", "block hash", args.HeaderHash)
	if args.TransactionsPool == nil {
		return ErrNilTransactionsPool
	}

	log.Debug("eventNotifier: checking if block has logs", "num logs", len(args.TransactionsPool.Logs))
	log.Debug("eventNotifier: checking if block has txs", "num txs", len(args.TransactionsPool.Txs))

	events := en.getLogEventsFromTransactionsPool(args.TransactionsPool.Logs)
	log.Debug("eventNotifier: extracted events from block logs", "num events", len(events))

	blockData := SaveBlockData{
		Hash:      hex.EncodeToString(args.HeaderHash),
		Txs:       args.TransactionsPool.Txs,
		Scrs:      args.TransactionsPool.Scrs,
		LogEvents: events,
	}

	err := en.httpClient.Post(pushEventEndpoint, blockData, nil)
	if err != nil {
		return fmt.Errorf("%w in eventNotifier.SaveBlock while posting block data", err)
	}

	return nil
}

func (en *eventNotifier) getLogEventsFromTransactionsPool(logs []*nodeData.LogData) []Event {
	var logEvents []nodeData.EventHandler
	for _, logData := range logs {
		if logData == nil {
			continue
		}
		if check.IfNil(logData.LogHandler) {
			continue
		}

		logEvents = append(logEvents, logData.LogHandler.GetLogEvents()...)
	}

	if len(logEvents) == 0 {
		return nil
	}

	var events []Event
	for _, eventHandler := range logEvents {
		if !eventHandler.IsInterfaceNil() {
			bech32Address := en.pubKeyConverter.Encode(eventHandler.GetAddress())
			eventIdentifier := string(eventHandler.GetIdentifier())

			log.Debug("eventNotifier: received event from address",
				"address", bech32Address,
				"identifier", eventIdentifier,
			)

			events = append(events, Event{
				Address:    bech32Address,
				Identifier: eventIdentifier,
				Topics:     eventHandler.GetTopics(),
				Data:       eventHandler.GetData(),
			})
		}
	}

	return events
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

	err = en.httpClient.Post(revertEventsEndpoint, revertBlock, nil)
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

	err := en.httpClient.Post(finalizedEventsEndpoint, finalizedBlock, nil)
	if err != nil {
		return fmt.Errorf("%w in eventNotifier.FinalizedBlock while posting event data", err)
	}

	return nil
}

// SaveRoundsInfo returns nil
func (en *eventNotifier) SaveRoundsInfo(_ []*indexer.RoundInfo) error {
	return nil
}

// SaveValidatorsRating returns nil
func (en *eventNotifier) SaveValidatorsRating(_ string, _ []*indexer.ValidatorRatingInfo) error {
	return nil
}

// SaveValidatorsPubKeys returns nil
func (en *eventNotifier) SaveValidatorsPubKeys(_ map[uint32][][]byte, _ uint32) error {
	return nil
}

// SaveAccounts does nothing
func (en *eventNotifier) SaveAccounts(_ uint64, _ map[string]*indexer.AlteredAccount) error {
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
