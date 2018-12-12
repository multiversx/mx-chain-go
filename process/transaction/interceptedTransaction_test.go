package transaction_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/stretchr/testify/assert"
)

//------- Check()

func TestInterceptedTransaction_CheckNilTransactionShouldRetFalse(t *testing.T) {
	t.Parallel()

	tx := transaction.NewInterceptedTransaction()

	tx.Transaction = nil
	assert.False(t, tx.Check())
}

func TestInterceptedTransaction_CheckNilSignatureShouldRetFalse(t *testing.T) {
	t.Parallel()

	tx := transaction.NewInterceptedTransaction()
	tx.Signature = nil
	tx.Challenge = make([]byte, 0)
	tx.RcvAddr = make([]byte, 0)
	tx.SndAddr = make([]byte, 0)

	assert.False(t, tx.Check())
}

func TestInterceptedTransaction_CheckNilChallengeShouldRetFalse(t *testing.T) {
	t.Parallel()

	tx := transaction.NewInterceptedTransaction()
	tx.Signature = make([]byte, 0)
	tx.Challenge = nil
	tx.RcvAddr = make([]byte, 0)
	tx.SndAddr = make([]byte, 0)

	assert.False(t, tx.Check())
}

func TestInterceptedTransaction_CheckNilRcvAddrShouldRetFalse(t *testing.T) {
	t.Parallel()

	tx := transaction.NewInterceptedTransaction()
	tx.Signature = make([]byte, 0)
	tx.Challenge = make([]byte, 0)
	tx.RcvAddr = nil
	tx.SndAddr = make([]byte, 0)

	assert.False(t, tx.Check())
}

func TestInterceptedTransaction_CheckNilSndAddrShouldRetFalse(t *testing.T) {
	t.Parallel()

	tx := transaction.NewInterceptedTransaction()
	tx.Signature = make([]byte, 0)
	tx.Challenge = make([]byte, 0)
	tx.RcvAddr = make([]byte, 0)
	tx.SndAddr = nil

	assert.False(t, tx.Check())
}

func TestTransactionInterceptor_CheckNegativeBalanceShouldRetFalse(t *testing.T) {
	t.Parallel()

	tx := transaction.NewInterceptedTransaction()
	tx.Signature = make([]byte, 0)
	tx.Challenge = make([]byte, 0)
	tx.RcvAddr = make([]byte, 0)
	tx.SndAddr = make([]byte, 0)
	tx.Value = *big.NewInt(-1)

	tx.SetAddressConverter(&mock.AddressConverterMock{})

	assert.False(t, tx.Check())

}

func TestTransactionInterceptor_CheckNilAddrConvertorShouldRetFalse(t *testing.T) {
	t.Parallel()

	tx := transaction.NewInterceptedTransaction()
	tx.Signature = make([]byte, 0)
	tx.Challenge = make([]byte, 0)
	tx.RcvAddr = make([]byte, 0)
	tx.SndAddr = make([]byte, 0)

	tx.SetAddressConverter(nil)

	assert.False(t, tx.Check())
}

func TestTransactionInterceptor_CheckInvalidSenderAddrShouldRetFalse(t *testing.T) {
	t.Parallel()

	tx := transaction.NewInterceptedTransaction()
	tx.Signature = make([]byte, 0)
	tx.Challenge = make([]byte, 0)
	tx.RcvAddr = make([]byte, 0)
	tx.SndAddr = []byte("please fail, addrConverter!")

	addrConv := &mock.AddressConverterMock{}
	addrConv.CreateAddressFromPublicKeyBytesRetErrForValue = []byte("please fail, addrConverter!")
	tx.SetAddressConverter(addrConv)

	assert.False(t, tx.Check())
}

func TestTransactionInterceptor_CheckInvalidReceiverAddrShouldRetFalse(t *testing.T) {
	t.Parallel()

	tx := transaction.NewInterceptedTransaction()
	tx.Signature = make([]byte, 0)
	tx.Challenge = make([]byte, 0)
	tx.RcvAddr = []byte("please fail, addrConverter!")
	tx.SndAddr = make([]byte, 0)

	addrConv := &mock.AddressConverterMock{}
	addrConv.CreateAddressFromPublicKeyBytesRetErrForValue = []byte("please fail, addrConverter!")
	tx.SetAddressConverter(addrConv)

	assert.False(t, tx.Check())
}

func TestTransactionInterceptor_CheckOkValsShouldRetTrue(t *testing.T) {
	t.Parallel()

	tx := transaction.NewInterceptedTransaction()
	tx.Signature = make([]byte, 0)
	tx.Challenge = make([]byte, 0)
	tx.RcvAddr = make([]byte, 0)
	tx.SndAddr = make([]byte, 0)

	tx.SetAddressConverter(&mock.AddressConverterMock{})

	assert.True(t, tx.Check())
}

//------- Getters and Setters

func TestTransactionInterceptor_AllGettersAndSettersShouldWork(t *testing.T) {
	t.Parallel()

	addrConv := &mock.AddressConverterMock{}

	tx := transaction.NewInterceptedTransaction()
	tx.SetAddressConverter(addrConv)
	assert.Equal(t, addrConv, tx.AddressConverter())

	tx.SetRcvShard(3)
	assert.Equal(t, uint32(3), tx.RcvShard())

	tx.SetSndShard(4)
	assert.Equal(t, uint32(4), tx.SndShard())

	tx.Nonce = 5
	assert.Equal(t, uint64(5), tx.GetTransaction().Nonce)

	tx.SetIsAddressedToOtherShards(true)
	assert.True(t, tx.IsAddressedToOtherShards())

	tx.SetHash([]byte("aaaa"))
	assert.Equal(t, "aaaa", tx.ID())
	assert.Equal(t, "aaaa", string(tx.Hash()))
}
