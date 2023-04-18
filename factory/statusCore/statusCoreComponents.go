package statusCore

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common/statistics"
	"github.com/multiversx/mx-chain-go/common/statistics/machine"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/debug/scquery"
	errErd "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/disabled"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/node/metrics"
	"github.com/multiversx/mx-chain-go/statusHandler"
	"github.com/multiversx/mx-chain-go/statusHandler/persister"
	trieStatistics "github.com/multiversx/mx-chain-go/trie/statistics"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("factory")

// StatusCoreComponentsFactoryArgs holds the arguments needed for creating a status core components factory
type StatusCoreComponentsFactoryArgs struct {
	Config          config.Config
	EpochConfig     config.EpochConfig
	RoundConfig     config.RoundConfig
	RatingsConfig   config.RatingsConfig
	EconomicsConfig config.EconomicsConfig
	CoreComp        factory.CoreComponentsHolder
}

// statusCoreComponentsFactory is responsible for creating the status core components
type statusCoreComponentsFactory struct {
	config          config.Config
	epochConfig     config.EpochConfig
	roundConfig     config.RoundConfig
	ratingsConfig   config.RatingsConfig
	economicsConfig config.EconomicsConfig
	coreComp        factory.CoreComponentsHolder
}

// statusCoreComponents is the DTO used for core components
type statusCoreComponents struct {
	resourceMonitor        factory.ResourceMonitor
	networkStatistics      factory.NetworkStatisticsProvider
	trieSyncStatistics     factory.TrieSyncStatisticsProvider
	appStatusHandler       core.AppStatusHandler
	statusMetrics          external.StatusMetricsHandler
	persistentHandler      factory.PersistentStatusHandler
	scQueryServiceDebugger factory.SCQueryServiceDebugger
}

// NewStatusCoreComponentsFactory initializes the factory which is responsible to creating status core components
func NewStatusCoreComponentsFactory(args StatusCoreComponentsFactoryArgs) (*statusCoreComponentsFactory, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &statusCoreComponentsFactory{
		config:          args.Config,
		epochConfig:     args.EpochConfig,
		roundConfig:     args.RoundConfig,
		ratingsConfig:   args.RatingsConfig,
		economicsConfig: args.EconomicsConfig,
		coreComp:        args.CoreComp,
	}, nil
}

func checkArgs(args StatusCoreComponentsFactoryArgs) error {
	if check.IfNil(args.CoreComp) {
		return errErd.ErrNilCoreComponents
	}
	if check.IfNil(args.CoreComp.EconomicsData()) {
		return errErd.ErrNilEconomicsData
	}
	if check.IfNil(args.CoreComp.GenesisNodesSetup()) {
		return errErd.ErrNilGenesisNodesSetupHandler
	}
	if check.IfNil(args.CoreComp.InternalMarshalizer()) {
		return errErd.ErrNilMarshalizer
	}
	if check.IfNil(args.CoreComp.Uint64ByteSliceConverter()) {
		return errErd.ErrNilUint64ByteSliceConverter
	}

	return nil
}

// Create creates the status core components
func (sccf *statusCoreComponentsFactory) Create() (*statusCoreComponents, error) {
	netStats := machine.NewNetStatistics()

	resourceMonitor, err := statistics.NewResourceMonitor(sccf.config, netStats)
	if err != nil {
		return nil, err
	}

	if sccf.config.ResourceStats.Enabled {
		resourceMonitor.StartMonitoring()
	}

	appStatusHandler, statusMetrics, persistentStatusHandler, err := sccf.createStatusHandler()
	if err != nil {
		return nil, err
	}

	scQueryServiceDebugger, err := createSCQueryServiceDebugger(sccf.config)
	if err != nil {
		return nil, err
	}

	ssc := &statusCoreComponents{
		resourceMonitor:        resourceMonitor,
		networkStatistics:      netStats,
		trieSyncStatistics:     trieStatistics.NewTrieSyncStatistics(),
		appStatusHandler:       appStatusHandler,
		statusMetrics:          statusMetrics,
		persistentHandler:      persistentStatusHandler,
		scQueryServiceDebugger: scQueryServiceDebugger,
	}

	return ssc, nil
}

func (sccf *statusCoreComponentsFactory) createStatusHandler() (core.AppStatusHandler, external.StatusMetricsHandler, factory.PersistentStatusHandler, error) {
	var appStatusHandlers []core.AppStatusHandler
	var handler core.AppStatusHandler
	statusMetrics := statusHandler.NewStatusMetrics()
	appStatusHandlers = append(appStatusHandlers, statusMetrics)

	persistentHandler, err := persister.NewPersistentStatusHandler(sccf.coreComp.InternalMarshalizer(), sccf.coreComp.Uint64ByteSliceConverter())
	if err != nil {
		return nil, nil, nil, err
	}
	appStatusHandlers = append(appStatusHandlers, persistentHandler)
	if len(appStatusHandlers) > 0 {
		handler, err = statusHandler.NewAppStatusFacadeWithHandlers(appStatusHandlers...)
		if err != nil {
			log.Warn("cannot init AppStatusFacade, will start with NilStatusHandler", "error", err)
			handler = statusHandler.NewNilStatusHandler()
		}
	} else {
		handler = statusHandler.NewNilStatusHandler()
		log.Debug("no AppStatusHandler used: will start with NilStatusHandler")
	}

	err = metrics.InitBaseMetrics(handler)
	if err != nil {
		return nil, nil, nil, err
	}

	err = metrics.InitConfigMetrics(handler, sccf.epochConfig, sccf.economicsConfig, sccf.coreComp.GenesisNodesSetup())
	if err != nil {
		return nil, nil, nil, err
	}

	err = metrics.InitRatingsMetrics(handler, sccf.ratingsConfig)
	if err != nil {
		return nil, nil, nil, err
	}

	err = sccf.coreComp.EconomicsData().SetStatusHandler(handler)
	if err != nil {
		log.Debug("cannot set status handler to economicsData", "error", err)
		return nil, nil, nil, err
	}

	return handler, statusMetrics, persistentHandler, nil
}

func createSCQueryServiceDebugger(config config.Config) (factory.SCQueryServiceDebugger, error) {
	if !config.Debug.SCQueryService.Enabled {
		return disabled.NewDisabledSCQueryServiceDebugger(), nil
	}

	args := scquery.ArgsSCQueryDebugger{
		IntervalAutoPrintInSeconds: config.Debug.SCQueryService.IntervalAutoPrintInSeconds,
		LoggerInstance:             logger.GetOrCreate("debug/scquery"),
	}
	return scquery.NewSCQueryDebugger(args)
}

// Close closes all underlying components
func (scc *statusCoreComponents) Close() error {
	var err error
	if !check.IfNil(scc.networkStatistics) {
		errLocal := scc.networkStatistics.Close()
		if errLocal != nil {
			err = fmt.Errorf("%w when closing networkStatistics component", errLocal)
		}
	}
	if !check.IfNil(scc.resourceMonitor) {
		errLocal := scc.resourceMonitor.Close()
		if errLocal != nil {
			err = fmt.Errorf("%w when closing resourceMonitor component", errLocal)
		}
	}
	if !check.IfNil(scc.scQueryServiceDebugger) {
		errLocal := scc.scQueryServiceDebugger.Close()
		if errLocal != nil {
			err = fmt.Errorf("%w when closing scQueryServiceDebugger component", errLocal)
		}
	}
	if !check.IfNil(scc.appStatusHandler) {
		scc.appStatusHandler.Close()
	}

	return err
}
