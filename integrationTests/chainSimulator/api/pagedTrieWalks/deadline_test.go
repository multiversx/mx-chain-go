package pagedTrieWalks

import (
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/config"
)

const (
	numDeadlinePairs      = 200_000
	numDeadlineTokens     = 3
	numConcurrentRequests = 16
)

// The walk budget of a request is 100ms (a 1s HTTP timeout with 900ms reserved for writing the response), the
// leaf and byte limits are above the fixture so only the deadline can end a page, and a single request is
// admitted at a time; the web server antiflood is disabled to show the trie operations limits do not depend on it
func TestChainSimulator_PagedTrieWalks_DeadlineAndThrottling(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs := startSimulator(t, 1, func(cfg *config.Configs) {
		cfg.GeneralConfig.TrieOperations.HTTPTimeoutMilliseconds = 1000
		cfg.GeneralConfig.TrieOperations.DeadlineReserveMilliseconds = 900
		cfg.GeneralConfig.TrieOperations.MaxConcurrentRequests = 1
		cfg.GeneralConfig.TrieOperations.MaxLeaves = 10 * numDeadlinePairs
		cfg.GeneralConfig.TrieOperations.MaxBytes = 100 * 1024 * 1024
		cfg.GeneralConfig.WebServerAntiflood.WebServerAntifloodEnabled = false
	})
	defer cs.Close()

	address := newWallet(t, cs, 0)
	pairs := randomPairs(numDeadlinePairs, "deadline", 16)
	tokens := fungibleTokenPairs(t, numDeadlineTokens)
	for key, value := range tokens {
		pairs[key] = value
	}
	setPairs(t, cs, address, pairs)

	t.Run("a page cut by the deadline is returned with its checkpoint", func(t *testing.T) {
		client := newRestClient(t, cs, 0)
		for _, method := range []string{http.MethodGet, http.MethodPost} {
			union, pageSizes := client.walkPairs(pageRequest{method: method, path: "/address/" + address + "/keys"}, "pairs")
			require.Equal(t, len(pairs), len(union), method)
			require.Equal(t, pairs, union, method)
			require.Greater(t, len(pageSizes), 1, method)
			t.Logf("%s: %d pages", method, len(pageSizes))
		}
	})

	t.Run("pages whose leaves are all filtered out are empty but carry a checkpoint", func(t *testing.T) {
		client := newRestClient(t, cs, 0)
		union := make(map[string]struct{})
		numEmptyPages := 0
		pages := client.walk(pageRequest{method: http.MethodGet, path: "/address/" + address + "/esdt"}, func(p page) {
			esdts := make(map[string]interface{})
			decodeField(t, p, "esdts", &esdts)
			if len(esdts) == 0 && p.iteratorState != nil {
				numEmptyPages++
			}
			for identifier := range esdts {
				_, duplicate := union[identifier]
				require.False(t, duplicate, "token %s returned twice", identifier)
				union[identifier] = struct{}{}
			}
		})
		require.Greater(t, len(pages), 1)
		require.Greater(t, numEmptyPages, 0)
		require.Len(t, union, numDeadlineTokens)
		for i := 0; i < numDeadlineTokens; i++ {
			require.Contains(t, union, stateTokenIdentifier(i))
		}
	})

	t.Run("concurrent requests above the limit are rejected as busy", func(t *testing.T) {
		client := newRestClient(t, cs, 0)
		results := make([]page, numConcurrentRequests)
		errs := make([]error, numConcurrentRequests)
		start := make(chan struct{})
		wg := sync.WaitGroup{}
		for i := 0; i < numConcurrentRequests; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				<-start
				method, body := http.MethodGet, []byte(nil)
				if index%2 == 1 {
					method, body = http.MethodPost, []byte(`{}`)
				}
				results[index], errs[index] = client.tryDo(method, "/address/"+address+"/esdt", body)
			}(i)
		}
		close(start)
		wg.Wait()

		numBusy, numServed := 0, 0
		for i, p := range results {
			require.Nil(t, errs[i])
			switch p.status {
			case http.StatusTooManyRequests:
				require.Equal(t, string(shared.ReturnCodeSystemBusy), p.response.Code)
				numBusy++
			case http.StatusOK:
				numServed++
			default:
				require.Fail(t, "unexpected status", "%d: %s", p.status, p.response.Error)
			}
		}
		require.Greater(t, numBusy, 0)
		require.Greater(t, numServed, 0)

		// the slot is released
		p := client.get("/address/"+address+"/esdt", "")
		require.Equal(t, http.StatusOK, p.status)
	})
}
