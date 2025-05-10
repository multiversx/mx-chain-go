package fetcher

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-go/state"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var log = logger.GetOrCreate("fetcher.gatewayFetcher")

const (
	dummyKey = "dummy"
)

type dataFetcher struct {
	mutex            sync.RWMutex
	baseURL          string
	addressConverter core.PubkeyConverter
}

// NewDataFetcher will create a new instance of data fetcher
func NewDataFetcher(serverURL string, converter core.PubkeyConverter) *dataFetcher {
	return &dataFetcher{
		baseURL:          serverURL,
		addressConverter: converter,
		mutex:            sync.RWMutex{},
	}
}

// GetAddressNonce will return the address nonce
func (df *dataFetcher) GetAddressNonce(address string) (uint64, error) {
	df.mutex.Lock()
	defer df.mutex.Unlock()

	accountFromGateway, err := df.getAccount(address)
	if err != nil {
		return 0, err
	}

	return accountFromGateway.Nonce, nil
}

// GetNetworkInfo will return the network info
func (df *dataFetcher) GetNetworkInfo() (*NetworkInfo, error) {
	df.mutex.Lock()
	defer df.mutex.Unlock()

	endpoint := fmt.Sprintf("%s/network/status/%d", df.baseURL, core.MetachainShardId)
	var response NetworkStatusResponse
	err := doGetRequest(endpoint, &response)
	if err != nil {
		return nil, err
	}

	if response.Error != "" {
		return nil, errors.New(response.Error)
	}

	return response.Data.Status, nil
}

// FetchAccount will fetch account from gateway and set data in AccountHandler
func (df *dataFetcher) FetchAccount(address []byte, newAccount vmcommon.AccountHandler) (state.UserAccountHandler, error) {
	df.mutex.Lock()
	defer df.mutex.Unlock()

	bechAddress, _ := df.addressConverter.Encode(address)

	accountFromGateway, err := df.getAccount(bechAddress)
	if err != nil {
		return nil, err
	}

	userAccount, ok := newAccount.(state.UserAccountHandler)
	if !ok {
		return nil, fmt.Errorf("invalid account type: %T", newAccount.(vmcommon.UserAccountHandler))
	}

	balanceBig, _ := big.NewInt(0).SetString(accountFromGateway.Balance, 10)
	err = userAccount.AddToBalance(balanceBig)
	if err != nil {
		return nil, err
	}

	userAccount.IncreaseNonce(accountFromGateway.Nonce)
	codeBytes, _ := hex.DecodeString(accountFromGateway.Code)

	userAccount.SetCode(codeBytes)
	userAccount.SetCodeHash(accountFromGateway.CodeHash)
	userAccount.SetCodeMetadata(accountFromGateway.CodeMetadata)

	err = userAccount.SaveKeyValue([]byte(dummyKey), []byte(dummyKey))
	if err != nil {
		return nil, err
	}

	return userAccount, nil
}

// FetchKeyFromGateway will fetch the value from a provided address and key from gateway
func (df *dataFetcher) FetchKeyFromGateway(address []byte, key []byte) ([]byte, error) {
	df.mutex.Lock()
	defer df.mutex.Unlock()

	bechAddress, _ := df.addressConverter.Encode(address)

	log.Warn("Fetch key", "key", hex.EncodeToString(key), "address", bechAddress)

	endpoint := fmt.Sprintf("%s/address/%s/key/%s", df.baseURL, bechAddress, hex.EncodeToString(key))
	var response FetchKeyResponse
	err := doGetRequest(endpoint, &response)
	if err != nil {
		return nil, err
	}

	if response.Error != "" {
		return nil, errors.New(response.Error)
	}

	decoded, err := hex.DecodeString(response.Data.Value)

	return decoded, err
}

func (df *dataFetcher) getAccount(address string) (*api.AccountResponse, error) {
	endpoint := fmt.Sprintf("%s/address/%s", df.baseURL, address)
	var response FetchAccountResponse
	err := doGetRequest(endpoint, &response)
	if err != nil {
		return nil, err
	}

	log.Warn("Fetch account gateway", "address", address)

	if response.Error != "" {
		return nil, errors.New(response.Error)
	}

	return &response.Data.Account, nil
}

func doGetRequest(endpoint string, response interface{}) error {
	resp, err := http.Get(endpoint)
	if err != nil {
		return fmt.Errorf("error making GET request to %s: %v", endpoint, err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-OK HTTP status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	return nil
}
