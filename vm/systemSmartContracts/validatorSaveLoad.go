package systemSmartContracts

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/common"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

func (v *validatorSC) setConfig(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ownerAddress := v.eei.GetStorage([]byte(ownerKey))
	if !bytes.Equal(ownerAddress, args.CallerAddr) {
		v.eei.AddReturnMessage("setConfig function was not called by the owner address")
		return vmcommon.UserError
	}

	if len(args.Arguments) != 6 {
		retMessage := fmt.Sprintf("setConfig function called with wrong number of arguments expected %d, got %d", 7, len(args.Arguments))
		v.eei.AddReturnMessage(retMessage)
		return vmcommon.UserError
	}

	validatorConfig := &ValidatorConfig{
		MinStakeValue: big.NewInt(0).SetBytes(args.Arguments[0]),
		TotalSupply:   big.NewInt(0).SetBytes(args.Arguments[1]),
		MinStep:       big.NewInt(0).SetBytes(args.Arguments[2]),
		NodePrice:     big.NewInt(0).SetBytes(args.Arguments[3]),
		UnJailPrice:   big.NewInt(0).SetBytes(args.Arguments[4]),
	}

	code := v.verifyConfig(validatorConfig)
	if code != vmcommon.Ok {
		return code
	}

	configData, err := v.marshalizer.Marshal(validatorConfig)
	if err != nil {
		v.eei.AddReturnMessage("setConfig marshal validatorConfig error")
		return vmcommon.UserError
	}

	epochBytes := args.Arguments[5]
	v.eei.SetStorage(epochBytes, configData)

	return vmcommon.Ok
}

func (v *validatorSC) getConfig(epoch uint32) ValidatorConfig {
	epochKey := big.NewInt(int64(epoch)).Bytes()
	configData := v.eei.GetStorage(epochKey)
	if len(configData) == 0 {
		return v.baseConfig
	}

	validatorConfig := &ValidatorConfig{}
	err := v.marshalizer.Unmarshal(validatorConfig, configData)
	if err != nil {
		log.Warn("unmarshal error on getConfig function, returning baseConfig",
			"error", err.Error(),
		)
		return v.baseConfig
	}

	if v.checkConfigCorrectness(*validatorConfig) != nil {
		baseConfigData, errMarshal := v.marshalizer.Marshal(&v.baseConfig)
		if errMarshal != nil {
			log.Warn("marshal error on getConfig function, returning baseConfig", "error", errMarshal)
			return v.baseConfig
		}
		v.eei.SetStorage(epochKey, baseConfigData)
		return v.baseConfig
	}

	return *validatorConfig
}

func (v *validatorSC) getOrCreateRegistrationData(key []byte) (*ValidatorDataV2, error) {
	data := v.eei.GetStorage(key)
	registrationData := &ValidatorDataV2{
		RewardAddress:   nil,
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      nil,
		TotalStakeValue: big.NewInt(0),
		LockedStake:     big.NewInt(0),
		MaxStakePerNode: big.NewInt(0),
		TotalUnstaked:   big.NewInt(0),
		UnstakedInfo:    make([]*UnstakedValue, 0),
	}

	if len(data) > 0 {
		err := v.marshalizer.Unmarshal(registrationData, data)
		if err != nil {
			log.Debug("unmarshal error on validator SC stake function",
				"error", err.Error(),
			)
			return nil, err
		}
	}

	if registrationData.TotalUnstaked == nil {
		registrationData.TotalUnstaked = big.NewInt(0)
	}
	if registrationData.UnstakedInfo == nil {
		registrationData.UnstakedInfo = make([]*UnstakedValue, 0)
	}

	return registrationData, nil
}

func (v *validatorSC) saveRegistrationData(key []byte, validator *ValidatorDataV2) error {
	if !v.flagEnableTopUp.IsSet() {
		return v.saveRegistrationDataV1(key, validator)
	}

	data, err := v.marshalizer.Marshal(validator)
	if err != nil {
		log.Debug("marshal error on staking SC stake function in saveRegistrationData",
			"error", err.Error(),
		)
		return err
	}

	v.eei.SetStorage(key, data)
	return nil
}

func (v *validatorSC) saveRegistrationDataV1(key []byte, validator *ValidatorDataV2) error {
	validatorDataV1 := &ValidatorDataV1{
		RegisterNonce:   validator.RegisterNonce,
		Epoch:           validator.Epoch,
		RewardAddress:   validator.RewardAddress,
		TotalStakeValue: validator.TotalStakeValue,
		LockedStake:     validator.LockedStake,
		MaxStakePerNode: validator.MaxStakePerNode,
		BlsPubKeys:      validator.BlsPubKeys,
		NumRegistered:   validator.NumRegistered,
	}

	data, err := v.marshalizer.Marshal(validatorDataV1)
	if err != nil {
		log.Debug("marshal error on staking SC stake function in saveRegistrationDataV1",
			"error", err.Error(),
		)
		return err
	}

	v.eei.SetStorage(key, data)
	return nil
}

func (v *validatorSC) getStakedData(key []byte) (*StakedDataV2_0, error) {
	data := v.eei.GetStorageFromAddress(v.stakingSCAddress, key)
	stakedData := &StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: 0,
		UnStakedEpoch: common.DefaultUnstakedEpoch,
		RewardAddress: nil,
		StakeValue:    big.NewInt(0),
		SlashValue:    big.NewInt(0),
	}

	if len(data) > 0 {
		err := v.marshalizer.Unmarshal(stakedData, data)
		if err != nil {
			log.Debug("unmarshal error on validator SC getStakedData function",
				"error", err.Error(),
			)
			return nil, err
		}

		if stakedData.SlashValue == nil {
			stakedData.SlashValue = big.NewInt(0)
		}
	}

	return stakedData, nil
}
