package exTransaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"math/big"
)

// RegisterAddress is the address used for a node to register in the Elrond network
const RegisterAddress = "0x0000000000000000000000000000000000000000000000000000000000000000"

// UnregisterAddress is the address used for a node to unregister from the Elrond network
const UnregisterAddress = "0x0000000000000000000000000000000000000000000000000000000000000001"

// RegistrationData holds the data which are sent in a register transaction
type RegistrationData struct {
	NodeId string
	Stake  uint64
}

// registerTransaction manages the register/unregister transactions received
type registerTransaction struct {
	txExecutor  execution.TransactionExecutor
	vldSyncer   execution.ValidatorSyncer
	marshalizer marshal.Marshalizer
}

// NewRegisterTransaction creates a new registerTransaction object
func NewRegisterTransaction(
	txExecutor execution.TransactionExecutor,
	vldSyncer execution.ValidatorSyncer,
	marshalizer marshal.Marshalizer,
) (*registerTransaction, error) {

	err := checkRegisterTransactionNilParameters(
		txExecutor,
		vldSyncer,
		marshalizer)

	if err != nil {
		return nil, err
	}

	rt := registerTransaction{
		txExecutor:  txExecutor,
		vldSyncer:   vldSyncer,
		marshalizer: marshalizer}

	rt.txExecutor.SetRegisterHandler(rt.register)
	rt.txExecutor.SetUnregisterHandler(rt.unregister)

	return &rt, nil
}

// checkRegisterTransactionNilParameters will check the imput parameters for nil values
func checkRegisterTransactionNilParameters(
	txExecutor execution.TransactionExecutor,
	vldSyncer execution.ValidatorSyncer,
	marshalizer marshal.Marshalizer,
) error {
	if txExecutor == nil {
		return execution.ErrNilTransactionExecutor
	}

	if vldSyncer == nil {
		return execution.ErrNilValidatorSyncer
	}

	if marshalizer == nil {
		return execution.ErrNilMarshalizer
	}

	return nil
}

// register is a call back method which is called when a register transaction is received and it
// unmarshal the data received into RegistrationData object and than it calls the AddValidator method
// which job is to add new validators into wait list
func (rt *registerTransaction) register(data []byte) error {
	rd := &RegistrationData{}

	err := rt.marshalizer.Unmarshal(rd, data)

	if err != nil {
		return err
	}

	rt.vldSyncer.AddValidator(rd.NodeId, *new(big.Int).SetUint64(rd.Stake))
	return nil
}

// unregister is a call back method which is called when a unregister transaction is received and it
// unmarshal the data received into RegistrationData object and than it calls the RemoveValidator method
// which job is to add new validators into unregister list
func (rt *registerTransaction) unregister(data []byte) error {
	rd := &RegistrationData{}

	err := rt.marshalizer.Unmarshal(rd, data)

	if err != nil {
		return err
	}

	rt.vldSyncer.RemoveValidator(rd.NodeId)
	return nil
}
