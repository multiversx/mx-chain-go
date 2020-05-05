package intermediate

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/genesis"
)

const stakedPlaceholder = "%sc_total_stake%"

var zero = big.NewInt(0)

type delegationDeployProcessor struct {
	genesis.DeployProcessor
	accountsParser genesis.AccountsParser
	pubkeyConv     state.PubkeyConverter
	nodePrice      *big.Int
}

// NewDelegationDeployProcessor returns a new deploy processor specialized for deploying delegation SC
func NewDelegationDeployProcessor(
	deployProcessor genesis.DeployProcessor,
	accountsParser genesis.AccountsParser,
	pubkeyConv state.PubkeyConverter,
	nodePrice *big.Int,
) (*delegationDeployProcessor, error) {
	if check.IfNil(deployProcessor) {
		return nil, genesis.ErrNilDeployProcessor
	}
	if check.IfNil(accountsParser) {
		return nil, genesis.ErrNilAccountsParser
	}
	if check.IfNil(pubkeyConv) {
		return nil, genesis.ErrNilPubkeyConverter
	}
	if nodePrice == nil {
		return nil, genesis.ErrNilInitialNodePrice
	}
	if nodePrice.Cmp(zero) < 1 {
		return nil, genesis.ErrInvalidInitialNodePrice
	}

	ddp := &delegationDeployProcessor{
		DeployProcessor: deployProcessor,
		accountsParser:  accountsParser,
		pubkeyConv:      pubkeyConv,
		nodePrice:       nodePrice,
	}
	ddp.SetReplacePlaceholders(ddp.replaceDelegationPlaceholders)

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

// IsInterfaceNil returns if underlying object is true
func (ddp *delegationDeployProcessor) IsInterfaceNil() bool {
	return ddp == nil || ddp.DeployProcessor == nil
}
