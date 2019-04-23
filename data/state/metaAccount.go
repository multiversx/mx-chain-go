package state

//
//import (
//	"math/big"
//
//	"github.com/ElrondNetwork/elrond-go-sandbox/data"
//)
//
//// MiniBlockData is the data to be saved in shard account for any shard
//type MiniBlockData struct {
//	Hash            []byte
//	ReceiverShardId uint32
//	SenderShardId   uint32
//	TxCount         uint32
//}
//
//// MetaAccount is the struct used in serialization/deserialization
//type MetaAccount struct {
//	Round            uint64
//	CodeHash         []byte
//	RootHash         []byte
//	PubKeyLeader     []byte
//	RegistrationData []RegistrationData
//	MiniBlockData    []MiniBlockData
//	TxCount          *big.Int
//}
//
//// NewAccount creates a new account object
//func NewMetaAccount() *MetaAccount {
//	return &MetaAccount{
//		TxCount: big.NewInt(0),
//	}
//}
//
//func (acc *MetaAccount) IntegrityAndValidity() error {
//	if acc.CodeHash == nil {
//		return data.ErrNilCodeHash
//	}
//	if acc.RootHash == nil {
//		return data.ErrNilRootHash
//	}
//	if acc.RegistrationData == nil {
//		return data.ErrNilRegistrationData
//	}
//	if acc.TxCount == nil {
//		return data.ErrNilTxCount
//	}
//	if acc.TxCount.Sign() < 0 {
//		return data.ErrNegativeTxCount
//	}
//	if len(acc.MiniBlockData) < 0 {
//		return data.ErrMiniBlockDataEmpty
//	}
//
//	return nil
//}
//
//func (acc *MetaAccount) GetRootHash() []byte {
//	return acc.RootHash
//}
//
//func (acc *MetaAccount) SetRootHash(roothash []byte) {
//	if roothash == nil {
//		return
//	}
//	acc.RootHash = roothash
//}
//
//func (acc *MetaAccount) GetCodeHash() []byte {
//	return acc.CodeHash
//}
//
//func (acc *MetaAccount) SetCodeHash(roothash []byte) {
//	if roothash == nil {
//		return
//	}
//	acc.CodeHash = roothash
//}
//
////TODO add Cap'N'Proto converter funcs
