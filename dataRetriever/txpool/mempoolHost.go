package txpool

import (
	"bytes"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
)

type argsMempoolHost struct {
	txGasHandler txGasHandler
	marshalizer  marshal.Marshalizer
}

type mempoolHost struct {
	txGasHandler        txGasHandler
	callArgumentsParser process.CallArgumentsParser
	esdtTransferParser  vmcommon.ESDTTransferParser
}

func newMempoolHost(args argsMempoolHost) (*mempoolHost, error) {
	if check.IfNil(args.txGasHandler) {
		return nil, dataRetriever.ErrNilTxGasHandler
	}
	if check.IfNil(args.marshalizer) {
		return nil, dataRetriever.ErrNilMarshalizer
	}

	argsParser := parsers.NewCallArgsParser()

	esdtTransferParser, err := parsers.NewESDTTransferParser(args.marshalizer)
	if err != nil {
		return nil, err
	}

	return &mempoolHost{
		txGasHandler:        args.txGasHandler,
		callArgumentsParser: argsParser,
		esdtTransferParser:  esdtTransferParser,
	}, nil
}

// ComputeTxFee computes the fee for a transaction.
func (host *mempoolHost) ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int {
	return host.txGasHandler.ComputeTxFee(tx)
}

// GetTransferredValue returns the value transferred by a transaction.
func (host *mempoolHost) GetTransferredValue(tx data.TransactionHandler) *big.Int {
	value := tx.GetValue()
	hasValue := value != nil && value.Sign() != 0
	if hasValue {
		// Early exit (optimization): a transaction can either bear a regular value or be a "MultiESDTNFTTransfer".
		return value
	}

	data := tx.GetData()
	hasData := len(data) > 0
	if !hasData {
		// Early exit (optimization): no "MultiESDTNFTTransfer" to parse.
		return tx.GetValue()
	}

	maybeMultiTransfer := bytes.HasPrefix(data, []byte(core.BuiltInFunctionMultiESDTNFTTransfer))
	if !maybeMultiTransfer {
		// Early exit (optimization).
		return big.NewInt(0)
	}

	function, args, err := host.callArgumentsParser.ParseData(string(data))
	if err != nil {
		return big.NewInt(0)
	}

	if function != core.BuiltInFunctionMultiESDTNFTTransfer {
		// Early exit (optimization).
		return big.NewInt(0)
	}

	esdtTransfers, err := host.esdtTransferParser.ParseESDTTransfers(tx.GetSndAddr(), tx.GetRcvAddr(), function, args)
	if err != nil {
		return big.NewInt(0)
	}

	accumulatedNativeValue := big.NewInt(0)

	for _, transfer := range esdtTransfers.ESDTTransfers {
		if transfer.ESDTTokenNonce != 0 {
			continue
		}
		if string(transfer.ESDTTokenName) != vmcommon.EGLDIdentifier {
			// We only care about native transfers.
			continue
		}

		_ = accumulatedNativeValue.Add(accumulatedNativeValue, transfer.ESDTValue)
	}

	return accumulatedNativeValue
}

// IsInterfaceNil returns true if there is no value under the interface
func (host *mempoolHost) IsInterfaceNil() bool {
	return host == nil
}
