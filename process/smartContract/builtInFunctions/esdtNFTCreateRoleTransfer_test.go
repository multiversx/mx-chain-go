package builtInFunctions

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestEsdtNFTCreateRoleTransfer_Constructor(t *testing.T) {
	t.Parallel()

	e, err := NewESDTNFTCreateRoleTransfer(nil, &mock.AccountsStub{}, mock.NewMultiShardsCoordinatorMock(2))
	assert.Nil(t, e)
	assert.Equal(t, err, process.ErrNilMarshalizer)

	e, err = NewESDTNFTCreateRoleTransfer(&mock.MarshalizerMock{}, &mock.AccountsStub{}, mock.NewMultiShardsCoordinatorMock(2))
	assert.Nil(t, e)
	assert.Equal(t, err, process.ErrNilAccountsAdapter)

	e, err := NewESDTNFTCreateRoleTransfer(nil, &mock.AccountsStub{}, mock.NewMultiShardsCoordinatorMock(2))
	assert.Nil(t, e)
	assert.Equal(t, err, process.ErrNilMarshalizer)
}
