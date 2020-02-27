package bootstrap

import (
	"fmt"
	"math"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
)

var log = logger.GetOrCreate("registration")
var _ process.Interceptor = (*simpleMetaBlockInterceptor)(nil)

const requestSuffix = "_REQUEST"

type metaBlockInterceptorHandler interface {
	process.Interceptor
	GetAllReceivedMetaBlocks() []block.MetaBlock
}

type shardHeaderInterceptorHandler interface {
	process.Interceptor
	GetAllReceivedShardHeaders() []block.ShardData
}

type EpochStartDataProvider struct {
	marshalizer            marshal.Marshalizer
	messenger              p2p.Messenger
	metaBlockInterceptor   metaBlockInterceptorHandler
	shardHeaderInterceptor shardHeaderInterceptorHandler
}

func NewEpochStartDataProvider(messenger p2p.Messenger, marshalizer marshal.Marshalizer) *EpochStartDataProvider {
	metaBlockInterceptor := NewSimpleMetaBlockInterceptor(marshalizer)
	shardHdrInterceptor := NewSimpleShardHeaderInterceptor(marshalizer)
	return &EpochStartDataProvider{
		marshalizer:            marshalizer,
		messenger:              messenger,
		metaBlockInterceptor:   metaBlockInterceptor,
		shardHeaderInterceptor: shardHdrInterceptor,
	}
}

func (ser *EpochStartDataProvider) RequestEpochStartMetaBlock(epoch uint32) (*block.MetaBlock, error) {
	err := ser.messenger.CreateTopic(factory.MetachainBlocksTopic+requestSuffix, false)
	if err != nil {
		return nil, err
	}

	err = ser.messenger.CreateTopic(factory.MetachainBlocksTopic, false)
	if err != nil {
		return nil, err
	}

	err = ser.messenger.RegisterMessageProcessor(factory.MetachainBlocksTopic, ser.metaBlockInterceptor)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = ser.messenger.UnregisterMessageProcessor(factory.MetachainBlocksTopic)
		if err != nil {
			log.Info("error unregistering message processor", "error", err)
		}
	}()

	err = ser.requestMetaBlock()
	if err != nil {
		return nil, err
	}

	// TODO: check if received block is correct by receiving the same block in majority
	threshold := 1
	for {
		if len(ser.metaBlockInterceptor.GetAllReceivedMetaBlocks()) >= threshold {
			break
		}

		time.Sleep(time.Second)
	}

	return &ser.metaBlockInterceptor.GetAllReceivedMetaBlocks()[0], nil
}

func (ser *EpochStartDataProvider) requestMetaBlock() error {
	rd := dataRetriever.RequestData{
		Type:  dataRetriever.EpochType,
		Epoch: 0,
		Value: []byte(fmt.Sprintf("epochStartBlock_%d", math.MaxUint32)),
	}
	rdBytes, err := ser.marshalizer.Marshal(rd)
	if err != nil {
		return err
	}

	ser.messenger.Broadcast(factory.MetachainBlocksTopic+requestSuffix, rdBytes)
	return nil
}
