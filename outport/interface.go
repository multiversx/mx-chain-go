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
	SaveAccounts(blockTimestamp uint64, acc []data.UserAccountHandler) error
	FinalizedBlock(headerHash []byte) error
	Close() error
	IsInterfaceNil() bool
}

// OutportHandler is interface that defines what a proxy implementation should be able to do
type OutportHandler interface {
	Driver
	SubscribeDriver(driver Driver) error
	HasDrivers() bool
}
