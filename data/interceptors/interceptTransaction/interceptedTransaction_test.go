package interceptTransaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/interceptors/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

//------- Check()

func TestInterceptedTransactionCheckTxNilTransactionShouldRetFalse(t *testing.T) {
	t.Parallel()

	tx := NewInterceptedTransaction()

	tx.Transaction = nil
	assert.False(t, tx.Check())
}

func TestInterceptedTransactionCheckTxNilSignatureShouldRetFalse(t *testing.T) {
	t.Parallel()

	tx := NewInterceptedTransaction()
	tx.Signature = nil
	tx.Challenge = make([]byte, 0)
	tx.RcvAddr = make([]byte, 0)
	tx.SndAddr = make([]byte, 0)

	assert.False(t, tx.Check())
}

func TestInterceptedTransactionCheckTxNilChallengeShouldRetFalse(t *testing.T) {
	t.Parallel()

	tx := NewInterceptedTransaction()
	tx.Signature = make([]byte, 0)
	tx.Challenge = nil
	tx.RcvAddr = make([]byte, 0)
	tx.SndAddr = make([]byte, 0)

	assert.False(t, tx.Check())
}

func TestInterceptedTransactionCheckTxNilRcvAddrShouldRetFalse(t *testing.T) {
	t.Parallel()

	tx := NewInterceptedTransaction()
	tx.Signature = make([]byte, 0)
	tx.Challenge = make([]byte, 0)
	tx.RcvAddr = nil
	tx.SndAddr = make([]byte, 0)

	assert.False(t, tx.Check())
}

func TestInterceptedTransactionCheckTxNilSndAddrShouldRetFalse(t *testing.T) {
	t.Parallel()

	tx := NewInterceptedTransaction()
	tx.Signature = make([]byte, 0)
	tx.Challenge = make([]byte, 0)
	tx.RcvAddr = make([]byte, 0)
	tx.SndAddr = nil

	assert.False(t, tx.Check())
}

//TODO add test with negative value as of current impl uint64 does not allow negative values

func TestTransactionInterceptorCheckNilAddrConvertorShouldRetFalse(t *testing.T) {
	t.Parallel()

	tx := NewInterceptedTransaction()
	tx.Signature = make([]byte, 0)
	tx.Challenge = make([]byte, 0)
	tx.RcvAddr = make([]byte, 0)
	tx.SndAddr = make([]byte, 0)

	tx.addrConv = nil

	assert.False(t, tx.Check())
}

func TestTransactionInterceptorCheckInvalidSenderAddrShouldRetFalse(t *testing.T) {
	t.Parallel()

	tx := NewInterceptedTransaction()
	tx.Signature = make([]byte, 0)
	tx.Challenge = make([]byte, 0)
	tx.RcvAddr = make([]byte, 0)
	tx.SndAddr = []byte("please fail, addrConverter!")

	addrConv := &mock.AddressConverterMock{}
	addrConv.CreateAddressFromPublicKeyBytesRetErrForValue = []byte("please fail, addrConverter!")
	tx.addrConv = addrConv

	assert.False(t, tx.Check())
}

func TestTransactionInterceptorCheckInvalidReceiverAddrShouldRetFalse(t *testing.T) {
	t.Parallel()

	tx := NewInterceptedTransaction()
	tx.Signature = make([]byte, 0)
	tx.Challenge = make([]byte, 0)
	tx.RcvAddr = []byte("please fail, addrConverter!")
	tx.SndAddr = make([]byte, 0)

	addrConv := &mock.AddressConverterMock{}
	addrConv.CreateAddressFromPublicKeyBytesRetErrForValue = []byte("please fail, addrConverter!")
	tx.addrConv = addrConv

	assert.False(t, tx.Check())
}

func TestTransactionInterceptorCheckOkValsShouldRetTrue(t *testing.T) {
	t.Parallel()

	tx := NewInterceptedTransaction()
	tx.Signature = make([]byte, 0)
	tx.Challenge = make([]byte, 0)
	tx.RcvAddr = make([]byte, 0)
	tx.SndAddr = make([]byte, 0)

	tx.addrConv = &mock.AddressConverterMock{}

	assert.True(t, tx.Check())
}

func TestTransactionInterceptorAllGettersAndSettersShouldWork(t *testing.T) {
	t.Parallel()

	addrConv := &mock.AddressConverterMock{}

	tx := NewInterceptedTransaction()
	tx.SetAddressConverter(addrConv)
	assert.Equal(t, addrConv, tx.AddressConverter())

	tx.rcvShard = 3
	assert.Equal(t, uint32(3), tx.RcvShard())

	tx.sndShard = 4
	assert.Equal(t, uint32(4), tx.SndShard())

	tx.Nonce = 5
	assert.Equal(t, uint64(5), tx.GetTransaction().Nonce)

	tx.isAddressedToOtherShards = true
	assert.True(t, tx.IsAddressedToOtherShards())

	tx.Signature = []byte("aaaa")
	assert.Equal(t, "aaaa", tx.ID())

}
