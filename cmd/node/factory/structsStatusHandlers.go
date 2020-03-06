package factory

import (
	"io"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	factoryViews "github.com/ElrondNetwork/elrond-go/statusHandler/factory"
	"github.com/ElrondNetwork/elrond-go/statusHandler/persister"
	"github.com/ElrondNetwork/elrond-go/statusHandler/view"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/urfave/cli"
)

// ArgStatusHandlers is a struct that stores arguments needed to create status handlers
type ArgStatusHandlers struct {
	LogViewName                  string
	ServersConfigurationFileName string
	Ctx                          *cli.Context
	Marshalizer                  marshal.Marshalizer
	Uint64ByteSliceConverter     typeConverters.Uint64ByteSliceConverter
}

// StatusHandlersInfo is struct that stores all components that are returned when status handlers are created
type statusHandlersInfo struct {
	UseTermUI                bool
	StatusHandler            core.AppStatusHandler
	StatusMetrics            external.StatusMetricsHandler
	PersistentHandler        *persister.PersistentStatusHandler
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
}

// NewStatusHandlersFactoryArgs will return arguments for status handlers
func NewStatusHandlersFactoryArgs(
	logViewName string,
	ctx *cli.Context,
	marshalizer marshal.Marshalizer,
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter,
) *ArgStatusHandlers {
	return &ArgStatusHandlers{
		LogViewName:              logViewName,
		Ctx:                      ctx,
		Marshalizer:              marshalizer,
		Uint64ByteSliceConverter: uint64ByteSliceConverter,
	}
}

// CreateStatusHandlers will return a slice of status handlers
func CreateStatusHandlers(arguments *ArgStatusHandlers) (*statusHandlersInfo, error) {
	var appStatusHandlers []core.AppStatusHandler
	var views []factoryViews.Viewer
	var err error
	var handler core.AppStatusHandler

	presenterStatusHandler := createStatusHandlerPresenter()

	useTermui := !arguments.Ctx.GlobalBool(arguments.LogViewName)
	if useTermui {
		views, err = createViews(presenterStatusHandler)
		if err != nil {
			return nil, err
		}

		writer, ok := presenterStatusHandler.(io.Writer)
		if ok {
			logger.ClearLogObservers()
			err = logger.AddLogObserver(writer, &logger.PlainFormatter{})
			if err != nil {
				return nil, err
			}
		}

		appStatusHandler, ok := presenterStatusHandler.(core.AppStatusHandler)
		if ok {
			appStatusHandlers = append(appStatusHandlers, appStatusHandler)
		}
	}

	if views == nil {
		log.Warn("No views for current node")
	}

	statusMetrics := statusHandler.NewStatusMetrics()
	appStatusHandlers = append(appStatusHandlers, statusMetrics)

	persistentHandler, err := persister.NewPersistentStatusHandler(arguments.Marshalizer, arguments.Uint64ByteSliceConverter)
	if err != nil {
		return nil, err
	}
	appStatusHandlers = append(appStatusHandlers, persistentHandler)

	if len(appStatusHandlers) > 0 {
		handler, err = statusHandler.NewAppStatusFacadeWithHandlers(appStatusHandlers...)
		if err != nil {
			log.Warn("Cannot init AppStatusFacade", err)
		}
	} else {
		handler = statusHandler.NewNilStatusHandler()
		log.Info("No AppStatusHandler used. Started with NilStatusHandler")
	}

	statusHandlersInfoObject := new(statusHandlersInfo)
	statusHandlersInfoObject.StatusHandler = handler
	statusHandlersInfoObject.UseTermUI = useTermui
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

// CreateStatusHandlerPresenter will return an instance of PresenterStatusHandler
func createStatusHandlerPresenter() view.Presenter {
	presenterStatusHandlerFactory := factoryViews.NewPresenterFactory()

	return presenterStatusHandlerFactory.Create()
}

// CreateViews will start an termui console  and will return an object if cannot create and start termuiConsole
func createViews(presenter view.Presenter) ([]factoryViews.Viewer, error) {
	viewsFactory, err := factoryViews.NewViewsFactory(presenter)
	if err != nil {
		return nil, err
	}

	views, err := viewsFactory.Create()
	if err != nil {
		return nil, err
	}

	for _, v := range views {
		err = v.Start()
		if err != nil {
			return nil, err
		}
	}

	return views, nil
}
