package vmValues

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	apiErrors "github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/mock"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/api/wrapper"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data/vm"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

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

const DummyScAddress = "00000000000000000500fabd9501b7e5353de57a4e319857c2fb99089770720a"

func TestGetHex_ShouldWork(t *testing.T) {
	t.Parallel()

	valueBuff, _ := hex.DecodeString("DEADBEEF")

	facade := mock.Facade{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vm.VMOutputApi, e error) {
			return &vm.VMOutputApi{
				ReturnData: [][]byte{valueBuff},
			}, nil
		},
	}

	request := VMValueRequest{
		ScAddress: DummyScAddress,
		FuncName:  "function",
		Args:      []string{},
	}

	response := simpleResponse{}
	statusCode := doPost(&facade, "/vm-values/hex", request, &response)

	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, "", response.Error)
	require.Equal(t, hex.EncodeToString(valueBuff), response.Data)
}

func TestGetString_ShouldWork(t *testing.T) {
	t.Parallel()

	valueBuff := "DEADBEEF"

	facade := mock.Facade{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vm.VMOutputApi, e error) {
			return &vm.VMOutputApi{
				ReturnData: [][]byte{[]byte(valueBuff)},
			}, nil
		},
	}

	request := VMValueRequest{
		ScAddress: DummyScAddress,
		FuncName:  "function",
		Args:      []string{},
	}

	response := simpleResponse{}
	statusCode := doPost(&facade, "/vm-values/string", request, &response)

	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, "", response.Error)
	require.Equal(t, valueBuff, response.Data)
}

func TestGetInt_ShouldWork(t *testing.T) {
	t.Parallel()

	value := "1234567"

	facade := mock.Facade{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vm.VMOutputApi, e error) {
			returnData := big.NewInt(0)
			returnData.SetString(value, 10)
			return &vm.VMOutputApi{
				ReturnData: [][]byte{returnData.Bytes()},
			}, nil
		},
	}

	request := VMValueRequest{
		ScAddress: DummyScAddress,
		FuncName:  "function",
		Args:      []string{},
	}

	response := simpleResponse{}
	statusCode := doPost(&facade, "/vm-values/int", request, &response)

	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, "", response.Error)
	require.Equal(t, value, response.Data)
}

func TestQuery_ShouldWork(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vm.VMOutputApi, e error) {

			return &vm.VMOutputApi{
				ReturnData: [][]byte{big.NewInt(42).Bytes()},
			}, nil
		},
	}

	request := VMValueRequest{
		ScAddress: DummyScAddress,
		FuncName:  "function",
		Args:      []string{},
	}

	response := vmOutputResponse{}
	statusCode := doPost(&facade, "/vm-values/query", request, &response)

	require.Equal(t, http.StatusOK, statusCode)
	require.Equal(t, "", response.Error)
	require.Equal(t, int64(42), big.NewInt(0).SetBytes(response.Data.ReturnData[0]).Int64())
}

func TestCreateSCQuery_ArgumentIsNotHexShouldErr(t *testing.T) {
	request := VMValueRequest{
		ScAddress: DummyScAddress,
		FuncName:  "function",
		Args:      []string{"bad arg"},
	}

	_, err := createSCQuery(&mock.Facade{}, &request)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "'bad arg' is not a valid hex string")
}

func TestAllRoutes_FacadeErrorsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("some random error")
	facade := mock.Facade{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vm.VMOutputApi, e error) {
			return nil, errExpected
		},
	}

	request := VMValueRequest{
		ScAddress: DummyScAddress,
		FuncName:  "function",
		Args:      []string{},
	}

	requireErrorOnAllRoutes(t, &facade, request, errExpected)
}

func TestAllRoutes_WhenBadAddressShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("not a valid address")
	facade := mock.Facade{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vm.VMOutputApi, e error) {
			return &vm.VMOutputApi{}, nil
		},
	}

	request := VMValueRequest{
		ScAddress: "DUMMY",
		FuncName:  "function",
		Args:      []string{},
	}

	requireErrorOnAllRoutes(t, &facade, request, errExpected)
}

func TestAllRoutes_WhenBadArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("not a valid hex string")
	facade := mock.Facade{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vm.VMOutputApi, e error) {
			return &vm.VMOutputApi{}, nil
		},
	}

	request := VMValueRequest{
		ScAddress: DummyScAddress,
		FuncName:  "function",
		Args:      []string{"AA", "ZZ"},
	}

	requireErrorOnAllRoutes(t, &facade, request, errExpected)
}

func TestAllRoutes_WhenNoVMReturnDataShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("no return data")
	facade := mock.Facade{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vm.VMOutputApi, e error) {
			return &vm.VMOutputApi{}, nil
		},
	}

	request := VMValueRequest{
		ScAddress: DummyScAddress,
		FuncName:  "function",
		Args:      []string{},
	}

	requireErrorOnGetSingleValueRoutes(t, &facade, request, errExpected)
}

func TestAllRoutes_WhenBadJsonShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vm.VMOutputApi, e error) {
			return &vm.VMOutputApi{}, nil
		},
	}

	requireErrorOnGetSingleValueRoutes(t, &facade, []byte("dummy"), apiErrors.ErrInvalidJSONRequest)
}

func TestAllRoutes_WhenNilFacadeShouldErr(t *testing.T) {
	t.Parallel()

	request := VMValueRequest{
		ScAddress: DummyScAddress,
		FuncName:  "function",
		Args:      []string{},
	}

	requireErrorOnAllRoutes(t, nil, request, apiErrors.ErrNilAppContext)
}

func TestAllRoutes_WhenBadFacadeShouldErr(t *testing.T) {
	t.Parallel()

	var facade interface{}

	request := VMValueRequest{
		ScAddress: DummyScAddress,
		FuncName:  "function",
		Args:      []string{},
	}

	requireErrorOnAllRoutes(t, &facade, request, apiErrors.ErrInvalidAppContext)
}

func doPost(facade interface{}, url string, request interface{}, response interface{}) int {
	// Serialize if not already
	requestAsBytes, ok := request.([]byte)
	if !ok {
		requestAsBytes, _ = json.Marshal(request)
	}

	server := startNodeServer(facade)
	httpRequest, _ := http.NewRequest("POST", url, bytes.NewBuffer(requestAsBytes))

	responseRecorder := httptest.NewRecorder()
	server.ServeHTTP(responseRecorder, httpRequest)

	responseI := shared.GenericAPIResponse{}
	parseResponse(responseRecorder.Body, &responseI)
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

func startNodeServer(handler interface{}) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	getValuesRoute := ws.Group("/vm-values")
	if handler != nil {
		getValuesRoute.Use(func(c *gin.Context) {
			c.Set("facade", handler)
			c.Next()
		})
	}
	vmValuesRoute, _ := wrapper.NewRouterWrapper("vm-values", getValuesRoute, getRoutesConfig())
	Routes(vmValuesRoute)

	return ws
}

func parseResponse(responseBody io.Reader, destination interface{}) {
	jsonParser := json.NewDecoder(responseBody)

	err := jsonParser.Decode(destination)
	if err != nil {
		fmt.Println(err)
	}
}

func requireErrorOnAllRoutes(t *testing.T, facade interface{}, request interface{}, errExpected error) {
	requireErrorOnGetSingleValueRoutes(t, facade, request, errExpected)

	response := simpleResponse{}
	statusCode := doPost(facade, "/vm-values/query", request, &response)
	require.Equal(t, http.StatusBadRequest, statusCode)
	require.Contains(t, response.Error, errExpected.Error())
}

func requireErrorOnGetSingleValueRoutes(t *testing.T, facade interface{}, request interface{}, errExpected error) {
	response := simpleResponse{}

	statusCode := doPost(facade, "/vm-values/hex", request, &response)
	require.Equal(t, http.StatusBadRequest, statusCode)
	require.Contains(t, response.Error, errExpected.Error())

	statusCode = doPost(facade, "/vm-values/string", request, &response)
	require.Equal(t, http.StatusBadRequest, statusCode)
	require.Contains(t, response.Error, errExpected.Error())

	statusCode = doPost(facade, "/vm-values/int", request, &response)
	require.Equal(t, http.StatusBadRequest, statusCode)
	require.Contains(t, response.Error, errExpected.Error())
}

func getRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"vm-values": {
				[]config.RouteConfig{
					{Name: "/hex", Open: true},
					{Name: "/string", Open: true},
					{Name: "/int", Open: true},
					{Name: "/query", Open: true},
				},
			},
		},
	}
}
