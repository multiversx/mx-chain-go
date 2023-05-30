package host

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/marshal"
)

// ArgsHostDriver holds the arguments needed for creating a new hostDriver
type ArgsHostDriver struct {
	Marshaller marshal.Marshalizer
	SenderHost SenderHost
	Log        core.Logger
}

type hostDriver struct {
	marshaller  marshal.Marshalizer
	senderHost  SenderHost
	isClosed    atomic.Flag
	log         core.Logger
	cfg         outport.OutportConfig
	payloadProc payloadProcessorHandler
}

// NewHostDriver will create a new instance of hostDriver
func NewHostDriver(args ArgsHostDriver) (*hostDriver, error) {
	if check.IfNil(args.SenderHost) {
		return nil, ErrNilHost
	}
	if check.IfNil(args.Marshaller) {
		return nil, core.ErrNilMarshalizer
	}
	if check.IfNil(args.Log) {
		return nil, core.ErrNilLogger
	}

	payloadProc, err := newPayloadProcessor()
	if err != nil {
		return nil, err
	}

	err = args.SenderHost.SetPayloadHandler(payloadProc)
	if err != nil {
		return nil, err
	}

	return &hostDriver{
		marshaller: args.Marshaller,
		senderHost: args.SenderHost,
		log:        args.Log,
		isClosed:   atomic.Flag{},
	}, nil
}

// SaveBlock will handle the saving of block
func (o *hostDriver) SaveBlock(outportBlock *outport.OutportBlock) error {
	return o.handleAction(outportBlock, outport.TopicSaveBlock)
}

// RevertIndexedBlock will handle the action of reverting the indexed block
func (o *hostDriver) RevertIndexedBlock(blockData *outport.BlockData) error {
	return o.handleAction(blockData, outport.TopicRevertIndexedBlock)
}

// SaveRoundsInfo will handle the saving of rounds
func (o *hostDriver) SaveRoundsInfo(roundsInfos *outport.RoundsInfo) error {
	return o.handleAction(roundsInfos, outport.TopicSaveRoundsInfo)
}

// SaveValidatorsPubKeys will handle the saving of the validators' public keys
func (o *hostDriver) SaveValidatorsPubKeys(validatorsPubKeys *outport.ValidatorsPubKeys) error {
	return o.handleAction(validatorsPubKeys, outport.TopicSaveValidatorsPubKeys)
}

// SaveValidatorsRating will handle the saving of the validators' rating
func (o *hostDriver) SaveValidatorsRating(validatorsRating *outport.ValidatorsRating) error {
	return o.handleAction(validatorsRating, outport.TopicSaveValidatorsRating)
}

// SaveAccounts will handle the accounts' saving
func (o *hostDriver) SaveAccounts(accounts *outport.Accounts) error {
	return o.handleAction(accounts, outport.TopicSaveAccounts)
}

// FinalizedBlock will handle the finalized block
func (o *hostDriver) FinalizedBlock(finalizedBlock *outport.FinalizedBlock) error {
	return o.handleAction(finalizedBlock, outport.TopicFinalizedBlock)
}

// GetMarshaller returns the internal marshaller
func (o *hostDriver) GetMarshaller() marshal.Marshalizer {
	return o.marshaller
}

func (o *hostDriver) handleAction(args interface{}, topic string) error {
	if o.isClosed.IsSet() {
		return ErrHostIsClosed
	}

	marshalledPayload, err := o.marshaller.Marshal(args)
	if err != nil {
		return fmt.Errorf("%w while marshaling block for topic %s", err, topic)
	}

	err = o.senderHost.Send(marshalledPayload, topic)
	if err != nil {
		return fmt.Errorf("%w while sending data on route for topic %s", err, topic)
	}

	return nil
}

// RegisterHandlerForSettingsRequest will register the handler function for the settings request
func (o *hostDriver) RegisterHandlerForSettingsRequest(handlerFunction func()) error {
	return o.payloadProc.SetHandlerFunc(handlerFunction)
}

// CurrentSettings will send the current settings
func (o *hostDriver) CurrentSettings(config outport.OutportConfig) error {
	configBytes, err := o.marshaller.Marshal(&config)
	if err != nil {
		return err
	}

	return o.senderHost.Send(configBytes, outport.TopicSettings)
}

// Close will handle the closing of the outport driver web socket sender
func (o *hostDriver) Close() error {
	o.isClosed.SetValue(true)
	return o.senderHost.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (o *hostDriver) IsInterfaceNil() bool {
	return o == nil
}
