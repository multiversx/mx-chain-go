package getValues_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/api/getValues"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/middleware"
	"github.com/ElrondNetwork/elrond-go-sandbox/api/mock"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/json"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type GeneralResponse struct {
	Data  string `json:"data"`
	Error string `json:"error"`
}

func init() {
	gin.SetMode(gin.TestMode)
}

func loadResponse(rsp io.Reader, destination interface{}) {
	jsonParser := json.NewDecoder(rsp)
	err := jsonParser.Decode(destination)
	if err != nil {
		logError(err)
	}
}

func logError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func startNodeServer(handler getValues.FacadeHandler) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	getValuesRoute := ws.Group("/get-values")
	if handler != nil {
		getValuesRoute.Use(middleware.WithElrondFacade(handler))
	}
	getValues.Routes(getValuesRoute)
	return ws
}

func startNodeServerWrongFacade() *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	ws.Use(func(c *gin.Context) {
		c.Set("elrondFacade", mock.WrongFacade{})
	})
	getValuesRoute := ws.Group("/get-values")
	getValues.Routes(getValuesRoute)
	return ws
}

//------- GetDataValueAsHexBytes

func TestGetDataValueAsHexBytes_WithWrongFacadeShouldErr(t *testing.T) {
	t.Parallel()

}

func TestGetDataValueAsHexBytes_WithParametersShouldReturnValueAsHex(t *testing.T) {
	t.Parallel()

	scAddress := "sc address"
	fName := "function"
	args := []string{"argument 1", "argument 2"}
	errUnexpected := errors.New("unexpected error")
	valueBuff, _ := hex.DecodeString("DEADBEEF")

	facade := mock.Facade{
		GetDataValueHandler: func(address string, funcName string, argsBuff ...[]byte) (bytes []byte, e error) {
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

	ws := startNodeServer(&facade)

	argsHex := make([]string, len(args))
	for i := 0; i < len(args); i++ {
		argsHex[i] = hex.EncodeToString([]byte(args[i]))
	}
	argsJson, _ := json.Marshal(argsHex)

	jsonStr := fmt.Sprintf(
		`{"scAddress":"%s",`+
			`"funcName":"%s",`+
			`"args":%s}`, scAddress, fName, argsJson)
	fmt.Printf("Request: %s\n", jsonStr)

	req, _ := http.NewRequest("POST", "/get-values/as-hex", bytes.NewBuffer([]byte(jsonStr)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := GeneralResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "", response.Error)
	assert.Equal(t, hex.EncodeToString(valueBuff), response.Data)
}
