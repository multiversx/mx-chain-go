package firehose

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/firehose"
	outportcore "github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/outport"
)

var log = logger.GetOrCreate("firehose")

const (
	firehosePrefix   = "FIRE"
	beginBlockPrefix = "BLOCK_BEGIN"
	endBlockPrefix   = "BLOCK_END"
)

type firehoseIndexer struct {
	writer     io.Writer
	marshaller marshal.Marshalizer
}

// NewFirehoseIndexer creates a new firehose instance which outputs block information
func NewFirehoseIndexer(writer io.Writer) (outport.Driver, error) {
	if writer == nil {
		return nil, errNilWriter
	}

	return &firehoseIndexer{
		writer:     writer,
		marshaller: &marshal.GogoProtoMarshalizer{},
	}, nil
}

// SaveBlock will write on stdout relevant block information for firehose
func (fi *firehoseIndexer) SaveBlock(args *outportcore.ArgsSaveBlockData) error {
	if check.IfNil(args.Header) {
		return errNilHeader
	}

	log.Debug("firehose: saving block", "nonce", args.Header.GetNonce(), "hash", args.HeaderHash)

	_, err := fmt.Fprintf(fi.writer, "%s %s %d\n",
		firehosePrefix,
		beginBlockPrefix,
		args.Header.GetNonce(),
	)
	if err != nil {
		return fmt.Errorf("could not write %s prefix , err: %w", beginBlockPrefix, err)
	}

	var headerBytes []byte
	var headerType core.HeaderType

	switch header := args.Header.(type) {
	case *block.MetaBlock:
		headerBytes, err = fi.marshaller.Marshal(header)
		headerType = core.MetaHeader
	case *block.Header:
		headerBytes, err = fi.marshaller.Marshal(header)
		headerType = core.ShardHeaderV1
	case *block.HeaderV2:
		headerBytes, err = fi.marshaller.Marshal(header)
		headerType = core.ShardHeaderV2
	default:
		return errInvalidHeaderType
	}

	if err != nil {
		return err
	}

	firehoseBlock := &firehose.FirehoseBlock{
		HeaderHash:  args.HeaderHash,
		HeaderType:  string(headerType),
		HeaderBytes: headerBytes,
	}

	marshalledBlock, err := fi.marshaller.Marshal(firehoseBlock)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(fi.writer, "%s %s %d %s %d %x\n",
		firehosePrefix,
		endBlockPrefix,
		args.Header.GetNonce(),
		hex.EncodeToString(args.Header.GetPrevHash()),
		args.Header.GetTimeStamp(),
		marshalledBlock,
	)
	if err != nil {
		return fmt.Errorf("could not write %s prefix , err: %w", endBlockPrefix, err)
	}

	return nil
}

// RevertIndexedBlock does nothing
func (fi *firehoseIndexer) RevertIndexedBlock(data.HeaderHandler, data.BodyHandler) error {
	return nil
}

// SaveRoundsInfo does nothing
func (fi *firehoseIndexer) SaveRoundsInfo([]*outportcore.RoundInfo) error {
	return nil
}

// SaveValidatorsPubKeys does nothing
func (fi *firehoseIndexer) SaveValidatorsPubKeys(map[uint32][][]byte, uint32) error {
	return nil
}

// SaveValidatorsRating does nothing
func (fi *firehoseIndexer) SaveValidatorsRating(string, []*outportcore.ValidatorRatingInfo) error {
	return nil
}

// SaveAccounts does nothing
func (fi *firehoseIndexer) SaveAccounts(uint64, map[string]*outportcore.AlteredAccount, uint32) error {
	return nil
}

// FinalizedBlock does nothing
func (fi *firehoseIndexer) FinalizedBlock([]byte) error {
	return nil
}

// Close does nothing
func (fi *firehoseIndexer) Close() error {
	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (fi *firehoseIndexer) IsInterfaceNil() bool {
	return fi == nil
}
