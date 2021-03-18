package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/vm"
)

// CheckIfNil verifies if contract call input is not nil
func CheckIfNil(args *vmcommon.ContractCallInput) error {
	if args == nil {
		return vm.ErrInputArgsIsNil
	}
	if args.CallValue == nil {
		return vm.ErrInputCallValueIsNil
	}
	if args.Function == "" {
		return vm.ErrInputFunctionIsNil
	}
	if args.CallerAddr == nil {
		return vm.ErrInputCallerAddrIsNil
	}
	if args.RecipientAddr == nil {
		return vm.ErrInputRecipientAddrIsNil
	}

	return nil
}

func verifyBLSPublicKeys(registrationData *ValidatorDataV2, arguments [][]byte) error {
	for _, argKey := range arguments {
		found := false
		for _, blsKey := range registrationData.BlsPubKeys {
			if bytes.Equal(argKey, blsKey) {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("%w, key %s not found", vm.ErrBLSPublicKeyMismatch, hex.EncodeToString(argKey))
		}
	}

	return nil
}

func isNumArgsCorrectToStake(args [][]byte) bool {
	maxNodesToRun := big.NewInt(0).SetBytes(args[0]).Uint64()
	areEnoughArgs := uint64(len(args)) >= 2*maxNodesToRun+1       // NumNodes + LIST(BLS_KEY+SignedMessage)
	areNotTooManyArgs := uint64(len(args)) <= 2*maxNodesToRun+1+2 // +2 are the optionals - reward address, maxStakePerNode
	return areEnoughArgs && areNotTooManyArgs
}
