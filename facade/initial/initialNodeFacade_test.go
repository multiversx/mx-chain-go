package initial

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
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
	inf := NewInitialNodeFacade(apiInterface, true)

	inf.SetSyncer(nil)
	b := inf.RestAPIServerDebugMode()
	assert.False(t, b)
	s1 := inf.RestApiInterface()
	assert.Equal(t, apiInterface, s1)
	s1, s2, err := inf.GetESDTBalance("", "")
	assert.Equal(t, emptyString, s1+s2)
	assert.Equal(t, errNodeStarting, err)
	v, err := inf.GetBalance("")
	assert.Nil(t, v)
	assert.Equal(t, errNodeStarting, err)

	s1, err = inf.GetUsername("")
	assert.Equal(t, emptyString, s1)
	assert.Equal(t, errNodeStarting, err)

	s1, err = inf.GetValueForKey("", "")
	assert.Equal(t, emptyString, s1)
	assert.Equal(t, errNodeStarting, err)

	s3, err := inf.GetAllESDTTokens("")
	assert.Nil(t, s3)
	assert.Equal(t, errNodeStarting, err)

	n1, n2, err := inf.CreateTransaction(uint64(0), "", "", []byte{0}, "",
		[]byte{0}, uint64(0), uint64(0), []byte{0}, "", "", uint32(0), uint32(0))
	assert.Nil(t, n1)
	assert.Nil(t, n2)
	assert.Equal(t, errNodeStarting, err)

	err = inf.ValidateTransaction(nil)
	assert.Equal(t, errNodeStarting, err)

	err = inf.ValidateTransactionForSimulation(nil, false)
	assert.Equal(t, errNodeStarting, err)

	v1, err := inf.ValidatorStatisticsApi()
	assert.Nil(t, v1)
	assert.Equal(t, errNodeStarting, err)

	u1, err := inf.SendBulkTransactions(nil)
	assert.Equal(t, uint64(0), u1)
	assert.Equal(t, errNodeStarting, err)

	u2, err := inf.SimulateTransactionExecution(nil)
	assert.Nil(t, u2)
	assert.Equal(t, errNodeStarting, err)

	t1, err := inf.GetTransaction("", false)
	assert.Nil(t, t1)
	assert.Equal(t, errNodeStarting, err)

	resp, err := inf.ComputeTransactionGasLimit(nil)
	assert.Nil(t, resp)
	assert.Equal(t, errNodeStarting, err)

	uac, err := inf.GetAccount("")
	assert.Equal(t, api.AccountResponse{}, uac)
	assert.Equal(t, errNodeStarting, err)

	hi, err := inf.GetHeartbeats()
	assert.Nil(t, hi)
	assert.NotNil(t, errNodeStarting, err)

	sm := inf.StatusMetrics()
	assert.NotNil(t, sm)

	vo, err := inf.ExecuteSCQuery(nil)
	assert.Nil(t, vo)
	assert.Equal(t, errNodeStarting, err)

	b = inf.PprofEnabled()
	assert.True(t, b)

	err = inf.Trigger(0, false)
	assert.Equal(t, errNodeStarting, err)

	b = inf.IsSelfTrigger()
	assert.False(t, b)

	s1, err = inf.EncodeAddressPubkey(nil)
	assert.Equal(t, emptyString, s1)
	assert.Equal(t, errNodeStarting, err)

	ba, err := inf.DecodeAddressPubkey("")
	assert.Nil(t, ba)
	assert.Equal(t, errNodeStarting, err)

	qh, err := inf.GetQueryHandler("")
	assert.Nil(t, qh)
	assert.Equal(t, errNodeStarting, err)

	qp, err := inf.GetPeerInfo("")
	assert.Nil(t, qp)
	assert.Equal(t, errNodeStarting, err)

	th, b := inf.GetThrottlerForEndpoint("")
	assert.Nil(t, th)
	assert.False(t, b)

	ab, err := inf.GetBlockByHash("", false)
	assert.Nil(t, ab)
	assert.Equal(t, errNodeStarting, err)

	c := inf.GetCode(nil)
	assert.Nil(t, c)

	ab, err = inf.GetBlockByNonce(0, false)
	assert.Nil(t, ab)
	assert.Equal(t, errNodeStarting, err)

	err = inf.Close()
	assert.Equal(t, errNodeStarting, err)

	u32 := inf.GetNumCheckpointsFromAccountState()
	assert.Equal(t, uint32(0), u32)

	u32 = inf.GetNumCheckpointsFromPeerState()
	assert.Equal(t, uint32(0), u32)

	proof, err := inf.GetProof("", "")
	assert.Nil(t, proof)
	assert.Equal(t, errNodeStarting, err)

	proof, err = inf.GetProofCurrentRootHash("")
	assert.Nil(t, proof)
	assert.Equal(t, errNodeStarting, err)

	b, err = inf.VerifyProof("", "", nil)
	assert.False(t, b)
	assert.Equal(t, errNodeStarting, err)

	sa, err := inf.GetNFTTokenIDsRegisteredByAddress("")
	assert.Nil(t, sa)
	assert.Equal(t, errNodeStarting, err)

	sa, err = inf.GetESDTsWithRole("", "")
	assert.Nil(t, sa)
	assert.Equal(t, errNodeStarting, err)

	err = inf.DirectTrigger(0, true)
	assert.Equal(t, errNodeStarting, err)

	asv, err := inf.GetDirectStakedList()
	assert.Nil(t, asv)
	assert.Equal(t, errNodeStarting, err)

	mss, err := inf.GetKeyValuePairs("")
	assert.Nil(t, mss)
	assert.Equal(t, errNodeStarting, err)

	ds, err := inf.GetDelegatorsList()
	assert.Nil(t, ds)
	assert.Equal(t, errNodeStarting, err)

	mssa, err := inf.GetESDTsRoles("")
	assert.Nil(t, mssa)
	assert.Equal(t, errNodeStarting, err)

	sa, err = inf.GetAllIssuedESDTs("")
	assert.Nil(t, sa)
	assert.Equal(t, errNodeStarting, err)

	s1, err = inf.GetTokenSupply("")
	assert.Empty(t, s1)
	assert.Equal(t, errNodeStarting, err)

	assert.False(t, check.IfNil(inf))
}
