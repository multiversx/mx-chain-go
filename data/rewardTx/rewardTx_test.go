package rewardTx_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/stretchr/testify/assert"
)

func TestRewardTx_GettersAndSetters(t *testing.T) {
	t.Parallel()

	rwdTx := rewardTx.RewardTx{}

	addr := []byte("address")
	value := big.NewInt(37)

	rwdTx.SetRcvAddr(addr)
	rwdTx.SetValue(value)

	assert.Equal(t, []byte(nil), rwdTx.GetSndAddr())
	assert.Equal(t, addr, rwdTx.GetRcvAddr())
	assert.Equal(t, value, rwdTx.GetValue())
	assert.Equal(t, []byte(""), rwdTx.GetData())
	assert.Equal(t, uint64(0), rwdTx.GetGasLimit())
	assert.Equal(t, uint64(0), rwdTx.GetGasPrice())
	assert.Equal(t, uint64(0), rwdTx.GetNonce())
}

func TestRewardTx_CheckIntegrityShouldWork(t *testing.T) {
	t.Parallel()

	rwdTx := &rewardTx.RewardTx{
		Round:   19,
		RcvAddr: []byte("rcv-address"),
		Value:   big.NewInt(10),
	}

	err := rwdTx.CheckIntegrity()
	assert.Nil(t, err)
}

func TestRewardTx_CheckIntegrityShouldErr(t *testing.T) {
	t.Parallel()

	rwdTx := &rewardTx.RewardTx{
		Round: 19,
	}

	err := rwdTx.CheckIntegrity()
	assert.Equal(t, data.ErrNilRcvAddr, err)

	rwdTx.RcvAddr = []byte("rcv-address")

	err = rwdTx.CheckIntegrity()
	assert.Equal(t, data.ErrNilValue, err)

	rwdTx.Value = big.NewInt(-1)

	err = rwdTx.CheckIntegrity()
	assert.Equal(t, data.ErrNegativeValue, err)
}
