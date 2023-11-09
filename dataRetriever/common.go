package dataRetriever

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process/factory"
)

// SetEpochHandlerToHdrResolver sets the epoch handler to the metablock hdr resolver
func SetEpochHandlerToHdrResolver(
	resolversContainer ResolversContainer,
	epochHandler EpochHandler,
) error {
	resolver, err := resolversContainer.Get(factory.MetachainBlocksTopic)
	if err != nil {
		return err
	}

	hdrResolver, ok := resolver.(HeaderResolver)
	if !ok {
		return ErrWrongTypeInContainer
	}

	err = hdrResolver.SetEpochHandler(epochHandler)
	if err != nil {
		return err
	}

	return nil
}

// SetEpochHandlerToHdrRequester sets the epoch handler to the metablock hdr requester
func SetEpochHandlerToHdrRequester(
	requestersContainer RequestersContainer,
	epochHandler EpochHandler,
) error {
	requester, err := requestersContainer.Get(factory.MetachainBlocksTopic)
	if err != nil {
		return err
	}

	hdrRequester, ok := requester.(HeaderRequester)
	if !ok {
		return ErrWrongTypeInContainer
	}

	err = hdrRequester.SetEpochHandler(epochHandler)
	if err != nil {
		return err
	}

	return nil
}

// GetHdrNonceHashDataUnit gets the HdrNonceHashDataUnit by shard
func GetHdrNonceHashDataUnit(shard uint32) UnitType {
	if shard == core.MetachainShardId {
		return MetaHdrNonceHashDataUnit
	}

	return ShardHdrNonceHashDataUnit + UnitType(shard)
}

// GetHeadersDataUnit gets the unit for block headers, by shard
func GetHeadersDataUnit(shard uint32) UnitType {
	if shard == core.MetachainShardId {
		return MetaBlockUnit
	}

	return BlockHeaderUnit
}
