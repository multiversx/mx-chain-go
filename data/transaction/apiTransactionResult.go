package transaction

// ApiTransactionResult is the DTO which will be returned on the get transaction by hash endpoint
type ApiTransactionResult struct {
	Type      string `json:"type"`
	Nonce     uint64 `json:"nonce"`
	Value     string `json:"value"`
	Receiver  string `json:"receiver"`
	Sender    string `json:"sender"`
	GasPrice  uint64 `json:"gasPrice"`
	GasLimit  uint64 `json:"gasLimit"`
	Data      string `json:"data"`
	Signature string `json:"signature"`
}
