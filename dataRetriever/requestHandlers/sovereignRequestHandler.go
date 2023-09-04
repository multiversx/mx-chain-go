package requestHandlers

import (
	"encoding/hex"
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

// TODO: Implement this in MX-14496

// RequestExtendedShardHeaderByNonce method asks for extended shard header from the connected peers by nonce
func (srrh *sovereignResolverRequestHandler) RequestExtendedShardHeaderByNonce(nonce uint64) {
	log.Error("RequestExtendedShardHeaderByNonce", "nonce", nonce)

	if nonce < 11 {
		log.Error("RequestExtendedShardHeaderByNonce REJECTED", "nonce", nonce)
		return
	}

	suffix := fmt.Sprintf("%s_%d", uniqueHeadersSuffix, srrh.shardID)
	key := []byte(fmt.Sprintf("%d-%d", srrh.shardID, nonce))
	if !srrh.testIfRequestIsNeeded(key, suffix) {
		return
	}

	log.Debug("requesting shard header by nonce from network",
		"shard", srrh.shardID,
		"nonce", nonce,
	)

	requester, err := srrh.getShardHeaderRequester(srrh.shardID)
	if err != nil {
		log.Error("RequestShardHeaderByNonce.getShardHeaderRequester",
			"error", err.Error(),
			"shard", srrh.shardID,
		)
		return
	}

	headerRequester, ok := requester.(NonceRequester)
	if !ok {
		log.Warn("wrong assertion type when creating header requester")
		return
	}

	srrh.whiteList.Add([][]byte{key})

	epoch := srrh.getEpoch()
	err = headerRequester.RequestDataFromNonce(nonce, epoch)
	if err != nil {
		log.Debug("RequestShardHeaderByNonce.RequestDataFromNonce",
			"error", err.Error(),
			"epoch", epoch,
			"nonce", nonce,
		)
		return
	}

	srrh.addRequestedItems([][]byte{key}, suffix)
}

func (srrh *sovereignResolverRequestHandler) getShardHeaderRequester(shardID uint32) (dataRetriever.Requester, error) {

	headerRequester, err := srrh.requestersFinder.CrossShardRequester(factory.ExtendedHeaderProofTopic, srrh.shardID) // CrossShardRequester(factory.ShardBlocksTopic, shardID)
	if err != nil {
		err = fmt.Errorf("%w, topic: %s, current shard ID: %d, cross shard ID: %d",
			err, factory.ExtendedHeaderProofTopic, srrh.shardID, shardID)

		log.Warn("available requesters in container",
			"requesters", srrh.requestersFinder.RequesterKeys(),
		)
		return nil, err
	}

	return headerRequester, nil
}

// RequestExtendedShardHeader method asks for extended shard header from the connected peers by nonce
func (srrh *sovereignResolverRequestHandler) RequestExtendedShardHeader(hash []byte) {
	//TODO: This method should be implemented for sovereign chain
	log.Error("sovereignResolverRequestHandler.RequestExtendedShardHeader", "hash", hex.EncodeToString(hash))

	suffix := fmt.Sprintf("%s_%d", uniqueHeadersSuffix, srrh.shardID)
	if !srrh.testIfRequestIsNeeded(hash, suffix) {
		return
	}

	log.Debug("requesting shard header from network",
		"shard", srrh.shardID,
		"hash", hash,
	)

	headerRequester, err := srrh.getShardHeaderRequester(srrh.shardID)
	if err != nil {
		log.Error("RequestShardHeader.getShardHeaderRequester",
			"error", err.Error(),
			"shard", srrh.shardID,
		)
		return
	}

	srrh.whiteList.Add([][]byte{hash})

	epoch := srrh.getEpoch()
	err = headerRequester.RequestDataFromHash(hash, epoch)
	if err != nil {
		log.Debug("RequestShardHeader.RequestDataFromHash",
			"error", err.Error(),
			"epoch", epoch,
			"hash", hash,
		)
		return
	}

	srrh.addRequestedItems([][]byte{hash}, suffix)
}
