package factory

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
)

// InterceptedDataType represents a type of intercepted data instantiated by the Create func
type InterceptedDataType string

// InterceptedTx is the type for intercepted transaction
const InterceptedTx InterceptedDataType = "intercepted transaction"

type interceptedDataFactory struct {
	argument            *InterceptedDataFactoryArgument
	interceptedDataType InterceptedDataType
}

// NewInterceptedDataFactory creates an instance of interceptedDataFactory that can create
// instances of process.InterceptedData
func NewInterceptedDataFactory(
	argument *InterceptedDataFactoryArgument,
	dataType InterceptedDataType,
) (*interceptedDataFactory, error) {

	if argument == nil {
		return nil, process.ErrNilArguments
	}

	return &interceptedDataFactory{
		argument:            argument,
		interceptedDataType: dataType,
	}, nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
// The type of the output instance is provided in the constructor
func (idf *interceptedDataFactory) Create(buff []byte) (process.InterceptedData, error) {
	switch idf.interceptedDataType {
	case InterceptedTx:
		return idf.createInterceptedTx(buff)

	default:
		return nil, process.ErrInterceptedDataTypeNotDefined
	}
}

func (idf *interceptedDataFactory) createInterceptedTx(buff []byte) (process.InterceptedData, error) {
	return transaction.NewInterceptedTransaction(
		buff,
		idf.argument.Marshalizer,
		idf.argument.Hasher,
		idf.argument.KeyGen,
		idf.argument.Signer,
		idf.argument.AddrConv,
		idf.argument.ShardCoordinator,
	)
}
