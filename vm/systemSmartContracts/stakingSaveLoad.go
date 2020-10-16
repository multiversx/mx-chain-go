package systemSmartContracts

import (
	"math"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
)

func (r *stakingSC) getConfig() *StakingNodesConfig {
	stakeConfig := &StakingNodesConfig{
		MinNumNodes: int64(r.minNumNodes),
		MaxNumNodes: int64(r.maxNumNodes),
	}
	configData := r.eei.GetStorage([]byte(nodesConfigKey))
	if len(configData) == 0 {
		return stakeConfig
	}

	err := r.marshalizer.Unmarshal(stakeConfig, configData)
	if err != nil {
		log.Warn("unmarshal error on getConfig function, returning baseConfig",
			"error", err.Error(),
		)
		return &StakingNodesConfig{
			MinNumNodes: int64(r.minNumNodes),
			MaxNumNodes: int64(r.maxNumNodes),
		}
	}

	return stakeConfig
}

func (r *stakingSC) setConfig(config *StakingNodesConfig) {
	configData, err := r.marshalizer.Marshal(config)
	if err != nil {
		log.Warn("marshal error on setConfig function",
			"error", err.Error(),
		)
		return
	}

	r.eei.SetStorage([]byte(nodesConfigKey), configData)
}

func (r *stakingSC) getOrCreateRegisteredData(key []byte) (*StakedDataV2_0, error) {
	registrationData := &StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: 0,
		UnStakedEpoch: core.DefaultUnstakedEpoch,
		RewardAddress: nil,
		StakeValue:    big.NewInt(0),
		JailedRound:   math.MaxUint64,
		UnJailedNonce: 0,
		JailedNonce:   0,
		StakedNonce:   math.MaxUint64,
		SlashValue:    big.NewInt(0),
	}

	data := r.eei.GetStorage(key)
	if len(data) > 0 {
		err := r.marshalizer.Unmarshal(registrationData, data)
		if err != nil {
			log.Debug("unmarshal error on staking SC stake function",
				"error", err.Error(),
			)
			return nil, err
		}
	}
	if registrationData.SlashValue == nil {
		registrationData.SlashValue = big.NewInt(0)
	}

	return registrationData, nil
}

func (r *stakingSC) saveStakingData(key []byte, stakedData *StakedDataV2_0) error {
	if !r.flagEnableStaking.IsSet() {
		return r.saveAsStakingDataV1P0(key, stakedData)
	}
	if !r.flagSetOwner.IsSet() {
		return r.saveAsStakingDataV1P1(key, stakedData)
	}

	data, err := r.marshalizer.Marshal(stakedData)
	if err != nil {
		log.Debug("marshal error in saveStakingData ",
			"error", err.Error(),
		)
		return err
	}

	r.eei.SetStorage(key, data)
	return nil
}

func (r *stakingSC) saveAsStakingDataV1P0(key []byte, stakedData *StakedDataV2_0) error {
	stakedDataV1 := &StakedDataV1_0{
		RegisterNonce: stakedData.RegisterNonce,
		StakedNonce:   stakedData.StakedNonce,
		Staked:        stakedData.Staked,
		UnStakedNonce: stakedData.UnStakedNonce,
		UnStakedEpoch: stakedData.UnStakedEpoch,
		RewardAddress: stakedData.RewardAddress,
		StakeValue:    stakedData.StakeValue,
		JailedRound:   stakedData.JailedRound,
		JailedNonce:   stakedData.JailedNonce,
		UnJailedNonce: stakedData.UnJailedNonce,
		Jailed:        stakedData.Jailed,
		Waiting:       stakedData.Waiting,
	}

	data, err := r.marshalizer.Marshal(stakedDataV1)
	if err != nil {
		log.Debug("marshal error in saveAsStakingDataV1P0 ",
			"error", err.Error(),
		)
		return err
	}

	r.eei.SetStorage(key, data)
	return nil
}

func (r *stakingSC) saveAsStakingDataV1P1(key []byte, stakedData *StakedDataV2_0) error {
	stakedDataV1 := &StakedDataV1_1{
		RegisterNonce: stakedData.RegisterNonce,
		StakedNonce:   stakedData.StakedNonce,
		Staked:        stakedData.Staked,
		UnStakedNonce: stakedData.UnStakedNonce,
		UnStakedEpoch: stakedData.UnStakedEpoch,
		RewardAddress: stakedData.RewardAddress,
		StakeValue:    stakedData.StakeValue,
		JailedRound:   stakedData.JailedRound,
		JailedNonce:   stakedData.JailedNonce,
		UnJailedNonce: stakedData.UnJailedNonce,
		Jailed:        stakedData.Jailed,
		Waiting:       stakedData.Waiting,
		NumJailed:     stakedData.NumJailed,
		SlashValue:    stakedData.SlashValue,
	}

	data, err := r.marshalizer.Marshal(stakedDataV1)
	if err != nil {
		log.Debug("marshal error in saveAsStakingDataV1P1 ",
			"error", err.Error(),
		)
		return err
	}

	r.eei.SetStorage(key, data)
	return nil
}
