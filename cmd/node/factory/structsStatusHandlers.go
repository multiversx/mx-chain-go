package factory

import (
	"io"
	"net/http"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	factoryViews "github.com/ElrondNetwork/elrond-go/statusHandler/factory"
	"github.com/ElrondNetwork/elrond-go/statusHandler/persistor"
	"github.com/ElrondNetwork/elrond-go/statusHandler/view"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// ArgStatusHandlers is a struct that store arguments needed for create status handlers
type ArgStatusHandlers struct {
	LogViewName                  string
	ServersConfigurationFileName string
	UserPrometheusName           string
	Ctx                          *cli.Context
	Marshalizer                  marshal.Marshalizer
}

// StatusHandlersInfo is a struct
type statusHandlersInfo struct {
	PrometheusJoinUrl string
	UsePrometheus     bool
	UserTermui        bool
	StatusHandler     core.AppStatusHandler
	StatusMetrics     external.StatusMetricsHandler
	PersistentHandler *persistor.PersistentStatusHandler
}

// NewStatusHandlersFactoryArgs will return arguments for status handlers
func NewStatusHandlersFactoryArgs(
	logViewName string,
	serversConfigurationFileName string,
	userPrometheusName string,
	ctx *cli.Context,
	marshalizer marshal.Marshalizer,
) *ArgStatusHandlers {
	return &ArgStatusHandlers{
		LogViewName:                  logViewName,
		ServersConfigurationFileName: serversConfigurationFileName,
		UserPrometheusName:           userPrometheusName,
		Ctx:                          ctx,
		Marshalizer:                  marshalizer,
	}
}

// CreateStatusHandlers will return a slice of status handlers
func CreateStatusHandlers(arguments *ArgStatusHandlers) (*statusHandlersInfo, error) {
	var appStatusHandlers []core.AppStatusHandler
	var views []factoryViews.Viewer
	var err error
	var handler core.AppStatusHandler

	prometheusJoinUrl, usePrometheusBool := getPrometheusJoinURLIfAvailable(arguments.Ctx, arguments.ServersConfigurationFileName, arguments.UserPrometheusName)
	if usePrometheusBool {
		prometheusStatusHandler := statusHandler.NewPrometheusStatusHandler()
		appStatusHandlers = append(appStatusHandlers, prometheusStatusHandler)
	}

	presenterStatusHandler := createStatusHandlerPresenter()

	useTermui := !arguments.Ctx.GlobalBool(arguments.LogViewName)
	if useTermui {

		views, err = createViews(presenterStatusHandler)
		if err != nil {
			return nil, err
		}

		writer, ok := presenterStatusHandler.(io.Writer)
		if ok {
			err = log.ChangePrinterHookWriter(writer)
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

	persistentHandler, err := persistor.NewPersistentStatusHandler(arguments.Marshalizer)
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

	statusHandlersInfo := new(statusHandlersInfo)
	statusHandlersInfo.StatusHandler = handler
	statusHandlersInfo.PrometheusJoinUrl = prometheusJoinUrl
	statusHandlersInfo.UsePrometheus = usePrometheusBool
	statusHandlersInfo.UserTermui = useTermui
	statusHandlersInfo.StatusMetrics = statusMetrics
	statusHandlersInfo.PersistentHandler = persistentHandler
	return statusHandlersInfo, nil
}

// UpdateStorerAndMetricsForPersistentHandler will set storer for persistent status handler
func (shi *statusHandlersInfo) UpdateStorerAndMetricsForPersistentHandler(store storage.Storer) error {
	err := shi.PersistentHandler.SetStorage(store)
	if err != nil {
		return err
	}

	uint64Metrics, stringMetrics := shi.PersistentHandler.LoadMetricsFromDb()
	shi.saveUint64Metrics(uint64Metrics)
	shi.saveStringMetrics(stringMetrics)

	shi.PersistentHandler.StartStoreMetricsInStorage()

	return nil
}

func (shi *statusHandlersInfo) saveUint64Metrics(metrics map[string]uint64) {
	if metrics == nil {
		return
	}

	for key, value := range metrics {
		shi.StatusHandler.SetUInt64Value(key, value)
	}
}

func (shi *statusHandlersInfo) saveStringMetrics(metrics map[string]string) {
	if metrics == nil {
		return
	}

	for key, value := range metrics {
		shi.StatusHandler.SetStringValue(key, value)
	}
}

func getPrometheusJoinURLIfAvailable(ctx *cli.Context, serversConfigurationFileName string, userPrometheusName string) (string, bool) {
	prometheusURLAvailable := true
	prometheusJoinUrl, err := getPrometheusJoinURL(ctx.GlobalString(serversConfigurationFileName))
	if err != nil || prometheusJoinUrl == "" {
		prometheusURLAvailable = false
	}
	usePrometheusBool := ctx.GlobalBool(userPrometheusName) && prometheusURLAvailable

	return prometheusJoinUrl, usePrometheusBool
}

func getPrometheusJoinURL(serversConfigurationFileName string) (string, error) {
	serversConfig, err := core.LoadServersPConfig(serversConfigurationFileName)
	if err != nil {
		return "", err
	}
	baseURL := serversConfig.Prometheus.PrometheusBaseURL
	statusURL := baseURL + serversConfig.Prometheus.StatusRoute
	resp, err := http.Get(statusURL)
	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusNotFound {
		return "", errors.New("prometheus URL not available")
	}
	joinURL := baseURL + serversConfig.Prometheus.JoinRoute
	return joinURL, nil
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
