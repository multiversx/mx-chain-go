package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ginFramework "github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/require"
)

type trieDeadlineRecorder struct{ *httptest.ResponseRecorder }

func (*trieDeadlineRecorder) SetReadDeadline(time.Time) error  { return nil }
func (*trieDeadlineRecorder) SetWriteDeadline(time.Time) error { return nil }

func TestTrieLimiterInstalledWithoutWebAntiflood(t *testing.T) {
	args := createMockArgsNewWebServer()
	args.AntiFloodConfig.WebServerAntifloodEnabled = false
	args.ApiConfig.Logging.LoggingEnabled = false
	args.TrieOperations = config.TrieOperationsConfig{MaxConcurrentRequests: 1}
	ws, err := NewGinWebServerHandler(args)
	require.NoError(t, err)
	processors, err := ws.createMiddlewareLimiters()
	require.NoError(t, err)
	r := ginFramework.New()
	for _, proc := range processors {
		r.Use(proc.MiddlewareHandlerFunc())
	}
	r.GET("/network/esdts", func(c *ginFramework.Context) {
		inner := &trieDeadlineRecorder{ResponseRecorder: httptest.NewRecorder()}
		r.ServeHTTP(inner, httptest.NewRequest(http.MethodGet, "/address/alice/keys", nil))
		require.Equal(t, http.StatusTooManyRequests, inner.Code)
		c.Status(http.StatusOK)
	})
	r.GET("/address/:address/keys", func(c *ginFramework.Context) { c.Status(http.StatusOK) })
	w := &trieDeadlineRecorder{ResponseRecorder: httptest.NewRecorder()}
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/network/esdts", nil))
	require.Equal(t, http.StatusOK, w.Code)
}

func TestTrieLimiterPreservesSourceAccounting(t *testing.T) {
	args := createMockArgsNewWebServer()
	args.ApiConfig.Logging.LoggingEnabled = false
	args.AntiFloodConfig.SameSourceRequests = 2
	args.AntiFloodConfig.SameSourceResetIntervalInSec = 3600
	args.AntiFloodConfig.SimultaneousRequests = 10
	args.TrieOperations.MaxConcurrentRequests = 1
	ws, err := NewGinWebServerHandler(args)
	require.NoError(t, err)
	processors, err := ws.createMiddlewareLimiters()
	require.NoError(t, err)
	t.Cleanup(ws.cancelFunc)
	r := ginFramework.New()
	for _, proc := range processors {
		r.Use(proc.MiddlewareHandlerFunc())
	}
	r.GET("/network/esdts", func(c *ginFramework.Context) {
		for _, message := range []string{"too many concurrent trie operations", "for address"} {
			inner := &trieDeadlineRecorder{ResponseRecorder: httptest.NewRecorder()}
			r.ServeHTTP(inner, httptest.NewRequest(http.MethodGet, "/address/alice/keys", nil))
			require.Equal(t, http.StatusTooManyRequests, inner.Code)
			require.Contains(t, inner.Body.String(), message)
			require.True(t, inner.Flushed)
		}
		c.Status(http.StatusOK)
	})
	r.GET("/address/:address/keys", func(*ginFramework.Context) { t.Fatal("rejected request reached handler") })
	w := &trieDeadlineRecorder{ResponseRecorder: httptest.NewRecorder()}
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/network/esdts", nil))
	require.Equal(t, http.StatusOK, w.Code)
}
