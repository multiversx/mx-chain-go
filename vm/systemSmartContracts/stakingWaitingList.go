package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strconv"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

const waitingListHeadKey = "waitingList"
const waitingElementPrefix = "w_"

type waitingListReturnData struct {
	blsKeys         [][]byte
	stakedDataList  []*StakedDataV2_0
	lastKey         []byte
	afterLastJailed bool
}

func (s *stakingSC) processStakeV1(blsKey []byte, registrationData *StakedDataV2_0, addFirst bool) error {
	if registrationData.Staked {
		return nil
	}

	registrationData.RegisterNonce = s.eei.BlockChainHook().CurrentNonce()
	if !s.canStake() {
		s.eei.AddReturnMessage(fmt.Sprintf("staking is full key put into waiting list %s", hex.EncodeToString(blsKey)))
		err := s.addToWaitingList(blsKey, addFirst)
		if err != nil {
			s.eei.AddReturnMessage("error while adding to waiting")
			return err
		}
		registrationData.Waiting = true
		s.eei.Finish([]byte{waiting})
		return nil
	}

	err := s.removeFromWaitingList(blsKey)
	if err != nil {
		s.eei.AddReturnMessage("error while removing from waiting")
		return err
	}

	s.addToStakedNodes(1)
	s.activeStakingFor(registrationData)

	return nil
}

func (s *stakingSC) unStakeV1(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	registrationData, retCode := s.checkUnStakeArgs(args)
	if retCode != vmcommon.Ok {
		return retCode
	}

	var err error
	if !registrationData.Staked {
		registrationData.Waiting = false
		err = s.removeFromWaitingList(args.Arguments[0])
		if err != nil {
			s.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}
		err = s.saveStakingData(args.Arguments[0], registrationData)
		if err != nil {
			s.eei.AddReturnMessage("cannot save staking data: error " + err.Error())
			return vmcommon.UserError
		}

		return vmcommon.Ok
	}

	addOneFromQueue := !s.enableEpochsHandler.IsFlagEnabled(common.CorrectLastUnJailedFlag) || s.canStakeIfOneRemoved()
	if addOneFromQueue {
		_, err = s.moveFirstFromWaitingToStaked()
		if err != nil {
			s.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}
	}

	return s.tryUnStake(args.Arguments[0], registrationData)
}

func (s *stakingSC) moveFirstFromWaitingToStakedIfNeeded(blsKey []byte) (bool, error) {
	waitingElementKey := createWaitingListKey(blsKey)
	_, err := s.getWaitingListElement(waitingElementKey)
	if err == nil {
		// node in waiting - remove from it - and that's it
		return false, s.removeFromWaitingList(blsKey)
	}

	return s.moveFirstFromWaitingToStaked()
}

func (s *stakingSC) moveFirstFromWaitingToStaked() (bool, error) {
	waitingList, err := s.getWaitingListHead()
	if err != nil {
		return false, err
	}
	if waitingList.Length == 0 {
		return false, nil
	}
	elementInList, err := s.getWaitingListElement(waitingList.FirstKey)
	if err != nil {
		return false, err
	}
	err = s.removeFromWaitingList(elementInList.BLSPublicKey)
	if err != nil {
		return false, err
	}

	nodeData, err := s.getOrCreateRegisteredData(elementInList.BLSPublicKey)
	if err != nil {
		return false, err
	}
	if len(nodeData.RewardAddress) == 0 || nodeData.Staked {
		return false, vm.ErrInvalidWaitingList
	}

	nodeData.Waiting = false
	nodeData.Staked = true
	nodeData.RegisterNonce = s.eei.BlockChainHook().CurrentNonce()
	nodeData.StakedNonce = s.eei.BlockChainHook().CurrentNonce()
	nodeData.UnStakedNonce = 0
	nodeData.UnStakedEpoch = common.DefaultUnstakedEpoch

	s.addToStakedNodes(1)
	return true, s.saveStakingData(elementInList.BLSPublicKey, nodeData)
}

func (s *stakingSC) addToWaitingList(blsKey []byte, addJailed bool) error {
	inWaitingListKey := createWaitingListKey(blsKey)
	marshaledData := s.eei.GetStorage(inWaitingListKey)
	if len(marshaledData) != 0 {
		return nil
	}

	waitingList, err := s.getWaitingListHead()
	if err != nil {
		return err
	}

	waitingList.Length += 1
	if waitingList.Length == 1 {
		return s.startWaitingList(waitingList, addJailed, blsKey)
	}

	if addJailed {
		return s.insertAfterLastJailed(waitingList, blsKey)
	}

	return s.addToEndOfTheList(waitingList, blsKey)
}

func (s *stakingSC) startWaitingList(
	waitingList *WaitingList,
	addJailed bool,
	blsKey []byte,
) error {
	inWaitingListKey := createWaitingListKey(blsKey)
	waitingList.FirstKey = inWaitingListKey
	waitingList.LastKey = inWaitingListKey
	if addJailed {
		waitingList.LastJailedKey = inWaitingListKey
	}

	elementInWaiting := &ElementInList{
		BLSPublicKey: blsKey,
		PreviousKey:  waitingList.LastKey,
		NextKey:      make([]byte, 0),
	}
	return s.saveElementAndList(inWaitingListKey, elementInWaiting, waitingList)
}

func (s *stakingSC) addToEndOfTheList(waitingList *WaitingList, blsKey []byte) error {
	inWaitingListKey := createWaitingListKey(blsKey)
	oldLastKey := make([]byte, len(waitingList.LastKey))
	copy(oldLastKey, waitingList.LastKey)

	lastElement, err := s.getWaitingListElement(waitingList.LastKey)
	if err != nil {
		return err
	}
	lastElement.NextKey = inWaitingListKey
	elementInWaiting := &ElementInList{
		BLSPublicKey: blsKey,
		PreviousKey:  oldLastKey,
		NextKey:      make([]byte, 0),
	}

	err = s.saveWaitingListElement(oldLastKey, lastElement)
	if err != nil {
		return err
	}

	waitingList.LastKey = inWaitingListKey
	return s.saveElementAndList(inWaitingListKey, elementInWaiting, waitingList)
}

func (s *stakingSC) insertAfterLastJailed(
	waitingList *WaitingList,
	blsKey []byte,
) error {
	inWaitingListKey := createWaitingListKey(blsKey)
	if len(waitingList.LastJailedKey) == 0 {
		previousFirstKey := make([]byte, len(waitingList.FirstKey))
		copy(previousFirstKey, waitingList.FirstKey)
		waitingList.FirstKey = inWaitingListKey
		waitingList.LastJailedKey = inWaitingListKey
		elementInWaiting := &ElementInList{
			BLSPublicKey: blsKey,
			PreviousKey:  inWaitingListKey,
			NextKey:      previousFirstKey,
		}

		if s.enableEpochsHandler.IsFlagEnabled(common.CorrectFirstQueuedFlag) && len(previousFirstKey) > 0 {
			previousFirstElement, err := s.getWaitingListElement(previousFirstKey)
			if err != nil {
				return err
			}
			previousFirstElement.PreviousKey = inWaitingListKey
			err = s.saveWaitingListElement(previousFirstKey, previousFirstElement)
			if err != nil {
				return err
			}
		}

		return s.saveElementAndList(inWaitingListKey, elementInWaiting, waitingList)
	}

	lastJailedElement, err := s.getWaitingListElement(waitingList.LastJailedKey)
	if err != nil {
		return err
	}

	if bytes.Equal(waitingList.LastKey, waitingList.LastJailedKey) {
		waitingList.LastJailedKey = inWaitingListKey
		return s.addToEndOfTheList(waitingList, blsKey)
	}

	firstNonJailedElement, err := s.getWaitingListElement(lastJailedElement.NextKey)
	if err != nil {
		return err
	}

	elementInWaiting := &ElementInList{
		BLSPublicKey: blsKey,
		PreviousKey:  make([]byte, len(inWaitingListKey)),
		NextKey:      make([]byte, len(inWaitingListKey)),
	}
	copy(elementInWaiting.PreviousKey, waitingList.LastJailedKey)
	copy(elementInWaiting.NextKey, lastJailedElement.NextKey)

	lastJailedElement.NextKey = inWaitingListKey
	firstNonJailedElement.PreviousKey = inWaitingListKey
	waitingList.LastJailedKey = inWaitingListKey

	err = s.saveWaitingListElement(elementInWaiting.PreviousKey, lastJailedElement)
	if err != nil {
		return err
	}
	err = s.saveWaitingListElement(elementInWaiting.NextKey, firstNonJailedElement)
	if err != nil {
		return err
	}
	err = s.saveWaitingListElement(inWaitingListKey, elementInWaiting)
	if err != nil {
		return err
	}
	return s.saveWaitingListHead(waitingList)
}

func (s *stakingSC) saveElementAndList(key []byte, element *ElementInList, waitingList *WaitingList) error {
	err := s.saveWaitingListElement(key, element)
	if err != nil {
		return err
	}

	return s.saveWaitingListHead(waitingList)
}

func (s *stakingSC) removeFromWaitingList(blsKey []byte) error {
	inWaitingListKey := createWaitingListKey(blsKey)
	marshaledData := s.eei.GetStorage(inWaitingListKey)
	if len(marshaledData) == 0 {
		return nil
	}
	s.eei.SetStorage(inWaitingListKey, nil)

	elementToRemove := &ElementInList{}
	err := s.marshalizer.Unmarshal(elementToRemove, marshaledData)
	if err != nil {
		return err
	}

	waitingList, err := s.getWaitingListHead()
	if err != nil {
		return err
	}
	if waitingList.Length == 0 {
		return vm.ErrInvalidWaitingList
	}
	waitingList.Length -= 1
	if waitingList.Length == 0 {
		s.eei.SetStorage([]byte(waitingListHeadKey), nil)
		return nil
	}

	// remove the first element
	isFirstElementBeforeFix := !s.enableEpochsHandler.IsFlagEnabled(common.CorrectFirstQueuedFlag) && bytes.Equal(elementToRemove.PreviousKey, inWaitingListKey)
	isFirstElementAfterFix := s.enableEpochsHandler.IsFlagEnabled(common.CorrectFirstQueuedFlag) && bytes.Equal(waitingList.FirstKey, inWaitingListKey)
	if isFirstElementBeforeFix || isFirstElementAfterFix {
		if bytes.Equal(inWaitingListKey, waitingList.LastJailedKey) {
			waitingList.LastJailedKey = make([]byte, 0)
		}

		nextElement, errGet := s.getWaitingListElement(elementToRemove.NextKey)
		if errGet != nil {
			return errGet
		}

		nextElement.PreviousKey = elementToRemove.NextKey
		waitingList.FirstKey = elementToRemove.NextKey
		return s.saveElementAndList(elementToRemove.NextKey, nextElement, waitingList)
	}

	if !s.enableEpochsHandler.IsFlagEnabled(common.CorrectLastUnJailedFlag) || bytes.Equal(inWaitingListKey, waitingList.LastJailedKey) {
		waitingList.LastJailedKey = make([]byte, len(elementToRemove.PreviousKey))
		copy(waitingList.LastJailedKey, elementToRemove.PreviousKey)
	}

	previousElement, _ := s.getWaitingListElement(elementToRemove.PreviousKey)
	// search the other way around for the element in front
	if s.enableEpochsHandler.IsFlagEnabled(common.CorrectFirstQueuedFlag) && previousElement == nil {
		previousElement, err = s.searchPreviousFromHead(waitingList, inWaitingListKey, elementToRemove)
		if err != nil {
			return err
		}
	}
	if previousElement == nil {
		previousElement, err = s.getWaitingListElement(elementToRemove.PreviousKey)
		if err != nil {
			return err
		}
	}
	if len(elementToRemove.NextKey) == 0 {
		waitingList.LastKey = elementToRemove.PreviousKey
		previousElement.NextKey = make([]byte, 0)
		return s.saveElementAndList(elementToRemove.PreviousKey, previousElement, waitingList)
	}

	nextElement, err := s.getWaitingListElement(elementToRemove.NextKey)
	if err != nil {
		return err
	}

	nextElement.PreviousKey = elementToRemove.PreviousKey
	previousElement.NextKey = elementToRemove.NextKey

	err = s.saveWaitingListElement(elementToRemove.NextKey, nextElement)
	if err != nil {
		return err
	}
	return s.saveElementAndList(elementToRemove.PreviousKey, previousElement, waitingList)
}

func (s *stakingSC) searchPreviousFromHead(waitingList *WaitingList, inWaitingListKey []byte, elementToRemove *ElementInList) (*ElementInList, error) {
	var previousElement *ElementInList
	index := uint32(1)
	nextKey := make([]byte, len(waitingList.FirstKey))
	copy(nextKey, waitingList.FirstKey)
	for len(nextKey) != 0 && index <= waitingList.Length {
		element, errGet := s.getWaitingListElement(nextKey)
		if errGet != nil {
			return nil, errGet
		}

		if bytes.Equal(inWaitingListKey, element.NextKey) {
			previousElement = element
			elementToRemove.PreviousKey = createWaitingListKey(previousElement.BLSPublicKey)
			return previousElement, nil
		}

		nextKey = make([]byte, len(element.NextKey))
		if len(element.NextKey) == 0 {
			break
		}
		index++
		copy(nextKey, element.NextKey)
	}
	return nil, vm.ErrElementNotFound
}

func (s *stakingSC) getWaitingListElement(key []byte) (*ElementInList, error) {
	marshaledData := s.eei.GetStorage(key)
	if len(marshaledData) == 0 {
		return nil, vm.ErrElementNotFound
	}

	element := &ElementInList{}
	err := s.marshalizer.Unmarshal(element, marshaledData)
	if err != nil {
		return nil, err
	}

	return element, nil
}

func (s *stakingSC) saveWaitingListElement(key []byte, element *ElementInList) error {
	marshaledData, err := s.marshalizer.Marshal(element)
	if err != nil {
		return err
	}

	s.eei.SetStorage(key, marshaledData)
	return nil
}

func (s *stakingSC) getWaitingListHead() (*WaitingList, error) {
	waitingList := &WaitingList{
		FirstKey:      make([]byte, 0),
		LastKey:       make([]byte, 0),
		Length:        0,
		LastJailedKey: make([]byte, 0),
	}
	marshaledData := s.eei.GetStorage([]byte(waitingListHeadKey))
	if len(marshaledData) == 0 {
		return waitingList, nil
	}

	err := s.marshalizer.Unmarshal(waitingList, marshaledData)
	if err != nil {
		return nil, err
	}

	return waitingList, nil
}

func (s *stakingSC) saveWaitingListHead(waitingList *WaitingList) error {
	marshaledData, err := s.marshalizer.Marshal(waitingList)
	if err != nil {
		return err
	}

	s.eei.SetStorage([]byte(waitingListHeadKey), marshaledData)
	return nil
}

func createWaitingListKey(blsKey []byte) []byte {
	return []byte(waitingElementPrefix + string(blsKey))
}

func (s *stakingSC) switchJailedWithWaiting(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if s.enableEpochsHandler.IsFlagEnabled(common.StakingV4StartedFlag) && !s.enableEpochsHandler.IsFlagEnabled(common.StakingV4Step1Flag) {
		s.eei.AddReturnMessage(vm.ErrWaitingListDisabled.Error())
		return vmcommon.UserError
	}
	if !bytes.Equal(args.CallerAddr, s.endOfEpochAccessAddr) {
		s.eei.AddReturnMessage("switchJailedWithWaiting function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		return vmcommon.UserError
	}

	blsKey := args.Arguments[0]
	registrationData, err := s.getOrCreateRegisteredData(blsKey)
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if len(registrationData.RewardAddress) == 0 {
		s.eei.AddReturnMessage("no need to jail as not a validator")
		return vmcommon.UserError
	}
	if !registrationData.Staked {
		s.eei.AddReturnMessage("no need to jail as not a validator")
		return vmcommon.UserError
	}
	if registrationData.Jailed {
		s.eei.AddReturnMessage(vm.ErrBLSPublicKeyAlreadyJailed.Error())
		return vmcommon.UserError
	}
	switched, err := s.moveFirstFromWaitingToStakedIfNeeded(blsKey)
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	registrationData.NumJailed++
	registrationData.Jailed = true
	registrationData.JailedNonce = s.eei.BlockChainHook().CurrentNonce()

	if !switched && !s.enableEpochsHandler.IsFlagEnabled(common.CorrectJailedNotUnStakedEmptyQueueFlag) {
		s.eei.AddReturnMessage("did not switch as nobody in waiting, but jailed")
	} else {
		s.tryRemoveJailedNodeFromStaked(registrationData)
	}

	err = s.saveStakingData(blsKey, registrationData)
	if err != nil {
		s.eei.AddReturnMessage("cannot save staking data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (s *stakingSC) getWaitingListIndex(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if s.enableEpochsHandler.IsFlagEnabled(common.StakingV4StartedFlag) {
		s.eei.AddReturnMessage(vm.ErrWaitingListDisabled.Error())
		return vmcommon.UserError
	}
	if !bytes.Equal(args.CallerAddr, s.stakeAccessAddr) {
		s.eei.AddReturnMessage("this is only a view function")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		s.eei.AddReturnMessage("number of arguments must be equal to 1")
		return vmcommon.UserError
	}

	waitingElementKey := createWaitingListKey(args.Arguments[0])
	_, err := s.getWaitingListElement(waitingElementKey)
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	waitingListHead, err := s.getWaitingListHead()
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	if bytes.Equal(waitingElementKey, waitingListHead.FirstKey) {
		s.eei.Finish([]byte(strconv.Itoa(1)))
		return vmcommon.Ok
	}
	if bytes.Equal(waitingElementKey, waitingListHead.LastKey) {
		s.eei.Finish([]byte(strconv.Itoa(int(waitingListHead.Length))))
		return vmcommon.Ok
	}

	prevElement, err := s.getWaitingListElement(waitingListHead.FirstKey)
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	index := uint32(2)
	nextKey := make([]byte, len(waitingElementKey))
	copy(nextKey, prevElement.NextKey)
	for len(nextKey) != 0 && index <= waitingListHead.Length {
		if bytes.Equal(nextKey, waitingElementKey) {
			s.eei.Finish([]byte(strconv.Itoa(int(index))))
			return vmcommon.Ok
		}

		prevElement, err = s.getWaitingListElement(nextKey)
		if err != nil {
			s.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}

		if len(prevElement.NextKey) == 0 {
			break
		}
		index++
		copy(nextKey, prevElement.NextKey)
	}

	s.eei.AddReturnMessage("element in waiting list not found")
	return vmcommon.UserError
}

func (s *stakingSC) getWaitingListSize(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if s.enableEpochsHandler.IsFlagEnabled(common.StakingV4StartedFlag) {
		s.eei.AddReturnMessage(vm.ErrWaitingListDisabled.Error())
		return vmcommon.UserError
	}

	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}

	err := s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.Get)
	if err != nil {
		s.eei.AddReturnMessage("insufficient gas")
		return vmcommon.OutOfGas
	}

	waitingListHead, err := s.getWaitingListHead()
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	s.eei.Finish([]byte(strconv.Itoa(int(waitingListHead.Length))))
	return vmcommon.Ok
}

func (s *stakingSC) getWaitingListRegisterNonceAndRewardAddress(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, s.stakeAccessAddr) {
		s.eei.AddReturnMessage("this is only a view function")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 0 {
		s.eei.AddReturnMessage("number of arguments must be equal to 0")
		return vmcommon.UserError
	}

	waitingListData, err := s.getFirstElementsFromWaitingList(math.MaxUint32)
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if len(waitingListData.stakedDataList) == 0 {
		s.eei.AddReturnMessage("no one in waitingList")
		return vmcommon.UserError
	}

	for index, stakedData := range waitingListData.stakedDataList {
		s.eei.Finish(waitingListData.blsKeys[index])
		s.eei.Finish(stakedData.RewardAddress)
		s.eei.Finish(big.NewInt(int64(stakedData.RegisterNonce)).Bytes())
	}

	return vmcommon.Ok
}

func (s *stakingSC) resetLastUnJailedFromQueue(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !s.enableEpochsHandler.IsFlagEnabled(common.CorrectLastUnJailedFlag) {
		// backward compatibility
		return vmcommon.UserError
	}
	if s.enableEpochsHandler.IsFlagEnabled(common.StakingV4StartedFlag) && !s.enableEpochsHandler.IsFlagEnabled(common.StakingV4Step1Flag) {
		s.eei.AddReturnMessage(vm.ErrWaitingListDisabled.Error())
		return vmcommon.UserError
	}
	if !bytes.Equal(args.CallerAddr, s.endOfEpochAccessAddr) {
		s.eei.AddReturnMessage("stake nodes from waiting list can be called by endOfEpochAccess address only")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 0 {
		s.eei.AddReturnMessage("number of arguments must be equal to 0")
		return vmcommon.UserError
	}

	waitingList, err := s.getWaitingListHead()
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	if len(waitingList.LastJailedKey) == 0 {
		return vmcommon.Ok
	}

	waitingList.LastJailedKey = make([]byte, 0)
	err = s.saveWaitingListHead(waitingList)
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (s *stakingSC) cleanAdditionalQueueNotEnoughFunds(
	waitingListData *waitingListReturnData,
) ([]string, map[string][][]byte, error) {

	listOfOwners := make([]string, 0)
	mapOwnersUnStakedNodes := make(map[string][][]byte)
	mapCheckedOwners := make(map[string]*validatorFundInfo)
	for i := len(waitingListData.blsKeys) - 1; i >= 0; i-- {
		stakedData := waitingListData.stakedDataList[i]
		validatorInfo, err := s.checkValidatorFunds(mapCheckedOwners, stakedData.OwnerAddress, s.stakeValue)
		if err != nil {
			return nil, nil, err
		}
		if validatorInfo.numNodesToUnstake == 0 {
			continue
		}

		validatorInfo.numNodesToUnstake--
		blsKey := waitingListData.blsKeys[i]
		err = s.removeFromWaitingList(blsKey)
		if err != nil {
			return nil, nil, err
		}

		registrationData, err := s.getOrCreateRegisteredData(blsKey)
		if err != nil {
			return nil, nil, err
		}

		registrationData.Staked = false
		registrationData.UnStakedEpoch = s.eei.BlockChainHook().CurrentEpoch()
		registrationData.UnStakedNonce = s.eei.BlockChainHook().CurrentNonce()
		registrationData.Waiting = false

		err = s.saveStakingData(blsKey, registrationData)
		if err != nil {
			return nil, nil, err
		}

		_, alreadyAdded := mapOwnersUnStakedNodes[string(stakedData.OwnerAddress)]
		if !alreadyAdded {
			listOfOwners = append(listOfOwners, string(stakedData.OwnerAddress))
		}

		mapOwnersUnStakedNodes[string(stakedData.OwnerAddress)] = append(mapOwnersUnStakedNodes[string(stakedData.OwnerAddress)], blsKey)
	}

	return listOfOwners, mapOwnersUnStakedNodes, nil
}

func (s *stakingSC) stakeNodesFromQueue(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !s.enableEpochsHandler.IsFlagEnabled(common.StakingV4Step2Flag) {
		s.eei.AddReturnMessage("invalid method to call")
		return vmcommon.UserError
	}
	if s.enableEpochsHandler.IsFlagEnabled(common.StakingV4StartedFlag) && !s.enableEpochsHandler.IsFlagEnabled(common.StakingV4Step1Flag) {
		s.eei.AddReturnMessage(vm.ErrWaitingListDisabled.Error())
		return vmcommon.UserError
	}
	if !bytes.Equal(args.CallerAddr, s.endOfEpochAccessAddr) {
		s.eei.AddReturnMessage("stake nodes from waiting list can be called by endOfEpochAccess address only")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 1 {
		s.eei.AddReturnMessage("number of arguments must be equal to 1")
		return vmcommon.UserError
	}

	numNodesToStake := big.NewInt(0).SetBytes(args.Arguments[0]).Uint64()
	waitingListData, err := s.getFirstElementsFromWaitingList(math.MaxUint32)
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if len(waitingListData.blsKeys) == 0 {
		s.eei.AddReturnMessage("no nodes in queue")
		return vmcommon.Ok
	}

	nodePriceToUse := big.NewInt(0).Set(s.minNodePrice)
	if s.enableEpochsHandler.IsFlagEnabled(common.CorrectLastUnJailedFlag) {
		nodePriceToUse.Set(s.stakeValue)
	}

	stakedNodes := uint64(0)
	mapCheckedOwners := make(map[string]*validatorFundInfo)
	for i, blsKey := range waitingListData.blsKeys {
		stakedData := waitingListData.stakedDataList[i]
		if stakedNodes >= numNodesToStake {
			break
		}

		validatorInfo, errCheck := s.checkValidatorFunds(mapCheckedOwners, stakedData.OwnerAddress, nodePriceToUse)
		if errCheck != nil {
			s.eei.AddReturnMessage(errCheck.Error())
			return vmcommon.UserError
		}
		if validatorInfo.numNodesToUnstake > 0 {
			continue
		}

		s.activeStakingFor(stakedData)
		err = s.saveStakingData(blsKey, stakedData)
		if err != nil {
			s.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}

		// remove from waiting list
		err = s.removeFromWaitingList(blsKey)
		if err != nil {
			s.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}

		stakedNodes++
		// return the change key
		s.eei.Finish(blsKey)
		s.eei.Finish(stakedData.RewardAddress)
	}

	s.addToStakedNodes(int64(stakedNodes))

	return vmcommon.Ok
}

func (s *stakingSC) cleanAdditionalQueue(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !s.enableEpochsHandler.IsFlagEnabled(common.CorrectLastUnJailedFlag) {
		s.eei.AddReturnMessage("invalid method to call")
		return vmcommon.UserError
	}
	if s.enableEpochsHandler.IsFlagEnabled(common.StakingV4StartedFlag) && !s.enableEpochsHandler.IsFlagEnabled(common.StakingV4Step1Flag) {
		s.eei.AddReturnMessage(vm.ErrWaitingListDisabled.Error())
		return vmcommon.UserError
	}
	if !bytes.Equal(args.CallerAddr, s.endOfEpochAccessAddr) {
		s.eei.AddReturnMessage("stake nodes from waiting list can be called by endOfEpochAccess address only")
		return vmcommon.UserError
	}
	if len(args.Arguments) != 0 {
		s.eei.AddReturnMessage("number of arguments must be 0")
		return vmcommon.UserError
	}

	waitingListData, err := s.getFirstElementsFromWaitingList(math.MaxUint32)
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if len(waitingListData.blsKeys) == 0 {
		s.eei.AddReturnMessage("no nodes in queue")
		return vmcommon.Ok
	}

	listOfOwners, mapOwnersAndBLSKeys, err := s.cleanAdditionalQueueNotEnoughFunds(waitingListData)
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	for _, owner := range listOfOwners {
		s.eei.Finish([]byte(owner))
		blsKeys := mapOwnersAndBLSKeys[owner]
		for _, blsKey := range blsKeys {
			s.eei.Finish(blsKey)
		}
	}

	return vmcommon.Ok
}

func (s *stakingSC) getFirstElementsFromWaitingList(numNodes uint32) (*waitingListReturnData, error) {
	waitingListData := &waitingListReturnData{}

	waitingListHead, err := s.getWaitingListHead()
	if err != nil {
		return nil, err
	}
	if waitingListHead.Length == 0 {
		return waitingListData, nil
	}

	blsKeysToStake := make([][]byte, 0)
	stakedDataList := make([]*StakedDataV2_0, 0)
	index := uint32(1)
	nextKey := make([]byte, len(waitingListHead.FirstKey))
	copy(nextKey, waitingListHead.FirstKey)
	for len(nextKey) != 0 && index <= waitingListHead.Length && index <= numNodes {
		element, errGet := s.getWaitingListElement(nextKey)
		if errGet != nil {
			return nil, errGet
		}

		if bytes.Equal(nextKey, waitingListHead.LastJailedKey) {
			waitingListData.afterLastJailed = true
		}

		stakedData, errGet := s.getOrCreateRegisteredData(element.BLSPublicKey)
		if errGet != nil {
			return nil, errGet
		}

		blsKeysToStake = append(blsKeysToStake, element.BLSPublicKey)
		stakedDataList = append(stakedDataList, stakedData)

		if len(element.NextKey) == 0 {
			break
		}
		index++
		copy(nextKey, element.NextKey)
	}

	if numNodes >= waitingListHead.Length && len(blsKeysToStake) != int(waitingListHead.Length) {
		log.Warn("mismatch length on waiting list elements in stakingSC.getFirstElementsFromWaitingList")
	}

	waitingListData.blsKeys = blsKeysToStake
	waitingListData.stakedDataList = stakedDataList
	waitingListData.lastKey = nextKey
	return waitingListData, nil
}

func (s *stakingSC) fixWaitingListQueueSize(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !s.enableEpochsHandler.IsFlagEnabled(common.CorrectFirstQueuedFlag) {
		s.eei.AddReturnMessage("invalid method to call")
		return vmcommon.UserError
	}
	if s.enableEpochsHandler.IsFlagEnabled(common.StakingV4StartedFlag) {
		s.eei.AddReturnMessage(vm.ErrWaitingListDisabled.Error())
		return vmcommon.UserError
	}

	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}

	err := s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.FixWaitingListSize)
	if err != nil {
		s.eei.AddReturnMessage("insufficient gas")
		return vmcommon.OutOfGas
	}

	waitingListHead, err := s.getWaitingListHead()
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	if waitingListHead.Length <= 1 {
		return vmcommon.Ok
	}

	foundLastJailedKey := len(waitingListHead.LastJailedKey) == 0

	index := uint32(1)
	nextKey := make([]byte, len(waitingListHead.FirstKey))
	copy(nextKey, waitingListHead.FirstKey)
	for len(nextKey) != 0 && index <= waitingListHead.Length {
		element, errGet := s.getWaitingListElement(nextKey)
		if errGet != nil {
			s.eei.AddReturnMessage(errGet.Error())
			return vmcommon.UserError
		}

		if bytes.Equal(waitingListHead.LastJailedKey, nextKey) {
			foundLastJailedKey = true
		}

		_, errGet = s.getOrCreateRegisteredData(element.BLSPublicKey)
		if errGet != nil {
			s.eei.AddReturnMessage(errGet.Error())
			return vmcommon.UserError
		}

		if len(element.NextKey) == 0 {
			break
		}
		index++
		copy(nextKey, element.NextKey)
	}

	waitingListHead.Length = index
	waitingListHead.LastKey = nextKey
	if !foundLastJailedKey {
		waitingListHead.LastJailedKey = make([]byte, 0)
	}

	err = s.saveWaitingListHead(waitingListHead)
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (s *stakingSC) addMissingNodeToQueue(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !s.enableEpochsHandler.IsFlagEnabled(common.CorrectFirstQueuedFlag) {
		s.eei.AddReturnMessage("invalid method to call")
		return vmcommon.UserError
	}
	if s.enableEpochsHandler.IsFlagEnabled(common.StakingV4StartedFlag) {
		s.eei.AddReturnMessage(vm.ErrWaitingListDisabled.Error())
		return vmcommon.UserError
	}
	if args.CallValue.Cmp(zero) != 0 {
		s.eei.AddReturnMessage(vm.TransactionValueMustBeZero)
		return vmcommon.UserError
	}
	err := s.eei.UseGas(s.gasCost.MetaChainSystemSCsCost.FixWaitingListSize)
	if err != nil {
		s.eei.AddReturnMessage("insufficient gas")
		return vmcommon.OutOfGas
	}
	if len(args.Arguments) != 1 {
		s.eei.AddReturnMessage("invalid number of arguments")
		return vmcommon.UserError
	}

	blsKey := args.Arguments[0]
	_, err = s.getWaitingListElement(createWaitingListKey(blsKey))
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	waitingListData, err := s.getFirstElementsFromWaitingList(math.MaxUint32)
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	for _, keyInList := range waitingListData.blsKeys {
		if bytes.Equal(keyInList, blsKey) {
			s.eei.AddReturnMessage("key is in queue, not missing")
			return vmcommon.UserError
		}
	}

	waitingList, err := s.getWaitingListHead()
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	waitingList.Length += 1
	if waitingList.Length == 1 {
		err = s.startWaitingList(waitingList, false, blsKey)
		if err != nil {
			s.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}

		return vmcommon.Ok
	}

	err = s.addToEndOfTheList(waitingList, blsKey)
	if err != nil {
		s.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}
