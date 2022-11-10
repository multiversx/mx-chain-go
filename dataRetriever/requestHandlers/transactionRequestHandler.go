package requestHandlers

import (
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// ArgTransactionRequestHandler is the argument structure used to create a new a transaction request handler instance
type ArgTransactionRequestHandler struct {
	ArgBaseRequestHandler
}

type transactionRequestHandler struct {
	*baseRequestHandler
}

// NewTransactionRequestHandler returns a new instance of transaction request handler
func NewTransactionRequestHandler(args ArgTransactionRequestHandler) (*transactionRequestHandler, error) {
	err := checkArgBase(args.ArgBaseRequestHandler)
	if err != nil {
		return nil, err
	}

	return &transactionRequestHandler{
		baseRequestHandler: createBaseRequestHandler(args.ArgBaseRequestHandler),
	}, nil
}

// RequestDataFromHashArray requests a list of tx hashes from other peers
func (handler *transactionRequestHandler) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	handler.printHashArray(hashes)

	b := &batch.Batch{
		Data: hashes,
	}
	buffHashes, err := handler.marshaller.Marshal(b)
	if err != nil {
		return err
	}

	return handler.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashArrayType,
			Value: buffHashes,
			Epoch: epoch,
		},
		hashes,
	)
}

func (handler *transactionRequestHandler) printHashArray(hashes [][]byte) {
	if log.GetLevel() > logger.LogTrace {
		return
	}

	for _, hash := range hashes {
		log.Trace("transactionRequestHandler.RequestDataFromHashArray", "hash", hash, "topic", handler.RequestTopic())
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *transactionRequestHandler) IsInterfaceNil() bool {
	return handler == nil
}
