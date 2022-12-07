package firehose

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/ElrondNetwork/elrond-go-core/data"
	outportcore "github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go/outport"
)

type firehoseIndexer struct {
}

func NewFirehoseIndexer() outport.Driver {
	return &firehoseIndexer{}
}

func (fi *firehoseIndexer) SaveBlock(args *outportcore.ArgsSaveBlockData) error {
	fmt.Fprintf(os.Stdout, "FIRE BLOCK_BEGIN %d\n", args.Header.GetNonce())
	fmt.Fprintf(os.Stdout, "FIRE BLOCK_END %d %s %s %d %d\n",
		args.Header.GetNonce(),
		hex.EncodeToString(args.HeaderHash),
		hex.EncodeToString(args.Header.GetPrevHash()),
		args.Header.GetTimeStamp(),
		0,
	)

	return nil
}

func (fi *firehoseIndexer) RevertIndexedBlock(header data.HeaderHandler, body data.BodyHandler) error {
	return nil
}

func (fi *firehoseIndexer) SaveRoundsInfo(roundsInfos []*outportcore.RoundInfo) error {
	return nil
}

func (fi *firehoseIndexer) SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32) error {
	return nil
}

func (fi *firehoseIndexer) SaveValidatorsRating(indexID string, infoRating []*outportcore.ValidatorRatingInfo) error {
	return nil
}

func (fi *firehoseIndexer) SaveAccounts(blockTimestamp uint64, acc map[string]*outportcore.AlteredAccount, shardID uint32) error {
	return nil
}

func (fi *firehoseIndexer) FinalizedBlock(headerHash []byte) error {
	return nil
}

func (fi *firehoseIndexer) Close() error {
	return nil
}

func (fi *firehoseIndexer) IsInterfaceNil() bool {
	return fi == nil
}
