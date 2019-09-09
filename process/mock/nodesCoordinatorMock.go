package mock

import (
    "bytes"
    "fmt"
    "math/big"

    "github.com/ElrondNetwork/elrond-go/sharding"
)

// NodesCoordinator defines the behaviour of a struct able to do validator group selection
type NodesCoordinatorMock struct {
    Validators                          map[uint32][]sharding.Validator
    ShardConsensusSize                  uint32
    MetaConsensusSize                   uint32
    ShardId                             uint32
    NbShards                            uint32
    GetSelectedPublicKeysCalled         func(selection []byte, shardId uint32) (publicKeys []string, err error)
    GetValidatorsPublicKeysCalled       func(randomness []byte, round uint64, shardId uint32) ([]string, error)
    GetValidatorsRewardsAddressesCalled func(randomness []byte, round uint64, shardId uint32) ([]string, error)
    LoadNodesPerShardsCalled            func(nodes map[uint32][]sharding.Validator) error
    ComputeValidatorsGroupCalled        func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error)
    GetValidatorWithPublicKeyCalled     func(publicKey []byte) (validator sharding.Validator, shardId uint32, err error)
}

func NewNodesCoordinatorMock() *NodesCoordinatorMock {
    nbShards := uint32(1)
    nodesPerShard := 2
    validatorsMap := make(map[uint32][]sharding.Validator)

    for sh := uint32(0); sh < nbShards; sh++ {
        validatorsList := make([]sharding.Validator, nodesPerShard)
        for v := 0; v < nodesPerShard; v++ {
            validatorsList[v], _ = sharding.NewValidator(
                big.NewInt(10),
                1,
                []byte(fmt.Sprintf("pubKey%d%d", sh, v)),
                []byte(fmt.Sprintf("address%d%d", sh, v)),
            )
        }
        validatorsMap[sh] = validatorsList
    }

    return &NodesCoordinatorMock{
        ShardConsensusSize: 1,
        MetaConsensusSize:  1,
        ShardId:            0,
        NbShards:           nbShards,
        Validators:         validatorsMap,
    }
}

func (ncm *NodesCoordinatorMock) GetSelectedPublicKeys(selection []byte, shardId uint32) (publicKeys []string, err error) {
    if ncm.GetSelectedPublicKeysCalled != nil {
        return ncm.GetSelectedPublicKeysCalled(selection, shardId)
    }

    if len(ncm.Validators) == 0 {
        return nil, sharding.ErrNilInputNodesMap
    }

    pubKeys := make([]string, 0)

    for _, v := range ncm.Validators[shardId] {
        pubKeys = append(pubKeys, string(v.PubKey()))
    }

    return pubKeys, nil
}

func (ncm *NodesCoordinatorMock) GetValidatorsPublicKeys(
    randomness []byte,
    round uint64,
    shardId uint32,
) ([]string, error) {
    if ncm.GetValidatorsPublicKeysCalled != nil {
        return ncm.GetValidatorsPublicKeysCalled(randomness, round, shardId)
    }

    validators, err := ncm.ComputeValidatorsGroup(randomness, round, shardId)
    if err != nil {
        return nil, err
    }

    valGrStr := make([]string, 0)

    for _, v := range validators {
        valGrStr = append(valGrStr, string(v.PubKey()))
    }

    return valGrStr, nil
}

func (ncm *NodesCoordinatorMock) GetValidatorsRewardsAddresses(
    randomness []byte,
    round uint64,
    shardId uint32,
) ([]string, error) {
    if ncm.GetValidatorsPublicKeysCalled != nil {
        return ncm.GetValidatorsRewardsAddressesCalled(randomness, round, shardId)
    }

    validators, err := ncm.ComputeValidatorsGroup(randomness, round, shardId)
    if err != nil {
        return nil, err
    }

    addresses := make([]string, 0)
    for _, v := range validators {
        addresses = append(addresses, string(v.Address()))
    }

    return addresses, nil
}

func (ncm *NodesCoordinatorMock) SetNodesPerShards(nodes map[uint32][]sharding.Validator) error {
    if ncm.LoadNodesPerShardsCalled != nil {
        return ncm.LoadNodesPerShardsCalled(nodes)
    }

    if nodes == nil {
        return sharding.ErrNilInputNodesMap
    }

    ncm.Validators = nodes

    return nil
}

func (ncm *NodesCoordinatorMock) ComputeValidatorsGroup(
    randomess []byte,
    round uint64,
    shardId uint32,
) ([]sharding.Validator, error) {
    var consensusSize uint32

    if ncm.ComputeValidatorsGroupCalled != nil {
        return ncm.ComputeValidatorsGroupCalled(randomess, round, shardId)
    }

    if ncm.ShardId == sharding.MetachainShardId {
        consensusSize = ncm.MetaConsensusSize
    } else {
        consensusSize = ncm.ShardConsensusSize
    }

    if randomess == nil {
        return nil, sharding.ErrNilRandomness
    }

    validatorsGroup := make([]sharding.Validator, 0)

    for i := uint32(0); i < consensusSize; i++ {
        validatorsGroup = append(validatorsGroup, ncm.Validators[shardId][i])
    }

    return validatorsGroup, nil
}

func (ncm *NodesCoordinatorMock) GetValidatorWithPublicKey(publicKey []byte) (sharding.Validator, uint32, error) {
    if ncm.GetValidatorWithPublicKeyCalled != nil {
        return ncm.GetValidatorWithPublicKeyCalled(publicKey)
    }

    if publicKey == nil {
        return nil, 0, sharding.ErrNilPubKey
    }

    for shardId, shardEligible := range ncm.Validators {
        for i := 0; i < len(shardEligible); i++ {
            if bytes.Equal(publicKey, shardEligible[i].PubKey()) {
                return shardEligible[i], shardId, nil
            }
        }
    }

    return nil, 0, sharding.ErrValidatorNotFound
}

func (ncm *NodesCoordinatorMock) IsInterfaceNil() bool {
	if ncm == nil {
		return true
	}
	return false
}
