package systemSmartContracts

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"testing"

	vmData "github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckIfNil_NilArgs(t *testing.T) {
	t.Parallel()

	err := CheckIfNil(nil)

	assert.Equal(t, vm.ErrInputArgsIsNil, err)
}

func TestTokenProperties(t *testing.T) {
	tokenID := "ETHWBTC-74e282"
	tokenIDHex := hex.EncodeToString([]byte(tokenID))

	proxy := "https://gateway.multiversx.com"
	esdtAddress := "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u"
	funcName := "getTokenProperties"

	type vmQueryReq struct {
		ScAddress string   `json:"scAddress"`
		FuncName  string   `json:"funcName"`
		Args      []string `json:"args"`
	}
	type vmQueryResp struct {
		Data struct {
			Data struct {
				ReturnData []string `json:"returnData"`
			} `json:"data"`
		} `json:"data"`
	}

	req := vmQueryReq{
		ScAddress: esdtAddress,
		FuncName:  funcName,
		Args:      []string{tokenIDHex},
	}
	reqJson, _ := json.Marshal(req)
	httpRequest, err := http.Post(fmt.Sprintf("%s/vm-values/query", proxy), "application/json", bytes.NewReader(reqJson))
	require.NoError(t, err)

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(httpRequest.Body)

	var response vmQueryResp
	responseBytes, err := ioutil.ReadAll(httpRequest.Body)
	require.NoError(t, err)

	err = json.Unmarshal(responseBytes, &response)
	require.NoError(t, err)

	for _, attr := range response.Data.Data.ReturnData {
		decodedAttr, err := base64.StdEncoding.DecodeString(attr)
		require.NoError(t, err)

		fmt.Println(string(decodedAttr))
	}
}

func TestCheckIfNil_NilCallerAddr(t *testing.T) {
	t.Parallel()

	err := CheckIfNil(&vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  nil,
			Arguments:   nil,
			CallValue:   big.NewInt(0),
			GasPrice:    0,
			GasProvided: 0,
			CallType:    vmData.DirectCall,
		},
		RecipientAddr: []byte("dummyAddress"),
		Function:      "something",
	})

	assert.Equal(t, vm.ErrInputCallerAddrIsNil, err)
}

func TestCheckIfNil_NilCallValue(t *testing.T) {
	t.Parallel()

	err := CheckIfNil(&vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("dummyAddress"),
			Arguments:   nil,
			CallValue:   nil,
			GasPrice:    0,
			GasProvided: 0,
			CallType:    vmData.DirectCall,
		},
		RecipientAddr: []byte("dummyAddress"),
		Function:      "something",
	})

	assert.Equal(t, vm.ErrInputCallValueIsNil, err)
}

func TestCheckIfNil_NilRecipientAddr(t *testing.T) {
	t.Parallel()

	err := CheckIfNil(&vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("dummyAddress"),
			Arguments:   nil,
			CallValue:   big.NewInt(0),
			GasPrice:    0,
			GasProvided: 0,
			CallType:    vmData.DirectCall,
		},
		RecipientAddr: nil,
		Function:      "something",
	})

	assert.Equal(t, vm.ErrInputRecipientAddrIsNil, err)
}

func TestCheckIfNil_NilFunction(t *testing.T) {
	t.Parallel()

	err := CheckIfNil(&vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("dummyAddress"),
			Arguments:   nil,
			CallValue:   big.NewInt(0),
			GasPrice:    0,
			GasProvided: 0,
			CallType:    vmData.DirectCall,
		},
		RecipientAddr: []byte("dummyAddress"),
		Function:      "",
	})

	assert.Equal(t, vm.ErrInputFunctionIsNil, err)
}

func TestCheckIfNil(t *testing.T) {
	t.Parallel()

	err := CheckIfNil(&vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("dummyAddress"),
			Arguments:   nil,
			CallValue:   big.NewInt(0),
			GasPrice:    0,
			GasProvided: 0,
			CallType:    vmData.DirectCall,
		},
		RecipientAddr: []byte("dummyAddress"),
		Function:      "something",
	})

	assert.Nil(t, err)
}
