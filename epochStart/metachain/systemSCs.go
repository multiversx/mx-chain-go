package metachain

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"sort"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/display"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// ArgsNewEpochStartSystemSCProcessing defines the arguments structure for the end of epoch system sc processor
type ArgsNewEpochStartSystemSCProcessing struct {
	SystemVM             vmcommon.VMExecutionHandler
	UserAccountsDB       state.AccountsAdapter
	PeerAccountsDB       state.AccountsAdapter
	Marshalizer          marshal.Marshalizer
	StartRating          uint32
	ValidatorInfoCreator epochStart.ValidatorInfoCreator
	ChanceComputer       nodesCoordinator.ChanceComputer
	ShardCoordinator     sharding.Coordinator
	EpochConfig          config.EpochConfig

	EndOfEpochCallerAddress []byte
	StakingSCAddress        []byte
	MaxNodesEnableConfig    []config.MaxNodesChangeConfig
	ESDTOwnerAddressBytes   []byte

	GenesisNodesConfig  sharding.GenesisNodesSetupHandler
	EpochNotifier       process.EpochNotifier
	NodesConfigProvider epochStart.NodesConfigProvider
	StakingDataProvider epochStart.StakingDataProvider
}

type systemSCProcessor struct {
	*legacySystemSCProcessor

	governanceEnableEpoch    uint32
	builtInOnMetaEnableEpoch uint32
	stakingV4EnableEpoch     uint32

	flagGovernanceEnabled    atomic.Flag
	flagBuiltInOnMetaEnabled atomic.Flag
	flagInitStakingV4Enabled atomic.Flag
	flagStakingV4Enabled     atomic.Flag
}

// NewSystemSCProcessor creates the end of epoch system smart contract processor
func NewSystemSCProcessor(args ArgsNewEpochStartSystemSCProcessing) (*systemSCProcessor, error) {
	if check.IfNil(args.EpochNotifier) {
		return nil, epochStart.ErrNilEpochStartNotifier
	}

	legacy, err := newLegacySystemSCProcessor(args)
	if err != nil {
		return nil, err
	}

	s := &systemSCProcessor{
		legacySystemSCProcessor:  legacy,
		governanceEnableEpoch:    args.EpochConfig.EnableEpochs.GovernanceEnableEpoch,
		builtInOnMetaEnableEpoch: args.EpochConfig.EnableEpochs.BuiltInFunctionOnMetaEnableEpoch,
		stakingV4EnableEpoch:     args.EpochConfig.EnableEpochs.StakingV4EnableEpoch,
	}

	log.Debug("systemSC: enable epoch for governanceV2 init", "epoch", s.governanceEnableEpoch)
	log.Debug("systemSC: enable epoch for create NFT on meta", "epoch", s.builtInOnMetaEnableEpoch)
	log.Debug("systemSC: enable epoch for staking v4", "epoch", s.stakingV4EnableEpoch)

	args.EpochNotifier.RegisterNotifyHandler(s)
	return s, nil
}

// ProcessSystemSmartContract does all the processing at end of epoch in case of system smart contract
func (s *systemSCProcessor) ProcessSystemSmartContract(
	validatorsInfoMap state.ShardValidatorsInfoMapHandler,
	header data.HeaderHandler,
) error {
	err := s.processLegacy(validatorsInfoMap, header.GetNonce(), header.GetEpoch())
	if err != nil {
		return err
	}
	return s.processWithNewFlags(validatorsInfoMap, header)
}

func (s *systemSCProcessor) processWithNewFlags(
	validatorsInfoMap state.ShardValidatorsInfoMapHandler,
	header data.HeaderHandler,
) error {
	if s.flagGovernanceEnabled.IsSet() {
		err := s.updateToGovernanceV2()
		if err != nil {
			return err
		}
	}

	if s.flagBuiltInOnMetaEnabled.IsSet() {
		tokenID, err := s.initTokenOnMeta()
		if err != nil {
			return err
		}

		err = s.initLiquidStakingSC(tokenID)
		if err != nil {
			return err
		}
	}

	if s.flagInitStakingV4Enabled.IsSet() {
		err := s.stakeNodesFromQueue(validatorsInfoMap, math.MaxUint32, header.GetNonce(), common.AuctionList)
		if err != nil {
			return err
		}
	}

	if s.flagStakingV4Enabled.IsSet() {
		err := s.prepareStakingDataForAllNodes(validatorsInfoMap)
		if err != nil {
			return err
		}

		err = s.fillStakingDataForNonEligible(validatorsInfoMap)
		if err != nil {
			return err
		}

		err = s.unStakeNodesWithNotEnoughFundsWithStakingV4(validatorsInfoMap, header.GetEpoch())
		if err != nil {
			return err
		}

		err = s.selectNodesFromAuctionList(validatorsInfoMap, header.GetPrevRandSeed())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *systemSCProcessor) unStakeNodesWithNotEnoughFundsWithStakingV4(
	validatorsInfoMap state.ShardValidatorsInfoMapHandler,
	epoch uint32,
) error {
	nodesToUnStake, mapOwnersKeys, err := s.stakingDataProvider.ComputeUnQualifiedNodes(validatorsInfoMap)
	if err != nil {
		return err
	}

	log.Debug("unStake nodes with not enough funds", "num", len(nodesToUnStake))
	for _, blsKey := range nodesToUnStake {
		log.Debug("unStake at end of epoch for node", "blsKey", blsKey)
		err = s.unStakeOneNode(blsKey, epoch)
		if err != nil {
			return err
		}

		validatorInfo := validatorsInfoMap.GetValidator(blsKey)
		if validatorInfo == nil {
			return fmt.Errorf(
				"%w in systemSCProcessor.unStakeNodesWithNotEnoughFundsWithStakingV4 because validator might be in additional queue after staking v4",
				epochStart.ErrNilValidatorInfo)
		}

		validatorLeaving := validatorInfo.ShallowClone()
		validatorLeaving.SetList(string(common.LeavingList))
		err = validatorsInfoMap.Replace(validatorInfo, validatorLeaving)
		if err != nil {
			return err
		}
	}

	return s.updateDelegationContracts(mapOwnersKeys)
}

// TODO: Staking v4: perhaps create a subcomponent which handles selection, which would be also very useful in tests
func (s *systemSCProcessor) selectNodesFromAuctionList(validatorsInfoMap state.ShardValidatorsInfoMapHandler, randomness []byte) error {
	auctionList, currNumOfValidators := getAuctionListAndNumOfValidators(validatorsInfoMap)
	numOfShuffledNodes := s.currentNodesEnableConfig.NodesToShufflePerShard * (s.shardCoordinator.NumberOfShards() + 1)

	numOfValidatorsAfterShuffling, err := safeSub(currNumOfValidators, numOfShuffledNodes)
	if err != nil {
		log.Warn(fmt.Sprintf("%v when trying to compute numOfValidatorsAfterShuffling = %v - %v (currNumOfValidators - numOfShuffledNodes)",
			err,
			currNumOfValidators,
			numOfShuffledNodes,
		))
		numOfValidatorsAfterShuffling = 0
	}

	availableSlots, err := safeSub(s.maxNodes, numOfValidatorsAfterShuffling)
	if availableSlots == 0 || err != nil {
		log.Info(fmt.Sprintf("%v or zero value when trying to compute availableSlots = %v - %v (maxNodes - numOfValidatorsAfterShuffling); skip selecting nodes from auction list",
			err,
			s.maxNodes,
			numOfValidatorsAfterShuffling,
		))
		return nil
	}

	auctionListSize := uint32(len(auctionList))
	log.Info("systemSCProcessor.selectNodesFromAuctionList",
		"max nodes", s.maxNodes,
		"current number of validators", currNumOfValidators,
		"num of nodes which will be shuffled out", numOfShuffledNodes,
		"num of validators after shuffling", numOfValidatorsAfterShuffling,
		"auction list size", auctionListSize,
		fmt.Sprintf("available slots (%v -%v)", s.maxNodes, numOfValidatorsAfterShuffling), availableSlots,
	)

	err = s.sortAuctionList(auctionList, randomness)
	if err != nil {
		return err
	}

	numOfAvailableNodeSlots := core.MinUint32(auctionListSize, availableSlots)
	s.displayAuctionList(auctionList, numOfAvailableNodeSlots)

	for i := uint32(0); i < numOfAvailableNodeSlots; i++ {
		newNode := auctionList[i]
		newNode.SetList(string(common.SelectedFromAuctionList))
		err = validatorsInfoMap.Replace(auctionList[i], newNode)
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: Move this in elrond-go-core
func safeSub(a, b uint32) (uint32, error) {
	if a < b {
		return 0, core.ErrSubtractionOverflow
	}
	return a - b, nil
}

func getAuctionListAndNumOfValidators(validatorsInfoMap state.ShardValidatorsInfoMapHandler) ([]state.ValidatorInfoHandler, uint32) {
	auctionList := make([]state.ValidatorInfoHandler, 0)
	numOfValidators := uint32(0)

	for _, validator := range validatorsInfoMap.GetAllValidatorsInfo() {
		if validator.GetList() == string(common.AuctionList) {
			auctionList = append(auctionList, validator)
			continue
		}
		if isValidator(validator) {
			numOfValidators++
		}
	}

	return auctionList, numOfValidators
}

func (s *systemSCProcessor) sortAuctionList(auctionList []state.ValidatorInfoHandler, randomness []byte) error {
	if len(auctionList) == 0 {
		return nil
	}

	validatorTopUpMap, err := s.getValidatorTopUpMap(auctionList)
	if err != nil {
		return fmt.Errorf("%w: %v", epochStart.ErrSortAuctionList, err)
	}

	pubKeyLen := len(auctionList[0].GetPublicKey())
	normRandomness := calcNormRand(randomness, pubKeyLen)
	sort.SliceStable(auctionList, func(i, j int) bool {
		pubKey1 := auctionList[i].GetPublicKey()
		pubKey2 := auctionList[j].GetPublicKey()

		nodeTopUpPubKey1 := validatorTopUpMap[string(pubKey1)]
		nodeTopUpPubKey2 := validatorTopUpMap[string(pubKey2)]

		if nodeTopUpPubKey1.Cmp(nodeTopUpPubKey2) == 0 {
			return compareByXORWithRandomness(pubKey1, pubKey2, normRandomness)
		}

		return nodeTopUpPubKey1.Cmp(nodeTopUpPubKey2) > 0
	})

	return nil
}

func (s *systemSCProcessor) getValidatorTopUpMap(validators []state.ValidatorInfoHandler) (map[string]*big.Int, error) {
	ret := make(map[string]*big.Int, len(validators))

	for _, validator := range validators {
		pubKey := validator.GetPublicKey()
		topUp, err := s.stakingDataProvider.GetNodeStakedTopUp(pubKey)
		if err != nil {
			return nil, fmt.Errorf("%w when trying to get top up per node for %s", err, hex.EncodeToString(pubKey))
		}

		ret[string(pubKey)] = topUp
	}

	return ret, nil
}

func calcNormRand(randomness []byte, expectedLen int) []byte {
	rand := randomness
	randLen := len(rand)

	if expectedLen > randLen {
		repeatedCt := expectedLen/randLen + 1 // todo: fix possible div by 0
		rand = bytes.Repeat(randomness, repeatedCt)
	}

	rand = rand[:expectedLen]
	return rand
}

func compareByXORWithRandomness(pubKey1, pubKey2, randomness []byte) bool {
	xorLen := len(randomness)

	key1Xor := make([]byte, xorLen)
	key2Xor := make([]byte, xorLen)

	for idx := 0; idx < xorLen; idx++ {
		key1Xor[idx] = pubKey1[idx] ^ randomness[idx]
		key2Xor[idx] = pubKey2[idx] ^ randomness[idx]
	}

	return bytes.Compare(key1Xor, key2Xor) == 1
}

func (s *systemSCProcessor) displayAuctionList(auctionList []state.ValidatorInfoHandler, numOfSelectedNodes uint32) {
	if log.GetLevel() > logger.LogDebug {
		return
	}

	tableHeader := []string{"Owner", "Registered key", "TopUp per node"}
	lines := make([]*display.LineData, 0, len(auctionList))
	horizontalLine := false
	for idx, validator := range auctionList {
		pubKey := validator.GetPublicKey()

		owner, err := s.stakingDataProvider.GetBlsKeyOwner(pubKey)
		log.LogIfError(err)

		topUp, err := s.stakingDataProvider.GetNodeStakedTopUp(pubKey)
		log.LogIfError(err)

		horizontalLine = uint32(idx) == numOfSelectedNodes-1
		line := display.NewLineData(horizontalLine, []string{
			hex.EncodeToString([]byte(owner)),
			hex.EncodeToString(pubKey),
			topUp.String(),
		})
		lines = append(lines, line)
	}

	table, err := display.CreateTableString(tableHeader, lines)
	if err != nil {
		log.Error("could not create table", "error", err)
		return
	}

	message := fmt.Sprintf("Auction list\n%s", table)
	log.Debug(message)
}

func (s *systemSCProcessor) prepareStakingDataForAllNodes(validatorsInfoMap state.ShardValidatorsInfoMapHandler) error {
	allNodes := s.getAllNodeKeys(validatorsInfoMap)
	return s.prepareStakingData(allNodes)
}

func (s *systemSCProcessor) getAllNodeKeys(
	validatorsInfo state.ShardValidatorsInfoMapHandler,
) map[uint32][][]byte {
	nodeKeys := make(map[uint32][][]byte)
	for shardID, validatorsInfoSlice := range validatorsInfo.GetShardValidatorsInfoMap() {
		nodeKeys[shardID] = make([][]byte, 0, s.nodesConfigProvider.ConsensusGroupSize(shardID))
		for _, validatorInfo := range validatorsInfoSlice {
			nodeKeys[shardID] = append(nodeKeys[shardID], validatorInfo.GetPublicKey())
		}
	}

	return nodeKeys
}

func (s *systemSCProcessor) updateToGovernanceV2() error {
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: vm.GovernanceSCAddress,
			CallValue:  big.NewInt(0),
			Arguments:  [][]byte{},
		},
		RecipientAddr: vm.GovernanceSCAddress,
		Function:      "initV2",
	}
	vmOutput, errRun := s.systemVM.RunSmartContractCall(vmInput)
	if errRun != nil {
		return fmt.Errorf("%w when updating to governanceV2", errRun)
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return fmt.Errorf("got return code %s when updating to governanceV2", vmOutput.ReturnCode)
	}

	err := s.processSCOutputAccounts(vmOutput)
	if err != nil {
		return err
	}

	return nil
}

func (s *systemSCProcessor) initTokenOnMeta() ([]byte, error) {
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  vm.ESDTSCAddress,
			CallValue:   big.NewInt(0),
			Arguments:   [][]byte{},
			GasProvided: math.MaxUint64,
		},
		RecipientAddr: vm.ESDTSCAddress,
		Function:      "initDelegationESDTOnMeta",
	}
	vmOutput, errRun := s.systemVM.RunSmartContractCall(vmInput)
	if errRun != nil {
		return nil, fmt.Errorf("%w when setting up NFTs on metachain", errRun)
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return nil, fmt.Errorf("got return code %s, return message %s when setting up NFTs on metachain", vmOutput.ReturnCode, vmOutput.ReturnMessage)
	}
	if len(vmOutput.ReturnData) != 1 {
		return nil, fmt.Errorf("invalid return data on initDelegationESDTOnMeta")
	}

	err := s.processSCOutputAccounts(vmOutput)
	if err != nil {
		return nil, err
	}

	return vmOutput.ReturnData[0], nil
}

func (s *systemSCProcessor) initLiquidStakingSC(tokenID []byte) error {
	codeMetaData := &vmcommon.CodeMetadata{
		Upgradeable: false,
		Payable:     false,
		Readable:    true,
	}

	vmInput := &vmcommon.ContractCreateInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: vm.LiquidStakingSCAddress,
			Arguments:  [][]byte{tokenID},
			CallValue:  big.NewInt(0),
		},
		ContractCode:         vm.LiquidStakingSCAddress,
		ContractCodeMetadata: codeMetaData.ToBytes(),
	}

	vmOutput, err := s.systemVM.RunSmartContractCreate(vmInput)
	if err != nil {
		return err
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return epochStart.ErrCouldNotInitLiquidStakingSystemSC
	}

	err = s.processSCOutputAccounts(vmOutput)
	if err != nil {
		return err
	}

	err = s.updateSystemSCContractsCode(vmInput.ContractCodeMetadata)
	if err != nil {
		return err
	}

	return nil
}

// IsInterfaceNil returns true if underlying object is nil
func (s *systemSCProcessor) IsInterfaceNil() bool {
	return s == nil
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (s *systemSCProcessor) EpochConfirmed(epoch uint32, _ uint64) {
	s.legacyEpochConfirmed(epoch)

	s.flagGovernanceEnabled.SetValue(epoch == s.governanceEnableEpoch)
	log.Debug("systemProcessor: governanceV2", "enabled", s.flagGovernanceEnabled.IsSet())

	s.flagBuiltInOnMetaEnabled.SetValue(epoch == s.builtInOnMetaEnableEpoch)
	log.Debug("systemProcessor: create NFT on meta", "enabled", s.flagBuiltInOnMetaEnabled.IsSet())

	s.flagInitStakingV4Enabled.SetValue(epoch == s.stakingV4InitEnableEpoch)
	log.Debug("systemProcessor: init staking v4", "enabled", s.flagInitStakingV4Enabled.IsSet())

	s.flagStakingV4Enabled.SetValue(epoch >= s.stakingV4EnableEpoch)
	log.Debug("systemProcessor: staking v4", "enabled", s.flagStakingV4Enabled.IsSet())
}
