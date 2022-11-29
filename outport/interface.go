package outport

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	outportcore "github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go/outport/process"
)

// Driver is an interface for saving node specific data to other storage.
// This could be an elastic search index, a MySql database or any other external services.
type Driver interface {
	SaveBlock(args *outportcore.ArgsSaveBlockData) error
	RevertIndexedBlock(header data.HeaderHandler, body data.BodyHandler) error
	SaveRoundsInfo(roundsInfos []*outportcore.RoundInfo) error
	SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32) error
	SaveValidatorsRating(indexID string, infoRating []*outportcore.ValidatorRatingInfo) error
	SaveAccounts(blockTimestamp uint64, acc map[string]*outportcore.AlteredAccount, shardID uint32) error
	FinalizedBlock(headerHash []byte) error
	Close() error
	IsInterfaceNil() bool
}

// OutportHandler is interface that defines what a proxy implementation should be able to do
// The node is able to talk only with this interface
type OutportHandler interface {
	SaveBlock(args *outportcore.ArgsSaveBlockData)
	RevertIndexedBlock(header data.HeaderHandler, body data.BodyHandler)
	SaveRoundsInfo(roundsInfos []*outportcore.RoundInfo)
	SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32)
	SaveValidatorsRating(indexID string, infoRating []*outportcore.ValidatorRatingInfo)
	SaveAccounts(blockTimestamp uint64, acc map[string]*outportcore.AlteredAccount, shardID uint32)
	FinalizedBlock(headerHash []byte)
	SubscribeDriver(driver Driver) error
	HasDrivers() bool
	Close() error
	IsInterfaceNil() bool
}

// DataProviderOutport is an interface that defines what an implementation of data provider outport should be able to do
type DataProviderOutport interface {
	PrepareOutportSaveBlockData(arg process.ArgPrepareOutportSaveBlockData) (*outportcore.ArgsSaveBlockData, error)
	IsInterfaceNil() bool
}
