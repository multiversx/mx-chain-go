package integrationtests

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"unicode/utf8"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const asciiSpace = byte(' ')
const asciiTab = byte('\t')
const asciiLineFeed = byte('\r')
const asciiNewLine = byte('\n')

var log = logger.GetOrCreate("stringers")

// TransactionHandlerToString will convert the transaction data slice provided to string
func TransactionHandlerToString(pubKeyConverter core.PubkeyConverter, txHandlers ...data.TransactionHandler) string {
	builder := &strings.Builder{}
	builder.WriteString("[\n")

	for _, txHandler := range txHandlers {
		switch tx := txHandler.(type) {
		case *smartContractResult.SmartContractResult:
			putSmartContractResultInBuilder(builder, tx, pubKeyConverter, "\t")
		default:
			// TODO implement the rest of the transaction handlers
			builder.WriteString(fmt.Sprintf("not implemented type %T\n", txHandler))
		}
	}

	builder.WriteString("]")
	return builder.String()
}

// SmartContractResultsToString will convert smartcontract results to an easy-to-understand string
func SmartContractResultsToString(pubKeyConverter core.PubkeyConverter, scrs ...*smartContractResult.SmartContractResult) string {
	builder := &strings.Builder{}
	builder.WriteString("[\n")
	for _, scr := range scrs {
		putSmartContractResultInBuilder(builder, scr, pubKeyConverter, "\t")
	}

	builder.WriteString("]")
	return builder.String()
}

func putSmartContractResultInBuilder(builder *strings.Builder, scr *smartContractResult.SmartContractResult, pubKeyConverter core.PubkeyConverter, indent string) {
	if scr == nil {
		builder.WriteString(fmt.Sprintf("%sSmartContractResult !NIL{}\n", indent))
		return
	}

	builder.WriteString(fmt.Sprintf("%sSmartContractResult %p{\n", indent, scr))
	defer builder.WriteString(fmt.Sprintf("%s}\n", indent))

	builder.WriteString(fmt.Sprintf("%s%sNonce: %d\n", indent, indent, scr.Nonce))
	putBigIntInBuilder(builder, "Value", indent, scr.Value)
	putAddressInBuilder(builder, "RcvAddr", indent, pubKeyConverter, scr.RcvAddr)
	putAddressInBuilder(builder, "SndAddr", indent, pubKeyConverter, scr.SndAddr)
	putAddressInBuilder(builder, "OriginalSender", indent, pubKeyConverter, scr.OriginalSender)
	builder.WriteString(fmt.Sprintf("%s%sReturnMessage: %s\n", indent, indent, convertStringIfNotASCII(scr.ReturnMessage)))
	builder.WriteString(fmt.Sprintf("%s%sData: %s\n", indent, indent, convertStringIfNotASCII(scr.Data)))
	builder.WriteString(fmt.Sprintf("%s%sCallType: %s\n", indent, indent, scr.CallType.ToString()))
	builder.WriteString(fmt.Sprintf("%s%sCode: %s\n", indent, indent, convertStringIfNotASCII(scr.Code)))
	builder.WriteString(fmt.Sprintf("%s%sGasLimit: %d\n", indent, indent, scr.GasLimit))
	builder.WriteString(fmt.Sprintf("%s%sGasPrice: %d\n", indent, indent, scr.GasPrice))
	builder.WriteString(fmt.Sprintf("%s%sOriginalTxHash: %s\n", indent, indent, convertStringIfNotASCII(scr.OriginalTxHash)))
	builder.WriteString(fmt.Sprintf("%s%sPrevTxHash: %s\n", indent, indent, convertStringIfNotASCII(scr.PrevTxHash)))
	putAddressInBuilder(builder, "RelayerAddr", indent, pubKeyConverter, scr.RelayerAddr)
	putBigIntInBuilder(builder, "RelayedValue", indent, scr.RelayedValue)
}

func putBigIntInBuilder(builder *strings.Builder, name string, indent string, value *big.Int) {
	strVal := "0"
	if value != nil {
		strVal = value.String()
	}

	builder.WriteString(fmt.Sprintf("%s%s%s: %s\n", indent, indent, name, strVal))
}

func putAddressInBuilder(builder *strings.Builder, name string, indent string, pubKeyConverter core.PubkeyConverter, slice []byte) {
	address := ""
	if len(slice) > 0 {
		if len(slice) != pubKeyConverter.Len() {
			// can not encode with the provided address
			address = hex.EncodeToString(slice) + " (!)"
		} else {
			address = pubKeyConverter.SilentEncode(slice, log)
		}
	}

	builder.WriteString(fmt.Sprintf("%s%s%s: %s\n", indent, indent, name, address))
}

func convertStringIfNotASCII(data []byte) string {
	if isASCII(string(data)) {
		return string(data)
	}

	return hex.EncodeToString(data)
}

func isASCII(data string) bool {
	for i := 0; i < len(data); i++ {
		if data[i] >= utf8.RuneSelf {
			return false
		}

		if data[i] >= asciiSpace {
			continue
		}

		if data[i] == asciiTab || data[i] == asciiLineFeed || data[i] == asciiNewLine {
			continue
		}

		return false
	}

	return true
}
