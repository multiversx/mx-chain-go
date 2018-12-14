package exTransaction_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution/exTransaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution/mock"
	"github.com/stretchr/testify/assert"
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

func TestRegisterShouldReturnNil(t *testing.T) {
	rt, _ := exTransaction.NewRegisterTransaction(&mock.ExecTransactionMock{}, &mock.SyncValidatorsMock{}, &mock.MarshalizerMock{})

	err := rt.Register(nil)

	assert.NotNil(t, err)
	assert.Equal(t, execution.ErrNilValue, err)
}

func TestRegisterShouldWork(t *testing.T) {
	rt, _ := exTransaction.NewRegisterTransaction(&mock.ExecTransactionMock{}, &mock.SyncValidatorsMock{}, &mock.MarshalizerMock{})

	rd1 := exTransaction.RegistrationData{NodeId: "node1", Stake: *big.NewInt(1), Action: exTransaction.ArRegister}
	rd2 := exTransaction.RegistrationData{NodeId: "node2", Stake: *big.NewInt(0), Action: exTransaction.ArUnregister}

	marsh := mock.MarshalizerMock{}

	dta1, _ := marsh.Marshal(&rd1)
	dta2, _ := marsh.Marshal(&rd2)

	rt.Register(dta1)
	rt.Register(dta2)

	assert.Equal(t, 2, len(rt.RegisterList()))
}

func TestCommitShouldWork(t *testing.T) {
	svm := mock.SyncValidatorsMock{}

	addWasCalled := false
	svm.AddValidatorCalled = func(nodeId string, stake big.Int) error {
		addWasCalled = true
		return nil
	}

	removeWasCalled := false
	svm.RemoveValidatorCalled = func(nodeId string) error {
		removeWasCalled = true
		return nil
	}

	rt, _ := exTransaction.NewRegisterTransaction(&mock.ExecTransactionMock{}, &svm, &mock.MarshalizerMock{})

	rd1 := exTransaction.RegistrationData{NodeId: "node1", Stake: *big.NewInt(1), Action: exTransaction.ArRegister}
	rd2 := exTransaction.RegistrationData{NodeId: "node2", Stake: *big.NewInt(0), Action: exTransaction.ArUnregister}

	marsh := mock.MarshalizerMock{}

	dta1, _ := marsh.Marshal(&rd1)
	dta2, _ := marsh.Marshal(&rd2)

	rt.Register(dta1)
	rt.Register(dta2)

	rt.Commit()

	assert.Equal(t, true, addWasCalled)
	assert.Equal(t, true, removeWasCalled)
	assert.Equal(t, 0, len(rt.RegisterList()))
}

func TestRevertAllShouldWork(t *testing.T) {
	rt, _ := exTransaction.NewRegisterTransaction(&mock.ExecTransactionMock{}, &mock.SyncValidatorsMock{}, &mock.MarshalizerMock{})

	rd1 := exTransaction.RegistrationData{NodeId: "node1", Stake: *big.NewInt(1), Action: exTransaction.ArRegister}
	rd2 := exTransaction.RegistrationData{NodeId: "node2", Stake: *big.NewInt(0), Action: exTransaction.ArUnregister}
	rd3 := exTransaction.RegistrationData{NodeId: "node3", Stake: *big.NewInt(2), Action: exTransaction.ArRegister}

	marsh := mock.MarshalizerMock{}

	dta1, _ := marsh.Marshal(&rd1)
	dta2, _ := marsh.Marshal(&rd2)
	dta3, _ := marsh.Marshal(&rd3)

	rt.Register(dta1)
	rt.Register(dta2)

	rt.RevertAll()

	rt.Register(dta3)

	assert.Equal(t, 1, len(rt.RegisterList()))
}

func TestRevertLastShouldWork(t *testing.T) {
	rt, _ := exTransaction.NewRegisterTransaction(&mock.ExecTransactionMock{}, &mock.SyncValidatorsMock{}, &mock.MarshalizerMock{})

	rd1 := exTransaction.RegistrationData{NodeId: "node1", Stake: *big.NewInt(1), Action: exTransaction.ArRegister}
	rd2 := exTransaction.RegistrationData{NodeId: "node2", Stake: *big.NewInt(0), Action: exTransaction.ArUnregister}
	rd3 := exTransaction.RegistrationData{NodeId: "node3", Stake: *big.NewInt(2), Action: exTransaction.ArRegister}
	rd4 := exTransaction.RegistrationData{NodeId: "node4", Stake: *big.NewInt(3), Action: exTransaction.ArRegister}

	marsh := mock.MarshalizerMock{}

	dta1, _ := marsh.Marshal(&rd1)
	dta2, _ := marsh.Marshal(&rd2)
	dta3, _ := marsh.Marshal(&rd3)
	dta4, _ := marsh.Marshal(&rd4)

	rt.Register(dta1)
	rt.Register(dta2)
	rt.Register(dta3)

	rt.RevertLast()

	rt.Register(dta4)

	assert.Equal(t, 3, len(rt.RegisterList()))
	assert.Equal(t, "node1", rt.RegisterList()[0].NodeId)
	assert.Equal(t, "node2", rt.RegisterList()[1].NodeId)
	assert.Equal(t, "node4", rt.RegisterList()[2].NodeId)
}
