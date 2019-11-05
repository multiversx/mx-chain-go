package vmValues_test

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
	"github.com/ElrondNetwork/elrond-go/api/vmValues"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/json"
	"github.com/stretchr/testify/assert"
)

type GeneralResponse struct {
	Data  string `json:"data"`
	Error string `json:"error"`
}

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGetDataValueAsHexBytes_BadRequestShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{
		GetDataValueHandler: func(address string, funcName string, argsBuff ...[]byte) (vmOutput interface{}, e error) {
			assert.Fail(t, "should have not called this")
			return nil, nil
		},
	}

	jsonBody := `{"this should error"}`
	response, _ := postRequest(&facade, "/get-values/hex", jsonBody)
	assert.Contains(t, response.Error, "invalid character")
}

func TestGetDataValueAsHexBytes_ArgumentIsNotHexShouldErr(t *testing.T) {
	t.Parallel()

	scAddress := "sc address"
	fName := "function"
	args := []string{"not a hex argument"}
	errUnexpected := errors.New("unexpected error")
	valueBuff, _ := hex.DecodeString("DEADBEEF")

	facade := mock.Facade{
		GetDataValueHandler: func(address string, funcName string, argsBuff ...[]byte) (vmOutput interface{}, e error) {
			if address == scAddress && funcName == fName && len(argsBuff) == len(args) {
				paramsOk := true
				for idx, arg := range args {
					if arg != string(argsBuff[idx]) {
						paramsOk = false
					}
				}

				if paramsOk {
					return valueBuff, nil
				}
			}

			return nil, errUnexpected
		},
	}

	argsJson, _ := json.Marshal(args)

	jsonBody := fmt.Sprintf(`{"scAddress":"%s", "funcName":"%s", "args":%s}`, scAddress, fName, argsJson)
	response, statusCode := postRequest(&facade, "/get-values/hex", jsonBody)

	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Contains(t, response.Error, "not a hex argument")
}

func testGetValueFacadeErrors(t *testing.T, route string) {
	t.Parallel()

	errExpected := errors.New("expected error")
	facade := mock.Facade{
		GetDataValueHandler: func(address string, funcName string, argsBuff ...[]byte) (vmOutput interface{}, e error) {
			return nil, errExpected
		},
	}

	jsonBody := `{}`
	response, statusCode := postRequest(&facade, route, jsonBody)

	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Contains(t, response.Error, errExpected.Error())
}

func TestGetDataValueAsHexBytes_FacadeErrorsShouldErr(t *testing.T) {
	testGetValueFacadeErrors(t, "/get-values/hex")
}

func TestGetDataValueAsHexBytes_WithParametersShouldReturnValueAsHex(t *testing.T) {
	t.Parallel()

	scAddress := "aaaa"
	fName := "function"
	args := []string{"argument 1", "argument 2"}
	errUnexpected := errors.New("unexpected error")
	valueBuff, _ := hex.DecodeString("DEADBEEF")

	facade := mock.Facade{
		GetDataValueHandler: func(address string, funcName string, argsBuff ...[]byte) (vmOutput interface{}, e error) {
			areArgumentsCorrect := hex.EncodeToString([]byte(address)) == scAddress &&
				funcName == fName &&
				len(argsBuff) == len(args)

			if areArgumentsCorrect {
				paramsOk := true
				for idx, arg := range args {
					if arg != string(argsBuff[idx]) {
						paramsOk = false
					}
				}

				if paramsOk {
					returnData := big.NewInt(0)
					returnData.SetBytes([]byte(valueBuff))
					return &vmcommon.VMOutput{
						ReturnData: []*big.Int{returnData},
					}, nil
				}
			}

			return nil, errUnexpected
		},
	}

	argsHex := make([]string, len(args))
	for i := 0; i < len(args); i++ {
		argsHex[i] = hex.EncodeToString([]byte(args[i]))
	}
	argsJson, _ := json.Marshal(argsHex)

	jsonBody := fmt.Sprintf(`{"scAddress":"%s", "funcName":"%s", "args":%s}`, scAddress, fName, argsJson)
	response, statusCode := postRequest(&facade, "/get-values/hex", jsonBody)

	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "", response.Error)
	assert.Equal(t, hex.EncodeToString(valueBuff), response.Data)
}

func TestGetDataValueAsString_FacadeErrorsShouldErr(t *testing.T) {
	testGetValueFacadeErrors(t, "/get-values/string")
}

func TestGetDataValueAsString_WithParametersShouldReturnValueAsHex(t *testing.T) {
	t.Parallel()

	scAddress := "aaaa"
	fName := "function"
	args := []string{"argument 1", "argument 2"}
	errUnexpected := errors.New("unexpected error")
	valueBuff := "DEADBEEF"

	facade := mock.Facade{
		GetDataValueHandler: func(address string, funcName string, argsBuff ...[]byte) (vmOutput interface{}, e error) {
			areArgumentsCorrect := hex.EncodeToString([]byte(address)) == scAddress &&
				funcName == fName &&
				len(argsBuff) == len(args)

			if areArgumentsCorrect {
				paramsOk := true
				for idx, arg := range args {
					if arg != string(argsBuff[idx]) {
						paramsOk = false
					}
				}

				if paramsOk {
					returnData := big.NewInt(0)
					returnData.SetBytes([]byte(valueBuff))
					return &vmcommon.VMOutput{
						ReturnData: []*big.Int{returnData},
					}, nil
				}
			}

			return nil, errUnexpected
		},
	}

	argsHex := make([]string, len(args))
	for i := 0; i < len(args); i++ {
		argsHex[i] = hex.EncodeToString([]byte(args[i]))
	}
	argsJson, _ := json.Marshal(argsHex)

	jsonBody := fmt.Sprintf(`{"scAddress":"%s", "funcName":"%s", "args":%s}`, scAddress, fName, argsJson)
	response, statusCode := postRequest(&facade, "/get-values/string", jsonBody)

	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "", response.Error)
	assert.Equal(t, valueBuff, response.Data)
}

func TestGetDataValueAsInt_FacadeErrorsShouldErr(t *testing.T) {
	testGetValueFacadeErrors(t, "/get-values/int")
}

func TestGetDataValueAsInt_WithParametersShouldReturnValueAsHex(t *testing.T) {
	t.Parallel()

	scAddress := "aaaa"
	fName := "function"
	args := []string{"argument 1", "argument 2"}
	errUnexpected := errors.New("unexpected error")
	valueBuff := "1234567"

	facade := mock.Facade{
		GetDataValueHandler: func(address string, funcName string, argsBuff ...[]byte) (vmOutput interface{}, e error) {
			areArgumentsCorrect := hex.EncodeToString([]byte(address)) == scAddress &&
				funcName == fName &&
				len(argsBuff) == len(args)

			if areArgumentsCorrect {
				paramsOk := true
				for idx, arg := range args {
					if arg != string(argsBuff[idx]) {
						paramsOk = false
					}
				}

				if paramsOk {
					returnData := big.NewInt(0)
					returnData.SetString(valueBuff, 10)
					return &vmcommon.VMOutput{
						ReturnData: []*big.Int{returnData},
					}, nil
				}
			}

			return nil, errUnexpected
		},
	}

	argsHex := make([]string, len(args))
	for i := 0; i < len(args); i++ {
		argsHex[i] = hex.EncodeToString([]byte(args[i]))
	}
	argsJson, _ := json.Marshal(argsHex)

	jsonBody := fmt.Sprintf(`{"scAddress":"%s", "funcName":"%s", "args":%s}`, scAddress, fName, argsJson)
	response, statusCode := postRequest(&facade, "/get-values/int", jsonBody)

	assert.Equal(t, http.StatusOK, statusCode)
	assert.Equal(t, "", response.Error)
	assert.Equal(t, valueBuff, response.Data)
}

func postRequest(facadeMock interface{}, url string, jsonBody string) (GeneralResponse, int) {
	server := startNodeServer(facadeMock)
	request, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonBody)))

	responseRecorder := httptest.NewRecorder()
	server.ServeHTTP(responseRecorder, request)

	response := GeneralResponse{}
	loadResponse(responseRecorder.Body, &response)

	return response, responseRecorder.Code
}

func startNodeServer(handler interface{}) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	getValuesRoute := ws.Group("/get-values")
	getValuesRoute.Use(middleware.WithElrondFacade(handler))
	vmValues.Routes(getValuesRoute)

	return ws
}

func loadResponse(rsp io.Reader, destination interface{}) {
	jsonParser := json.NewDecoder(rsp)
	err := jsonParser.Decode(destination)
	if err != nil {
		fmt.Println(err)
	}
}
