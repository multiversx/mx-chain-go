package hooks

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type sovereignBlockChainHook struct {
	*BlockChainHookImpl
}

// NewSovereignBlockChainHook creates a sovereign blockchain hook
func NewSovereignBlockChainHook(blockChainHook *BlockChainHookImpl) (*sovereignBlockChainHook, error) {
	if check.IfNil(blockChainHook) {
		return nil, ErrNilBlockChainHook
	}

	sbh := &sovereignBlockChainHook{
		BlockChainHookImpl: blockChainHook,
	}

	sbh.getUserAccountsFunc = sbh.getUserAccounts
	return sbh, nil
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
