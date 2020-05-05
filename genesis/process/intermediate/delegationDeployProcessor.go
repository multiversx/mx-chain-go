package intermediate

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/genesis"
)

const stakedPlaceholder = "%sc_total_stake%"

var zero = big.NewInt(0)

type delegationDeployProcessor struct {
	*deployProcessor
	accountsParser genesis.AccountsParser
	nodePrice      *big.Int
}

// NewDelegationDeployProcessor returns a new deploy processor specialized for deploying delegation SC
func NewDelegationDeployProcessor(
	deployProcessor *deployProcessor,
	accountsParser genesis.AccountsParser,
	nodePrice *big.Int,
) (*delegationDeployProcessor, error) {
	if check.IfNil(deployProcessor) {
		return nil, genesis.ErrNilDeployProcessor
	}
	if check.IfNil(accountsParser) {
		return nil, genesis.ErrNilAccountsParser
	}
	if nodePrice == nil {
		return nil, genesis.ErrNilInitialNodePrice
	}
	if nodePrice.Cmp(zero) < 1 {
		return nil, genesis.ErrInvalidInitialNodePrice
	}

	ddp := &delegationDeployProcessor{
		deployProcessor: deployProcessor,
		accountsParser:  accountsParser,
		nodePrice:       nodePrice,
	}
	ddp.replacePlaceholders = ddp.replaceDelegationPlaceholders

	return ddp, nil
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
