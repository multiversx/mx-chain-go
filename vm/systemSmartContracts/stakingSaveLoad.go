package systemSmartContracts

import (
	"math"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
)

func (s *stakingSC) getConfig() *StakingNodesConfig {
	stakeConfig := &StakingNodesConfig{
		MinNumNodes: int64(s.minNumNodes),
		MaxNumNodes: int64(s.maxNumNodes),
	}
	configData := s.eei.GetStorage([]byte(nodesConfigKey))
	if len(configData) == 0 {
		return stakeConfig
	}

	err := s.marshalizer.Unmarshal(stakeConfig, configData)
	if err != nil {
		log.Warn("unmarshal error on getConfig function, returning baseConfig",
			"error", err.Error(),
		)
		return &StakingNodesConfig{
			MinNumNodes: int64(s.minNumNodes),
			MaxNumNodes: int64(s.maxNumNodes),
		}
	}

	return stakeConfig
}

func (s *stakingSC) setConfig(config *StakingNodesConfig) {
	configData, err := s.marshalizer.Marshal(config)
	if err != nil {
		log.Warn("marshal error on setConfig function",
			"error", err.Error(),
		)
		return
	}

	s.eei.SetStorage([]byte(nodesConfigKey), configData)
}

func (s *stakingSC) getOrCreateRegisteredData(key []byte) (*StakedDataV2_0, error) {
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

	data := s.eei.GetStorage(key)
	if len(data) > 0 {
		err := s.marshalizer.Unmarshal(registrationData, data)
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

func (s *stakingSC) saveStakingData(key []byte, stakedData *StakedDataV2_0) error {
	if !s.flagEnableStaking.IsSet() {
		return s.saveAsStakingDataV1P0(key, stakedData)
	}
	if !s.flagStakingV2.IsSet() {
		return s.saveAsStakingDataV1P1(key, stakedData)
	}

	data, err := s.marshalizer.Marshal(stakedData)
	if err != nil {
		log.Debug("marshal error in saveStakingData ",
			"error", err.Error(),
		)
		return err
	}

	s.eei.SetStorage(key, data)
	return nil
}

func (s *stakingSC) saveAsStakingDataV1P0(key []byte, stakedData *StakedDataV2_0) error {
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

	data, err := s.marshalizer.Marshal(stakedDataV1)
	if err != nil {
		log.Debug("marshal error in saveAsStakingDataV1P0 ",
			"error", err.Error(),
		)
		return err
	}

	s.eei.SetStorage(key, data)
	return nil
}

func (s *stakingSC) saveAsStakingDataV1P1(key []byte, stakedData *StakedDataV2_0) error {
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

	data, err := s.marshalizer.Marshal(stakedDataV1)
	if err != nil {
		log.Debug("marshal error in saveAsStakingDataV1P1 ",
			"error", err.Error(),
		)
		return err
	}

	s.eei.SetStorage(key, data)
	return nil
}
