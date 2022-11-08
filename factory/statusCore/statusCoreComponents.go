package statusCore

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common/statistics"
	"github.com/ElrondNetwork/elrond-go/common/statistics/machine"
	"github.com/ElrondNetwork/elrond-go/config"
	errErd "github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/node/metrics"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/ElrondNetwork/elrond-go/statusHandler/persister"
	trieStatistics "github.com/ElrondNetwork/elrond-go/trie/statistics"
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
	resourceMonitor    factory.ResourceMonitor
	networkStatistics  factory.NetworkStatisticsProvider
	trieSyncStatistics factory.TrieSyncStatisticsProvider
	appStatusHandler   core.AppStatusHandler
	statusMetrics      external.StatusMetricsHandler
	persistentHandler  factory.PersistentStatusHandler
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

	ssc := &statusCoreComponents{
		resourceMonitor:    resourceMonitor,
		networkStatistics:  netStats,
		trieSyncStatistics: trieStatistics.NewTrieSyncStatistics(),
		appStatusHandler:   appStatusHandler,
		statusMetrics:      statusMetrics,
		persistentHandler:  persistentStatusHandler,
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

// Close closes all underlying components
func (scc *statusCoreComponents) Close() error {
	var errNetStats error
	var errResourceMonitor error
	if !check.IfNil(scc.networkStatistics) {
		errNetStats = scc.networkStatistics.Close()
	}
	if !check.IfNil(scc.resourceMonitor) {
		errResourceMonitor = scc.resourceMonitor.Close()
	}
	if !check.IfNil(scc.appStatusHandler) {
		scc.appStatusHandler.Close()
	}

	if errNetStats != nil {
		return errNetStats
	}
	return errResourceMonitor
}
