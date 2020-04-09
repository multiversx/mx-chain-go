package bootstrap

import (
	"bytes"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func TestPrepareEpochFromStorage(t *testing.T) {
	args := createMockEpochStartBootstrapArgs()
	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.initializeFromLocalStorage()

	_, err := epochStartProvider.prepareEpochFromStorage()
	assert.Error(t, err)
}

func TestGetEpochStartMetaFromStorage(t *testing.T) {
	args := createMockEpochStartBootstrapArgs()
	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.initializeFromLocalStorage()

	meta := &block.MetaBlock{Nonce: 1}
	storer := &mock.StorerStub{
		GetCalled: func(key []byte) (bytes []byte, err error) {
			metaBytes, _ := json.Marshal(meta)
			return metaBytes, nil
		},
	}
	metaBlock, err := epochStartProvider.getEpochStartMetaFromStorage(storer)
	assert.Nil(t, err)
	assert.Equal(t, meta, metaBlock)
}

func TestGetLastBootstrapData(t *testing.T) {
	args := createMockEpochStartBootstrapArgs()
	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.initializeFromLocalStorage()

	round := int64(10)

	roundNum := bootstrapStorage.RoundNum{
		Num: round,
	}
	roundBytes, _ := json.Marshal(&roundNum)
	nodesCoordinatorConfigKey := []byte("key")

	nodesConfigRegistry := sharding.NodesCoordinatorRegistry{
		CurrentEpoch: 10,
	}
	bootstrapData := bootstrapStorage.BootstrapData{
		NodesCoordinatorConfigKey: nodesCoordinatorConfigKey,
	}

	storer := &mock.StorerStub{
		GetCalled: func(key []byte) (b []byte, err error) {
			switch {
			case bytes.Equal([]byte(core.HighestRoundFromBootStorage), key):
				return roundBytes, nil
			case bytes.Equal([]byte(strconv.FormatInt(round, 10)), key):

				bootstrapDataBytes, _ := json.Marshal(bootstrapData)
				return bootstrapDataBytes, nil
			default:
				nodesConfigRegistryBytes, _ := json.Marshal(nodesConfigRegistry)
				return nodesConfigRegistryBytes, nil
			}
		},
	}

	bootData, nodesRegistry, err := epochStartProvider.getLastBootstrapData(storer)
	assert.Nil(t, err)
	assert.Equal(t, &bootstrapData, bootData)
	assert.Equal(t, &nodesConfigRegistry, nodesRegistry)
}
