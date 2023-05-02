package groups_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	apiErrors "github.com/multiversx/mx-chain-go/api/errors"
	"github.com/multiversx/mx-chain-go/api/groups"
	"github.com/multiversx/mx-chain-go/api/mock"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestNewVmValuesGroup(t *testing.T) {
	t.Parallel()

	t.Run("nil facade", func(t *testing.T) {
		hg, err := groups.NewVmValuesGroup(nil)
		require.True(t, errors.Is(err, apiErrors.ErrNilFacadeHandler))
		require.Nil(t, hg)
	})

	t.Run("should work", func(t *testing.T) {
		hg, err := groups.NewVmValuesGroup(&mock.FacadeStub{})
		require.NoError(t, err)
		require.NotNil(t, hg)
	})
}

type simpleResponse struct {
	Data  string `json:"data"`
	Error string `json:"error"`
}

type vmOutputResponse struct {
	Data  *vmcommon.VMOutput `json:"data"`
	Error string             `json:"error"`
}

func init() {
	gin.SetMode(gin.TestMode)
}

const dummyScAddress = "00000000000000000500fabd9501b7e5353de57a4e319857c2fb99089770720a"

func TestGetHex_ShouldWork(t *testing.T) {
	t.Parallel()

	valueBuff, _ := hex.DecodeString("DEADBEEF")

	facade := mock.FacadeStub{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vm.VMOutputApi, e error) {
			return &vm.VMOutputApi{
				ReturnData: [][]byte{valueBuff},
			}, nil
		},
	}

	request := groups.VMValueRequest{
		ScAddress: dummyScAddress,
		FuncName:  "function",
		Args:      []string{},
	}

	response := simpleResponse{}
	statusCode := doPost(t, &facade, "/vm-values/hex", request, &response)

	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, "", response.Error)
	require.Equal(t, hex.EncodeToString(valueBuff), response.Data)
}

func TestGetString_ShouldWork(t *testing.T) {
	t.Parallel()

	valueBuff := "DEADBEEF"

	facade := mock.FacadeStub{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vm.VMOutputApi, e error) {
			return &vm.VMOutputApi{
				ReturnData: [][]byte{[]byte(valueBuff)},
			}, nil
		},
	}

	request := groups.VMValueRequest{
		ScAddress: dummyScAddress,
		FuncName:  "function",
		Args:      []string{},
	}

	response := simpleResponse{}
	statusCode := doPost(t, &facade, "/vm-values/string", request, &response)

	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, "", response.Error)
	require.Equal(t, valueBuff, response.Data)
}

func TestGetInt_ShouldWork(t *testing.T) {
	t.Parallel()

	value := "1234567"

	facade := mock.FacadeStub{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vm.VMOutputApi, e error) {
			returnData := big.NewInt(0)
			returnData.SetString(value, 10)
			return &vm.VMOutputApi{
				ReturnData: [][]byte{returnData.Bytes()},
			}, nil
		},
	}

	request := groups.VMValueRequest{
		ScAddress: dummyScAddress,
		FuncName:  "function",
		Args:      []string{},
	}

	response := simpleResponse{}
	statusCode := doPost(t, &facade, "/vm-values/int", request, &response)

	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, "", response.Error)
	require.Equal(t, value, response.Data)
}

func TestQuery_ShouldWork(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vm.VMOutputApi, e error) {

			return &vm.VMOutputApi{
				ReturnData: [][]byte{big.NewInt(42).Bytes()},
			}, nil
		},
	}

	request := groups.VMValueRequest{
		ScAddress: dummyScAddress,
		FuncName:  "function",
		Args:      []string{},
	}

	response := vmOutputResponse{}
	statusCode := doPost(t, &facade, "/vm-values/query", request, &response)

	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, "", response.Error)
	require.Equal(t, int64(42), big.NewInt(0).SetBytes(response.Data.ReturnData[0]).Int64())
}

func TestCreateSCQuery_ArgumentIsNotHexShouldErr(t *testing.T) {
	request := groups.VMValueRequest{
		ScAddress: dummyScAddress,
		FuncName:  "function",
		Args:      []string{"bad arg"},
	}

	group, _ := groups.NewVmValuesGroup(&mock.FacadeStub{})
	_, err := group.CreateSCQuery(&request)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "'bad arg' is not a valid hex string")
}

func TestAllRoutes_FacadeErrorsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("some random error")
	facade := mock.FacadeStub{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vm.VMOutputApi, e error) {
			return nil, errExpected
		},
	}

	request := groups.VMValueRequest{
		ScAddress: dummyScAddress,
		FuncName:  "function",
		Args:      []string{},
	}

	requireErrorOnAllRoutes(t, &facade, request, errExpected)
}

func TestAllRoutes_WhenBadAddressShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("not a valid address")
	facade := mock.FacadeStub{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vm.VMOutputApi, e error) {
			return &vm.VMOutputApi{}, nil
		},
	}

	request := groups.VMValueRequest{
		ScAddress: "DUMMY",
		FuncName:  "function",
		Args:      []string{},
	}

	requireErrorOnAllRoutes(t, &facade, request, errExpected)
}

func TestAllRoutes_WhenBadArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("not a valid hex string")
	facade := mock.FacadeStub{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vm.VMOutputApi, e error) {
			return &vm.VMOutputApi{}, nil
		},
	}

	request := groups.VMValueRequest{
		ScAddress: dummyScAddress,
		FuncName:  "function",
		Args:      []string{"AA", "ZZ"},
	}

	requireErrorOnAllRoutes(t, &facade, request, errExpected)
}

func TestAllRoutes_WhenNoVMReturnDataShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("no return data")
	facade := &mock.FacadeStub{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vm.VMOutputApi, e error) {
			return &vm.VMOutputApi{}, nil
		},
	}

	request := groups.VMValueRequest{
		ScAddress: dummyScAddress,
		FuncName:  "function",
		Args:      []string{},
		CallValue: "1",
	}

	response := simpleResponse{}

	statusCode := doPost(t, facade, "/vm-values/hex", request, &response)
	require.Equal(t, http.StatusOK, statusCode)
	require.Contains(t, response.Error, errExpected.Error())

	statusCode = doPost(t, facade, "/vm-values/string", request, &response)
	require.Equal(t, http.StatusOK, statusCode)
	require.Contains(t, response.Error, errExpected.Error())

	statusCode = doPost(t, facade, "/vm-values/int", request, &response)
	require.Equal(t, http.StatusOK, statusCode)
	require.Contains(t, response.Error, errExpected.Error())
}

func TestAllRoutes_WhenBadJsonShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vm.VMOutputApi, e error) {
			return &vm.VMOutputApi{}, nil
		},
	}

	requireErrorOnGetSingleValueRoutes(t, &facade, []byte("dummy"), apiErrors.ErrInvalidJSONRequest)
}

func TestAllRoutes_DecodeAddressPubkeyFailsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	cnt := 0
	facade := mock.FacadeStub{
		DecodeAddressPubkeyCalled: func(pk string) ([]byte, error) {
			cnt++
			if cnt > 1 {
				return nil, expectedErr
			}
			return hex.DecodeString(pk)
		},
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vm.VMOutputApi, e error) {
			return &vm.VMOutputApi{}, nil
		},
	}

	request := groups.VMValueRequest{
		ScAddress:  dummyScAddress,
		FuncName:   "function",
		Args:       []string{},
		CallerAddr: dummyScAddress,
	}
	requireErrorOnGetSingleValueRoutes(t, &facade, request, expectedErr)
}

func TestAllRoutes_SetStringFailsShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vm.VMOutputApi, e error) {
			return &vm.VMOutputApi{}, nil
		},
	}

	request := groups.VMValueRequest{
		ScAddress:  dummyScAddress,
		FuncName:   "function",
		Args:       []string{},
		CallerAddr: dummyScAddress, // coverage
		CallValue:  "not an int",
	}
	requireErrorOnGetSingleValueRoutes(t, &facade, request, errors.New("non numeric call value"))
}

func TestVMValuesGroup_UpdateFacade(t *testing.T) {
	t.Parallel()

	t.Run("nil facade should error", func(t *testing.T) {
		t.Parallel()

		group, err := groups.NewVmValuesGroup(&mock.FacadeStub{})
		require.NoError(t, err)

		err = group.UpdateFacade(nil)
		require.Equal(t, apiErrors.ErrNilFacadeHandler, err)
	})
	t.Run("cast failure should error", func(t *testing.T) {
		t.Parallel()

		group, err := groups.NewVmValuesGroup(&mock.FacadeStub{})
		require.NoError(t, err)

		err = group.UpdateFacade("this is not a facade handler")
		require.True(t, errors.Is(err, apiErrors.ErrFacadeWrongTypeAssertion))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		valueBuff, _ := hex.DecodeString("DEADBEEF")
		facade := &mock.FacadeStub{
			ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vm.VMOutputApi, e error) {

				return &vm.VMOutputApi{
					ReturnData: [][]byte{valueBuff},
					ReturnCode: "NOK", // coverage
				}, nil
			},
		}

		request := groups.VMValueRequest{
			ScAddress: dummyScAddress,
			FuncName:  "function",
			Args:      []string{},
		}
		requestAsBytes, _ := json.Marshal(request)
		group, err := groups.NewVmValuesGroup(facade)
		require.NoError(t, err)

		server := startWebServer(group, "vm-values", getVmValuesRoutesConfig())

		httpRequest, _ := http.NewRequest("POST", "/vm-values/hex", bytes.NewBuffer(requestAsBytes))
		responseRecorder := httptest.NewRecorder()
		server.ServeHTTP(responseRecorder, httpRequest)

		responseI := shared.GenericAPIResponse{}
		loadResponse(responseRecorder.Body, &responseI)
		responseDataMap := responseI.Data.(map[string]interface{})
		responseDataMapBytes, _ := json.Marshal(responseDataMap)
		response := &simpleResponse{}
		_ = json.Unmarshal(responseDataMapBytes, response)
		require.Equal(t, http.StatusOK, responseRecorder.Code)
		require.Contains(t, responseI.Error, "NOK")
		require.Contains(t, "", response.Error)
		require.Equal(t, hex.EncodeToString(valueBuff), response.Data)

		expectedErr := errors.New("expected error")
		newFacade := &mock.FacadeStub{
			ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vm.VMOutputApi, e error) {

				return &vm.VMOutputApi{
					ReturnData: nil,
				}, expectedErr
			},
		}

		err = group.UpdateFacade(newFacade)
		require.NoError(t, err)

		httpRequest, _ = http.NewRequest("POST", "/vm-values/hex", bytes.NewBuffer(requestAsBytes))
		responseRecorder = httptest.NewRecorder()
		server.ServeHTTP(responseRecorder, httpRequest)
		loadResponse(responseRecorder.Body, &responseI)
		require.Equal(t, http.StatusBadRequest, responseRecorder.Code)
		require.Contains(t, responseI.Error, expectedErr.Error())
	})
}

func TestVMValuesGroup_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	group, _ := groups.NewVmValuesGroup(nil)
	require.True(t, group.IsInterfaceNil())

	group, _ = groups.NewVmValuesGroup(&mock.FacadeStub{})
	require.False(t, group.IsInterfaceNil())
}

func doPost(t *testing.T, facade interface{}, url string, request interface{}, response interface{}) int {
	// Serialize if not already
	requestAsBytes, ok := request.([]byte)
	if !ok {
		requestAsBytes, _ = json.Marshal(request)
	}

	vmValuesFacade, ok := facade.(groups.VmValuesFacadeHandler)
	require.True(t, ok)

	group, err := groups.NewVmValuesGroup(vmValuesFacade)
	require.NoError(t, err)

	server := startWebServer(group, "vm-values", getVmValuesRoutesConfig())

	httpRequest, _ := http.NewRequest("POST", url, bytes.NewBuffer(requestAsBytes))

	responseRecorder := httptest.NewRecorder()
	server.ServeHTTP(responseRecorder, httpRequest)

	responseI := shared.GenericAPIResponse{}
	loadResponse(responseRecorder.Body, &responseI)
	if responseI.Error == "" {
		responseDataMap := responseI.Data.(map[string]interface{})
		responseDataMapBytes, _ := json.Marshal(responseDataMap)
		_ = json.Unmarshal(responseDataMapBytes, response)
	} else {
		resp := response.(*simpleResponse)
		resp.Error = responseI.Error
	}

	return responseRecorder.Code
}

func requireErrorOnAllRoutes(t *testing.T, facade interface{}, request interface{}, errExpected error) {
	requireErrorOnGetSingleValueRoutes(t, facade, request, errExpected)

	response := simpleResponse{}
	statusCode := doPost(t, facade, "/vm-values/query", request, &response)
	require.Equal(t, http.StatusBadRequest, statusCode)
	require.Contains(t, response.Error, errExpected.Error())
}

func requireErrorOnGetSingleValueRoutes(t *testing.T, facade interface{}, request interface{}, errExpected error) {
	response := simpleResponse{}

	statusCode := doPost(t, facade, "/vm-values/hex", request, &response)
	require.Equal(t, http.StatusBadRequest, statusCode)
	require.Contains(t, response.Error, errExpected.Error())

	statusCode = doPost(t, facade, "/vm-values/string", request, &response)
	require.Equal(t, http.StatusBadRequest, statusCode)
	require.Contains(t, response.Error, errExpected.Error())

	statusCode = doPost(t, facade, "/vm-values/int", request, &response)
	require.Equal(t, http.StatusBadRequest, statusCode)
	require.Contains(t, response.Error, errExpected.Error())
}

func getVmValuesRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"vm-values": {
				Routes: []config.RouteConfig{
					{Name: "/hex", Open: true},
					{Name: "/string", Open: true},
					{Name: "/int", Open: true},
					{Name: "/query", Open: true},
				},
			},
		},
	}
}
