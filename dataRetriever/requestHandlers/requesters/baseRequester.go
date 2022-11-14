package requesters

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

var log = logger.GetOrCreate("requesters")

// ArgBaseRequester is the argument structure used as base to create a new requester instance
type ArgBaseRequester struct {
	SenderResolver dataRetriever.TopicResolverSender
	Marshaller     marshal.Marshalizer
}

type baseRequester struct {
	dataRetriever.TopicResolverSender
	marshaller marshal.Marshalizer
}

func createBaseRequester(args ArgBaseRequester) *baseRequester {
	return &baseRequester{
		TopicResolverSender: args.SenderResolver,
		marshaller:          args.Marshaller,
	}
}

func checkArgBase(arg ArgBaseRequester) error {
	if check.IfNil(arg.SenderResolver) {
		return dataRetriever.ErrNilResolverSender
	}
	if check.IfNil(arg.Marshaller) {
		return dataRetriever.ErrNilMarshalizer
	}
	return nil
}

// RequestDataFromHash requests data from other peers by having a hash and the epoch as input
func (requester *baseRequester) RequestDataFromHash(hash []byte, epoch uint32) error {
	return requester.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashType,
			Value: hash,
			Epoch: epoch,
		},
		[][]byte{hash},
	)
}

// SetNumPeersToQuery sets the number of intra shard and cross shard number of peer to query
func (requester *baseRequester) SetNumPeersToQuery(intra int, cross int) {
	requester.TopicResolverSender.SetNumPeersToQuery(intra, cross)
}

// NumPeersToQuery returns the number of intra shard and cross shard number of peer to query
func (requester *baseRequester) NumPeersToQuery() (int, int) {
	return requester.TopicResolverSender.NumPeersToQuery()
}

// SetResolverDebugHandler sets a resolver debug requester
func (requester *baseRequester) SetResolverDebugHandler(debugHandler dataRetriever.ResolverDebugHandler) error {
	return requester.TopicResolverSender.SetResolverDebugHandler(debugHandler)
}

func (requester *baseRequester) requestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	requester.printHashArray(hashes)

	b := &batch.Batch{
		Data: hashes,
	}
	batchBytes, err := requester.marshaller.Marshal(b)
	if err != nil {
		return err
	}

	return requester.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashArrayType,
			Value: batchBytes,
			Epoch: epoch,
		},
		hashes,
	)
}

func (requester *baseRequester) printHashArray(hashes [][]byte) {
	if log.GetLevel() > logger.LogTrace {
		return
	}

	for _, hash := range hashes {
		log.Trace("baseRequester.requestDataFromHashArray", "hash", hash, "topic", requester.RequestTopic())
	}
}
