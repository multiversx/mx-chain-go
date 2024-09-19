package bootstrap

import (
	"context"
	"math"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	dataCore "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process/factory"
)

type epochStartSovereignBlockProcessor struct {
	*epochStartMetaBlockProcessor
}

// GetEpochStartMetaBlock will return the sovereign block after it is confirmed or an error if the number of tries was exceeded
// This is a blocking method which will end after the consensus for the sovereign block is obtained or the context is done
func (e *epochStartSovereignBlockProcessor) GetEpochStartMetaBlock(ctx context.Context) (dataCore.MetaHeaderHandler, error) {
	topic := getTopic()
	originalIntra, _, err := e.requestHandler.GetNumPeersToQuery(topic)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = e.requestHandler.SetNumPeersToQuery(topic, originalIntra, 0)
		if err != nil {
			log.Warn("sovereign epoch bootstrapper: error setting num of peers intra/cross for resolver",
				"resolver", factory.ShardBlocksTopic,
				"error", err)
		}
	}()

	err = e.requestMetaBlock()
	if err != nil {
		return nil, err
	}

	chanRequests := time.After(durationBetweenReRequests)
	chanCheckMaps := time.After(durationBetweenChecks)
	for {
		select {
		case <-e.chanConsensusReached:
			return e.metaBlock, nil
		case <-ctx.Done():
			return e.getMostReceivedMetaBlock()
		case <-chanRequests:
			err = e.requestMetaBlock()
			if err != nil {
				return nil, err
			}
			chanRequests = time.After(durationBetweenReRequests)
		case <-chanCheckMaps:
			e.checkMaps()
			chanCheckMaps = time.After(durationBetweenChecks)
		}
	}
}

func (e *epochStartSovereignBlockProcessor) requestMetaBlock() error {
	numConnectedPeers := len(e.messenger.ConnectedPeers())
	err := e.requestHandler.SetNumPeersToQuery(getTopic(), numConnectedPeers, numConnectedPeers)
	if err != nil {
		return err
	}

	unknownEpoch := uint32(math.MaxUint32)
	e.requestHandler.RequestStartOfEpochMetaBlock(unknownEpoch)
	return nil
}

func getTopic() string {
	return factory.ShardBlocksTopic + core.CommunicationIdentifierBetweenShards(core.SovereignChainShardId, core.SovereignChainShardId)
}
