package exTransaction_test

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution/exTransaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution/mock"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestNewRegisterTransactionShouldThrowNilTransactionExecutor(t *testing.T) {
	rt, err := exTransaction.NewRegisterTransaction(nil, nil, nil)

	assert.Nil(t, rt)
	assert.Equal(t, execution.ErrNilTransactionExecutor, err)
}

func TestNewRegisterTransactionShouldThrowNilValidatorSyncer(t *testing.T) {
	rt, err := exTransaction.NewRegisterTransaction(&mock.ExecTransactionMock{}, nil, nil)

	assert.Nil(t, rt)
	assert.Equal(t, execution.ErrNilValidatorSyncer, err)
}

func TestNewRegisterTransactionShouldThrowNilMarshalizer(t *testing.T) {
	rt, err := exTransaction.NewRegisterTransaction(&mock.ExecTransactionMock{}, &mock.SyncValidatorsMock{}, nil)

	assert.Nil(t, rt)
	assert.Equal(t, execution.ErrNilMarshalizer, err)
}

func TestNewRegisterTransactionShouldWork(t *testing.T) {
	rt, err := exTransaction.NewRegisterTransaction(&mock.ExecTransactionMock{}, &mock.SyncValidatorsMock{}, &mock.MarshalizerMock{})

	assert.NotNil(t, rt)
	assert.Nil(t, err)
}

func TestRegisterShouldWork(t *testing.T) {
	svm := mock.SyncValidatorsMock{}

	addWasCalled := false
	svm.AddValidatorCalled = func(nodeId string, stake big.Int) error {
		addWasCalled = true
		return nil
	}

	rt, _ := exTransaction.NewRegisterTransaction(&mock.ExecTransactionMock{}, &svm, &mock.MarshalizerMock{})
	rd := exTransaction.RegistrationData{NodeId: "node1", Stake: 1}

	marsh := mock.MarshalizerMock{}

	data, err := marsh.Marshal(rd)

	assert.Nil(t, err)

	err = rt.Register(data)

	assert.Nil(t, err)
	assert.Equal(t, true, addWasCalled)
}

func TestUnregisterShouldWork(t *testing.T) {
	svm := mock.SyncValidatorsMock{}

	removeWasCalled := false
	svm.RemoveValidatorCalled = func(nodeId string) error {
		removeWasCalled = true
		return nil
	}

	rt, _ := exTransaction.NewRegisterTransaction(&mock.ExecTransactionMock{}, &svm, &mock.MarshalizerMock{})
	rd := exTransaction.RegistrationData{NodeId: "node1", Stake: 1}

	marsh := mock.MarshalizerMock{}

	data, err := marsh.Marshal(rd)

	assert.Nil(t, err)

	err = rt.Unregister(data)

	assert.Nil(t, err)
	assert.Equal(t, true, removeWasCalled)
}
