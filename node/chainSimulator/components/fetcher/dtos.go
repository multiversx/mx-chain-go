package fetcher

import "github.com/multiversx/mx-chain-core-go/data/api"

// FetchKeyResponse represents the structure of the key response
type FetchKeyResponse struct {
	Data struct {
		Value string `json:"value"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}
type FetchAccountResponse struct {
	Data struct {
		Account api.AccountResponse `json:"account"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}

type NetworkStatusResponse struct {
	Data struct {
		Status *NetworkInfo `json:"status"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}

type NetworkInfo struct {
	CurrentRound int64  `json:"erd_current_round"`
	CurrentNonce uint64 `json:"erd_nonce"`
	CurrentEpoch uint32 `json:"erd_epoch_number"`
}
