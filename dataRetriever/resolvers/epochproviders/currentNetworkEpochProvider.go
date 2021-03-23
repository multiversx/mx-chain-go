package epochproviders

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
)

var log = logger.GetOrCreate("resolvers/epochproviders")

const durationBetweenChecks = 200 * time.Millisecond
const maxNumTriesForFetchingEpoch = 50
const metaBlocksTopicIdentifier = "network epoch"

// ArgsCurrentNetworkProvider holds the arguments needed for creating a new currentNetworkEpochProvider
type ArgsCurrentNetworkProvider struct {
	RequestHandler                 process.RequestHandler
	Messenger                      dataRetriever.Messenger
	EpochStartMetaBlockInterceptor process.Interceptor
	NumActivePersisters            int
	DurationBetweenMetablockChecks time.Duration
}

type currentNetworkEpochProvider struct {
	mutLastMetablockCheckpoint     sync.RWMutex
	lastMetablockCheckpoint        time.Time
	durationBetweenMetablockChecks time.Duration
	currentEpoch                   uint32
	mutCurrentEpoch                sync.RWMutex
	requestHandler                 process.RequestHandler
	epochStartMetaBlockInterceptor process.Interceptor
	messenger                      dataRetriever.Messenger
	mutReceivedMetablock           sync.RWMutex
	receivedMetaBlock              *block.MetaBlock
	numActivePersisters            int
	ctx                            context.Context
	cancelFunc                     func()
}

// NewCurrentNetworkEpochProvider will return a new instance of currentNetworkEpochProvider
//TODO: refactor the code from this file - for now it is not used: remove possible deadlocks due to the locking mutexes
// on possible infinite running functions, L176: the set instruction requires locking on the mutex
func NewCurrentNetworkEpochProvider(args ArgsCurrentNetworkProvider) (*currentNetworkEpochProvider, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	return &currentNetworkEpochProvider{
		currentEpoch:                   uint32(0),
		numActivePersisters:            args.NumActivePersisters,
		requestHandler:                 args.RequestHandler,
		messenger:                      args.Messenger,
		epochStartMetaBlockInterceptor: args.EpochStartMetaBlockInterceptor,
		ctx:                            ctx,
		cancelFunc:                     cancelFunc,
		durationBetweenMetablockChecks: args.DurationBetweenMetablockChecks,
	}, nil
}

// SetNetworkEpochAtBootstrap will update the component's current epoch at bootstrap
func (cnep *currentNetworkEpochProvider) SetNetworkEpochAtBootstrap(epoch uint32) {
	cnep.updateCurrentEpoch(epoch)
	cnep.updateLastMetablockCheckpoint()
}

func (cnep *currentNetworkEpochProvider) updateCurrentEpoch(epoch uint32) {
	cnep.mutCurrentEpoch.Lock()
	cnep.currentEpoch = epoch
	cnep.mutCurrentEpoch.Unlock()
}

// EpochIsActiveInNetwork returns true if the persister for the given epoch is active in the network
func (cnep *currentNetworkEpochProvider) EpochIsActiveInNetwork(epoch uint32) bool {
	cnep.mutCurrentEpoch.RLock()
	defer cnep.mutCurrentEpoch.RUnlock()

	if cnep.shouldCheckCurrentMetablockOnNetwork() {
		err := cnep.syncCurrentEpochFromNetwork()
		if err != nil {
			log.Warn("cannot sync current epoch from network", "error", err)
			return false
		}
		//update only if the operation succeeded
		cnep.updateLastMetablockCheckpoint()
	}

	return cnep.isEpochInRange(epoch)
}

func (cnep *currentNetworkEpochProvider) shouldCheckCurrentMetablockOnNetwork() bool {
	cnep.mutLastMetablockCheckpoint.RLock()
	checkpoint := cnep.lastMetablockCheckpoint.Add(cnep.durationBetweenMetablockChecks)
	cnep.mutLastMetablockCheckpoint.RUnlock()

	return checkpoint.Unix() < time.Now().Unix()
}

func (cnep *currentNetworkEpochProvider) updateLastMetablockCheckpoint() {
	cnep.mutLastMetablockCheckpoint.Lock()
	cnep.lastMetablockCheckpoint = time.Now()
	cnep.mutLastMetablockCheckpoint.Unlock()
}

func (cnep *currentNetworkEpochProvider) isEpochInRange(epoch uint32) bool {
	lower := core.MaxInt(int(cnep.currentEpoch)-cnep.numActivePersisters+1, 0)
	upper := cnep.currentEpoch

	return epoch >= uint32(lower) && epoch <= upper
}

// CurrentEpoch returns the current epoch in the network
func (cnep *currentNetworkEpochProvider) CurrentEpoch() uint32 {
	cnep.mutCurrentEpoch.RLock()
	defer cnep.mutCurrentEpoch.RUnlock()

	return cnep.currentEpoch
}

func (cnep *currentNetworkEpochProvider) syncCurrentEpochFromNetwork() error {
	cnep.epochStartMetaBlockInterceptor.RegisterHandler(cnep.handlerEpochStartMetaBlock)

	err := cnep.messenger.RegisterMessageProcessor(factory.MetachainBlocksTopic, metaBlocksTopicIdentifier, cnep.epochStartMetaBlockInterceptor)
	if err != nil {
		return err
	}

	defer func() {
		err = cnep.messenger.UnregisterMessageProcessor(factory.MetachainBlocksTopic, metaBlocksTopicIdentifier)
		log.LogIfError(err)
	}()

	err = cnep.requestEpochStartMetaBlockFromNetwork()
	if err != nil {
		return err
	}

	numTries := 0
	for {
		numTries++
		cnep.mutReceivedMetablock.RLock()
		currentMetablock := cnep.receivedMetaBlock
		cnep.mutReceivedMetablock.RUnlock()

		if currentMetablock == nil {
			if numTries > maxNumTriesForFetchingEpoch {
				return ErrCannotGetLatestEpochStartMetaBlock
			}

			select {
			case <-time.After(durationBetweenChecks):
			case <-cnep.ctx.Done():
				return ErrComponentClosing
			}

			err = cnep.requestEpochStartMetaBlockFromNetwork()
			if err != nil {
				return err
			}

			continue
		}

		cnep.currentEpoch = currentMetablock.Epoch
		return nil
	}
}

func (cnep *currentNetworkEpochProvider) requestEpochStartMetaBlockFromNetwork() error {
	log.Debug("currentNetworkEpochProvider requesting current start of epoch metablock from network")

	originalIntra, originalCross, err := cnep.requestHandler.GetNumPeersToQuery(factory.MetachainBlocksTopic)
	if err != nil {
		return err
	}

	numConnectedPeers := len(cnep.messenger.ConnectedPeers())
	err = cnep.requestHandler.SetNumPeersToQuery(factory.MetachainBlocksTopic, numConnectedPeers, numConnectedPeers)
	if err != nil {
		return err
	}

	defer func() {
		err = cnep.requestHandler.SetNumPeersToQuery(factory.MetachainBlocksTopic, originalIntra, originalCross)
		if err != nil {
			log.Warn("currentNetworkEpochProvider: error setting num of peers intra/cross for resolver",
				"resolver", factory.MetachainBlocksTopic,
				"error", err)
		}
	}()

	cnep.requestHandler.RequestStartOfEpochMetaBlock(uint32(math.MaxUint32))

	return nil
}

func (cnep *currentNetworkEpochProvider) handlerEpochStartMetaBlock(_ string, _ []byte, data interface{}) {
	metaBlock, ok := data.(*block.MetaBlock)
	if !ok {
		log.Warn("currentNetworkEpochProvider-handlerEpochStartMetaBlock: cannot cast received epoch start meta block")
		return
	}
	if !metaBlock.IsStartOfEpochBlock() {
		log.Warn("received meta block is not of type epoch start meta block")
		return
	}

	cnep.mutReceivedMetablock.Lock()
	cnep.receivedMetaBlock = metaBlock
	cnep.mutReceivedMetablock.Unlock()
}

// Close will interrupt any check current epoch processes
func (cnep *currentNetworkEpochProvider) Close() error {
	cnep.cancelFunc()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (cnep *currentNetworkEpochProvider) IsInterfaceNil() bool {
	return cnep == nil
}

func checkArgs(args ArgsCurrentNetworkProvider) error {
	if check.IfNil(args.RequestHandler) {
		return wrapArgsError(ErrNilRequestHandler)
	}
	if check.IfNil(args.Messenger) {
		return wrapArgsError(ErrNilMessenger)
	}
	if check.IfNil(args.EpochStartMetaBlockInterceptor) {
		return wrapArgsError(ErrNilEpochStartMetaBlockInterceptor)
	}

	return nil
}

func wrapArgsError(err error) error {
	return fmt.Errorf("%w when creating currentNetworkEpochProvider", err)
}
