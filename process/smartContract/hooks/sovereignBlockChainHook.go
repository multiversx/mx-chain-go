package hooks

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type sovereignBlockChainHook struct {
	*BlockChainHookImpl
	metaAddresses map[string]struct{}
}

// NewSovereignBlockChainHook creates a sovereign blockchain hook
func NewSovereignBlockChainHook(blockChainHook *BlockChainHookImpl) (*sovereignBlockChainHook, error) {
	if check.IfNil(blockChainHook) {
		return nil, ErrNilBlockChainHook
	}

	sbh := &sovereignBlockChainHook{
		BlockChainHookImpl: blockChainHook,
		metaAddresses:      initSystemSCsAddresses(),
	}

	sbh.getUserAccountsFunc = sbh.getUserAccounts
	return sbh, nil
}

func initSystemSCsAddresses() map[string]struct{} {
	addresses := make([][]byte, 0)
	addresses = append(addresses, vm.StakingSCAddress)
	addresses = append(addresses, vm.ValidatorSCAddress)
	addresses = append(addresses, vm.GovernanceSCAddress)
	addresses = append(addresses, vm.ESDTSCAddress)
	addresses = append(addresses, vm.DelegationManagerSCAddress)
	addresses = append(addresses, vm.FirstDelegationSCAddress)

	addressMap := make(map[string]struct{})
	for _, addr := range addresses {
		addressMap[string(addr)] = struct{}{}
	}

	return addressMap
}

func (sbh *sovereignBlockChainHook) getUserAccounts(
	input *vmcommon.ContractCallInput,
) (vmcommon.UserAccountHandler, vmcommon.UserAccountHandler, error) {
	var err error
	var sndAccount vmcommon.UserAccountHandler

	// If not incoming sovereign scr from main chain
	if !bytes.Equal(input.CallerAddr, core.ESDTSCAddress) {
		sndAccount, err = sbh.getAccount(input.CallerAddr, sbh.accounts.GetExistingAccount)
		if err != nil {
			return nil, nil, err
		}
	}

	dstAccount, err := sbh.getAccount(input.RecipientAddr, sbh.accounts.LoadAccount)
	if err != nil {
		return nil, nil, err
	}

	return sndAccount, dstAccount, nil
}

// IsPayable checks whether the provided address can receive payments
func (sbh *sovereignBlockChainHook) IsPayable(sndAddress []byte, rcvAddress []byte) (bool, error) {
	if core.IsSystemAccountAddress(rcvAddress) {
		return false, nil
	}

	if !sbh.IsSmartContract(rcvAddress) {
		return true, nil
	}

	if sbh.isNotSystemAccountAndCrossShardInSovereign(sndAddress, rcvAddress) {
		return true, nil
	}

	return sbh.baseIsPayable(sndAddress, rcvAddress)
}

func (sbh *sovereignBlockChainHook) isNotSystemAccountAndCrossShardInSovereign(sndAddress []byte, rcvAddress []byte) bool {
	return !core.IsSystemAccountAddress(rcvAddress) && sbh.isCrossShardForPayableCheck(sndAddress, rcvAddress)
}

func (sbh *sovereignBlockChainHook) isCrossShardForPayableCheck(sndAddress []byte, rcvAddress []byte) bool {
	_, isSenderMetaAddr := sbh.metaAddresses[string(sndAddress)]
	_, isReceiverMetaAddr := sbh.metaAddresses[string(rcvAddress)]

	return isSenderMetaAddr != isReceiverMetaAddr
}
