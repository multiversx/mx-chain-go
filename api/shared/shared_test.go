package shared

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/stretchr/testify/require"
)

func TestRespondWithInternalErrorRecognizesWrappedTrieBusy(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	RespondWithInternalError(c, fmt.Errorf("listing failed"), fmt.Errorf("wrapped: %w", common.ErrTrieOperationsBusy))
	require.Equal(t, http.StatusTooManyRequests, w.Code)
	require.Contains(t, w.Body.String(), string(ReturnCodeSystemBusy))
}
