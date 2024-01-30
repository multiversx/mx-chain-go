package notifier

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// TODO: remove custom http event notifier integration in the following iterations

var log = logger.GetOrCreate("outport/eventNotifier")

const (
	pushEventEndpoint       = "/events/push"
	revertEventsEndpoint    = "/events/revert"
	finalizedEventsEndpoint = "/events/finalized"
)

type eventNotifier struct {
	httpClient     httpClientHandler
	marshalizer    marshal.Marshalizer
	blockContainer BlockContainerHandler
}

// ArgsEventNotifier defines the arguments needed for event notifier creation
type ArgsEventNotifier struct {
	HttpClient     httpClientHandler
	Marshaller     marshal.Marshalizer
	BlockContainer BlockContainerHandler
}

// NewEventNotifier creates a new instance of the eventNotifier
// It implements all methods of process.Indexer
func NewEventNotifier(args ArgsEventNotifier) (*eventNotifier, error) {
	err := checkEventNotifierArgs(args)
	if err != nil {
		return nil, err
	}

	return &eventNotifier{
		httpClient:     args.HttpClient,
		marshalizer:    args.Marshaller,
		blockContainer: args.BlockContainer,
	}, nil
}

func checkEventNotifierArgs(args ArgsEventNotifier) error {
	if check.IfNil(args.HttpClient) {
		return ErrNilHTTPClientWrapper
	}
	if check.IfNil(args.Marshaller) {
		return ErrNilMarshaller
	}
	if check.IfNilReflect(args.BlockContainer) {
		return ErrNilBlockContainerHandler
	}

	return nil
}

// SaveBlock converts block data in order to be pushed to subscribers
func (en *eventNotifier) SaveBlock(args *outport.OutportBlock) error {
	if args.BlockData != nil {
		log.Debug("eventNotifier: SaveBlock called at block", "block hash", args.BlockData.HeaderHash)
	}

	err := en.httpClient.Post(pushEventEndpoint, args)
	if err != nil {
		return fmt.Errorf("%w in eventNotifier.SaveBlock while posting block data", err)
	}

	return nil
}

// RevertIndexedBlock converts revert data in order to be pushed to subscribers
func (en *eventNotifier) RevertIndexedBlock(blockData *outport.BlockData) error {
	err := en.httpClient.Post(revertEventsEndpoint, blockData)
	if err != nil {
		return fmt.Errorf("%w in eventNotifier.RevertIndexedBlock while posting event data", err)
	}

	return nil
}

// FinalizedBlock converts finalized block data in order to push it to subscribers
func (en *eventNotifier) FinalizedBlock(finalizedBlock *outport.FinalizedBlock) error {
	err := en.httpClient.Post(finalizedEventsEndpoint, finalizedBlock)
	if err != nil {
		return fmt.Errorf("%w in eventNotifier.FinalizedBlock while posting event data", err)
	}

	return nil
}

// SaveRoundsInfo returns nil
func (en *eventNotifier) SaveRoundsInfo(_ *outport.RoundsInfo) error {
	return nil
}

// SaveValidatorsRating returns nil
func (en *eventNotifier) SaveValidatorsRating(_ *outport.ValidatorsRating) error {
	return nil
}

// SaveValidatorsPubKeys returns nil
func (en *eventNotifier) SaveValidatorsPubKeys(_ *outport.ValidatorsPubKeys) error {
	return nil
}

// SaveAccounts does nothing
func (en *eventNotifier) SaveAccounts(_ *outport.Accounts) error {
	return nil
}

// GetMarshaller returns internal marshaller
func (en *eventNotifier) GetMarshaller() marshal.Marshalizer {
	return en.marshalizer
}

// IsInterfaceNil returns whether the interface is nil
func (en *eventNotifier) IsInterfaceNil() bool {
	return en == nil
}

// Close returns nil
func (en *eventNotifier) Close() error {
	return nil
}

// RegisterHandler will do nothing
func (en *eventNotifier) RegisterHandler(_ func() error, _ string) error {
	return nil
}

// SetCurrentSettings will do nothing
func (en *eventNotifier) SetCurrentSettings(_ outport.OutportConfig) error {
	return nil
}
