package groups_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	apiErrors "github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/api/groups"
	"github.com/multiversx/mx-chain-go/api/mock"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type triggerResponse struct {
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

func TestHardforkGroup_TriggerCannotExecuteShouldErr(t *testing.T) {
	t.Parallel()

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

func TestHardforkGroup_TriggerWrongRequestTypeShouldErr(t *testing.T) {
	t.Parallel()

	hardforkGroup, err := groups.NewHardforkGroup(&mock.HardforkFacade{})
	require.NoError(t, err)

	ws := startWebServer(hardforkGroup, "hardfork", getHardforkRoutesConfig())

	req, _ := http.NewRequest("POST", "/hardfork/trigger", bytes.NewBuffer([]byte("wrong buffer")))
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	triggerResp := triggerResponse{}
	loadResponse(resp.Body, &triggerResp)

	assert.Equal(t, resp.Code, http.StatusBadRequest)
}

func TestHardforkGroup_ManualShouldWork(t *testing.T) {
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

	triggerResp := triggerResponse{}
	mapResponseData := response.Data.(map[string]interface{})
	mapResponseBytes, _ := json.Marshal(&mapResponseData)
	_ = json.Unmarshal(mapResponseBytes, &triggerResp)

	assert.Equal(t, resp.Code, http.StatusOK)
	assert.Equal(t, groups.ExecManualTrigger, triggerResp.Status)
	assert.Equal(t, hr.Epoch, atomic.LoadUint32(&recoveredEpoch))
}

func TestHardforkGroup_BroadcastShouldWork(t *testing.T) {
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

	triggerResp := triggerResponse{}
	mapResponseData := response.Data.(map[string]interface{})
	mapResponseBytes, _ := json.Marshal(&mapResponseData)
	_ = json.Unmarshal(mapResponseBytes, &triggerResp)

	assert.Equal(t, resp.Code, http.StatusOK)
	assert.Equal(t, groups.ExecBroadcastTrigger, triggerResp.Status)
}

func TestHardforkGroup_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	hardforkGroup, _ := groups.NewHardforkGroup(nil)
	require.True(t, hardforkGroup.IsInterfaceNil())

	hardforkGroup, _ = groups.NewHardforkGroup(&mock.FacadeStub{})
	require.False(t, hardforkGroup.IsInterfaceNil())
}

func TestHardforkGroup_UpdateFacadeStub(t *testing.T) {
	t.Parallel()

	t.Run("nil facade should error", func(t *testing.T) {
		t.Parallel()

		hardforkGroup, err := groups.NewHardforkGroup(&mock.FacadeStub{})
		require.NoError(t, err)

		err = hardforkGroup.UpdateFacade(nil)
		require.Equal(t, apiErrors.ErrNilFacadeHandler, err)
	})
	t.Run("cast failure should error", func(t *testing.T) {
		t.Parallel()

		hardforkGroup, err := groups.NewHardforkGroup(&mock.FacadeStub{})
		require.NoError(t, err)

		err = hardforkGroup.UpdateFacade("this is not a facade handler")
		require.True(t, errors.Is(err, apiErrors.ErrFacadeWrongTypeAssertion))
	})
	t.Run("should work", func(t *testing.T) {
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

		triggerResp := triggerResponse{}
		mapResponseData := response.Data.(map[string]interface{})
		mapResponseBytes, _ := json.Marshal(&mapResponseData)
		_ = json.Unmarshal(mapResponseBytes, &triggerResp)

		assert.Equal(t, resp.Code, http.StatusOK)
		assert.Equal(t, groups.ExecBroadcastTrigger, triggerResp.Status)

		newFacade := &mock.HardforkFacade{
			TriggerCalled: func(_ uint32, _ bool) error {
				return expectedErr
			},
		}
		err = hardforkGroup.UpdateFacade(newFacade)
		require.NoError(t, err)

		req, _ = http.NewRequest("POST", "/hardfork/trigger", bytes.NewBuffer(buffHr))
		resp = httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response = shared.GenericAPIResponse{}
		loadResponse(resp.Body, &response)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.True(t, strings.Contains(response.Error, expectedErr.Error()))
	})
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
