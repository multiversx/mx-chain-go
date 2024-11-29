package checking_test

import (
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/checking"
	"github.com/multiversx/mx-chain-go/genesis/data"
	"github.com/multiversx/mx-chain-go/genesis/mock"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func createEmptyInitialAccount() *data.InitialAccount {
	return &data.InitialAccount{
		Address:      "",
		Supply:       big.NewInt(0),
		Balance:      big.NewInt(0),
		StakingValue: big.NewInt(0),
		Delegation: &data.DelegationData{
			Address: "",
			Value:   big.NewInt(0),
		},
	}
}

func createArgs() checking.ArgsNodesSetupChecker {
	return checking.ArgsNodesSetupChecker{
		AccountsParser:           &mock.AccountsParserStub{},
		InitialNodePrice:         big.NewInt(0),
		ValidatorPubKeyConverter: testscommon.NewPubkeyConverterMock(32),
		KeyGenerator:             &mock.KeyGeneratorStub{},
	}
}

//------- NewNodesSetupChecker

func TestNewNodesSetupChecker_NilGenesisParserShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgs()
	args.AccountsParser = nil
	nsc, err := checking.NewNodesSetupChecker(args)

	assert.True(t, check.IfNil(nsc))
	assert.Equal(t, genesis.ErrNilAccountsParser, err)
}

func TestNewNodesSetupChecker_NilInitialNodePriceShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgs()
	args.InitialNodePrice = nil
	nsc, err := checking.NewNodesSetupChecker(args)

	assert.True(t, check.IfNil(nsc))
	assert.Equal(t, genesis.ErrNilInitialNodePrice, err)
}

func TestNewNodesSetupChecker_InvalidInitialNodePriceShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgs()
	args.InitialNodePrice = big.NewInt(-1)
	nsc, err := checking.NewNodesSetupChecker(args)

	assert.True(t, check.IfNil(nsc))
	assert.True(t, errors.Is(err, genesis.ErrInvalidInitialNodePrice))
}

func TestNewNodesSetupChecker_NilValidatorPubKeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgs()
	args.ValidatorPubKeyConverter = nil
	nsc, err := checking.NewNodesSetupChecker(args)

	assert.True(t, check.IfNil(nsc))
	assert.Equal(t, genesis.ErrNilPubkeyConverter, err)
}

func TestNewNodesSetupChecker_NilKeyGeneratorShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgs()
	args.KeyGenerator = nil
	nsc, err := checking.NewNodesSetupChecker(args)

	assert.True(t, check.IfNil(nsc))
	assert.Equal(t, genesis.ErrNilKeyGenerator, err)
}

func TestNewNodesSetupChecker_ShouldWork(t *testing.T) {
	t.Parallel()

	args := createArgs()
	nsc, err := checking.NewNodesSetupChecker(args)

	assert.False(t, check.IfNil(nsc))
	assert.Nil(t, err)
}

//------- Check

func TestNewNodesSetupChecker_CheckNotAValidPubkeyShouldErr(t *testing.T) {
	t.Parallel()

	ia := createEmptyInitialAccount()
	ia.SetAddressBytes([]byte("staked address"))

	expectedErr := errors.New("expected error")
	args := createArgs()
	args.AccountsParser = &mock.AccountsParserStub{
		InitialAccountsCalled: func() []genesis.InitialAccountHandler {
			return []genesis.InitialAccountHandler{ia}
		},
	}
	args.KeyGenerator = &mock.KeyGeneratorStub{
		CheckPublicKeyValidCalled: func(b []byte) error {
			return expectedErr
		},
	}
	nsc, _ := checking.NewNodesSetupChecker(args)

	err := nsc.Check(
		[]nodesCoordinator.GenesisNodeInfoHandler{
			&mock.GenesisNodeInfoHandlerMock{
				AssignedShardValue: 0,
				AddressBytesValue:  []byte("staked address"),
				PubKeyBytesValue:   []byte("pubkey"),
			},
		},
	)

	assert.True(t, errors.Is(err, genesis.ErrInvalidPubKey))
}

func TestNewNodeSetupChecker_CheckNotStakedShouldErr(t *testing.T) {
	t.Parallel()

	ia := createEmptyInitialAccount()
	ia.SetAddressBytes([]byte("staked address"))

	args := createArgs()
	args.AccountsParser = &mock.AccountsParserStub{
		InitialAccountsCalled: func() []genesis.InitialAccountHandler {
			return []genesis.InitialAccountHandler{ia}
		},
	}
	nsc, _ := checking.NewNodesSetupChecker(args)

	err := nsc.Check(
		[]nodesCoordinator.GenesisNodeInfoHandler{
			&mock.GenesisNodeInfoHandlerMock{
				AssignedShardValue: 0,
				AddressBytesValue:  []byte("not-staked-address"),
				PubKeyBytesValue:   []byte("pubkey"),
			},
		},
	)

	assert.True(t, errors.Is(err, genesis.ErrNodeNotStaked))
}

func TestNewNodeSetupChecker_CheckNotEnoughStakedShouldErr(t *testing.T) {
	t.Parallel()

	nodePrice := big.NewInt(32)
	ia := createEmptyInitialAccount()
	ia.StakingValue = big.NewInt(0).Set(nodePrice)
	ia.SetAddressBytes([]byte("staked address"))

	args := createArgs()
	args.AccountsParser = &mock.AccountsParserStub{
		InitialAccountsCalled: func() []genesis.InitialAccountHandler {
			return []genesis.InitialAccountHandler{ia}
		},
	}
	args.InitialNodePrice = big.NewInt(nodePrice.Int64() + 1)
	nsc, _ := checking.NewNodesSetupChecker(args)

	err := nsc.Check(
		[]nodesCoordinator.GenesisNodeInfoHandler{
			&mock.GenesisNodeInfoHandlerMock{
				AssignedShardValue: 0,
				AddressBytesValue:  []byte("staked address"),
				PubKeyBytesValue:   []byte("pubkey"),
			},
		},
	)

	assert.True(t, errors.Is(err, genesis.ErrStakingValueIsNotEnough))
}

func TestNewNodeSetupChecker_CheckTooMuchStakedShouldErr(t *testing.T) {
	t.Parallel()

	nodePrice := big.NewInt(32)
	ia := createEmptyInitialAccount()
	ia.StakingValue = big.NewInt(0).Set(nodePrice)
	ia.SetAddressBytes([]byte("staked address"))

	args := createArgs()
	args.AccountsParser = &mock.AccountsParserStub{
		InitialAccountsCalled: func() []genesis.InitialAccountHandler {
			return []genesis.InitialAccountHandler{ia}
		},
	}
	args.InitialNodePrice = big.NewInt(nodePrice.Int64() - 1)
	nsc, _ := checking.NewNodesSetupChecker(args)

	err := nsc.Check(
		[]nodesCoordinator.GenesisNodeInfoHandler{
			&mock.GenesisNodeInfoHandlerMock{
				AssignedShardValue: 0,
				AddressBytesValue:  []byte("staked address"),
				PubKeyBytesValue:   []byte("pubkey"),
			},
		},
	)

	assert.True(t, errors.Is(err, genesis.ErrInvalidStakingBalance))
}

func TestNewNodeSetupChecker_CheckNotEnoughDelegatedShouldErr(t *testing.T) {
	t.Parallel()

	nodePrice := big.NewInt(32)
	ia := createEmptyInitialAccount()
	ia.Delegation.SetAddressBytes([]byte("delegated address"))
	ia.Delegation.Value = big.NewInt(0).Set(nodePrice)

	args := createArgs()
	args.AccountsParser = &mock.AccountsParserStub{
		InitialAccountsCalled: func() []genesis.InitialAccountHandler {
			return []genesis.InitialAccountHandler{ia}
		},
	}
	args.InitialNodePrice = big.NewInt(nodePrice.Int64() + 1)
	nsc, _ := checking.NewNodesSetupChecker(args)

	err := nsc.Check(
		[]nodesCoordinator.GenesisNodeInfoHandler{
			&mock.GenesisNodeInfoHandlerMock{
				AssignedShardValue: 0,
				AddressBytesValue:  []byte("delegated address"),
				PubKeyBytesValue:   []byte("pubkey"),
			},
		},
	)

	assert.True(t, errors.Is(err, genesis.ErrDelegationValueIsNotEnough))
}

func TestNewNodeSetupChecker_CheckTooMuchDelegatedShouldErr(t *testing.T) {
	t.Parallel()

	nodePrice := big.NewInt(32)
	ia := createEmptyInitialAccount()
	ia.Delegation.SetAddressBytes([]byte("delegated address"))
	ia.Delegation.Value = big.NewInt(0).Set(nodePrice)

	args := createArgs()
	args.AccountsParser = &mock.AccountsParserStub{
		InitialAccountsCalled: func() []genesis.InitialAccountHandler {
			return []genesis.InitialAccountHandler{ia}
		},
	}
	args.InitialNodePrice = big.NewInt(nodePrice.Int64() - 1)
	nsc, _ := checking.NewNodesSetupChecker(args)

	err := nsc.Check(
		[]nodesCoordinator.GenesisNodeInfoHandler{
			&mock.GenesisNodeInfoHandlerMock{
				AssignedShardValue: 0,
				AddressBytesValue:  []byte("delegated address"),
				PubKeyBytesValue:   []byte("pubkey"),
			},
		},
	)

	assert.True(t, errors.Is(err, genesis.ErrInvalidDelegationValue))
}

func TestNewNodeSetupChecker_CheckStakedAndDelegatedShouldWork(t *testing.T) {
	t.Parallel()

	nodePrice := big.NewInt(32)
	iaStaked := createEmptyInitialAccount()
	iaStaked.StakingValue = big.NewInt(0).Set(nodePrice)
	iaStaked.SetAddressBytes([]byte("staked address"))

	iaDelegated := createEmptyInitialAccount()
	iaDelegated.Delegation.Value = big.NewInt(0).Set(nodePrice)
	iaDelegated.Delegation.SetAddressBytes([]byte("delegated address"))

	args := createArgs()
	args.AccountsParser = &mock.AccountsParserStub{
		InitialAccountsCalled: func() []genesis.InitialAccountHandler {
			return []genesis.InitialAccountHandler{iaDelegated, iaStaked}
		},
	}
	args.InitialNodePrice = nodePrice
	nsc, _ := checking.NewNodesSetupChecker(args)

	err := nsc.Check(
		[]nodesCoordinator.GenesisNodeInfoHandler{
			&mock.GenesisNodeInfoHandlerMock{
				AssignedShardValue: 0,
				AddressBytesValue:  []byte("delegated address"),
				PubKeyBytesValue:   []byte("pubkey"),
			},
			&mock.GenesisNodeInfoHandlerMock{
				AssignedShardValue: 0,
				AddressBytesValue:  []byte("staked address"),
				PubKeyBytesValue:   []byte("pubkey"),
			},
		},
	)

	assert.Nil(t, err)
	// the following 2 asserts assure that the original values did not change
	assert.Equal(t, nodePrice, iaStaked.StakingValue)
	assert.Equal(t, nodePrice, iaDelegated.Delegation.Value)
}
