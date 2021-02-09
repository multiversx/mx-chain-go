package facade

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestDisabledNodeFacade_AllMethodes(t *testing.T) {
	t.Parallel()
	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panic: %v", r))
		}
	}()

	dnf, _ := NewDisableNodeFacade()
	dnf.SetSyncer(nil)
	dnf.SetTpsBenchmark(nil)
	r := dnf.TpsBenchmark()
	assert.Nil(t, r)
	dnf.StartBackgroundServices()
	b := dnf.RestAPIServerDebugMode()
	assert.Equal(t, b, false)
	s1 := dnf.RestApiInterface()
	assert.Equal(t, s1, emptStr)
	s1, s2, err := dnf.GetESDTBalance("", "")
	assert.Equal(t, s1+s2, emptStr)
	assert.Equal(t, err, errNodeStarting)
	v, err := dnf.GetBalance("")
	assert.Nil(t, v)
	assert.Equal(t, err, errNodeStarting)

	s1, err = dnf.GetUsername("")
	assert.Equal(t, s1, "")
	assert.Equal(t, err, errNodeStarting)

	s1, err = dnf.GetValueForKey("", "")
	assert.Equal(t, s1, "")
	assert.Equal(t, err, errNodeStarting)

	s3, err := dnf.GetAllESDTTokens("")
	assert.Nil(t, s3)
	assert.Equal(t, err, errNodeStarting)

	n1, n2, err := dnf.CreateTransaction(0, "", "", "", 0, 0, nil, "", "", 0, 0)
	assert.Nil(t, n1)
	assert.Nil(t, n2)
	assert.Equal(t, err, errNodeStarting)

	err = dnf.ValidateTransaction(nil)
	assert.Equal(t, err, errNodeStarting)

	err = dnf.ValidateTransactionForSimulation(nil)
	assert.Equal(t, err, errNodeStarting)

	v1, err := dnf.ValidatorStatisticsApi()
	assert.Nil(t, v1)
	assert.Equal(t, err, errNodeStarting)

	u1, err := dnf.SendBulkTransactions(nil)
	assert.Equal(t, u1, uint64(0))
	assert.Equal(t, err, errNodeStarting)

	u2, err := dnf.SimulateTransactionExecution(nil)
	assert.Nil(t, u2)
	assert.Equal(t, err, errNodeStarting)

	t1, err := dnf.GetTransaction("", false)
	assert.Nil(t, t1)
	assert.Equal(t, err, errNodeStarting)

	u1, err = dnf.ComputeTransactionGasLimit(nil)
	assert.Equal(t, u1, uint64(0))
	assert.Equal(t, err, errNodeStarting)

	uac, err := dnf.GetAccount("")
	assert.Nil(t, uac)
	assert.Equal(t, err, errNodeStarting)

	hi, err := dnf.GetHeartbeats()
	assert.Nil(t, hi)
	assert.NotNil(t, err, errNodeStarting)

	sm := dnf.StatusMetrics()
	assert.Nil(t, sm)

	vo, err := dnf.ExecuteSCQuery(nil)
	assert.Nil(t, vo)
	assert.Equal(t, err, errNodeStarting)

	b = dnf.PprofEnabled()
	assert.Equal(t, b, false)

	err = dnf.Trigger(0, false)
	assert.Equal(t, err, errNodeStarting)

	b = dnf.IsSelfTrigger()
	assert.Equal(t, b, false)

	s1, err = dnf.EncodeAddressPubkey(nil)
	assert.Equal(t, s1, emptStr)
	assert.Equal(t, err, errNodeStarting)

	ba, err := dnf.DecodeAddressPubkey("")
	assert.Nil(t, ba)
	assert.Equal(t, err, errNodeStarting)

	qh, err := dnf.GetQueryHandler("")
	assert.Nil(t, qh)
	assert.Equal(t, err, errNodeStarting)

	qp, err := dnf.GetPeerInfo("")
	assert.Nil(t, qp)
	assert.Equal(t, err, errNodeStarting)

	th, b := dnf.GetThrottlerForEndpoint("")
	assert.Nil(t, th)
	assert.Equal(t, b, false)

	ab, err := dnf.GetBlockByHash("", false)
	assert.Nil(t, ab)
	assert.Equal(t, err, errNodeStarting)

	ab, err = dnf.GetBlockByNonce(0, false)
	assert.Nil(t, ab)
	assert.Equal(t, err, errNodeStarting)

	err = dnf.Close()
	assert.Equal(t, err, errNodeStarting)

	u32 := dnf.GetNumCheckpointsFromAccountState()
	assert.Equal(t, u32, uint32(0))

	u32 = dnf.GetNumCheckpointsFromPeerState()
	assert.Equal(t, u32, uint32(0))

	vmoa := dnf.convertVmOutputToApiResponse(nil)
	assert.Nil(t, vmoa)

	assert.False(t, check.IfNil(dnf))

}
