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

func TestRewardTx_GettersAndSetters(t *testing.T) {
	t.Parallel()

	rwdTx := rewardTx.RewardTx{}
	assert.False(t, rwdTx.IsInterfaceNil())

	addr := []byte("address")
	value := big.NewInt(37)

	rwdTx.SetRecvAddress(addr)
	rwdTx.SetValue(value)

	assert.Equal(t, []byte(nil), rwdTx.GetSndAddress())
	assert.Equal(t, addr, rwdTx.GetRecvAddress())
	assert.Equal(t, value, rwdTx.GetValue())
	assert.Equal(t, []byte(""), rwdTx.GetData())
	assert.Equal(t, uint64(0), rwdTx.GetGasLimit())
	assert.Equal(t, uint64(0), rwdTx.GetGasPrice())
	assert.Equal(t, uint64(0), rwdTx.GetNonce())
}
