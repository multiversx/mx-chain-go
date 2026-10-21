package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/require"
)

type deadlineRecorder struct {
	*httptest.ResponseRecorder
	readDeadline  time.Time
	writeDeadline time.Time
}

func newDeadlineRecorder() *deadlineRecorder {
	return &deadlineRecorder{ResponseRecorder: httptest.NewRecorder()}
}

func (w *deadlineRecorder) SetReadDeadline(deadline time.Time) error {
	w.readDeadline = deadline
	return nil
}
func (w *deadlineRecorder) SetWriteDeadline(deadline time.Time) error {
	w.writeDeadline = deadline
	return nil
}

type blockedTrieResponse struct {
	*deadlineRecorder
	started chan struct{}
	release chan struct{}
	once    sync.Once
}

func (w *blockedTrieResponse) Write(b []byte) (int, error) {
	w.once.Do(func() { close(w.started) })
	<-w.release
	return w.ResponseRecorder.Write(b)
}

func TestTrieHTTPAdmissionCoversResponseAndRoutes(t *testing.T) {
	r := gin.New()
	installTrieLimiters(r, config.TrieOperationsConfig{MaxConcurrentRequests: 1})
	handler := func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }
	paths := []string{"/address/:address/keys", "/address/:address/esdt", "/address/:address/esdts/roles", "/address/:address/esdts-with-role/:role", "/address/:address/registered-nfts", "/network/esdts", "/network/esdt/fungible-tokens", "/network/esdt/semi-fungible-tokens", "/network/esdt/non-fungible-tokens", "/address/:address"}
	for _, path := range paths {
		r.GET(path, handler)
	}
	r.POST("/address/iterate-keys", handler)
	w := &blockedTrieResponse{deadlineRecorder: newDeadlineRecorder(), started: make(chan struct{}), release: make(chan struct{})}
	done := make(chan struct{})
	go func() {
		defer close(done)
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/address/alice/keys", nil))
	}()
	select {
	case <-w.started:
	case <-time.After(5 * time.Second):
		t.Fatal("response did not start")
	}
	for _, path := range []string{"/address/bob/keys", "/address/bob/esdt", "/address/bob/esdts/roles", "/address/bob/esdts-with-role/role", "/address/bob/registered-nfts", "/network/esdts", "/network/esdt/fungible-tokens", "/network/esdt/semi-fungible-tokens", "/network/esdt/non-fungible-tokens", "/address/bob?withKeys=true", "/address/bob?withKeys=1", "/address/iterate-keys"} {
		method := http.MethodGet
		if path == "/address/iterate-keys" {
			method = http.MethodPost
		}
		recorder := newDeadlineRecorder()
		r.ServeHTTP(recorder, httptest.NewRequest(method, path, nil))
		require.Equal(t, http.StatusTooManyRequests, recorder.Code, path)
		require.Contains(t, recorder.Body.String(), "system_busy")
	}
	for _, path := range []string{"/address/bob", "/address/bob?withKeys=false"} {
		recorder := newDeadlineRecorder()
		r.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		require.Equal(t, http.StatusOK, recorder.Code)
	}
	close(w.release)
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("response did not finish")
	}
	recorder := newDeadlineRecorder()
	r.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/network/esdts", nil))
	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestTrieHTTPAdmissionReleasedOnPanic(t *testing.T) {
	r := gin.New()
	limiter := installTrieLimiters(r, config.TrieOperationsConfig{MaxConcurrentRequests: 1})
	r.GET("/network/esdts", func(*gin.Context) { panic("test") })
	require.Panics(t, func() {
		r.ServeHTTP(newDeadlineRecorder(), httptest.NewRequest(http.MethodGet, "/network/esdts", nil))
	})
	require.Empty(t, limiter.slots)
}

type deadlineBody struct{ recorder *deadlineRecorder }

func (b *deadlineBody) Read([]byte) (int, error) {
	if b.recorder.readDeadline.IsZero() {
		return 0, io.ErrUnexpectedEOF
	}
	<-time.After(time.Until(b.recorder.readDeadline))
	return 0, os.ErrDeadlineExceeded
}

func (b *deadlineBody) Close() error { return nil }

type deadlineWriteRecorder struct{ *deadlineRecorder }

func (w *deadlineWriteRecorder) Write([]byte) (int, error) {
	if w.writeDeadline.IsZero() {
		return 0, io.ErrUnexpectedEOF
	}
	<-time.After(time.Until(w.writeDeadline))
	return 0, os.ErrDeadlineExceeded
}

func TestTrieHTTPDeadlinesReleaseAdmission(t *testing.T) {
	for _, direction := range []string{"read", "write"} {
		t.Run(direction, func(t *testing.T) {
			r := gin.New()
			limiter := installTrieLimiters(r, config.TrieOperationsConfig{MaxConcurrentRequests: 1, HTTPTimeoutMilliseconds: 20})
			var operationErr error
			r.POST("/address/iterate-keys", func(c *gin.Context) {
				if direction == "read" {
					_, operationErr = io.ReadAll(c.Request.Body)
				} else {
					_, operationErr = c.Writer.Write([]byte("response"))
				}
			})
			w := newDeadlineRecorder()
			req := httptest.NewRequest(http.MethodPost, "/address/iterate-keys", nil)
			if direction == "read" {
				req.Body = &deadlineBody{recorder: w}
				r.ServeHTTP(w, req)
			} else {
				r.ServeHTTP(&deadlineWriteRecorder{deadlineRecorder: w}, req)
			}
			require.ErrorIs(t, operationErr, os.ErrDeadlineExceeded)
			require.Empty(t, limiter.slots)
			require.False(t, w.readDeadline.IsZero())
			require.False(t, w.writeDeadline.IsZero())
			require.True(t, w.writeDeadline.Before(time.Now()))
		})
	}
}

func TestTrieHTTPBodyLimit(t *testing.T) {
	for _, knownLength := range []bool{true, false} {
		r := gin.New()
		limiter := installTrieLimiters(r, config.TrieOperationsConfig{MaxRequestBodyBytes: 4})
		var readErr error
		called := false
		r.POST("/address/iterate-keys", func(c *gin.Context) { called = true; _, readErr = io.ReadAll(c.Request.Body) })
		req := httptest.NewRequest(http.MethodPost, "/address/iterate-keys", strings.NewReader("12345"))
		if !knownLength {
			req.ContentLength = -1
		}
		w := newDeadlineRecorder()
		r.ServeHTTP(w, req)
		if knownLength {
			require.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
			require.False(t, called)
		} else {
			var limitErr *http.MaxBytesError
			require.ErrorAs(t, readErr, &limitErr)
		}
		require.Empty(t, limiter.slots)
	}
}

func TestTrieHTTPRequiresDeadlineSupport(t *testing.T) {
	r := gin.New()
	installTrieLimiters(r, config.TrieOperationsConfig{})
	r.GET("/network/esdts", func(*gin.Context) { t.Fatal("unsupported writer reached handler") })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/network/esdts", nil))
	require.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestTrieHTTPDeadlinesWorkWithGinServer(t *testing.T) {
	r := gin.New()
	installTrieLimiters(r, config.TrieOperationsConfig{})
	r.GET("/network/esdts", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"tokens": []string{}}) })
	server := httptest.NewServer(r)
	defer server.Close()
	for i := 0; i < 2; i++ {
		response, err := server.Client().Get(server.URL + "/network/esdts")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)
		_, err = io.Copy(io.Discard, response.Body)
		require.NoError(t, err)
		require.NoError(t, response.Body.Close())
	}
}

func installTrieLimiters(r *gin.Engine, cfg config.TrieOperationsConfig) *trieOperationsThrottler {
	limiter := NewTrieOperationsThrottler(cfg)
	r.Use(NewTrieOperationsTransport(cfg).MiddlewareHandlerFunc(), limiter.MiddlewareHandlerFunc())
	return limiter
}

type checkedDeadlineRecorder struct {
	*deadlineRecorder
	check func()
}

func (w *checkedDeadlineRecorder) Write(data []byte) (int, error) {
	w.check()
	return w.ResponseRecorder.Write(data)
}

func (w *checkedDeadlineRecorder) Flush() {
	w.check()
	w.ResponseRecorder.Flush()
}

func TestTrieRejectionKeepsDeadlinesThroughFlush(t *testing.T) {
	r := gin.New()
	limiter := installTrieLimiters(r, config.TrieOperationsConfig{MaxConcurrentRequests: 1})
	limiter.slots <- struct{}{}
	defer func() { <-limiter.slots }()
	r.GET("/network/esdts", func(*gin.Context) { t.Fatal("busy request reached handler") })
	w := &checkedDeadlineRecorder{deadlineRecorder: newDeadlineRecorder()}
	checks := 0
	w.check = func() {
		checks++
		require.False(t, w.readDeadline.IsZero())
		require.False(t, w.writeDeadline.IsZero())
	}
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/network/esdts", nil))
	require.Equal(t, http.StatusTooManyRequests, w.Code)
	require.GreaterOrEqual(t, checks, 2)
	require.True(t, w.Flushed)
	require.True(t, w.readDeadline.IsZero())
	require.True(t, w.writeDeadline.IsZero())
	require.Len(t, limiter.slots, 1)
}

func TestTrieAdmissionHeldThroughFinalFlush(t *testing.T) {
	r := gin.New()
	limiter := installTrieLimiters(r, config.TrieOperationsConfig{MaxConcurrentRequests: 1})
	r.GET("/network/esdts", func(c *gin.Context) { c.Status(http.StatusOK) })
	w := &checkedDeadlineRecorder{deadlineRecorder: newDeadlineRecorder()}
	checks := 0
	w.check = func() {
		if checks == 0 {
			require.Len(t, limiter.slots, 1)
		}
		checks++
		require.False(t, w.writeDeadline.IsZero())
	}
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/network/esdts", nil))
	require.Positive(t, checks)
	require.Empty(t, limiter.slots)
}
