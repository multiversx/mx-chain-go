package requesters

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("requesters")

// ArgBaseRequester is the argument structure used as base to create a new requester instance
type ArgBaseRequester struct {
	RequestSender dataRetriever.TopicRequestSender
	Marshaller    marshal.Marshalizer
}

type baseRequester struct {
	dataRetriever.TopicRequestSender
	marshaller marshal.Marshalizer
}

func createBaseRequester(args ArgBaseRequester) *baseRequester {
	return &baseRequester{
		TopicRequestSender: args.RequestSender,
		marshaller:         args.Marshaller,
	}
}

func checkArgBase(arg ArgBaseRequester) error {
	if check.IfNil(arg.RequestSender) {
		return dataRetriever.ErrNilRequestSender
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
	requester.TopicRequestSender.SetNumPeersToQuery(intra, cross)
}

// NumPeersToQuery returns the number of intra shard and cross shard number of peer to query
func (requester *baseRequester) NumPeersToQuery() (int, int) {
	return requester.TopicRequestSender.NumPeersToQuery()
}

// SetDebugHandler sets a debug requester
func (requester *baseRequester) SetDebugHandler(debugHandler dataRetriever.DebugHandler) error {
	return requester.TopicRequestSender.SetDebugHandler(debugHandler)
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
