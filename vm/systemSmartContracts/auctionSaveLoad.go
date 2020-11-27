package systemSmartContracts

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
)

func (s *stakingAuctionSC) setConfig(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ownerAddress := s.eei.GetStorage([]byte(ownerKey))
	if !bytes.Equal(ownerAddress, args.CallerAddr) {
		s.eei.AddReturnMessage("setConfig function was not called by the owner address")
		return vmcommon.UserError
	}

	if len(args.Arguments) != 7 {
		retMessage := fmt.Sprintf("setConfig function called with wrong number of arguments expected %d, got %d", 7, len(args.Arguments))
		s.eei.AddReturnMessage(retMessage)
		return vmcommon.UserError
	}

	auctionConfig := &AuctionConfig{
		MinStakeValue: big.NewInt(0).SetBytes(args.Arguments[0]),
		NumNodes:      uint32(big.NewInt(0).SetBytes(args.Arguments[1]).Uint64()),
		TotalSupply:   big.NewInt(0).SetBytes(args.Arguments[2]),
		MinStep:       big.NewInt(0).SetBytes(args.Arguments[3]),
		NodePrice:     big.NewInt(0).SetBytes(args.Arguments[4]),
		UnJailPrice:   big.NewInt(0).SetBytes(args.Arguments[5]),
	}

	code := s.verifyConfig(auctionConfig)
	if code != vmcommon.Ok {
		return code
	}

	configData, err := s.marshalizer.Marshal(auctionConfig)
	if err != nil {
		s.eei.AddReturnMessage("setConfig marshal auctionConfig error")
		return vmcommon.UserError
	}

	epochBytes := args.Arguments[6]
	s.eei.SetStorage(epochBytes, configData)

	return vmcommon.Ok
}

func (s *stakingAuctionSC) getConfig(epoch uint32) AuctionConfig {
	epochKey := big.NewInt(int64(epoch)).Bytes()
	configData := s.eei.GetStorage(epochKey)
	if len(configData) == 0 {
		return s.baseConfig
	}

	auctionConfig := &AuctionConfig{}
	err := s.marshalizer.Unmarshal(auctionConfig, configData)
	if err != nil {
		log.Warn("unmarshal error on getConfig function, returning baseConfig",
			"error", err.Error(),
		)
		return s.baseConfig
	}

	if s.checkConfigCorrectness(*auctionConfig) != nil {
		baseConfigData, errMarshal := s.marshalizer.Marshal(&s.baseConfig)
		if errMarshal != nil {
			log.Warn("marshal error on getConfig function, returning baseConfig", "error", errMarshal)
			return s.baseConfig
		}
		s.eei.SetStorage(epochKey, baseConfigData)
		return s.baseConfig
	}

	return *auctionConfig
}

func (s *stakingAuctionSC) getOrCreateRegistrationData(key []byte) (*AuctionDataV2, error) {
	data := s.eei.GetStorage(key)
	registrationData := &AuctionDataV2{
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
		err := s.marshalizer.Unmarshal(registrationData, data)
		if err != nil {
			log.Debug("unmarshal error on auction SC stake function",
				"error", err.Error(),
			)
			return nil, err
		}
	}

	return registrationData, nil
}

func (s *stakingAuctionSC) saveRegistrationData(key []byte, auction *AuctionDataV2) error {
	if !s.flagEnableTopUp.IsSet() {
		return s.saveRegistrationDataV1(key, auction)
	}

	data, err := s.marshalizer.Marshal(auction)
	if err != nil {
		log.Debug("marshal error on staking SC stake function in saveRegistrationData",
			"error", err.Error(),
		)
		return err
	}

	s.eei.SetStorage(key, data)
	return nil
}

func (s *stakingAuctionSC) saveRegistrationDataV1(key []byte, auction *AuctionDataV2) error {
	auctionDataV1 := &AuctionDataV1{
		RegisterNonce:   auction.RegisterNonce,
		Epoch:           auction.Epoch,
		RewardAddress:   auction.RewardAddress,
		TotalStakeValue: auction.TotalStakeValue,
		LockedStake:     auction.LockedStake,
		MaxStakePerNode: auction.MaxStakePerNode,
		BlsPubKeys:      auction.BlsPubKeys,
		NumRegistered:   auction.NumRegistered,
	}

	data, err := s.marshalizer.Marshal(auctionDataV1)
	if err != nil {
		log.Debug("marshal error on staking SC stake function in saveRegistrationDataV1",
			"error", err.Error(),
		)
		return err
	}

	s.eei.SetStorage(key, data)
	return nil
}

func (s *stakingAuctionSC) getStakedData(key []byte) (*StakedDataV2_0, error) {
	data := s.eei.GetStorageFromAddress(s.stakingSCAddress, key)
	stakedData := &StakedDataV2_0{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: 0,
		UnStakedEpoch: core.DefaultUnstakedEpoch,
		RewardAddress: nil,
		StakeValue:    big.NewInt(0),
		SlashValue:    big.NewInt(0),
	}

	if len(data) > 0 {
		err := s.marshalizer.Unmarshal(stakedData, data)
		if err != nil {
			log.Debug("unmarshal error on auction SC getStakedData function",
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
