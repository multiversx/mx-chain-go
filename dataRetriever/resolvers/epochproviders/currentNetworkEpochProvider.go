package epochproviders

import (
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
}

type currentNetworkEpochProvider struct {
	currentEpoch                   uint32
	isSynced                       bool
	mutCurrentEpoch                sync.RWMutex
	requestHandler                 process.RequestHandler
	epochStartMetaBlockInterceptor process.Interceptor
	messenger                      dataRetriever.Messenger
	receivedMetaBlock              *block.MetaBlock
	numActivePersisters            int
}

// NewCurrentNetworkEpochProvider will return a new instance of currentNetworkEpochProvider
func NewCurrentNetworkEpochProvider(args ArgsCurrentNetworkProvider) (*currentNetworkEpochProvider, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &currentNetworkEpochProvider{
		currentEpoch:                   uint32(0),
		isSynced:                       false,
		numActivePersisters:            args.NumActivePersisters,
		requestHandler:                 args.RequestHandler,
		messenger:                      args.Messenger,
		epochStartMetaBlockInterceptor: args.EpochStartMetaBlockInterceptor,
	}, nil
}

// SetNetworkEpochAtBootstrap will update the component's current epoch at bootstrap
func (cnep *currentNetworkEpochProvider) SetNetworkEpochAtBootstrap(epoch uint32) {
	cnep.mutCurrentEpoch.Lock()
	cnep.currentEpoch = epoch
	cnep.mutCurrentEpoch.Unlock()
}

// EpochIsActiveInNetwork returns true if the persister for the given epoch is active in the network
func (cnep *currentNetworkEpochProvider) EpochIsActiveInNetwork(epoch uint32) bool {
	if cnep.isSynced {
		return true
	}

	cnep.mutCurrentEpoch.RLock()
	defer cnep.mutCurrentEpoch.RUnlock()

	isSynced := cnep.isEpochInRange(epoch)
	if !isSynced {
		return false
	}

	// epoch is in range. confirm with the network if it is already synced or in process of syncing
	err := cnep.syncCurrentEpochFromNetwork()
	if err != nil {
		log.Warn("cannot sync current epoch from network", "error", err)
		return false
	}

	if cnep.isEpochInRange(epoch) {
		cnep.isSynced = true
		return true
	}

	return false
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

// SetRequestHandler will update the inner request handler
func (cnep *currentNetworkEpochProvider) SetRequestHandler(rh process.RequestHandler) {
	cnep.requestHandler = rh
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
		if cnep.receivedMetaBlock == nil {
			if numTries > maxNumTriesForFetchingEpoch {
				return ErrCannotGetLatestEpochStartMetaBlock
			}
			time.Sleep(durationBetweenChecks)
			continue
		}

		cnep.currentEpoch = cnep.receivedMetaBlock.Epoch
		return nil

	}
}

func (cnep *currentNetworkEpochProvider) requestEpochStartMetaBlockFromNetwork() error {
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

	cnep.receivedMetaBlock = metaBlock
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
