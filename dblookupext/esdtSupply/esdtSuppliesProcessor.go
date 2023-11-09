//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. supplyESDT.proto

package esdtSupply

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("dblookupext/esdtSupply")

type suppliesProcessor struct {
	logsProc *logsProcessor
	logsGet  *logsGetter
	mutex    sync.Mutex
}

// NewSuppliesProcessor will create a new instance of the supplies processor
func NewSuppliesProcessor(
	marshalizer marshal.Marshalizer,
	suppliesStorer storage.Storer,
	logsStorer storage.Storer,
) (*suppliesProcessor, error) {
	if check.IfNil(marshalizer) {
		return nil, core.ErrNilMarshalizer
	}
	if check.IfNil(suppliesStorer) {
		return nil, core.ErrNilStore
	}
	if check.IfNil(logsStorer) {
		return nil, core.ErrNilStore
	}

	logsGet := newLogsGetter(marshalizer, logsStorer)
	logsProc := newLogsProcessor(marshalizer, suppliesStorer)

	return &suppliesProcessor{
		logsProc: logsProc,
		logsGet:  logsGet,
	}, nil
}

// ProcessLogs will process the provided logs
func (sp *suppliesProcessor) ProcessLogs(blockNonce uint64, logs []*data.LogData) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	logsMap := make(map[string]*data.LogData)
	for _, logData := range logs {
		if logData != nil {
			logsMap[logData.TxHash] = logData
		}
	}

	return sp.logsProc.processLogs(blockNonce, logsMap, false)
}

// RevertChanges will revert supplies changes based on the provided block body
func (sp *suppliesProcessor) RevertChanges(header data.HeaderHandler, body data.BodyHandler) error {
	if check.IfNil(header) || check.IfNil(body) {
		return nil
	}

	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	logsFromDB, err := sp.logsGet.getLogsBasedOnBody(body)
	if err != nil {
		return err
	}

	return sp.logsProc.processLogs(header.GetNonce(), logsFromDB, true)
}

// GetESDTSupply will return the supply from the storage for the given token
func (sp *suppliesProcessor) GetESDTSupply(token string) (*SupplyESDT, error) {
	return sp.logsProc.getESDTSupply([]byte(token))
}

// IsInterfaceNil returns true if there is no value under the interface
func (sp *suppliesProcessor) IsInterfaceNil() bool {
	return sp == nil
}
