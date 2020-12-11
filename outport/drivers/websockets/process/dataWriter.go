package process

import (
	"encoding/json"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/outport/drivers/websockets/types"
	outportTypes "github.com/ElrondNetwork/elrond-go/outport/types"
)

// ArgsDataWriter -
type ArgsDataWriter struct {
	Hasher                   hashing.Hasher
	Marshalizer              marshal.Marshalizer
	ValidatorPubkeyConverter core.PubkeyConverter
}

type dataWriter struct {
	chanWrite                  chan []byte
	blockProcessor             *blockProcessor
	validatorsPubKeysProcessor *validatorsPubKeysProcessor
}

// NewDataWriter -
func NewDataWriter(args *ArgsDataWriter) (*dataWriter, error) {
	bp, err := newBlockProcessor(args.Hasher, args.Marshalizer)
	if err != nil {
		return nil, err
	}

	vpp, err := newValidatorsPubKeyProcessor(args.ValidatorPubkeyConverter)
	if err != nil {
		return nil, err
	}

	return &dataWriter{
		chanWrite:                  make(chan []byte, 100),
		blockProcessor:             bp,
		validatorsPubKeysProcessor: vpp,
	}, nil
}

// SaveBlock -
func (dw *dataWriter) SaveBlock(args outportTypes.ArgsSaveBlocks) error {
	block, err := dw.blockProcessor.prepareBlock(args.Header, args.SignersIndexes, args.Body, args.NotarizedHeadersHashes, args.TxsFromPool)
	if err != nil {
		return err
	}

	err = dw.prepareAndWrite(types.MessageSaveBlock, block)
	if err != nil {
		return err
	}

	return nil
}

// SaveValidatorsPubKeys -
func (dw *dataWriter) SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32) error {
	validatorsKeys := dw.validatorsPubKeysProcessor.prepareValidatorPubKey(validatorsPubKeys, epoch)
	err := dw.prepareAndWrite(types.MessageValidatorsPubKeys, validatorsKeys)
	if err != nil {
		return err
	}

	return nil
}

func (dw *dataWriter) prepareAndWrite(messageType int, data interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	message := &types.DataToSend{
		MessageType: messageType,
		Info:        dataBytes,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	dw.chanWrite <- messageBytes
	return nil
}

// ReadBlocking -
func (dw *dataWriter) ReadBlocking() ([]byte, bool) {
	data, ok := <-dw.chanWrite

	return data, ok
}

// Close -
func (dw *dataWriter) Close() error {
	return nil
}
