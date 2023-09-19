package requestHandlers

import (
	"fmt"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
)

type sovereignResolverRequestHandler struct {
	*resolverRequestHandler
}

// NewSovereignResolverRequestHandler creates a sovereignRequestHandler interface implementation with request functions
func NewSovereignResolverRequestHandler(resolverRequestHandler *resolverRequestHandler) (*sovereignResolverRequestHandler, error) {
	if resolverRequestHandler == nil {
		return nil, process.ErrNilRequestHandler
	}

	srrh := &sovereignResolverRequestHandler{
		resolverRequestHandler,
	}

	return srrh, nil
}

// RequestExtendedShardHeaderByNonce method asks for extended shard header from the connected peers by nonce
func (srrh *sovereignResolverRequestHandler) RequestExtendedShardHeaderByNonce(nonce uint64) {
	suffix := fmt.Sprintf("%s_%d", uniqueHeadersSuffix, srrh.shardID)
	key := []byte(fmt.Sprintf("%d-%d", srrh.shardID, nonce))
	if !srrh.testIfRequestIsNeeded(key, suffix) {
		return
	}

	log.Debug("RequestExtendedShardHeaderByNonce.getExtendedShardHeaderRequester: requesting extended shard header by nonce from network",
		"shard", srrh.shardID,
		"nonce", nonce,
	)

	requester, err := srrh.getExtendedShardHeaderRequester()
	if err != nil {
		log.Error("RequestExtendedShardHeaderByNonce.getExtendedShardHeaderRequester",
			"error", err.Error(),
			"shard", srrh.shardID,
		)
		return
	}

	headerRequester, ok := requester.(NonceRequester)
	if !ok {
		log.Error("sovereignResolverRequestHandler.RequestExtendedShardHeaderByNonce: wrong assertion type when creating header requester")
		return
	}

	srrh.whiteList.Add([][]byte{key})

	epoch := srrh.getEpoch()
	err = headerRequester.RequestDataFromNonce(nonce, epoch)
	if err != nil {
		log.Debug("RequestExtendedShardHeaderByNonce.RequestDataFromNonce",
			"error", err.Error(),
			"epoch", epoch,
			"nonce", nonce,
		)
		return
	}

	srrh.addRequestedItems([][]byte{key}, suffix)
}

func (srrh *sovereignResolverRequestHandler) getExtendedShardHeaderRequester() (dataRetriever.Requester, error) {
	headerRequester, err := srrh.requestersFinder.IntraShardRequester(factory.ExtendedHeaderProofTopic)
	if err != nil {
		log.Warn("extended header proof container not found, available requesters in container",
			"requesters", srrh.requestersFinder.RequesterKeys(),
		)
		return nil, fmt.Errorf("%w, topic: %s", err, factory.ExtendedHeaderProofTopic)
	}

	return headerRequester, nil
}

// RequestExtendedShardHeader method asks for extended shard header from the connected peers by nonce
func (srrh *sovereignResolverRequestHandler) RequestExtendedShardHeader(hash []byte) {
	suffix := fmt.Sprintf("%s_%d", uniqueHeadersSuffix, srrh.shardID)
	if !srrh.testIfRequestIsNeeded(hash, suffix) {
		return
	}

	log.Debug("sovereignResolverRequestHandler.RequestExtendedShardHeader: requesting extended shard header from network by hash",
		"shard", srrh.shardID,
		"hash", hash,
	)

	headerRequester, err := srrh.getExtendedShardHeaderRequester()
	if err != nil {
		log.Error("RequestExtendedShardHeader.getExtendedShardHeaderRequester",
			"error", err.Error(),
			"shard", srrh.shardID,
		)
		return
	}

	srrh.whiteList.Add([][]byte{hash})

	epoch := srrh.getEpoch()
	err = headerRequester.RequestDataFromHash(hash, epoch)
	if err != nil {
		log.Debug("RequestExtendedShardHeader.RequestDataFromHash",
			"error", err.Error(),
			"epoch", epoch,
			"hash", hash,
		)
		return
	}

	srrh.addRequestedItems([][]byte{hash}, suffix)
}
