package smartContract

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-vm-common"
)

func CreateVMInput(account state.Account, tx *transaction.Transaction) (*vmcommon.VMInput, error) {
	vmInput := &vmcommon.VMInput{}

	return vmInput, nil
}
