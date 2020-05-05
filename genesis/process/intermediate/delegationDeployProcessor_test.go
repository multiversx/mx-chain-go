package intermediate

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/genesis/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewDelegationDeployProcessor_NilDeployProcessorShouldErr(t *testing.T) {
	t.Parallel()

	ddp, err := NewDelegationDeployProcessor(
		nil,
		&mock.AccountsParserStub{},
		mock.NewPubkeyConverterMock(32),
		big.NewInt(1),
	)

	assert.True(t, check.IfNil(ddp))
	assert.Equal(t, genesis.ErrNilDeployProcessor, err)
}

func TestNewDelegationDeployProcessor_NilAccountsParserShouldErr(t *testing.T) {
	t.Parallel()

	ddp, err := NewDelegationDeployProcessor(
		&mock.DeployProcessorStub{},
		nil,
		mock.NewPubkeyConverterMock(32),
		big.NewInt(1),
	)

	assert.True(t, check.IfNil(ddp))
	assert.Equal(t, genesis.ErrNilAccountsParser, err)
}

func TestNewDelegationDeployProcessor_NilPubkeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	ddp, err := NewDelegationDeployProcessor(
		&mock.DeployProcessorStub{},
		&mock.AccountsParserStub{},
		nil,
		big.NewInt(1),
	)

	assert.True(t, check.IfNil(ddp))
	assert.Equal(t, genesis.ErrNilPubkeyConverter, err)
}

func TestNewDelegationDeployProcessor_NilInitialNodePriceShouldErr(t *testing.T) {
	t.Parallel()

	ddp, err := NewDelegationDeployProcessor(
		&mock.DeployProcessorStub{},
		&mock.AccountsParserStub{},
		mock.NewPubkeyConverterMock(32),
		nil,
	)

	assert.True(t, check.IfNil(ddp))
	assert.Equal(t, genesis.ErrNilInitialNodePrice, err)
}

func TestNewDelegationDeployProcessor_InvalidInitialNodePriceShouldErr(t *testing.T) {
	t.Parallel()

	ddp, err := NewDelegationDeployProcessor(
		&mock.DeployProcessorStub{},
		&mock.AccountsParserStub{},
		mock.NewPubkeyConverterMock(32),
		big.NewInt(0),
	)

	assert.True(t, check.IfNil(ddp))
	assert.Equal(t, genesis.ErrInvalidInitialNodePrice, err)
}

func TestNewDelegationDeployProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	ddp, err := NewDelegationDeployProcessor(
		&mock.DeployProcessorStub{},
		&mock.AccountsParserStub{},
		mock.NewPubkeyConverterMock(32),
		big.NewInt(1),
	)

	assert.False(t, check.IfNil(ddp))
	assert.Nil(t, err)
}
