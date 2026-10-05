package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"

	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
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
	if len(args) == 0 {
		return false
	}
	maxNodesBig := big.NewInt(0).SetBytes(args[0])
	if !maxNodesBig.IsUint64() {
		return false
	}
	maxNodesToRun := maxNodesBig.Uint64()
	if maxNodesToRun > (math.MaxUint64-3)/2 {
		return false
	}
	areEnoughArgs := uint64(len(args)) >= 2*maxNodesToRun+1       // NumNodes + LIST(BLS_KEY+SignedMessage)
	areNotTooManyArgs := uint64(len(args)) <= 2*maxNodesToRun+1+2 // +2 are the optionals - reward address, maxStakePerNode
	return areEnoughArgs && areNotTooManyArgs
}
