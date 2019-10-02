package rewardTx_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/stretchr/testify/assert"
)

func TestRewardTx_SaveLoad(t *testing.T) {
	smrS := rewardTx.RewardTx{
		Round:   uint64(1),
		Epoch:   uint32(1),
		Value:   big.NewInt(1),
		RcvAddr: []byte("receiver_address"),
		ShardId: 10,
	}

	var b bytes.Buffer
	err := smrS.Save(&b)
	assert.Nil(t, err)

	loadSMR := rewardTx.RewardTx{}
	err = loadSMR.Load(&b)
	assert.Nil(t, err)

	assert.Equal(t, smrS, loadSMR)
}

func TestRewardTx_GetRecvAddr(t *testing.T) {
	t.Parallel()

	data := []byte("data")
	scr := &rewardTx.RewardTx{RcvAddr: data}

	assert.Equal(t, data, scr.RcvAddr)
}

func TestRewardTx_GetValue(t *testing.T) {
	t.Parallel()

	value := big.NewInt(10)
	scr := &rewardTx.RewardTx{Value: value}

	assert.Equal(t, value, scr.Value)
}

func TestRewardTx_SetRecvAddr(t *testing.T) {
	t.Parallel()

	data := []byte("data")
	scr := &rewardTx.RewardTx{}
	scr.SetRecvAddress(data)

	assert.Equal(t, data, scr.RcvAddr)
}

func TestRewardTx_SetValue(t *testing.T) {
	t.Parallel()

	value := big.NewInt(10)
	scr := &rewardTx.RewardTx{}
	scr.SetValue(value)

	assert.Equal(t, value, scr.Value)
}
