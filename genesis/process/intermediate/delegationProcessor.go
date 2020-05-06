package intermediate

import (
	"fmt"

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
	nodesListSplitter    genesis.NodesListSplitter
}

// NewDelegationProcessor returns a new delegation processor instance
func NewDelegationProcessor(
	executor genesis.TxExecutionProcessor,
	shardCoordinator sharding.Coordinator,
	accountsParser genesis.AccountsParser,
	smartContractParser genesis.InitialSmartContractParser,
	nodesListSplitter genesis.NodesListSplitter,
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
	if check.IfNil(nodesListSplitter) {
		return nil, genesis.ErrNilNodesListSplitter
	}

	return &delegationProcessor{
		TxExecutionProcessor: executor,
		shardCoordinator:     shardCoordinator,
		accuntsParser:        accountsParser,
		smartContractsParser: smartContractParser,
		nodesListSplitter:    nodesListSplitter,
	}, nil
}

// ExecuteDelegation will execute stake, set bls keys and activate on all delegation contracts from this shard
func (dp *delegationProcessor) ExecuteDelegation() (genesis.DelegationResult, error) {
	smartContracts, err := dp.getDelegationScOnCurrentShard()
	if err != nil {
		return genesis.DelegationResult{}, err
	}

	if len(smartContracts) == 0 {
		return genesis.DelegationResult{}, nil
	}

	dr := genesis.DelegationResult{}
	dr.NumTotalStaked, err = dp.executeStake(smartContracts)
	if err != nil {
		return genesis.DelegationResult{}, err
	}

	dr.NumTotalDelegated, err = dp.activateBlsKeys(smartContracts)
	return dr, err
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

func (dp *delegationProcessor) executeStake(smartContracts []genesis.InitialSmartContractHandler) (int, error) {
	stakedOnDelegation := 0

	for _, sc := range smartContracts {
		accounts := dp.accuntsParser.GetInitialAccountsForDelegated(sc.AddressBytes())
		for _, ac := range accounts {
			err := dp.stake(ac, sc)
			if err != nil {
				return 0, fmt.Errorf("%w while calling stake function from account %s", err, ac.GetAddress())
			}
		}
		stakedOnDelegation += len(accounts)
	}

	return stakedOnDelegation, nil
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

	return nil
}

func (dp *delegationProcessor) activateBlsKeys(smartContracts []genesis.InitialSmartContractHandler) (int, error) {
	//mockSignature := "genesis"

	totalDelegated := 0
	for _, sc := range smartContracts {
		delegatedNodes := dp.nodesListSplitter.GetDelegatedNodes(sc.AddressBytes())

		lenDelegated := len(delegatedNodes)
		if lenDelegated == 0 {
			continue
		}
		totalDelegated += lenDelegated

		//TODO refactor this: use the new delegation contract version that will only activate the nodes internally
		//setBlsKeys := make([]string, 0, lenDelegated)
		//activateKeys := make([]string, 0, lenDelegated)
		//for _, node := range delegatedNodes {
		//	setBlsKeys = append(setBlsKeys, hex.EncodeToString(node.PubKeyBytes()))
		//	activateKeys = append(activateKeys, mockSignature)
		//}
		//
		//nonce, err := dp.GetNonce(sc.OwnerBytes())
		//if err != nil {
		//	return 0, err
		//}
		//
		//setString := fmt.Sprintf("setBlsKeys@%d@%s", lenDelegated, strings.Join(setBlsKeys, "@"))
		//err = dp.ExecuteTransaction(
		//	nonce,
		//	sc.OwnerBytes(),
		//	sc.AddressBytes(),
		//	big.NewInt(0),
		//	[]byte(setString),
		//)
		//if err != nil {
		//	return 0, err
		//}
		//
		//nonce++
		//
		//hexLenDelegated := hex.EncodeToString(big.NewInt(int64(lenDelegated)).Bytes())
		//activateString := fmt.Sprintf("activate@%s@%s", hexLenDelegated, strings.Join(activateKeys, "@"))
		//err = dp.ExecuteTransaction(
		//	nonce,
		//	sc.OwnerBytes(),
		//	sc.AddressBytes(),
		//	big.NewInt(0),
		//	[]byte(activateString),
		//)
		//if err != nil {
		//	return 0, err
		//}
	}

	return totalDelegated, nil
}

// IsInterfaceNil returns if underlying object is true
func (dp *delegationProcessor) IsInterfaceNil() bool {
	return dp == nil || dp.TxExecutionProcessor == nil
}
