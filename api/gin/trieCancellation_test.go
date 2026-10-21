package gin

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ginFramework "github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/facade"
	facadeMock "github.com/multiversx/mx-chain-go/facade/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/stretchr/testify/require"
)

func TestTrieRequestCancellationReachesNode(t *testing.T) {
	paths := []string{
		"/address/alice?withKeys=true", "/address/alice/keys", "/address/alice/esdt",
		"/address/alice/esdts/roles", "/address/alice/esdts-with-role/ESDTRoleLocalMint",
		"/address/alice/registered-nfts", "/address/iterate-keys", "/network/esdts",
		"/network/esdt/fungible-tokens", "/network/esdt/semi-fungible-tokens", "/network/esdt/non-fungible-tokens",
	}
	for _, antiflood := range []bool{false, true} {
		for _, path := range paths {
			t.Run(fmt.Sprintf("antiflood=%t%s", antiflood, path), func(t *testing.T) {
				testTrieContext(t, antiflood, path, false)
			})
		}
		t.Run(fmt.Sprintf("deadline/antiflood=%t", antiflood), func(t *testing.T) {
			testTrieContext(t, antiflood, "/address/alice/keys", true)
		})
	}
}

func testTrieContext(t *testing.T, antiflood bool, path string, expire bool) {
	t.Helper()
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	called := false
	observe := func(ctx context.Context) error {
		called = true
		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		require.LessOrEqual(t, time.Until(deadline), time.Second)
		if !expire {
			cancel()
		}
		select {
		case <-ctx.Done():
		case <-time.After(2 * time.Second):
			t.Fatal("request cancellation did not reach node")
		}
		want := context.Canceled
		if expire {
			want = context.DeadlineExceeded
		}
		require.ErrorIs(t, ctx.Err(), want)
		return ctx.Err()
	}
	n := &facadeMock.NodeStub{
		GetAccountWithKeysCalled: func(_ string, _ api.AccountQueryOptions, ctx context.Context) (api.AccountResponse, api.BlockInfo, error) {
			return api.AccountResponse{}, api.BlockInfo{}, observe(ctx)
		},
		GetKeyValuePairsCalled: func(_ string, _ api.AccountQueryOptions, ctx context.Context) (map[string]string, api.BlockInfo, error) {
			return nil, api.BlockInfo{}, observe(ctx)
		},
		IterateKeysCalled: func(_ string, _ uint, _ [][]byte, _ api.AccountQueryOptions, ctx context.Context) (map[string]string, [][]byte, api.BlockInfo, error) {
			return nil, nil, api.BlockInfo{}, observe(ctx)
		},
		GetAllESDTTokensCalled: func(_ string, _ api.AccountQueryOptions, ctx context.Context) (map[string]*esdt.ESDigitalToken, api.BlockInfo, error) {
			return nil, api.BlockInfo{}, observe(ctx)
		},
		GetESDTsRolesCalled: func(_ string, _ api.AccountQueryOptions, ctx context.Context) (map[string][]string, api.BlockInfo, error) {
			return nil, api.BlockInfo{}, observe(ctx)
		},
		GetNFTTokenIDsRegisteredByAddressCalled: func(_ string, _ api.AccountQueryOptions, ctx context.Context) ([]string, api.BlockInfo, error) {
			return nil, api.BlockInfo{}, observe(ctx)
		},
		GetESDTsWithRoleCalled: func(_ string, _ string, _ api.AccountQueryOptions, ctx context.Context) ([]string, api.BlockInfo, error) {
			return nil, api.BlockInfo{}, observe(ctx)
		},
		GetAllIssuedESDTsCalled: func(_ string, ctx context.Context) ([]string, error) {
			return nil, observe(ctx)
		},
	}
	args := createMockArgsNewWebServer()
	args.ApiConfig.Logging.LoggingEnabled = false
	args.ApiConfig.APIPackages = make(map[string]config.APIPackageConfig)
	for group, routes := range map[string][]string{
		"address": {"/:address", "/:address/keys", "/:address/esdt", "/:address/esdts/roles", "/:address/esdts-with-role/:role", "/:address/registered-nfts", "/iterate-keys"},
		"network": {"/esdts", "/esdt/fungible-tokens", "/esdt/semi-fungible-tokens", "/esdt/non-fungible-tokens"},
	} {
		pkg := config.APIPackageConfig{}
		for _, route := range routes {
			pkg.Routes = append(pkg.Routes, config.RouteConfig{Name: route, Open: true})
		}
		args.ApiConfig.APIPackages[group] = pkg
	}
	args.AntiFloodConfig.WebServerAntifloodEnabled = antiflood
	args.AntiFloodConfig.TrieOperationsDeadlineMilliseconds = 10000
	args.TrieOperations.HTTPTimeoutMilliseconds = 1000
	if expire {
		args.TrieOperations.HTTPTimeoutMilliseconds = 20
	}
	var err error
	args.Facade, err = facade.NewNodeFacade(facade.ArgNodeFacade{
		Node: n, ApiResolver: &facadeMock.ApiResolverStub{}, ApiRoutesConfig: args.ApiConfig,
		WsAntifloodConfig: args.AntiFloodConfig, AccountsStateAPI: &stateMock.AccountsStub{}, Blockchain: &testscommon.ChainHandlerStub{},
	})
	require.NoError(t, err)
	ws, err := NewGinWebServerHandler(args)
	require.NoError(t, err)
	processors, err := ws.createMiddlewareLimiters()
	require.NoError(t, err)
	if ws.cancelFunc != nil {
		t.Cleanup(ws.cancelFunc)
	}
	r := ginFramework.New()
	for _, proc := range processors {
		r.Use(proc.MiddlewareHandlerFunc())
	}
	require.NoError(t, ws.createGroups())
	ws.registerRoutes(r)
	method, body := http.MethodGet, ""
	if path == "/address/iterate-keys" {
		method, body = http.MethodPost, `{"address":"alice","numKeys":1}`
	}
	req := httptest.NewRequest(method, path, strings.NewReader(body)).WithContext(parent)
	w := &trieDeadlineRecorder{ResponseRecorder: httptest.NewRecorder()}
	r.ServeHTTP(w, req)
	require.True(t, called)
	require.Equal(t, http.StatusInternalServerError, w.Code)
}
