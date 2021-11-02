package groups_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	apiErrors "github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/groups"
	"github.com/ElrondNetwork/elrond-go/api/mock"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


type TriggerResponse struct {
	Status string `json:"status"`
}

func TestNewHardforkGroup(t *testing.T) {
	t.Parallel()

	t.Run("nil facade", func(t *testing.T) {
		hg, err := groups.NewHardforkGroup(nil)
		require.True(t, errors.Is(err, apiErrors.ErrNilFacadeHandler))
		require.Nil(t, hg)
	})

	t.Run("should work", func(t *testing.T) {
		hg, err := groups.NewHardforkGroup(&mock.HardforkFacade{})
		require.NoError(t, err)
		require.NotNil(t, hg)
	})
}

func TestTrigger_TriggerCanNotExecuteShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	hardforkFacade := &mock.HardforkFacade{
		TriggerCalled: func(_ uint32, _ bool) error {
			return expectedErr
		},
	}

	hardforkGroup, err := groups.NewHardforkGroup(hardforkFacade)
	require.NoError(t, err)

	ws := startWebServer(hardforkGroup, "hardfork", getHardforkRoutesConfig())

	hr := &groups.HardforkRequest{
		Epoch: 4,
	}

	buffHr, _ := json.Marshal(hr)
	req, _ := http.NewRequest("POST", "/hardfork/trigger", bytes.NewBuffer(buffHr))
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Contains(t, response.Error, expectedErr.Error())
}

func TestTrigger_TriggerWrongRequestTypeShouldErr(t *testing.T) {
	t.Parallel()

	hardforkGroup, err := groups.NewHardforkGroup(&mock.HardforkFacade{})
	require.NoError(t, err)

	ws := startWebServer(hardforkGroup, "hardfork", getHardforkRoutesConfig())

	req, _ := http.NewRequest("POST", "/hardfork/trigger", bytes.NewBuffer([]byte("wrong buffer")))
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	triggerResponse := TriggerResponse{}
	loadResponse(resp.Body, &triggerResponse)

	assert.Equal(t, resp.Code, http.StatusBadRequest)
}

func TestTrigger_ManualShouldWork(t *testing.T) {
	t.Parallel()

	recoveredEpoch := uint32(0)
	hr := &groups.HardforkRequest{
		Epoch: 4,
	}
	buffHr, _ := json.Marshal(hr)

	hardforkFacade := &mock.HardforkFacade{
		TriggerCalled: func(epoch uint32, _ bool) error {
			atomic.StoreUint32(&recoveredEpoch, epoch)

			return nil
		},
		IsSelfTriggerCalled: func() bool {
			return false
		},
	}
	hardforkGroup, err := groups.NewHardforkGroup(hardforkFacade)
	require.NoError(t, err)

	ws := startWebServer(hardforkGroup, "hardfork", getHardforkRoutesConfig())

	req, _ := http.NewRequest("POST", "/hardfork/trigger", bytes.NewBuffer(buffHr))
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	triggerResponse := TriggerResponse{}
	mapResponseData := response.Data.(map[string]interface{})
	mapResponseBytes, _ := json.Marshal(&mapResponseData)
	_ = json.Unmarshal(mapResponseBytes, &triggerResponse)

	assert.Equal(t, resp.Code, http.StatusOK)
	assert.Equal(t, groups.ExecManualTrigger, triggerResponse.Status)
	assert.Equal(t, hr.Epoch, atomic.LoadUint32(&recoveredEpoch))
}

func TestTrigger_BroadcastShouldWork(t *testing.T) {
	t.Parallel()

	hardforkFacade := &mock.HardforkFacade{
		TriggerCalled: func(_ uint32, _ bool) error {
			return nil
		},
		IsSelfTriggerCalled: func() bool {
			return true
		},
	}

	hardforkGroup, err := groups.NewHardforkGroup(hardforkFacade)
	require.NoError(t, err)

	ws := startWebServer(hardforkGroup, "hardfork", getHardforkRoutesConfig())

	hr := &groups.HardforkRequest{
		Epoch: 4,
	}
	buffHr, _ := json.Marshal(hr)
	req, _ := http.NewRequest("POST", "/hardfork/trigger", bytes.NewBuffer(buffHr))
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	triggerResponse := TriggerResponse{}
	mapResponseData := response.Data.(map[string]interface{})
	mapResponseBytes, _ := json.Marshal(&mapResponseData)
	_ = json.Unmarshal(mapResponseBytes, &triggerResponse)

	assert.Equal(t, resp.Code, http.StatusOK)
	assert.Equal(t, groups.ExecBroadcastTrigger, triggerResponse.Status)
}

func getHardforkRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"hardfork": {
				Routes: []config.RouteConfig{
					{Name: "/trigger", Open: true},
				},
			},
		},
	}
}
