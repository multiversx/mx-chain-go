package stakingcommon

import (
	"math/big"
	"strconv"

	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/genesis/process/disabled"
	"github.com/multiversx/mx-chain-go/process"
	economicsHandler "github.com/multiversx/mx-chain-go/process/economics"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("testscommon/stakingCommon")

// RegisterValidatorKeys will register validator's staked key in the provided accounts db
func RegisterValidatorKeys(
	accountsDB state.AccountsAdapter,
	ownerAddress []byte,
	rewardAddress []byte,
	stakedKeys [][]byte,
	totalStake *big.Int,
	marshaller marshal.Marshalizer,
) {
	AddValidatorData(accountsDB, ownerAddress, stakedKeys, totalStake, marshaller)
	AddStakingData(accountsDB, ownerAddress, rewardAddress, stakedKeys, marshaller)
	_, err := accountsDB.Commit()
	log.LogIfError(err)
}

// AddValidatorData will add the validator's registered keys in the provided accounts db
func AddValidatorData(
	accountsDB state.AccountsAdapter,
	ownerKey []byte,
	registeredKeys [][]byte,
	totalStake *big.Int,
	marshaller marshal.Marshalizer,
) {
	validatorSC := LoadUserAccount(accountsDB, vm.ValidatorSCAddress)
	ownerStoredData, _, _ := validatorSC.RetrieveValue(ownerKey)
	validatorData := &systemSmartContracts.ValidatorDataV2{}
	if len(ownerStoredData) != 0 {
		_ = marshaller.Unmarshal(validatorData, ownerStoredData)
		validatorData.BlsPubKeys = append(validatorData.BlsPubKeys, registeredKeys...)
		validatorData.TotalStakeValue = totalStake
	} else {
		validatorData = &systemSmartContracts.ValidatorDataV2{
			RegisterNonce:   0,
			Epoch:           0,
			RewardAddress:   ownerKey,
			TotalStakeValue: totalStake,
			LockedStake:     big.NewInt(0),
			TotalUnstaked:   big.NewInt(0),
			BlsPubKeys:      registeredKeys,
			NumRegistered:   uint32(len(registeredKeys)),
		}
	}

	marshaledData, _ := marshaller.Marshal(validatorData)
	_ = validatorSC.SaveKeyValue(ownerKey, marshaledData)

	_ = accountsDB.SaveAccount(validatorSC)
}

// AddStakingData will add the owner's staked keys in the provided accounts db
func AddStakingData(
	accountsDB state.AccountsAdapter,
	ownerAddress []byte,
	rewardAddress []byte,
	stakedKeys [][]byte,
	marshaller marshal.Marshalizer,
) {
	stakedData := &systemSmartContracts.StakedDataV2_0{
		Staked:        true,
		RewardAddress: rewardAddress,
		OwnerAddress:  ownerAddress,
		StakeValue:    big.NewInt(100),
	}
	marshaledData, _ := marshaller.Marshal(stakedData)

	stakingSCAcc := LoadUserAccount(accountsDB, vm.StakingSCAddress)
	for _, key := range stakedKeys {
		_ = stakingSCAcc.SaveKeyValue(key, marshaledData)
	}

	_ = accountsDB.SaveAccount(stakingSCAcc)
}

// AddKeysToWaitingList will add the owner's provided bls keys in the staking queue list
func AddKeysToWaitingList(
	accountsDB state.AccountsAdapter,
	waitingKeys [][]byte,
	marshaller marshal.Marshalizer,
	rewardAddress []byte,
	ownerAddress []byte,
) {
	if len(waitingKeys) == 0 {
		return
	}

	stakingSCAcc := LoadUserAccount(accountsDB, vm.StakingSCAddress)
	waitingList := getWaitingList(stakingSCAcc, marshaller)

	waitingListAlreadyHasElements := waitingList.Length > 0
	waitingListLastKeyBeforeAddingNewKeys := waitingList.LastKey
	previousKey := waitingList.LastKey
	if !waitingListAlreadyHasElements {
		waitingList.FirstKey = getPrefixedWaitingKey(waitingKeys[0])
		previousKey = waitingList.FirstKey
	}

	numWaitingKeys := len(waitingKeys)
	waitingList.LastKey = getPrefixedWaitingKey(waitingKeys[numWaitingKeys-1])
	waitingList.Length += uint32(numWaitingKeys)
	saveWaitingList(stakingSCAcc, marshaller, waitingList)

	for i, waitingKey := range waitingKeys {
		waitingListElement := &systemSmartContracts.ElementInList{
			BLSPublicKey: waitingKey,
			PreviousKey:  previousKey,
			NextKey:      make([]byte, 0),
		}

		if i < numWaitingKeys-1 {
			nextKey := getPrefixedWaitingKey(waitingKeys[i+1])
			waitingListElement.NextKey = nextKey
		}

		prefixedWaitingKey := getPrefixedWaitingKey(waitingKey)
		saveStakedWaitingKey(stakingSCAcc, marshaller, rewardAddress, ownerAddress, waitingKey)
		saveElemInList(stakingSCAcc, marshaller, waitingListElement, prefixedWaitingKey)

		previousKey = prefixedWaitingKey
	}

	if waitingListAlreadyHasElements {
		lastElem, _ := GetWaitingListElement(stakingSCAcc, marshaller, waitingListLastKeyBeforeAddingNewKeys)
		lastElem.NextKey = getPrefixedWaitingKey(waitingKeys[0])
		saveElemInList(stakingSCAcc, marshaller, lastElem, waitingListLastKeyBeforeAddingNewKeys)
	}

	_ = accountsDB.SaveAccount(stakingSCAcc)
}

func getWaitingList(
	stakingSCAcc state.UserAccountHandler,
	marshaller marshal.Marshalizer,
) *systemSmartContracts.WaitingList {
	marshaledData, _, _ := stakingSCAcc.RetrieveValue([]byte("waitingList"))
	waitingList := &systemSmartContracts.WaitingList{}
	_ = marshaller.Unmarshal(waitingList, marshaledData)

	return waitingList
}

func saveWaitingList(
	stakingSCAcc state.UserAccountHandler,
	marshaller marshal.Marshalizer,
	waitingList *systemSmartContracts.WaitingList,
) {
	marshaledData, _ := marshaller.Marshal(waitingList)
	_ = stakingSCAcc.SaveKeyValue([]byte("waitingList"), marshaledData)
}

func getPrefixedWaitingKey(key []byte) []byte {
	return []byte("w_" + string(key))
}

func saveStakedWaitingKey(
	stakingSCAcc state.UserAccountHandler,
	marshaller marshal.Marshalizer,
	rewardAddress []byte,
	ownerAddress []byte,
	key []byte,
) {
	stakedData := &systemSmartContracts.StakedDataV2_0{
		Waiting:       true,
		RewardAddress: rewardAddress,
		OwnerAddress:  ownerAddress,
		StakeValue:    big.NewInt(100),
	}

	marshaledData, _ := marshaller.Marshal(stakedData)
	_ = stakingSCAcc.SaveKeyValue(key, marshaledData)
}

func saveElemInList(
	stakingSCAcc state.UserAccountHandler,
	marshaller marshal.Marshalizer,
	elem *systemSmartContracts.ElementInList,
	key []byte,
) {
	marshaledData, _ := marshaller.Marshal(elem)
	_ = stakingSCAcc.SaveKeyValue(key, marshaledData)
}

// GetWaitingListElement returns the element in waiting list saved at the provided key
func GetWaitingListElement(
	stakingSCAcc state.UserAccountHandler,
	marshaller marshal.Marshalizer,
	key []byte,
) (*systemSmartContracts.ElementInList, error) {
	marshaledData, _, _ := stakingSCAcc.RetrieveValue(key)
	if len(marshaledData) == 0 {
		return nil, vm.ErrElementNotFound
	}

	element := &systemSmartContracts.ElementInList{}
	err := marshaller.Unmarshal(element, marshaledData)
	if err != nil {
		return nil, err
	}

	return element, nil
}

// LoadUserAccount returns address's state.UserAccountHandler from the provided db
func LoadUserAccount(accountsDB state.AccountsAdapter, address []byte) state.UserAccountHandler {
	acc, _ := accountsDB.LoadAccount(address)
	return acc.(state.UserAccountHandler)
}

// CreateEconomicsData returns an initialized process.EconomicsDataHandler
func CreateEconomicsData() process.EconomicsDataHandler {
	maxGasLimitPerBlock := strconv.FormatUint(1500000000, 10)
	minGasPrice := strconv.FormatUint(10, 10)
	minGasLimit := strconv.FormatUint(10, 10)

	argsNewEconomicsData := economicsHandler.ArgsNewEconomicsData{
		Economics: &config.EconomicsConfig{
			GlobalSettings: config.GlobalSettings{
				GenesisTotalSupply: "2000000000000000000000",
				MinimumInflation:   0,
				YearSettings: []*config.YearSetting{
					{
						Year:             0,
						MaximumInflation: 0.01,
					},
				},
			},
			RewardsSettings: config.RewardsSettings{
				RewardsConfigByEpoch: []config.EpochRewardSettings{
					{
						LeaderPercentage:                 0.1,
						DeveloperPercentage:              0.1,
						ProtocolSustainabilityPercentage: 0.1,
						ProtocolSustainabilityAddress:    "protocol",
						TopUpGradientPoint:               "300000000000000000000",
						TopUpFactor:                      0.25,
					},
				},
			},
			FeeSettings: config.FeeSettings{
				GasLimitSettings: []config.GasLimitSetting{
					{
						MaxGasLimitPerBlock:         maxGasLimitPerBlock,
						MaxGasLimitPerMiniBlock:     maxGasLimitPerBlock,
						MaxGasLimitPerMetaBlock:     maxGasLimitPerBlock,
						MaxGasLimitPerMetaMiniBlock: maxGasLimitPerBlock,
						MaxGasLimitPerTx:            maxGasLimitPerBlock,
						MinGasLimit:                 minGasLimit,
						ExtraGasLimitGuardedTx:      maxGasLimitPerBlock,
					},
				},
				MinGasPrice:            minGasPrice,
				GasPerDataByte:         "1",
				GasPriceModifier:       1.0,
				MaxGasPriceSetGuardian: minGasPrice,
			},
		},
		EpochNotifier:       &epochNotifier.EpochNotifierStub{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		TxVersionChecker:    &disabled.TxVersionChecker{},
	}
	economicsData, _ := economicsHandler.NewEconomicsData(argsNewEconomicsData)
	return economicsData
}

// SaveNodesConfig saves the nodes config in accounts db under "nodesConfig" key with provided params
func SaveNodesConfig(
	accountsDB state.AccountsAdapter,
	marshaller marshal.Marshalizer,
	stakedNodes,
	minNumNodes,
	maxNumNodes int64,
) {
	nodesConfigData := &systemSmartContracts.StakingNodesConfig{
		StakedNodes: stakedNodes,
		MinNumNodes: minNumNodes,
		MaxNumNodes: maxNumNodes,
	}
	nodesDataBytes, err := marshaller.Marshal(nodesConfigData)
	log.LogIfError(err)

	account, err := accountsDB.LoadAccount(vm.StakingSCAddress)
	log.LogIfError(err)

	userAccount, _ := account.(state.UserAccountHandler)
	err = userAccount.SaveKeyValue([]byte("nodesConfig"), nodesDataBytes)
	log.LogIfError(err)
	err = accountsDB.SaveAccount(account)
	log.LogIfError(err)
	_, err = accountsDB.Commit()
	log.LogIfError(err)
}

// SaveDelegationManagerConfig will save a mock configuration for the delegation manager SC
func SaveDelegationManagerConfig(accountsDB state.AccountsAdapter, marshaller marshal.Marshalizer) {
	managementData := &systemSmartContracts.DelegationManagement{
		MinDeposit:          big.NewInt(100),
		LastAddress:         vm.FirstDelegationSCAddress,
		MinDelegationAmount: big.NewInt(1),
	}
	marshaledData, err := marshaller.Marshal(managementData)
	log.LogIfError(err)

	acc, err := accountsDB.LoadAccount(vm.DelegationManagerSCAddress)
	log.LogIfError(err)
	delegationAcc, _ := acc.(state.UserAccountHandler)

	err = delegationAcc.SaveKeyValue([]byte("delegationManagement"), marshaledData)
	log.LogIfError(err)
	err = accountsDB.SaveAccount(delegationAcc)
	log.LogIfError(err)
	_, err = accountsDB.Commit()
	log.LogIfError(err)
}
