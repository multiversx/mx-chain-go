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
	"github.com/stretchr/testify/require"
)

func TestPrepareEpochFromStorage(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	epochStartProvider, err := NewEpochStartBootstrap(args)
	require.Nil(t, err)
	epochStartProvider.initializeFromLocalStorage()

	epochStartProvider.baseData.lastEpoch = 10
	_, err = epochStartProvider.prepareEpochFromStorage()
	assert.Error(t, err)
}

func TestGetEpochStartMetaFromStorage(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.initializeFromLocalStorage()

	meta := &block.MetaBlock{Nonce: 1}
	metaBytes, _ := json.Marshal(meta)
	storer := &mock.StorerStub{
		GetCalled: func(key []byte) (bytes []byte, err error) {
			return metaBytes, nil
		},
		SearchFirstCalled: func(key []byte) ([]byte, error) {
			return metaBytes, nil
		},
	}
	metaBlock, err := epochStartProvider.getEpochStartMetaFromStorage(storer)
	assert.Nil(t, err)
	assert.Equal(t, meta, metaBlock)
}

func TestGetLastBootstrapData(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
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
				return nil, nil
			}
		},
		SearchFirstCalled: func(key []byte) ([]byte, error) {
			nodesConfigRegistryBytes, _ := json.Marshal(nodesConfigRegistry)
			return nodesConfigRegistryBytes, nil
		},
	}

	bootData, nodesRegistry, err := epochStartProvider.getLastBootstrapData(storer)
	assert.Nil(t, err)
	assert.Equal(t, &bootstrapData, bootData)
	assert.Equal(t, &nodesConfigRegistry, nodesRegistry)
}

func TestCheckIfShuffledOut_ValidatorIsInWaitingList(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.initializeFromLocalStorage()
	epochStartProvider.baseData.lastEpoch = 0

	publicKey := []byte("pubKey")
	nodesConfig := &sharding.NodesCoordinatorRegistry{
		CurrentEpoch: 1,
		EpochsConfig: map[string]*sharding.EpochValidators{
			"0": {
				WaitingValidators: map[string][]*sharding.SerializableValidator{
					"0": {{PubKey: publicKey, Chances: 0, Index: 0}},
				},
			},
		},
	}

	shardId, result := epochStartProvider.checkIfShuffledOut(publicKey, nodesConfig)
	assert.False(t, result)
	assert.Equal(t, shardId, epochStartProvider.baseData.shardId)
}

func TestCheckIfShuffledOut_ValidatorIsInEligibleList(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.initializeFromLocalStorage()
	epochStartProvider.baseData.lastEpoch = 0

	publicKey := []byte("pubKey")
	nodesConfig := &sharding.NodesCoordinatorRegistry{
		CurrentEpoch: 1,
		EpochsConfig: map[string]*sharding.EpochValidators{
			"0": {
				EligibleValidators: map[string][]*sharding.SerializableValidator{
					"0": {{PubKey: publicKey, Chances: 0, Index: 0}},
				},
			},
		},
	}

	shardId, result := epochStartProvider.checkIfShuffledOut(publicKey, nodesConfig)
	assert.False(t, result)
	assert.Equal(t, shardId, epochStartProvider.baseData.shardId)
}

func TestCheckIfShuffledOut_ValidatorIsShuffledToEligibleList(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.initializeFromLocalStorage()
	epochStartProvider.baseData.lastEpoch = 0
	epochStartProvider.baseData.shardId = 1

	publicKey := []byte("pubKey")
	nodesConfig := &sharding.NodesCoordinatorRegistry{
		CurrentEpoch: 1,
		EpochsConfig: map[string]*sharding.EpochValidators{
			"0": {
				EligibleValidators: map[string][]*sharding.SerializableValidator{
					"0": {{PubKey: publicKey, Chances: 0, Index: 0}},
				},
			},
		},
	}

	shardId, result := epochStartProvider.checkIfShuffledOut(publicKey, nodesConfig)
	assert.True(t, result)
	assert.NotEqual(t, shardId, epochStartProvider.baseData.shardId)
}

func TestCheckIfShuffledOut_ValidatorNotInEligibleOrWaiting(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.initializeFromLocalStorage()
	epochStartProvider.baseData.lastEpoch = 0

	publicKey := []byte("pubKey")
	nodesConfig := &sharding.NodesCoordinatorRegistry{
		CurrentEpoch: 1,
		EpochsConfig: map[string]*sharding.EpochValidators{
			"0": {
				EligibleValidators: map[string][]*sharding.SerializableValidator{},
				WaitingValidators:  map[string][]*sharding.SerializableValidator{},
			},
		},
	}

	shardId, result := epochStartProvider.checkIfShuffledOut(publicKey, nodesConfig)
	assert.False(t, result)
	assert.Equal(t, epochStartProvider.baseData.shardId, shardId)
}
