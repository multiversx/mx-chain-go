package trieIterators

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type delegatedListProcessor struct {
	*commonStakingProcessor
	publicKeyConverter core.PubkeyConverter
}

// NewDelegatedListProcessor will create a new instance of delegatedListProc
func NewDelegatedListProcessor(arg ArgTrieIteratorProcessor) (*delegatedListProcessor, error) {
	err := checkArguments(arg)
	if err != nil {
		return nil, err
	}

	return &delegatedListProcessor{
		commonStakingProcessor: &commonStakingProcessor{
			queryService: arg.QueryService,
			accounts:     arg.Accounts,
		},
		publicKeyConverter: arg.PublicKeyConverter,
	}, nil
}

// GetDelegatorsList will return the delegators list
func (dlp *delegatedListProcessor) GetDelegatorsList(ctx context.Context) ([]*api.Delegator, error) {
	dlp.accounts.Lock()
	defer dlp.accounts.Unlock()

	delegationScAddresses, err := dlp.getAllDelegationContractAddresses()
	if err != nil {
		return nil, err
	}

	delegatorsInfo := make(map[string]*api.Delegator)
	for _, delegationSC := range delegationScAddresses {
		err = dlp.getDelegatorsInfo(delegationSC, delegatorsInfo, ctx)
		if err != nil {
			return nil, err
		}
	}

	return dlp.mapToSlice(delegatorsInfo), nil
}

func (dlp *delegatedListProcessor) getAllDelegationContractAddresses() ([][]byte, error) {
	scQuery := &process.SCQuery{
		ScAddress:  vm.DelegationManagerSCAddress,
		FuncName:   "getAllContractAddresses",
		CallerAddr: vm.DelegationManagerSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  make([][]byte, 0),
	}

	vmOutput, err := dlp.queryService.ExecuteQuery(scQuery)
	if err != nil {
		return nil, err
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return nil, fmt.Errorf("%w, return code: %v, message: %s", epochStart.ErrExecutingSystemScCode, vmOutput.ReturnCode, vmOutput.ReturnMessage)
	}

	return vmOutput.ReturnData, nil
}

func (dlp *delegatedListProcessor) getDelegatorsInfo(delegationSC []byte, delegatorsMap map[string]*api.Delegator, ctx context.Context) error {
	delegatorsList, err := dlp.getDelegatorsList(delegationSC, ctx)
	if err != nil {
		return err
	}

	var value *big.Int
	for _, delegatorAddress := range delegatorsList {
		value, err = dlp.getActiveFund(delegationSC, delegatorAddress)
		if err != nil {
			// delegatorAddress byte slice might not represent a real delegator address
			continue
		}

		delegatorInfo, ok := delegatorsMap[string(delegatorAddress)]
		if !ok {
			encodedDelegatorAddress, err := dlp.publicKeyConverter.Encode(delegatorAddress)
			if err != nil {
				return fmt.Errorf("%w encoding the address of the delegator %s", err, hex.EncodeToString(delegatorAddress))
			}

			delegatorInfo = &api.Delegator{
				DelegatorAddress: encodedDelegatorAddress,
				DelegatedTo:      make([]*api.DelegatedValue, 0),
				TotalAsBigInt:    big.NewInt(0),
			}

			delegatorsMap[string(delegatorAddress)] = delegatorInfo
		}

		delegatorInfo.TotalAsBigInt = big.NewInt(0).Add(delegatorInfo.TotalAsBigInt, value)
		delegatorInfo.Total = delegatorInfo.TotalAsBigInt.String()
		delegationSCAddress, err := dlp.publicKeyConverter.Encode(delegationSC)
		if err != nil {
			return fmt.Errorf("%w encoding delegation SC address %s", err, hex.EncodeToString(delegationSC))
		}

		delegatorInfo.DelegatedTo = append(delegatorInfo.DelegatedTo, &api.DelegatedValue{
			DelegationScAddress: delegationSCAddress,
			Value:               value.String(),
		})
	}

	return nil
}

func (dlp *delegatedListProcessor) getDelegatorsList(delegationSC []byte, ctx context.Context) ([][]byte, error) {
	delegatorAccount, err := dlp.getAccount(delegationSC)
	if err != nil {
		return nil, fmt.Errorf("%w for delegationSC %s", err, hex.EncodeToString(delegationSC))
	}

	chLeaves := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	err = delegatorAccount.GetAllLeaves(chLeaves, ctx)
	if err != nil {
		return nil, err
	}

	delegators := make([][]byte, 0)
	for leaf := range chLeaves.LeavesChan {
		leafKey := leaf.Key()
		if len(leafKey) != dlp.publicKeyConverter.Len() {
			continue
		}

		delegators = append(delegators, leafKey)
	}

	err = chLeaves.ErrChan.ReadFromChanNonBlocking()
	if err != nil {
		return nil, err
	}

	if common.IsContextDone(ctx) {
		return nil, ErrTrieOperationsTimeout
	}

	return delegators, nil
}

func (dlp *delegatedListProcessor) getActiveFund(delegationSC []byte, delegator []byte) (*big.Int, error) {
	scQuery := &process.SCQuery{
		ScAddress:  delegationSC,
		FuncName:   "getUserActiveStake",
		CallerAddr: delegationSC,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{delegator},
	}

	vmOutput, err := dlp.queryService.ExecuteQuery(scQuery)
	if err != nil {
		return nil, err
	}
	if vmOutput.ReturnCode != vmcommon.Ok {
		return nil, fmt.Errorf("%w, return code: %v, message: %s", epochStart.ErrExecutingSystemScCode, vmOutput.ReturnCode, vmOutput.ReturnMessage)
	}

	if len(vmOutput.ReturnData) != 1 {
		return nil, fmt.Errorf("%w, getActiveFund function should have returned one value", epochStart.ErrExecutingSystemScCode)
	}

	value := big.NewInt(0).SetBytes(vmOutput.ReturnData[0])

	return value, nil
}

func (dlp *delegatedListProcessor) mapToSlice(mapDelegators map[string]*api.Delegator) []*api.Delegator {
	keys := make([]string, 0, len(mapDelegators))
	for key := range mapDelegators {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return strings.Compare(keys[i], keys[j]) < 0
	})

	delegatorsInfo := make([]*api.Delegator, 0, len(keys))
	for _, key := range keys {
		delegatorsInfo = append(delegatorsInfo, mapDelegators[key])
	}

	return delegatorsInfo
}

// IsInterfaceNil returns true if there is no value under the interface
func (dlp *delegatedListProcessor) IsInterfaceNil() bool {
	return dlp == nil
}
