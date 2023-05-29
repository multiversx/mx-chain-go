package initial

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-go/facade"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestInitialNodeFacade(t *testing.T) {
	t.Parallel()

	t.Run("nil status metrics should error", func(t *testing.T) {
		t.Parallel()

		inf, err := NewInitialNodeFacade("127.0.0.1:8080", true, nil)
		assert.Equal(t, facade.ErrNilStatusMetrics, err)
		assert.Nil(t, inf)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		inf, err := NewInitialNodeFacade("127.0.0.1:8080", true, &testscommon.StatusMetricsStub{})
		assert.Nil(t, err)
		assert.NotNil(t, inf)
	})
}

func TestInitialNodeFacade_AllMethodsShouldNotPanic(t *testing.T) {
	t.Parallel()
	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panic: %v", r))
		}
	}()

	apiInterface := "127.0.0.1:7799"
	inf, err := NewInitialNodeFacade(apiInterface, true, &testscommon.StatusMetricsStub{})
	assert.Nil(t, err)

	inf.SetSyncer(nil)
	b := inf.RestAPIServerDebugMode()
	assert.False(t, b)
	s1 := inf.RestApiInterface()
	assert.Equal(t, apiInterface, s1)
	s1, s2, _, err := inf.GetESDTBalance("", "", api.AccountQueryOptions{})
	assert.Equal(t, emptyString, s1+s2)
	assert.Equal(t, errNodeStarting, err)
	v, _, err := inf.GetBalance("", api.AccountQueryOptions{})
	assert.Nil(t, v)
	assert.Equal(t, errNodeStarting, err)

	s1, _, err = inf.GetUsername("", api.AccountQueryOptions{})
	assert.Equal(t, emptyString, s1)
	assert.Equal(t, errNodeStarting, err)

	s1, _, err = inf.GetValueForKey("", "", api.AccountQueryOptions{})
	assert.Equal(t, emptyString, s1)
	assert.Equal(t, errNodeStarting, err)

	s3, _, err := inf.GetAllESDTTokens("", api.AccountQueryOptions{})
	assert.Nil(t, s3)
	assert.Equal(t, errNodeStarting, err)

	n1, n2, err := inf.CreateTransaction(&external.ArgsCreateTransaction{})
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

	uac, _, err := inf.GetAccount("", api.AccountQueryOptions{})
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

	ab, err := inf.GetBlockByHash("", api.BlockQueryOptions{})
	assert.Nil(t, ab)
	assert.Equal(t, errNodeStarting, err)

	c := inf.GetCode(nil, api.AccountQueryOptions{})
	assert.Nil(t, c)

	ab, err = inf.GetBlockByNonce(0, api.BlockQueryOptions{})
	assert.Nil(t, ab)
	assert.Equal(t, errNodeStarting, err)

	ab, err = inf.GetBlockByRound(0, api.BlockQueryOptions{})
	assert.Nil(t, ab)
	assert.Equal(t, errNodeStarting, err)

	err = inf.Close()
	assert.Equal(t, errNodeStarting, err)

	proof, err := inf.GetProof("", "")
	assert.Nil(t, proof)
	assert.Equal(t, errNodeStarting, err)

	proof, err = inf.GetProofCurrentRootHash("")
	assert.Nil(t, proof)
	assert.Equal(t, errNodeStarting, err)

	b, err = inf.VerifyProof("", "", nil)
	assert.False(t, b)
	assert.Equal(t, errNodeStarting, err)

	sa, _, err := inf.GetNFTTokenIDsRegisteredByAddress("", api.AccountQueryOptions{})
	assert.Nil(t, sa)
	assert.Equal(t, errNodeStarting, err)

	sa, _, err = inf.GetESDTsWithRole("", "", api.AccountQueryOptions{})
	assert.Nil(t, sa)
	assert.Equal(t, errNodeStarting, err)

	err = inf.DirectTrigger(0, true)
	assert.Equal(t, errNodeStarting, err)

	asv, err := inf.GetDirectStakedList()
	assert.Nil(t, asv)
	assert.Equal(t, errNodeStarting, err)

	mss, _, err := inf.GetKeyValuePairs("", api.AccountQueryOptions{})
	assert.Nil(t, mss)
	assert.Equal(t, errNodeStarting, err)

	ds, err := inf.GetDelegatorsList()
	assert.Nil(t, ds)
	assert.Equal(t, errNodeStarting, err)

	mssa, _, err := inf.GetESDTsRoles("", api.AccountQueryOptions{})
	assert.Nil(t, mssa)
	assert.Equal(t, errNodeStarting, err)

	sa, err = inf.GetAllIssuedESDTs("")
	assert.Nil(t, sa)
	assert.Equal(t, errNodeStarting, err)

	supply, err := inf.GetTokenSupply("")
	assert.Nil(t, supply)
	assert.Equal(t, errNodeStarting, err)

	txPool, err := inf.GetTransactionsPool("")
	assert.Nil(t, txPool)
	assert.Equal(t, errNodeStarting, err)

	eligible, waiting, err := inf.GetGenesisNodesPubKeys()
	assert.Nil(t, eligible)
	assert.Nil(t, waiting)
	assert.Equal(t, errNodeStarting, err)

	gasConfig, err := inf.GetGasConfigs()
	assert.Nil(t, gasConfig)
	assert.Equal(t, errNodeStarting, err)

	txs, err := inf.GetTransactionsPoolForSender("", "")
	assert.Nil(t, txs)
	assert.Equal(t, errNodeStarting, err)

	nonce, err := inf.GetLastPoolNonceForSender("")
	assert.Equal(t, uint64(0), nonce)
	assert.Equal(t, errNodeStarting, err)

	guardianData, _, err := inf.GetGuardianData("", api.AccountQueryOptions{})
	assert.Equal(t, api.GuardianData{}, guardianData)
	assert.Equal(t, errNodeStarting, err)

	mainTrieResponse, dataTrieResponse, err := inf.GetProofDataTrie("", "", "")
	assert.Nil(t, mainTrieResponse)
	assert.Nil(t, dataTrieResponse)
	assert.Equal(t, errNodeStarting, err)

	codeHash, blockInfo, err := inf.GetCodeHash("", api.AccountQueryOptions{})
	assert.Nil(t, codeHash)
	assert.Equal(t, api.BlockInfo{}, blockInfo)
	assert.Equal(t, errNodeStarting, err)

	accountsResponse, blockInfo, err := inf.GetAccounts([]string{}, api.AccountQueryOptions{})
	assert.Nil(t, accountsResponse)
	assert.Equal(t, api.BlockInfo{}, blockInfo)
	assert.Equal(t, errNodeStarting, err)

	stakeValue, err := inf.GetTotalStakedValue()
	assert.Nil(t, stakeValue)
	assert.Equal(t, errNodeStarting, err)

	ratings := inf.GetConnectedPeersRatings()
	assert.Equal(t, "", ratings)

	epochStartData, err := inf.GetEpochStartDataAPI(0)
	assert.Nil(t, epochStartData)
	assert.Equal(t, errNodeStarting, err)

	alteredAcc, err := inf.GetAlteredAccountsForBlock(api.GetAlteredAccountsForBlockOptions{})
	assert.Nil(t, alteredAcc)
	assert.Equal(t, errNodeStarting, err)

	block, err := inf.GetInternalMetaBlockByHash(0, "")
	assert.Nil(t, block)
	assert.Equal(t, errNodeStarting, err)

	block, err = inf.GetInternalMetaBlockByNonce(0, 0)
	assert.Nil(t, block)
	assert.Equal(t, errNodeStarting, err)

	block, err = inf.GetInternalMetaBlockByRound(0, 0)
	assert.Nil(t, block)
	assert.Equal(t, errNodeStarting, err)

	block, err = inf.GetInternalStartOfEpochMetaBlock(0, 0)
	assert.Nil(t, block)
	assert.Equal(t, errNodeStarting, err)

	validatorsInfo, err := inf.GetInternalStartOfEpochValidatorsInfo(0)
	assert.Nil(t, validatorsInfo)
	assert.Equal(t, errNodeStarting, err)

	block, err = inf.GetInternalShardBlockByHash(0, "")
	assert.Nil(t, block)
	assert.Equal(t, errNodeStarting, err)

	block, err = inf.GetInternalShardBlockByNonce(0, 0)
	assert.Nil(t, block)
	assert.Equal(t, errNodeStarting, err)

	block, err = inf.GetInternalShardBlockByRound(0, 0)
	assert.Nil(t, block)
	assert.Equal(t, errNodeStarting, err)

	block, err = inf.GetInternalMiniBlockByHash(0, "", 0)
	assert.Nil(t, block)
	assert.Equal(t, errNodeStarting, err)

	esdtData, blockInfo, err := inf.GetESDTData("", "", 0, api.AccountQueryOptions{})
	assert.Nil(t, esdtData)
	assert.Equal(t, api.BlockInfo{}, blockInfo)
	assert.Equal(t, errNodeStarting, err)

	genesisBalances, err := inf.GetGenesisBalances()
	assert.Nil(t, genesisBalances)
	assert.Equal(t, errNodeStarting, err)

	txPoolGaps, err := inf.GetTransactionsPoolNonceGapsForSender("")
	assert.Nil(t, txPoolGaps)
	assert.Equal(t, errNodeStarting, err)

	assert.NotNil(t, inf)
}

func TestInitialNodeFacade_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var inf *initialNodeFacade
	assert.True(t, inf.IsInterfaceNil())

	inf, _ = NewInitialNodeFacade("127.0.0.1:7799", true, &testscommon.StatusMetricsStub{})
	assert.False(t, inf.IsInterfaceNil())
}
