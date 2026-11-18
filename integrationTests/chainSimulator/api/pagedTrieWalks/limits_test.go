package pagedTrieWalks

import (
	"encoding/hex"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
)

const (
	limitedMaxBytes  = 4096
	limitedMaxLeaves = 25
)

func TestChainSimulator_PagedTrieWalks_Limits(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	cs := startSimulator(t, 1, func(cfg *config.Configs) {
		cfg.GeneralConfig.TrieOperations.MaxBytes = limitedMaxBytes
		cfg.GeneralConfig.TrieOperations.MaxLeaves = limitedMaxLeaves
		closeRoute(cfg, "address", "/:address/esdt")
	})
	defer cs.Close()

	t.Run("pages stop on the byte budget", func(t *testing.T) {
		client := newRestClient(t, cs, 0)
		address := newWallet(t, cs, 0)
		pairs := randomPairs(100, "bytes", 300) // ~630 bytes charged per pair, ~6 pairs per page
		setPairs(t, cs, address, pairs)

		union, pageSizes := client.walkPairs(pageRequest{method: http.MethodGet, path: "/address/" + address + "/keys"}, "pairs")
		require.Equal(t, pairs, union)
		require.GreaterOrEqual(t, len(pageSizes), 10)
		for i, size := range pageSizes {
			require.LessOrEqual(t, size, 7, "page %d", i) // the budget is checked before every leaf: ~6 pairs, plus the one that reaches it
		}
	})

	t.Run("a value larger than the byte budget is returned", func(t *testing.T) {
		client := newRestClient(t, cs, 0)
		address := newWallet(t, cs, 0)
		pairs := randomPairs(10, "small", 8)
		bigKey := hex.EncodeToString([]byte("big"))
		pairs[bigKey] = hex.EncodeToString(make([]byte, 2*limitedMaxBytes))
		setPairs(t, cs, address, pairs)

		union, pageSizes := client.walkPairs(pageRequest{method: http.MethodPost, path: "/address/" + address + "/keys"}, "pairs")
		require.Equal(t, pairs, union)
		require.GreaterOrEqual(t, len(pageSizes), 2)
	})

	t.Run("pages stop on the configured maximum number of leaves", func(t *testing.T) {
		client := newRestClient(t, cs, 0)
		address := newWallet(t, cs, 0)
		pairs := randomPairs(100, "leaves", 4) // 100 pairs, ~40 bytes charged each, far below the byte budget
		setPairs(t, cs, address, pairs)

		for _, method := range []string{http.MethodGet, http.MethodPost} {
			union, pageSizes := client.walkPairs(pageRequest{method: method, path: "/address/" + address + "/keys", numKeys: 1000}, "pairs")
			require.Equal(t, pairs, union, method)
			require.GreaterOrEqual(t, len(pageSizes), 4, method)
			for i, size := range pageSizes {
				require.LessOrEqual(t, size, limitedMaxLeaves, "page %d", i)
			}
		}
	})

	t.Run("a closed route is closed for both methods", func(t *testing.T) {
		client := newRestClient(t, cs, 0)
		address := newWallet(t, cs, 0)
		for _, method := range []string{http.MethodGet, http.MethodPost} {
			p := client.requestPage(pageRequest{method: method, path: "/address/" + address + "/esdt"}, nil)
			require.Equal(t, http.StatusNotFound, p.status, method)
		}

		p := client.get("/address/"+address+"/keys", "")
		require.Equal(t, http.StatusOK, p.status)
	})
}

// closeRoute closes the route of the API package in the configuration
func closeRoute(cfg *config.Configs, packageName string, routeName string) {
	apiPackage := cfg.ApiRoutesConfig.APIPackages[packageName]
	for i := range apiPackage.Routes {
		if apiPackage.Routes[i].Name == routeName {
			apiPackage.Routes[i].Open = false
		}
	}
	cfg.ApiRoutesConfig.APIPackages[packageName] = apiPackage
}
