package latestData

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart/metachain"
	"github.com/multiversx/mx-chain-go/epochStart/shardchain"
	"github.com/multiversx/mx-chain-go/storage"
)

type epochStartRoundLoader struct {
	registryHandler MetaEpochStartTriggerRegistryHandler
}

func newEpochStartRoundLoader() *epochStartRoundLoader {
	return &epochStartRoundLoader{
		registryHandler: metachain.NewMetaTriggerRegistryCreator(),
	}
}

// loadEpochStartRound will return the epoch start round from the bootstrap unit
func (l *epochStartRoundLoader) loadEpochStartRound(
	shardID uint32,
	key []byte,
	storer storage.Storer,
) (uint64, error) {
	trigInternalKey := append([]byte(common.TriggerRegistryKeyPrefix), key...)
	trigData, err := storer.Get(trigInternalKey)
	if err != nil {
		return 0, err
	}

	marshaller := &marshal.GogoProtoMarshalizer{}
	if shardID == core.MetachainShardId {
		state, err := l.registryHandler.UnmarshalTrigger(marshaller, trigData)
		if err != nil {
			return 0, err
		}

		return state.GetCurrEpochStartRound(), nil
	}

	var trigHandler data.TriggerRegistryHandler
	trigHandler, err = shardchain.UnmarshalTrigger(marshaller, trigData)
	if err != nil {
		return 0, err
	}

	return trigHandler.GetEpochStartRound(), nil
}
