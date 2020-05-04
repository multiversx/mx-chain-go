package intermediate

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const stakeFunction = "stake"

type delegationProcessor struct {
	genesis.TxExecutionProcessor
	shardCoordinator     sharding.Coordinator
	accuntsParser        genesis.AccountsParser
	smartContractsParser genesis.InitialSmartContractParser
	nodesHandler         genesis.NodesHandler
}

// NewDelegationProcessor returns a new delegation processor instance
func NewDelegationProcessor(
	executor genesis.TxExecutionProcessor,
	shardCoordinator sharding.Coordinator,
	accountsParser genesis.AccountsParser,
	smartContractParser genesis.InitialSmartContractParser,
	nodesHandler genesis.NodesHandler,
) (*delegationProcessor, error) {
	if check.IfNil(executor) {
		return nil, genesis.ErrNilTxExecutionProcessor
	}
	if check.IfNil(shardCoordinator) {
		return nil, genesis.ErrNilShardCoordinator
	}
	if check.IfNil(accountsParser) {
		return nil, genesis.ErrNilAccountsParser
	}
	if check.IfNil(smartContractParser) {
		return nil, genesis.ErrNilSmartContractParser
	}
	if check.IfNil(nodesHandler) {
		return nil, genesis.ErrNilNodesHandler
	}

	return &delegationProcessor{
		TxExecutionProcessor: executor,
		shardCoordinator:     shardCoordinator,
		accuntsParser:        accountsParser,
		smartContractsParser: smartContractParser,
		nodesHandler:         nodesHandler,
	}, nil
}

// ExecuteDelegation will execute stake, set bls keys and activate on all delegation contracts from this shard
func (dp *delegationProcessor) ExecuteDelegation() error {
	smartContracts, err := dp.getDelegationScOnCurrentShard()
	if err != nil {
		return err
	}

	if len(smartContracts) == 0 {
		return nil
	}

	err = dp.executeStake(smartContracts)
	if err != nil {
		return err
	}

	return dp.activateBlsKeys(smartContracts)
}

func (dp *delegationProcessor) getDelegationScOnCurrentShard() ([]genesis.InitialSmartContractHandler, error) {
	allSmartContracts, err := dp.smartContractsParser.InitialSmartContractsSplitOnOwnersShards(dp.shardCoordinator)
	if err != nil {
		return nil, err
	}

	smartContracts := make([]genesis.InitialSmartContractHandler, 0)
	smartContractsForCurrentShard := allSmartContracts[dp.shardCoordinator.SelfId()]
	for _, sc := range smartContractsForCurrentShard {
		if sc.GetType() == genesis.DelegationType {
			smartContracts = append(smartContracts, sc)
		}
	}

	return smartContracts, nil
}

func (dp *delegationProcessor) executeStake(smartContracts []genesis.InitialSmartContractHandler) error {
	for _, sc := range smartContracts {
		accounts := dp.accuntsParser.GetInitialAccountsForDelegated(sc.AddressBytes())
		for _, ac := range accounts {
			err := dp.stake(ac, sc)
			if err != nil {
				return fmt.Errorf("%w while calling stake function from account %s", err, ac.GetAddress())
			}
		}
	}

	return nil
}

func (dp *delegationProcessor) stake(ac genesis.InitialAccountHandler, sc genesis.InitialSmartContractHandler) error {
	isIntraShardCall := dp.shardCoordinator.SameShard(ac.AddressBytes(), sc.AddressBytes())

	dh := ac.GetDelegationHandler()
	if check.IfNil(dh) {
		return genesis.ErrNilDelegationHandler
	}

	var err error
	if isIntraShardCall {
		//intra shard transaction, get current nonce, add to balance the delegation value
		// in order to make the tx processor work
		nonce, errGetNonce := dp.GetNonce(ac.AddressBytes())
		if errGetNonce != nil {
			return errGetNonce
		}

		err = dp.AddBalance(ac.AddressBytes(), dh.GetValue())
		if err != nil {
			return err
		}

		return dp.ExecuteTransaction(
			nonce,
			ac.AddressBytes(),
			sc.AddressBytes(),
			dh.GetValue(),
			[]byte(stakeFunction),
		)
	}

	//cross shard transaction, just increment the nonce offset internally after executing stake function on delegation SC
	err = dp.ExecuteTransaction(
		0,
		ac.AddressBytes(),
		sc.AddressBytes(),
		dh.GetValue(),
		[]byte(stakeFunction),
	)
	if err != nil {
		return err
	}

	ac.IncrementNonceOffset()

	return nil
}

func (dp *delegationProcessor) activateBlsKeys(smartContracts []genesis.InitialSmartContractHandler) error {
	mockSignature := "genesis"

	for _, sc := range smartContracts {
		delegatedNodes := dp.nodesHandler.GetDelegatedNodes(sc.AddressBytes())

		lenDelegated := len(delegatedNodes)
		if lenDelegated == 0 {
			continue
		}

		setBlsKeys := make([]string, 0, lenDelegated)
		activateKeys := make([]string, 0, lenDelegated)
		for _, node := range delegatedNodes {
			setBlsKeys = append(setBlsKeys, hex.EncodeToString(node.PubKeyBytes()))
			activateKeys = append(activateKeys, mockSignature)
		}

		nonce, err := dp.GetNonce(sc.OwnerBytes())
		if err != nil {
			return err
		}

		setString := fmt.Sprintf("setBlsKeys@%d@%s", lenDelegated, strings.Join(setBlsKeys, "@"))
		err = dp.ExecuteTransaction(
			nonce,
			sc.OwnerBytes(),
			sc.AddressBytes(),
			big.NewInt(0),
			[]byte(setString),
		)
		if err != nil {
			return err
		}

		nonce++

		activateString := fmt.Sprintf("activate@%d@%s", lenDelegated, strings.Join(activateKeys, "@"))
		err = dp.ExecuteTransaction(
			nonce,
			sc.OwnerBytes(),
			sc.AddressBytes(),
			big.NewInt(0),
			[]byte(activateString),
		)
		if err != nil {
			return err
		}
	}

	return nil
}
