package indexer

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// WorkItemType defines the different work item types that a work queue can handle
type WorkItemType int8

const (
	// WorkTypeSaveBlock defines the work type for saving a block in the indexer
	WorkTypeSaveBlock = iota
	// WorkTypeUpdateTPSBenchmark defines the work type for updating tps benchmark
	WorkTypeUpdateTPSBenchmark
	//WorkTypeSaveRoundsInfo defines the work type for saving rounds information
	WorkTypeSaveRoundsInfo
	// WorkTypeSaveValidatorsRating defines the work type for saving validators rating information
	WorkTypeSaveValidatorsRating
	// WorkTypeSaveValidatorsPubKeys defines the work type for saving validators public key when at epoch start
	WorkTypeSaveValidatorsPubKeys
	// WorkTypeRemoveBlock defines the work type for removing a block from indexer database
	WorkTypeRemoveBlock
)

type workItem struct {
	Data interface{}
	Type WorkItemType
}

type removeBlockData struct {
	bodyHandler   data.BodyHandler
	headerHandler data.HeaderHandler
}

type saveBlockData struct {
	bodyHandler            data.BodyHandler
	headerHandler          data.HeaderHandler
	txPool                 map[string]data.TransactionHandler
	signersIndexes         []uint64
	notarizedHeadersHashes []string
}

type saveValidatorsRatingData struct {
	indexID    string
	infoRating []ValidatorRatingInfo
}

type saveValidatorsPubKeysData struct {
	epoch             uint32
	validatorsPubKeys map[uint32][][]byte
}

// NewWorkItem creates a new work item given a type and it's internal data
func NewWorkItem(workItemType WorkItemType, data interface{}) *workItem {
	return &workItem{
		Type: workItemType,
		Data: data,
	}
}
