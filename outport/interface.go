package outport

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
)

// Driver is an interface for saving node specific data to other storage.
// This could be an elastic search index, a MySql database or any other external services.
type Driver interface {
	SaveBlock(args *indexer.ArgsSaveBlockData) error
	RevertIndexedBlock(header data.HeaderHandler, body data.BodyHandler) error
	SaveRoundsInfo(roundsInfos []*indexer.RoundInfo) error
	SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32) error
	SaveValidatorsRating(indexID string, infoRating []*indexer.ValidatorRatingInfo) error
	SaveAccounts(blockTimestamp uint64, acc map[string]*indexer.AlteredAccount) error
	FinalizedBlock(headerHash []byte) error
	Close() error
	IsInterfaceNil() bool
}

// OutportHandler is interface that defines what a proxy implementation should be able to do
// The node is able to talk only with this interface
type OutportHandler interface {
	SaveBlock(args *indexer.ArgsSaveBlockData)
	RevertIndexedBlock(header data.HeaderHandler, body data.BodyHandler)
	SaveRoundsInfo(roundsInfos []*indexer.RoundInfo)
	SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32)
	SaveValidatorsRating(indexID string, infoRating []*indexer.ValidatorRatingInfo)
	SaveAccounts(blockTimestamp uint64, acc map[string]*indexer.AlteredAccount)
	FinalizedBlock(headerHash []byte)
	SubscribeDriver(driver Driver) error
	HasDrivers() bool
	Close() error
	IsInterfaceNil() bool
}

type DataProviderOutport interface {
	PrepareOutportSaveBlockData(
		headerHash []byte,
		body data.BodyHandler,
		header data.HeaderHandler,
		rewardsTxs map[string]data.TransactionHandler,
		notarizedHeadersHashes []string,
	) (*indexer.ArgsSaveBlockData, error)
}
