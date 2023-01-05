package firehose

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/alteredAccount"
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

	headerBytes, headerType, err := fi.getHeaderBytes(args.Header)
	if err != nil {
		return err
	}

	pool, err := getTxPool(args.TransactionsPool)
	if err != nil {
		return fmt.Errorf("getTxPool error: %w, header hash %s", err, hex.EncodeToString(args.HeaderHash))
	}

	body, err := getBody(args.Body)
	if err != nil && err != errNilBlockBody {
		return fmt.Errorf("%w, header hash: %s", err, hex.EncodeToString(args.HeaderHash))
	}

	firehoseBlock := &firehose.FirehoseBlock{
		HeaderBytes:         headerBytes,
		HeaderType:          string(headerType),
		HeaderHash:          args.HeaderHash,
		Body:                body,
		AlteredAccounts:     getAlteredAccounts(args.AlteredAccounts),
		Transactions:        pool.transactions,
		SmartContractResult: pool.smartContractResult,
		Rewards:             pool.rewards,
		Receipts:            pool.receipts,
		InvalidTxs:          pool.invalidTxs,
		Logs:                pool.logs,
		SignersIndexes:      args.SignersIndexes,
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

func (fi *firehoseIndexer) getHeaderBytes(headerHandler data.HeaderHandler) ([]byte, core.HeaderType, error) {
	var err error
	var headerBytes []byte
	var headerType core.HeaderType

	switch header := headerHandler.(type) {
	case *block.MetaBlock:
		headerType = core.MetaHeader
		headerBytes, err = fi.marshaller.Marshal(header)
	case *block.Header:
		headerType = core.ShardHeaderV1
		headerBytes, err = fi.marshaller.Marshal(header)
	case *block.HeaderV2:
		headerType = core.ShardHeaderV2
		headerBytes, err = fi.marshaller.Marshal(header)
	default:
		return nil, "", errInvalidHeaderType
	}

	return headerBytes, headerType, err
}

func getBody(bodyHandler data.BodyHandler) (*block.Body, error) {
	if check.IfNil(bodyHandler) {
		return nil, errNilBlockBody
	}

	body, castOk := bodyHandler.(*block.Body)
	if !castOk {
		return nil, errCannotCastBlockBody
	}

	return body, nil
}

func getAlteredAccounts(accounts map[string]*outportcore.AlteredAccount) []*alteredAccount.AlteredAccount {
	ret := make([]*alteredAccount.AlteredAccount, len(accounts))

	idx := 0
	for _, acc := range accounts {
		ret[idx] = &alteredAccount.AlteredAccount{
			Address: acc.Address,
			Nonce:   acc.Nonce,
			Balance: acc.Balance,
			Tokens:  getTokens(acc.Tokens),
		}
		idx++
	}

	return ret
}

func getTokens(tokens []*outportcore.AccountTokenData) []*alteredAccount.AccountTokenData {
	ret := make([]*alteredAccount.AccountTokenData, len(tokens))

	for idx, token := range tokens {
		ret[idx] = &alteredAccount.AccountTokenData{
			Nonce:      token.Nonce,
			Identifier: token.Identifier,
			Balance:    token.Balance,
			Properties: token.Properties,
		}
	}

	return ret
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
