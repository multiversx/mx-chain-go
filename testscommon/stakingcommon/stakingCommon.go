package stakingcommon

import (
	"math/big"
	"strconv"

	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process"
	economicsHandler "github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon/epochNotifier"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
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
	validatorData := &systemSmartContracts.ValidatorDataV2{
		RegisterNonce:   0,
		Epoch:           0,
		RewardAddress:   ownerKey,
		TotalStakeValue: totalStake,
		LockedStake:     big.NewInt(0),
		TotalUnstaked:   big.NewInt(0),
		BlsPubKeys:      registeredKeys,
		NumRegistered:   uint32(len(registeredKeys)),
	}

	marshaledData, _ := marshaller.Marshal(validatorData)
	_ = validatorSC.DataTrieTracker().SaveKeyValue(ownerKey, marshaledData)

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
		_ = stakingSCAcc.DataTrieTracker().SaveKeyValue(key, marshaledData)
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
		waitingList.FirstKey = []byte("w_" + string(waitingKeys[0]))
		previousKey = waitingList.FirstKey
	}

	numWaitingKeys := len(waitingKeys)
	waitingList.LastKey = []byte("w_" + string(waitingKeys[numWaitingKeys-1]))
	waitingList.Length += uint32(numWaitingKeys)
	saveWaitingList(stakingSCAcc, marshaller, waitingList)

	for i, waitingKey := range waitingKeys {
		waitingListElement := &systemSmartContracts.ElementInList{
			BLSPublicKey: waitingKey,
			PreviousKey:  previousKey,
			NextKey:      make([]byte, 0),
		}

		if i < numWaitingKeys-1 {
			nextKey := []byte("w_" + string(waitingKeys[i+1]))
			waitingListElement.NextKey = nextKey
		}

		saveStakedData(stakingSCAcc, marshaller, rewardAddress, ownerAddress, waitingKey)
		previousKey = saveElemInList(stakingSCAcc, marshaller, waitingListElement, waitingKey)
	}

	if waitingListAlreadyHasElements {
		lastKeyWithoutPrefix := waitingListLastKeyBeforeAddingNewKeys[2:]

		lastElem := getElemInList(stakingSCAcc, marshaller, lastKeyWithoutPrefix)
		lastElem.NextKey = []byte("w_" + string(waitingKeys[0]))
		saveElemInList(stakingSCAcc, marshaller, lastElem, lastKeyWithoutPrefix)
	}

	_ = accountsDB.SaveAccount(stakingSCAcc)
}

func getWaitingList(
	stakingSCAcc state.UserAccountHandler,
	marshaller marshal.Marshalizer,
) *systemSmartContracts.WaitingList {
	marshaledData, _ := stakingSCAcc.DataTrieTracker().RetrieveValue([]byte("waitingList"))
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
	_ = stakingSCAcc.DataTrieTracker().SaveKeyValue([]byte("waitingList"), marshaledData)
}

func saveStakedData(
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
	_ = stakingSCAcc.DataTrieTracker().SaveKeyValue(key, marshaledData)
}

func saveElemInList(
	stakingSCAcc state.UserAccountHandler,
	marshaller marshal.Marshalizer,
	elem *systemSmartContracts.ElementInList,
	key []byte,
) []byte {
	marshaledData, _ := marshaller.Marshal(elem)
	waitingKeyInList := []byte("w_" + string(key))
	_ = stakingSCAcc.DataTrieTracker().SaveKeyValue(waitingKeyInList, marshaledData)

	return waitingKeyInList
}

func getElemInList(
	stakingSCAcc state.UserAccountHandler,
	marshaller marshal.Marshalizer,
	key []byte,
) *systemSmartContracts.ElementInList {
	marshaledData, _ := stakingSCAcc.DataTrieTracker().RetrieveValue([]byte("w_" + string(key)))
	waitingListElement := &systemSmartContracts.ElementInList{}
	_ = marshaller.Unmarshal(waitingListElement, marshaledData)

	return waitingListElement
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
					},
				},
				MinGasPrice:      minGasPrice,
				GasPerDataByte:   "1",
				GasPriceModifier: 1.0,
			},
		},
		PenalizedTooMuchGasEnableEpoch: 0,
		EpochNotifier:                  &epochNotifier.EpochNotifierStub{},
		BuiltInFunctionsCostHandler:    &mock.BuiltInCostHandlerStub{},
	}
	economicsData, _ := economicsHandler.NewEconomicsData(argsNewEconomicsData)
	return economicsData
}
