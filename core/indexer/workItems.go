package indexer

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// WorkItemType defines the different work item types that a work queue can handle
type WorkItemType int8
const (
	// WorkTypeUnknown is the invalid work type
	WorkTypeUnknown WorkItemType = iota
	// WorkTypeSaveBlock defines the work type for saving a block in the indexer
	WorkTypeSaveBlock
)

type workItem struct {
	Data interface{}
	Type WorkItemType
}

type saveBlockData struct {
	bodyHandler data.BodyHandler
	headerHandler data.HeaderHandler
	txPool map[string]data.TransactionHandler
	signersIndexes []uint64
	notarizedHeadersHashes []string
}

// NewWorkItem creates a new work item given a type and it's internal data
func NewWorkItem(workItemType WorkItemType, data interface{}) (*workItem, error) {
	if workItemType == WorkTypeUnknown {
		return nil, ErrInvalidWorkItemType
	}

	return &workItem{
		Type: workItemType,
		Data: data,
	}, nil
}
