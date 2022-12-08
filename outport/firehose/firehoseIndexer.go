package firehose

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/ElrondNetwork/elrond-go-core/data"
	outportcore "github.com/ElrondNetwork/elrond-go-core/data/outport"
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
	writer io.Writer
}

// NewFirehoseIndexer creates a new firehose instance which outputs block information
func NewFirehoseIndexer(writer io.Writer) (outport.Driver, error) {
	if writer == nil {
		return nil, errNilWriter
	}

	return &firehoseIndexer{
		writer: writer,
	}, nil
}

// SaveBlock will write on stdout relevant block information for firehose
func (fi *firehoseIndexer) SaveBlock(args *outportcore.ArgsSaveBlockData) error {
	log.Debug("firehose: saving block", "nonce", args.Header.GetNonce(), "hash", args.HeaderHash)

	_, err := fmt.Fprintf(fi.writer, "%s %s %d\n",
		firehosePrefix,
		beginBlockPrefix,
		args.Header.GetNonce(),
	)
	if err != nil {
		return fmt.Errorf("could not write %s prefix , err: %w", beginBlockPrefix, err)
	}

	_, err = fmt.Fprintf(fi.writer, "%s %s %d %s %s %d %d\n",
		firehosePrefix,
		endBlockPrefix,
		args.Header.GetNonce(),
		hex.EncodeToString(args.HeaderHash),
		hex.EncodeToString(args.Header.GetPrevHash()),
		args.Header.GetTimeStamp(),
		0, // num transactions, implementation will follow
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
