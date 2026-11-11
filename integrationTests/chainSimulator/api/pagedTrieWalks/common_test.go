package pagedTrieWalks

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common/checkpoint"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/vm"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	chainSimulatorAPI "github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
)

const (
	defaultPathToInitialConfig = "../../../../cmd/node/config/"
	maxPagesPerWalk            = 5000
)

// testSimulator is a chain simulator whose nodes serve the REST API
type testSimulator interface {
	chainSimulatorIntegrationTests.ChainSimulator
	GetRestAPIInterfaces() map[uint32]string
}

// startSimulator starts a chain simulator with a free port REST API on every node. The state of the tests is
// set through the simulator, which the Supernova block processing does not pick up, so Supernova stays off
func startSimulator(t *testing.T, numOfShards uint32, alterConfigs func(cfg *config.Configs)) testSimulator {
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:         true,
		BypassCreateBlockTimeCheck:     true,
		TempDir:                        t.TempDir(),
		PathToInitialConfig:            defaultPathToInitialConfig,
		NumOfShards:                    numOfShards,
		RoundDurationInMillis:          vm.RoundDurationInMillis,
		SupernovaRoundDurationInMillis: vm.SupernovaRoundDurationInMillis,
		RoundsPerEpoch:                 vm.RoundsPerEpoch,
		SupernovaRoundsPerEpoch:        vm.SupernovaRoundsPerEpoch,
		ApiInterface:                   chainSimulatorAPI.NewFreePortAPIConfigurator("localhost"),
		MinNodesPerShard:               3,
		MetaChainMinNodes:              3,
		NumNodesWaitingListMeta:        0,
		NumNodesWaitingListShard:       0,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.SupernovaEnableEpoch = integrationTests.UnreachableEpoch
			alterConfigs(cfg)
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	return cs
}

// restClient sends requests to the REST API of one node; it is bound to the test it is created in
type restClient struct {
	t       *testing.T
	baseURL string
	client  *http.Client
}

func newRestClient(t *testing.T, cs testSimulator, shardID uint32) *restClient {
	interfaces := cs.GetRestAPIInterfaces()
	require.Contains(t, interfaces, shardID)

	return &restClient{
		t:       t,
		baseURL: "http://" + interfaces[shardID],
		client:  &http.Client{Timeout: time.Minute},
	}
}

// genericResponse is the wire format of every REST response
type genericResponse struct {
	Data  map[string]json.RawMessage `json:"data"`
	Error string                     `json:"error"`
	Code  string                     `json:"code"`
}

// page is one decoded response of a paged trie walk
type page struct {
	status        int
	response      genericResponse
	iteratorState *checkpoint.Checkpoint
	blockInfo     api.BlockInfo
}

// tryDo sends the request and decodes the response; it can be called from any goroutine
func (rc *restClient) tryDo(method string, path string, body []byte) (page, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, rc.baseURL+path, reader)
	if err != nil {
		return page{}, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := rc.client.Do(req)
	if err != nil {
		return page{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return page{}, err
	}

	result := page{status: resp.StatusCode}
	if resp.StatusCode == http.StatusNotFound {
		result.response.Error = string(raw) // gin answers an unknown route with a plain text body
		return result, nil
	}
	err = json.Unmarshal(raw, &result.response)
	if err != nil {
		return page{}, fmt.Errorf("%w: status %d body %s", err, resp.StatusCode, string(raw))
	}

	if encodedState, found := result.response.Data["iteratorState"]; found {
		result.iteratorState = &checkpoint.Checkpoint{}
		err = json.Unmarshal(encodedState, result.iteratorState)
		if err != nil {
			return page{}, err
		}
	}
	if encodedBlockInfo, found := result.response.Data["blockInfo"]; found {
		err = json.Unmarshal(encodedBlockInfo, &result.blockInfo)
		if err != nil {
			return page{}, err
		}
	}

	return result, nil
}

func (rc *restClient) do(method string, path string, body []byte) page {
	result, err := rc.tryDo(method, path, body)
	require.Nil(rc.t, err)
	return result
}

// get sends a GET request with the given query string (without the leading "?")
func (rc *restClient) get(path string, query string) page {
	return rc.do(http.MethodGet, withQuery(path, query), nil)
}

// post sends a POST request with the given JSON body
func (rc *restClient) post(path string, body map[string]interface{}) page {
	encoded, err := json.Marshal(body)
	require.Nil(rc.t, err)
	return rc.do(http.MethodPost, path, encoded)
}

func withQuery(path string, query string) string {
	if query == "" {
		return path
	}
	return path + "?" + query
}

func joinQuery(query string, param string) string {
	if query == "" {
		return param
	}
	return query + "&" + param
}

// pageRequest describes the requests of a paged walk; the first request carries the start state, if any, and the
// continuation requests carry the returned state
type pageRequest struct {
	method  string
	path    string
	numKeys uint
	query   string                 // extra GET query parameters, without the leading "?"
	body    map[string]interface{} // extra POST body fields
	start   *checkpoint.Checkpoint // the state the walk starts from
}

// requestPage sends one page request, through the query for GET and through the body for POST
func (rc *restClient) requestPage(req pageRequest, state *checkpoint.Checkpoint) page {
	if req.method == http.MethodGet {
		query := req.query
		if req.numKeys != 0 {
			query = joinQuery(query, fmt.Sprintf("numKeys=%d", req.numKeys))
		}
		if state != nil {
			encoded, err := checkpoint.Encode(state)
			require.Nil(rc.t, err)
			query = joinQuery(query, "iteratorState="+encoded)
		}
		return rc.get(req.path, query)
	}

	body := make(map[string]interface{})
	for key, value := range req.body {
		body[key] = value
	}
	if req.numKeys != 0 {
		body["numKeys"] = req.numKeys
	}
	if state != nil {
		body["iteratorState"] = state
	}
	return rc.post(withQuery(req.path, req.query), body)
}

// walk pages an endpoint to the end, handing every page to collect, and checks that every page is served from
// the block of the first one and that a returned checkpoint carries the coordinates of that block (a request by
// block root hash knows only the root hash)
func (rc *restClient) walk(req pageRequest, collect func(p page)) []page {
	pages := make([]page, 0)
	state := req.start
	for {
		require.Less(rc.t, len(pages), maxPagesPerWalk, "the walk does not finish")
		p := rc.requestPage(req, state)
		require.Equal(rc.t, http.StatusOK, p.status, "page %d: %s", len(pages), p.response.Error)
		require.NotEmpty(rc.t, p.blockInfo.RootHash, "page %d has no block info", len(pages))
		if len(pages) > 0 {
			require.Equal(rc.t, pages[0].blockInfo, p.blockInfo, "page %d served from another block", len(pages))
		}
		if p.iteratorState != nil {
			require.Equal(rc.t, p.blockInfo.Nonce, p.iteratorState.BlockNonce)
			require.Equal(rc.t, p.blockInfo.Hash, p.iteratorState.BlockHash)
			require.Equal(rc.t, p.blockInfo.RootHash, p.iteratorState.BlockRootHash)
		}
		pages = append(pages, p)
		collect(p)

		if p.iteratorState == nil {
			return pages
		}
		state = p.iteratorState
	}
}

// decodeField decodes one field of the data of a page
func decodeField(t *testing.T, p page, field string, into interface{}) {
	raw, found := p.response.Data[field]
	require.True(t, found, "field %s missing from %v", field, p.response)
	require.Nil(t, json.Unmarshal(raw, into))
}

// walkPairs pages an endpoint whose pages hold key-value pairs under the given field; it returns their union and
// the number of pairs of every page
func (rc *restClient) walkPairs(req pageRequest, field string) (map[string]string, []int) {
	union := make(map[string]string)
	pageSizes := make([]int, 0)
	rc.walk(req, func(p page) {
		pairs := make(map[string]string)
		decodeField(rc.t, p, field, &pairs)
		pageSizes = append(pageSizes, len(pairs))
		for key, value := range pairs {
			_, duplicate := union[key]
			require.False(rc.t, duplicate, "key %s returned twice", key)
			union[key] = value
		}
	})

	return union, pageSizes
}

// walkTokens pages an endpoint whose pages hold a list of token identifiers under the given field
func (rc *restClient) walkTokens(req pageRequest, field string) ([]string, []int) {
	union := make([]string, 0)
	seen := make(map[string]struct{})
	pageSizes := make([]int, 0)
	rc.walk(req, func(p page) {
		tokens := make([]string, 0)
		decodeField(rc.t, p, field, &tokens)
		pageSizes = append(pageSizes, len(tokens))
		for _, token := range tokens {
			_, duplicate := seen[token]
			require.False(rc.t, duplicate, "token %s returned twice", token)
			seen[token] = struct{}{}
			union = append(union, token)
		}
	})

	return union, pageSizes
}

// walkFunc pages an endpoint and returns the union of the pages and the number of items of every page
type walkFunc func(req pageRequest) (interface{}, []int)

// samePagesForBothMethods walks the endpoint through GET and through POST with the given page sizes and checks
// that every walk returns the union of the single page walk, that no page holds more than numKeys items (every
// leaf is resolved as a node of its own, so the count is exact) and that the walk is split in pages when the items
// do not fit in one
func samePagesForBothMethods(t *testing.T, req pageRequest, pageSizes []uint, walk walkFunc) {
	req.numKeys = 0
	req.method = http.MethodGet
	expected, _ := walk(req)

	for _, method := range []string{http.MethodGet, http.MethodPost} {
		for _, numKeys := range pageSizes {
			req.method = method
			req.numKeys = numKeys
			union, sizes := walk(req)
			require.Equal(t, expected, union, "method %s numKeys %d", method, numKeys)

			numItems := 0
			for i, size := range sizes {
				require.LessOrEqual(t, size, int(numKeys), "method %s numKeys %d page %d", method, numKeys, i)
				numItems += size
			}
			if numItems > int(numKeys) {
				require.Greater(t, len(sizes), 1, "method %s numKeys %d", method, numKeys)
			}
		}
	}
}

// randomPageSizes returns a few fixed page sizes and a random one, logged for reproduction
func randomPageSizes(t *testing.T, max int) []uint {
	seed := time.Now().UnixNano()
	random := rand.New(rand.NewSource(seed))
	randomSize := uint(1 + random.Intn(max))
	t.Logf("random numKeys %d (seed %d)", randomSize, seed)

	return []uint{1, 3, randomSize}
}

// randomPairs makes numPairs hex encoded key-value pairs with values of valueSize bytes
func randomPairs(numPairs int, keyPrefix string, valueSize int) map[string]string {
	pairs := make(map[string]string, numPairs)
	for i := 0; i < numPairs; i++ {
		key := []byte(fmt.Sprintf("%s%08d", keyPrefix, i))
		value := make([]byte, valueSize)
		copy(value, fmt.Sprintf("value%08d", i))
		pairs[hex.EncodeToString(key)] = hex.EncodeToString(value)
	}

	return pairs
}

// setPairs sets the key-value pairs on the account and executes one block so the API serves them
func setPairs(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, address string, pairs map[string]string) {
	err := cs.SetKeyValueForAddress(address, pairs)
	require.Nil(t, err)
	err = cs.GenerateBlocks(1)
	require.Nil(t, err)
}

// newWallet creates a funded wallet on the given shard
func newWallet(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, shardID uint32) string {
	mintValue := big.NewInt(0).Mul(chainSimulatorIntegrationTests.OneEGLD, big.NewInt(100))
	wallet, err := cs.GenerateAndMintWalletAddress(shardID, mintValue)
	require.Nil(t, err)
	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	return wallet.Bech32
}
