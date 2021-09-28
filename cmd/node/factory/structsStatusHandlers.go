package factory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/ElrondNetwork/elrond-go/statusHandler/persister"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("main")

// StatusHandlersInfo is struct that stores all components that are returned when status handlers are created
type statusHandlersInfo struct {
	AppStatusHandler  core.AppStatusHandler
	StatusMetrics     external.StatusMetricsHandler
	PersistentHandler *persister.PersistentStatusHandler
}

type statusHandlerUtilsFactory struct {
}

// NewStatusHandlersFactory will return the status handler factory
func NewStatusHandlersFactory() (*statusHandlerUtilsFactory, error) {
	return &statusHandlerUtilsFactory{}, nil
}

// Create will return a slice of status handlers
func (shuf *statusHandlerUtilsFactory) Create(
	marshalizer marshal.Marshalizer,
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter,
) (StatusHandlersUtils, error) {
	var appStatusHandlers []core.AppStatusHandler
	var err error
	var handler core.AppStatusHandler

	baseErrMessage := "error creating status handler"
	if check.IfNil(marshalizer) {
		return nil, fmt.Errorf("%s: nil marshalizer", baseErrMessage)
	}
	if check.IfNil(uint64ByteSliceConverter) {
		return nil, fmt.Errorf("%s: nil uint64 byte slice converter", baseErrMessage)
	}

	statusMetrics := statusHandler.NewStatusMetrics()
	appStatusHandlers = append(appStatusHandlers, statusMetrics)

	persistentHandler, err := persister.NewPersistentStatusHandler(marshalizer, uint64ByteSliceConverter)
	if err != nil {
		return nil, err
	}
	appStatusHandlers = append(appStatusHandlers, persistentHandler)

	if len(appStatusHandlers) > 0 {
		handler, err = statusHandler.NewAppStatusFacadeWithHandlers(appStatusHandlers...)
		if err != nil {
			log.Warn("cannot init AppStatusFacade", "error", err)
		}
	} else {
		handler = statusHandler.NewNilStatusHandler()
		log.Debug("no AppStatusHandler used: started with NilStatusHandler")
	}

	statusHandlersInfoObject := new(statusHandlersInfo)
	statusHandlersInfoObject.AppStatusHandler = handler
	statusHandlersInfoObject.StatusMetrics = statusMetrics
	statusHandlersInfoObject.PersistentHandler = persistentHandler

	return statusHandlersInfoObject, nil
}

// UpdateStorerAndMetricsForPersistentHandler will set storer for persistent status handler
func (shi *statusHandlersInfo) UpdateStorerAndMetricsForPersistentHandler(store storage.Storer) error {
	err := shi.PersistentHandler.SetStorage(store)
	if err != nil {
		return err
	}

	return nil
}

// StatusHandler returns the status handler
func (shi *statusHandlersInfo) StatusHandler() core.AppStatusHandler {
	return shi.AppStatusHandler
}

// Metrics returns the status metrics
func (shi *statusHandlersInfo) Metrics() external.StatusMetricsHandler {
	return shi.StatusMetrics
}

// IsInterfaceNil returns true if the interface is nil
func (shi *statusHandlersInfo) IsInterfaceNil() bool {
	return shi == nil
}
