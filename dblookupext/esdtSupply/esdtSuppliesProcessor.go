//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. supplyESDT.proto

package esdtSupply

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/storage"
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
func (sp *suppliesProcessor) ProcessLogs(logs map[string]data.LogHandler) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	return sp.logsProc.processLogs(logs, false)
}

// RevertChanges will revert supplies changes based on the provided block body
func (sp *suppliesProcessor) RevertChanges(_ data.HeaderHandler, body data.BodyHandler) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	logsFromDB, err := sp.logsGet.getLogsBasedOnBody(body)
	if err != nil {
		return err
	}

	return sp.logsProc.processLogs(logsFromDB, true)
}

func (sp *suppliesProcessor) GetESDTSupply(token string) (string, error) {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	return sp.logsProc.getESDTSupply(token)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sp *suppliesProcessor) IsInterfaceNil() bool {
	return sp == nil
}
