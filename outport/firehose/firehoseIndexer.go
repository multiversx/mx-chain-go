package firehose

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/outport"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("firehose")

const (
	firehosePrefix   = "FIRE"
	beginBlockPrefix = "BLOCK_BEGIN"
	endBlockPrefix   = "BLOCK_END"
)

type firehoseIndexer struct {
	writer       io.Writer
	marshaller   marshal.Marshalizer
	blockCreator block.EmptyBlockCreator
}

// NewFirehoseIndexer creates a new firehose instance which outputs block information
func NewFirehoseIndexer(writer io.Writer, blockCreator block.EmptyBlockCreator) (outport.Driver, error) {
	if writer == nil {
		return nil, errNilWriter
	}
	if check.IfNil(blockCreator) {
		return nil, errNilBlockCreator
	}

	return &firehoseIndexer{
		writer:       writer,
		marshaller:   &marshal.GogoProtoMarshalizer{},
		blockCreator: blockCreator,
	}, nil
}

// SaveBlock will write on stdout relevant block information for firehose
func (fi *firehoseIndexer) SaveBlock(outportBlock *outportcore.OutportBlock) error {
	if outportBlock == nil || outportBlock.BlockData == nil {
		return errOutportBlock
	}

	header, err := block.GetHeaderFromBytes(fi.marshaller, fi.blockCreator, outportBlock.BlockData.HeaderBytes)
	if err != nil {
		return err
	}

	log.Debug("firehose: saving block", "nonce", header.GetNonce(), "hash", outportBlock.BlockData.HeaderHash)

	_, err = fmt.Fprintf(fi.writer, "%s %s %d\n",
		firehosePrefix,
		beginBlockPrefix,
		header.GetNonce(),
	)
	if err != nil {
		return fmt.Errorf("could not write %s prefix , err: %w", beginBlockPrefix, err)
	}

	marshalledBlock, err := fi.marshaller.Marshal(outportBlock)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(fi.writer, "%s %s %d %s %d %x\n",
		firehosePrefix,
		endBlockPrefix,
		header.GetNonce(),
		hex.EncodeToString(header.GetPrevHash()),
		header.GetTimeStamp(),
		marshalledBlock,
	)
	if err != nil {
		return fmt.Errorf("could not write %s prefix , err: %w", endBlockPrefix, err)
	}

	return nil
}

// RevertIndexedBlock does nothing
func (fi *firehoseIndexer) RevertIndexedBlock(*outportcore.BlockData) error {
	return nil
}

// SaveRoundsInfo does nothing
func (fi *firehoseIndexer) SaveRoundsInfo(*outportcore.RoundsInfo) error {
	return nil
}

// SaveValidatorsPubKeys does nothing
func (fi *firehoseIndexer) SaveValidatorsPubKeys(*outportcore.ValidatorsPubKeys) error {
	return nil
}

// SaveValidatorsRating does nothing
func (fi *firehoseIndexer) SaveValidatorsRating(*outportcore.ValidatorsRating) error {
	return nil
}

// SaveAccounts does nothing
func (fi *firehoseIndexer) SaveAccounts(*outportcore.Accounts) error {
	return nil
}

// FinalizedBlock does nothing
func (fi *firehoseIndexer) FinalizedBlock(*outportcore.FinalizedBlock) error {
	return nil
}

// GetMarshaller returns internal marshaller
func (fi *firehoseIndexer) GetMarshaller() marshal.Marshalizer {
	return fi.marshaller
}

// Close does nothing
func (fi *firehoseIndexer) Close() error {
	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (fi *firehoseIndexer) IsInterfaceNil() bool {
	return fi == nil
}
