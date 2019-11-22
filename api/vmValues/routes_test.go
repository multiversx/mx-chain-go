package vmValues

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/json"
	"github.com/stretchr/testify/assert"
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

func TestGetDataValueAsHexBytes(t *testing.T) {
	t.Parallel()

	valueBuff, _ := hex.DecodeString("DEADBEEF")

	facade := mock.Facade{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vmcommon.VMOutput, e error) {
			return &vmcommon.VMOutput{
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

	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "", response.Error)
	assert.Equal(t, hex.EncodeToString(valueBuff), response.Data)
}

func TestGetDataValueAsString(t *testing.T) {
	t.Parallel()

	valueBuff := "DEADBEEF"

	facade := mock.Facade{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vmcommon.VMOutput, e error) {
			return &vmcommon.VMOutput{
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

	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "", response.Error)
	assert.Equal(t, valueBuff, response.Data)
}

func TestGetDataValueAsInt(t *testing.T) {
	t.Parallel()

	value := "1234567"

	facade := mock.Facade{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vmcommon.VMOutput, e error) {
			returnData := big.NewInt(0)
			returnData.SetString(value, 10)
			return &vmcommon.VMOutput{
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

	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "", response.Error)
	assert.Equal(t, value, response.Data)
}

func TestExecuteQuery(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vmcommon.VMOutput, e error) {

			return &vmcommon.VMOutput{
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

	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "", response.Error)
	assert.Equal(t, int64(42), big.NewInt(0).SetBytes(response.Data.ReturnData[0]).Int64())
}

func TestCreateSCQuery_ArgumentIsNotHexShouldErr(t *testing.T) {
	request := VMValueRequest{
		ScAddress: DummyScAddress,
		FuncName:  "function",
		Args:      []string{"bad arg"},
	}

	_, err := createSCQuery(&request)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "'bad arg' is not a valid hex string")
}

func TestGetDataValue_FacadeErrorsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	facade := mock.Facade{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (vmOutput *vmcommon.VMOutput, e error) {
			return nil, errExpected
		},
	}

	request := VMValueRequest{
		ScAddress: DummyScAddress,
		FuncName:  "function",
		Args:      []string{},
	}

	response := simpleResponse{}

	statusCode := doPost(&facade, "/vm-values/hex", request, &response)
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Contains(t, response.Error, errExpected.Error())

	statusCode = doPost(&facade, "/vm-values/string", request, &response)
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Contains(t, response.Error, errExpected.Error())

	statusCode = doPost(&facade, "/vm-values/int", request, &response)
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Contains(t, response.Error, errExpected.Error())
}

func doPost(facadeMock interface{}, url string, request VMValueRequest, response interface{}) int {
	jsonBody, _ := json.Marshal(request)

	server := startNodeServer(facadeMock)
	httpRequest, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonBody)))

	responseRecorder := httptest.NewRecorder()
	server.ServeHTTP(responseRecorder, httpRequest)

	parseResponse(responseRecorder.Body, &response)
	return responseRecorder.Code
}

func startNodeServer(handler interface{}) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	getValuesRoute := ws.Group("/vm-values")
	getValuesRoute.Use(middleware.WithElrondFacade(handler))
	Routes(getValuesRoute)

	return ws
}

func parseResponse(responseBody io.Reader, destination interface{}) {
	jsonParser := json.NewDecoder(responseBody)

	err := jsonParser.Decode(destination)
	if err != nil {
		fmt.Println(err)
	}
}
