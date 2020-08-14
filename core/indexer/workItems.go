package indexer

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// WorkItemType -
type WorkItemType int8
const (
	// WorkTypeUnknown -
	WorkTypeUnknown WorkItemType = iota
	// WorkTypeSaveBlock -
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

// NewWorkItem creates a new work items
func NewWorkItem(workItemType WorkItemType, data interface{}) (*workItem, error) {
	if workItemType == WorkTypeUnknown {
		return nil, ErrInvalidWorkItemType
	}

	return &workItem{
		Type: workItemType,
		Data: data,
	}, nil
}
