package groups_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/validator"
	apiErrors "github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/api/groups"
	"github.com/multiversx/mx-chain-go/api/mock"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidatorGroup(t *testing.T) {
	t.Parallel()

	t.Run("nil facade", func(t *testing.T) {
		hg, err := groups.NewValidatorGroup(nil)
		require.True(t, errors.Is(err, apiErrors.ErrNilFacadeHandler))
		require.Nil(t, hg)
	})

	t.Run("should work", func(t *testing.T) {
		hg, err := groups.NewValidatorGroup(&mock.FacadeStub{})
		require.NoError(t, err)
		require.NotNil(t, hg)
	})
}

// ValidatorStatisticsResponse is the response for the validator statistics endpoint.
type ValidatorStatisticsResponse struct {
	Result map[string]*validator.ValidatorStatistics `json:"statistics"`
	Error  string                                    `json:"error"`
}

func TestValidatorStatistics_ErrorWhenFacadeFails(t *testing.T) {
	t.Parallel()

	errStr := "error in facade"

	facade := mock.FacadeStub{
		ValidatorStatisticsHandler: func() (map[string]*validator.ValidatorStatistics, error) {
			return nil, errors.New(errStr)
		},
	}

	validatorGroup, err := groups.NewValidatorGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(validatorGroup, "validator", getValidatorRoutesConfig())

	req, _ := http.NewRequest("GET", "/validator/statistics", nil)

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := ValidatorStatisticsResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, response.Error, errStr)
}

func TestValidatorStatistics_ReturnsSuccessfully(t *testing.T) {
	t.Parallel()

	mapToReturn := make(map[string]*validator.ValidatorStatistics)
	mapToReturn["test"] = &validator.ValidatorStatistics{
		NumLeaderSuccess:    5,
		NumLeaderFailure:    2,
		NumValidatorSuccess: 7,
		NumValidatorFailure: 3,
	}

	facade := mock.FacadeStub{
		ValidatorStatisticsHandler: func() (map[string]*validator.ValidatorStatistics, error) {
			return mapToReturn, nil
		},
	}

	validatorGroup, err := groups.NewValidatorGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(validatorGroup, "validator", getValidatorRoutesConfig())

	req, _ := http.NewRequest("GET", "/validator/statistics", nil)

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	validatorStatistics := ValidatorStatisticsResponse{}
	mapResponseData := response.Data.(map[string]interface{})
	mapResponseDataBytes, _ := json.Marshal(mapResponseData)
	_ = json.Unmarshal(mapResponseDataBytes, &validatorStatistics)

	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, validatorStatistics.Result, mapToReturn)
}

func TestValidatorGroup_UpdateFacade(t *testing.T) {
	t.Parallel()

	t.Run("nil facade should error", func(t *testing.T) {
		t.Parallel()

		validatorGroup, err := groups.NewValidatorGroup(&mock.FacadeStub{})
		require.NoError(t, err)

		err = validatorGroup.UpdateFacade(nil)
		require.Equal(t, apiErrors.ErrNilFacadeHandler, err)
	})
	t.Run("cast failure should error", func(t *testing.T) {
		t.Parallel()

		validatorGroup, err := groups.NewValidatorGroup(&mock.FacadeStub{})
		require.NoError(t, err)

		err = validatorGroup.UpdateFacade("this is not a facade handler")
		require.True(t, errors.Is(err, apiErrors.ErrFacadeWrongTypeAssertion))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		mapToReturn := make(map[string]*validator.ValidatorStatistics)
		mapToReturn["test"] = &validator.ValidatorStatistics{
			NumLeaderSuccess:    5,
			NumLeaderFailure:    2,
			NumValidatorSuccess: 7,
			NumValidatorFailure: 3,
		}
		facade := mock.FacadeStub{
			ValidatorStatisticsHandler: func() (map[string]*validator.ValidatorStatistics, error) {
				return mapToReturn, nil
			},
		}
		validatorGroup, err := groups.NewValidatorGroup(&facade)
		require.NoError(t, err)

		ws := startWebServer(validatorGroup, "validator", getValidatorRoutesConfig())

		req, _ := http.NewRequest("GET", "/validator/statistics", nil)
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		response := shared.GenericAPIResponse{}
		loadResponse(resp.Body, &response)
		validatorStatistics := ValidatorStatisticsResponse{}
		mapResponseData := response.Data.(map[string]interface{})
		mapResponseDataBytes, _ := json.Marshal(mapResponseData)
		_ = json.Unmarshal(mapResponseDataBytes, &validatorStatistics)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, validatorStatistics.Result, mapToReturn)

		expectedErr := errors.New("expected error")
		newFacade := mock.FacadeStub{
			ValidatorStatisticsHandler: func() (map[string]*validator.ValidatorStatistics, error) {
				return nil, expectedErr
			},
		}

		err = validatorGroup.UpdateFacade(&newFacade)
		require.NoError(t, err)

		req, _ = http.NewRequest("GET", "/validator/statistics", nil)
		resp = httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		loadResponse(resp.Body, &response)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, response.Error, expectedErr.Error())
	})
}

func TestValidatorGroup_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	validatorGroup, _ := groups.NewValidatorGroup(nil)
	require.True(t, validatorGroup.IsInterfaceNil())

	validatorGroup, _ = groups.NewValidatorGroup(&mock.FacadeStub{})
	require.False(t, validatorGroup.IsInterfaceNil())
}

func getValidatorRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"validator": {
				Routes: []config.RouteConfig{
					{Name: "/statistics", Open: true},
				},
			},
		},
	}
}
