package exTransaction

//
//import (
//	"sync"
//
//	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
//	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
//	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
//)
//
//// registerTransaction manages the register/unregister transactions received
//type registerTransaction struct {
//	txExecutor  execution.TransactionExecutor
//	vldSyncer   execution.ValidatorSyncer
//	marshalizer marshal.Marshalizer
//
//	registerList []*state.RegistrationData
//
//	mut sync.RWMutex
//}
//
//// NewRegisterTransaction creates a new registerTransaction object
//func NewRegisterTransaction(
//	txExecutor execution.TransactionExecutor,
//	vldSyncer execution.ValidatorSyncer,
//	marshalizer marshal.Marshalizer,
//) (*registerTransaction, error) {
//
//	err := checkRegisterTransactionNilParameters(
//		txExecutor,
//		vldSyncer,
//		marshalizer)
//
//	if err != nil {
//		return nil, err
//	}
//
//	rt := registerTransaction{
//		txExecutor:  txExecutor,
//		vldSyncer:   vldSyncer,
//		marshalizer: marshalizer}
//
//	rt.txExecutor.SetRegisterHandler(rt.register)
//
//	rt.registerList = make([]*RegistrationData, 0)
//
//	return &rt, nil
//}
//
//// checkRegisterTransactionNilParameters will check the imput parameters for nil values
//func checkRegisterTransactionNilParameters(
//	txExecutor execution.TransactionExecutor,
//	vldSyncer execution.ValidatorSyncer,
//	marshalizer marshal.Marshalizer,
//) error {
//	if txExecutor == nil {
//		return execution.ErrNilTransactionExecutor
//	}
//
//	if vldSyncer == nil {
//		return execution.ErrNilValidatorSyncer
//	}
//
//	if marshalizer == nil {
//		return execution.ErrNilMarshalizer
//	}
//
//	return nil
//}
//
//// register is a call back method which is called when a register transaction is received. This method
//// unmarshals the data received into RegistrationData object and put it into register list
//func (rt *registerTransaction) register(data []byte) error {
//	if data == nil {
//		return execution.ErrNilValue
//	}
//
//	rd := &RegistrationData{}
//
//	err := rt.marshalizer.Unmarshal(rd, data)
//
//	if err != nil {
//		return err
//	}
//
//	rt.mut.Lock()
//	rt.registerList = append(rt.registerList, rd)
//	rt.mut.Unlock()
//
//	return nil
//}
//
//// Commit method calls the AddValidator / RemoveValidator methods which job are to add new validators into
//// wait / unregister list
//func (rt *registerTransaction) Commit() {
//	rt.mut.RLock()
//
//	for _, rd := range rt.registerList {
//		switch rd.Action {
//		case ArRegister:
//			rt.vldSyncer.AddValidator(rd.NodeId, rd.Stake)
//		case ArUnregister:
//			rt.vldSyncer.RemoveValidator(rd.NodeId)
//		}
//	}
//
//	rt.registerList = nil
//
//	rt.mut.RUnlock()
//}
//
//// RevertAll method deletes all the registration requests from register list
//func (rt *registerTransaction) RevertAll() {
//	rt.mut.Lock()
//	rt.registerList = nil
//	rt.mut.Unlock()
//}
//
//// RevertLast method deletes the last registration request added in register list
//func (rt *registerTransaction) RevertLast() {
//	rt.mut.Lock()
//
//	if len(rt.registerList) > 0 {
//		rt.registerList = rt.registerList[0 : len(rt.registerList)-1]
//	}
//
//	rt.mut.Unlock()
//}
