package rewardTx_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/stretchr/testify/assert"
)

func TestRewardTx_GettersAndSetters(t *testing.T) {
	t.Parallel()

	rwdTx := rewardTx.RewardTx{}
	assert.False(t, check.IfNil(&rwdTx))

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
