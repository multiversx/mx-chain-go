package factory

import (
	"fmt"
	"io"
	"math/big"
	"os"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data/metrics"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
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
	ChanStartViews               chan struct{}
	ChanLogRewrite               chan struct{}
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
	chanStartViews chan struct{},
	chanLogRewrite chan struct{},
) (*ArgStatusHandlers, error) {
	baseErrMessage := "error creating status handler factory arguments"
	if ctx == nil {
		return nil, fmt.Errorf("%s: nil context", baseErrMessage)
	}
	if check.IfNil(marshalizer) {
		return nil, fmt.Errorf("%s: nil marshalizer", baseErrMessage)
	}
	if check.IfNil(uint64ByteSliceConverter) {
		return nil, fmt.Errorf("%s: nil uint64 byte slice converter", baseErrMessage)
	}
	if chanLogRewrite == nil {
		return nil, fmt.Errorf("%s: nil log rewrite channel", baseErrMessage)
	}
	if chanStartViews == nil {
		return nil, fmt.Errorf("%s: nil views start channel", baseErrMessage)
	}

	return &ArgStatusHandlers{
		LogViewName:              logViewName,
		Ctx:                      ctx,
		Marshalizer:              marshalizer,
		Uint64ByteSliceConverter: uint64ByteSliceConverter,
		ChanStartViews:           chanStartViews,
		ChanLogRewrite:           chanLogRewrite,
	}, nil
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
		views, err = createViews(presenterStatusHandler, arguments.ChanStartViews)
		if err != nil {
			return nil, err
		}

		go func() {
			<-arguments.ChanLogRewrite
			writer, ok := presenterStatusHandler.(io.Writer)
			if !ok {
				return
			}
			err = logger.RemoveLogObserver(os.Stdout)
			if err != nil {
				log.Warn("cannot remove the log observer for std out", "error", err)
			}

			err = logger.AddLogObserver(writer, &logger.PlainFormatter{})
			if err != nil {
				log.Warn("cannot add log observer for TermUI", "error", err)
			}
		}()

		appStatusHandler, ok := presenterStatusHandler.(core.AppStatusHandler)
		if ok {
			appStatusHandlers = append(appStatusHandlers, appStatusHandler)
		}
	}

	if len(views) == 0 {
		log.Info("current mode is log-view")
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

// LoadTpsBenchmarkFromStorage will try to load tps benchmark from storage or zero values otherwise
func (shi *statusHandlersInfo) LoadTpsBenchmarkFromStorage(
	store storage.Storer,
	marshalizer marshal.Marshalizer,
) *statistics.TpsPersistentData {
	emptyTpsBenchmarks := &statistics.TpsPersistentData{
		BlockNumber:           0,
		RoundNumber:           0,
		PeakTPS:               0,
		AverageBlockTxCount:   big.NewInt(0),
		TotalProcessedTxCount: big.NewInt(0),
		LastBlockTxCount:      0,
	}
	lastNonceBytes, err := store.Get([]byte(core.LastNonceKeyMetricsStorage))
	if err != nil {
		log.Debug("cannot load last nonce from metrics storage", "error", err, "key", []byte("lastNonce"))
		return emptyTpsBenchmarks
	}

	lastDataList, err := store.Get(lastNonceBytes)
	if err != nil {
		log.Debug("cannot load metrics from storage", "error", err, "key", lastNonceBytes)
		return emptyTpsBenchmarks
	}

	metricsList := &metrics.MetricsList{}
	err = marshalizer.Unmarshal(metricsList, lastDataList)
	if err != nil {
		log.Debug("cannot unmarshal persistent metrics", "error", err)
		return emptyTpsBenchmarks
	}

	metricsMap := metrics.MapFromList(metricsList)

	okTpsBenchmarks := &statistics.TpsPersistentData{}

	okTpsBenchmarks.BlockNumber = persister.GetUint64(metricsMap[core.MetricNonceForTPS])
	okTpsBenchmarks.RoundNumber = persister.GetUint64(metricsMap[core.MetricCurrentRound])
	okTpsBenchmarks.LastBlockTxCount = uint32(persister.GetUint64(metricsMap[core.MetricLastBlockTxCount]))
	okTpsBenchmarks.PeakTPS = float64(persister.GetUint64(metricsMap[core.MetricPeakTPS]))
	okTpsBenchmarks.TotalProcessedTxCount = big.NewInt(int64(persister.GetUint64(metricsMap[core.MetricNumProcessedTxs])))
	okTpsBenchmarks.AverageBlockTxCount = persister.GetBigIntFromString(metricsMap[core.MetricAverageBlockTxCount])

	shi.updateTpsMetrics(metricsMap)

	log.Debug("loaded tps benchmark from storage",
		"block number", okTpsBenchmarks.BlockNumber,
		"round number", okTpsBenchmarks.RoundNumber,
		"peak tps", okTpsBenchmarks.PeakTPS,
		"last block tx count", okTpsBenchmarks.LastBlockTxCount,
		"average block tx count", okTpsBenchmarks.AverageBlockTxCount,
		"total txs processed", okTpsBenchmarks.TotalProcessedTxCount.String())

	return okTpsBenchmarks
}

func (shi *statusHandlersInfo) updateTpsMetrics(metricsMap map[string]interface{}) {
	for key, value := range metricsMap {
		if key == core.MetricAverageBlockTxCount {
			log.Trace("setting metric value", "key", key, "value string", value.(string))
			shi.StatusHandler.SetStringValue(key, value.(string))
			continue
		}
		log.Trace("setting metric value", "key", key, "value uint64", value.(uint64))
		shi.StatusHandler.SetUInt64Value(key, value.(uint64))
	}
}

// CreateStatusHandlerPresenter will return an instance of PresenterStatusHandler
func createStatusHandlerPresenter() view.Presenter {
	presenterStatusHandlerFactory := factoryViews.NewPresenterFactory()

	return presenterStatusHandlerFactory.Create()
}

// CreateViews will start an termui console  and will return an object if cannot create and start termuiConsole
func createViews(presenter view.Presenter, chanStart chan struct{}) ([]factoryViews.Viewer, error) {
	viewsFactory, err := factoryViews.NewViewsFactory(presenter)
	if err != nil {
		return nil, err
	}

	views, err := viewsFactory.Create()
	if err != nil {
		return nil, err
	}

	for _, v := range views {
		err = v.Start(chanStart)
		if err != nil {
			return nil, err
		}
	}

	return views, nil
}
