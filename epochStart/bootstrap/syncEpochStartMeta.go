package bootstrap

import (
	"math"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process/factory"
)

type epochStartMetaSyncer struct {
	requestHandler                 epochStart.RequestHandler
	messenger                      p2p.Messenger
	epochStartMetaBlockInterceptor EpochStartInterceptor
	marshalizer                    marshal.Marshalizer
	hasher                         hashing.Hasher
}

// ArgsNewEpochStartMetaSyncer -
type ArgsNewEpochStartMetaSyncer struct {
	RequestHandler epochStart.RequestHandler
	Messenger      p2p.Messenger
	Marshalizer    marshal.Marshalizer
	Hasher         hashing.Hasher
}

const delayBetweenRequests = 1 * time.Second
const thresholdForConsideringMetaBlockCorrect = 0.2
const numRequestsToSendOnce = 4
const maxNumTimesToRetry = 100

func NewEpochStartMetaSyncer(args ArgsNewEpochStartMetaSyncer) (*epochStartMetaSyncer, error) {
	e := &epochStartMetaSyncer{
		requestHandler: args.RequestHandler,
		messenger:      args.Messenger,
		marshalizer:    args.Marshalizer,
		hasher:         args.Hasher,
	}

	var err error
	e.epochStartMetaBlockInterceptor, err = NewSimpleEpochStartMetaBlockInterceptor(e.marshalizer, e.hasher)
	if err != nil {
		return nil, err
	}

	return e, nil
}

// SyncEpochStartMeta syncs the latest epoch start metablock
func (e *epochStartMetaSyncer) SyncEpochStartMeta(_ time.Duration) (*block.MetaBlock, error) {
	err := e.initTopicForEpochStartMetaBlockInterceptor()
	if err != nil {
		return nil, err
	}
	defer func() {
		e.resetTopicsAndInterceptors()
	}()

	e.requestEpochStartMetaBlock()

	unknownEpoch := uint32(math.MaxUint32)
	count := 0
	for {
		if count > maxNumTimesToRetry {
			return nil, epochStart.ErrNumTriesExceeded
		}

		count++
		numConnectedPeers := len(e.messenger.ConnectedPeers())
		threshold := int(thresholdForConsideringMetaBlockCorrect * float64(numConnectedPeers))

		mb, errConsensusNotReached := e.epochStartMetaBlockInterceptor.GetEpochStartMetaBlock(threshold, unknownEpoch)
		if errConsensusNotReached == nil {
			return mb, nil
		}

		log.Info("consensus not reached for meta block. re-requesting and trying again...")
		e.requestEpochStartMetaBlock()
	}
}

func (e *epochStartMetaSyncer) requestEpochStartMetaBlock() {
	// send more requests
	unknownEpoch := uint32(math.MaxUint32)
	for i := 0; i < numRequestsToSendOnce; i++ {
		time.Sleep(delayBetweenRequests)
		e.requestHandler.RequestStartOfEpochMetaBlock(unknownEpoch)
	}
}

func (e *epochStartMetaSyncer) resetTopicsAndInterceptors() {
	err := e.messenger.UnregisterMessageProcessor(factory.MetachainBlocksTopic)
	if err != nil {
		log.Info("error unregistering message processors", "error", err)
	}
}

func (e *epochStartMetaSyncer) initTopicForEpochStartMetaBlockInterceptor() error {
	err := e.messenger.UnregisterMessageProcessor(factory.MetachainBlocksTopic)
	if err != nil {
		log.Info("error unregistering message processor", "error", err)
		return err
	}

	err = e.messenger.CreateTopic(factory.MetachainBlocksTopic, true)
	if err != nil {
		log.Info("error registering message processor", "error", err)
		return err
	}

	err = e.messenger.RegisterMessageProcessor(factory.MetachainBlocksTopic, e.epochStartMetaBlockInterceptor)
	if err != nil {
		return err
	}

	return nil
}

// IsInterfaceNil returns true if underlying object is nil
func (e *epochStartMetaSyncer) IsInterfaceNil() bool {
	return e == nil
}
