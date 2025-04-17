package fetcher

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-go/state"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"io"
	"math/big"
	"net/http"
)

var log = logger.GetOrCreate("fetcher.gatewayFetcher")

type dataFetcher struct {
	baseURL          string
	addressConverter core.PubkeyConverter
}

func NewDataFetcher(serverURL string, converter core.PubkeyConverter) *dataFetcher {
	return &dataFetcher{baseURL: serverURL,
		addressConverter: converter}
}

type FetchAccountResponse struct {
	Data struct {
		Account api.AccountResponse `json:"account"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}

func (df *dataFetcher) FetchAccount(address []byte, newAccount vmcommon.AccountHandler) (vmcommon.AccountHandler, error) {
	bechAddress, _ := df.addressConverter.Encode(address)

	endpoint := fmt.Sprintf("%s/address/%s", df.baseURL, bechAddress)

	log.Warn("Fetch account gateway", "address", bechAddress)

	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("error making GET request to %s: %v", endpoint, err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-OK HTTP status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// Define a variable to hold the unmarshalled data
	var response FetchAccountResponse

	// Unmarshal the JSON into the struct
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	userAccount, ok := newAccount.(state.UserAccountHandler)
	if !ok {
		return nil, fmt.Errorf("invalid account type: %T", newAccount.(vmcommon.UserAccountHandler))
	}

	balanceBig, _ := big.NewInt(0).SetString(response.Data.Account.Balance, 10)
	err = userAccount.AddToBalance(balanceBig)
	if err != nil {
		return nil, err
	}

	userAccount.IncreaseNonce(response.Data.Account.Nonce)
	codeBytes, _ := hex.DecodeString(response.Data.Account.Code)

	userAccount.SetCode(codeBytes)
	userAccount.SetCodeHash(response.Data.Account.CodeHash)
	userAccount.SetCodeMetadata(response.Data.Account.CodeMetadata)

	return newAccount, nil
}

// FetchKeyResponse represents the structure of the key response
type FetchKeyResponse struct {
	Data struct {
		Value string `json:"value"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}

func (df *dataFetcher) FetchKeyFromGateway(address []byte, key []byte) ([]byte, error) {
	bechAddress, _ := df.addressConverter.Encode(address)

	log.Warn("Fetch key", "key", hex.EncodeToString(key), "address", bechAddress)

	endpoint := fmt.Sprintf("%s/address/%s/key/%s", df.baseURL, bechAddress, hex.EncodeToString(key))

	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("error making GET request to %s: %v", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-OK HTTP status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// Define a variable to hold the unmarshalled data
	var response FetchKeyResponse

	// Unmarshal the JSON into the struct
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Printf("Error unmarshalling JSON: %v\n", err)
		return nil, err
	}

	decoded, err := hex.DecodeString(response.Data.Value)

	return decoded, err
}
