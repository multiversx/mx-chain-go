package disabled

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestDisabledNodeFacade_AllMethodsShouldNotPanic(t *testing.T) {
	t.Parallel()
	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panic: %v", r))
		}
	}()

	apiInterface := "127.0.0.1:7799"
	dnf := NewDisabledNodeFacade(apiInterface)

	dnf.SetSyncer(nil)
	dnf.SetTpsBenchmark(nil)
	r := dnf.TpsBenchmark()
	assert.Nil(t, r)
	b := dnf.RestAPIServerDebugMode()
	assert.Equal(t, false, b)
	s1 := dnf.RestApiInterface()
	assert.Equal(t, apiInterface, s1)
	s1, s2, err := dnf.GetESDTBalance("", "")
	assert.Equal(t, emptyString, s1+s2)
	assert.Equal(t, errNodeStarting, err)
	v, err := dnf.GetBalance("")
	assert.Nil(t, v)
	assert.Equal(t, errNodeStarting, err)

	s1, err = dnf.GetUsername("")
	assert.Equal(t, emptyString, s1)
	assert.Equal(t, errNodeStarting, err)

	s1, err = dnf.GetValueForKey("", "")
	assert.Equal(t, emptyString, s1)
	assert.Equal(t, errNodeStarting, err)

	s3, err := dnf.GetAllESDTTokens("")
	assert.Nil(t, s3)
	assert.Equal(t, errNodeStarting, err)

	n1, n2, err := dnf.CreateTransaction(uint64(0), "", "", []byte{0}, "",
		[]byte{0}, uint64(0), uint64(0), []byte{0}, "", "", uint32(0), uint32(0))
	assert.Nil(t, n1)
	assert.Nil(t, n2)
	assert.Equal(t, errNodeStarting, err)

	err = dnf.ValidateTransaction(nil)
	assert.Equal(t, errNodeStarting, err)

	err = dnf.ValidateTransactionForSimulation(nil, false)
	assert.Equal(t, errNodeStarting, err)

	v1, err := dnf.ValidatorStatisticsApi()
	assert.Nil(t, v1)
	assert.Equal(t, errNodeStarting, err)

	u1, err := dnf.SendBulkTransactions(nil)
	assert.Equal(t, uint64(0), u1)
	assert.Equal(t, errNodeStarting, err)

	u2, err := dnf.SimulateTransactionExecution(nil)
	assert.Nil(t, u2)
	assert.Equal(t, errNodeStarting, err)

	t1, err := dnf.GetTransaction("", false)
	assert.Nil(t, t1)
	assert.Equal(t, errNodeStarting, err)

	resp, err := dnf.ComputeTransactionGasLimit(nil)
	assert.Nil(t, resp)
	assert.Equal(t, errNodeStarting, err)

	uac, err := dnf.GetAccount("")
	assert.Nil(t, uac)
	assert.Equal(t, errNodeStarting, err)

	hi, err := dnf.GetHeartbeats()
	assert.Nil(t, hi)
	assert.NotNil(t, errNodeStarting, err)

	sm := dnf.StatusMetrics()
	assert.NotNil(t, sm)

	vo, err := dnf.ExecuteSCQuery(nil)
	assert.Nil(t, vo)
	assert.Equal(t, errNodeStarting, err)

	b = dnf.PprofEnabled()
	assert.Equal(t, false, b)

	err = dnf.Trigger(0, false)
	assert.Equal(t, errNodeStarting, err)

	b = dnf.IsSelfTrigger()
	assert.Equal(t, false, b)

	s1, err = dnf.EncodeAddressPubkey(nil)
	assert.Equal(t, emptyString, s1)
	assert.Equal(t, errNodeStarting, err)

	ba, err := dnf.DecodeAddressPubkey("")
	assert.Nil(t, ba)
	assert.Equal(t, errNodeStarting, err)

	qh, err := dnf.GetQueryHandler("")
	assert.Nil(t, qh)
	assert.Equal(t, errNodeStarting, err)

	qp, err := dnf.GetPeerInfo("")
	assert.Nil(t, qp)
	assert.Equal(t, errNodeStarting, err)

	th, b := dnf.GetThrottlerForEndpoint("")
	assert.Nil(t, th)
	assert.Equal(t, false, b)

	ab, err := dnf.GetBlockByHash("", false)
	assert.Nil(t, ab)
	assert.Equal(t, errNodeStarting, err)

	c := dnf.GetCode(nil)
	assert.Nil(t, c)

	ab, err = dnf.GetBlockByNonce(0, false)
	assert.Nil(t, ab)
	assert.Equal(t, errNodeStarting, err)

	err = dnf.Close()
	assert.Equal(t, errNodeStarting, err)

	u32 := dnf.GetNumCheckpointsFromAccountState()
	assert.Equal(t, uint32(0), u32)

	u32 = dnf.GetNumCheckpointsFromPeerState()
	assert.Equal(t, uint32(0), u32)

	assert.False(t, check.IfNil(dnf))
}
