package intermediate

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ElrondNetwork/elrond-go/genesis"
)

const stakedPlaceholder = "%sc_total_stake%"

var zero = big.NewInt(0)

type delegationDeployProcessor struct {
	*deployProcessor
	accountsParser genesis.AccountsParser
	nodePrice      *big.Int
}

func NewDelegationDeployProcessor(
	processor *deployProcessor,
	accountsParser genesis.AccountsParser,
	nodePrice *big.Int,
) (*deployProcessor, error) {
	//TODO add constructor
	return nil, nil
}

func (ddp *delegationDeployProcessor) replaceDelegationPlaceholders(
	txData string,
	scResultingAddressBytes []byte,
) (string, error) {

	scResultingAddress := ddp.pubkeyConv.Encode(scResultingAddressBytes)
	val := ddp.accountsParser.GetTotalStakedForDelegationAddress(scResultingAddress)
	if val.Cmp(zero) == 0 {
		return "", fmt.Errorf("%w, 0 delegated value for resulting address %s",
			genesis.ErrInvalidDelegationValue, scResultingAddress)
	}

	exactDiv := big.NewInt(0).Set(val)
	exactDiv.Mod(exactDiv, ddp.nodePrice)
	if exactDiv.Cmp(zero) != 0 {
		return "", fmt.Errorf("%w, not a node price multiple value, for resulting address %s",
			genesis.ErrInvalidDelegationValue, scResultingAddress)
	}

	txData = strings.Replace(txData, stakedPlaceholder, val.Text(16), -1)

	return txData, nil
}
