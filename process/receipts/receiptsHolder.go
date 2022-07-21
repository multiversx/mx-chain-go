package receipts

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
)

func newReceiptsHolder() *process.ReceiptsHolder {
	return &process.ReceiptsHolder{
		Miniblocks: make([]*block.MiniBlock, 0),
	}
}

func marshalReceiptsHolder(holder *process.ReceiptsHolder, marshaller marshal.Marshalizer) ([]byte, error) {
	receiptsBatch := &batch.Batch{}

	for _, miniBlock := range holder.Miniblocks {
		miniblockBytes, err := marshaller.Marshal(miniBlock)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", errCannotMarshalReceipts, err)
		}

		receiptsBatch.Data = append(receiptsBatch.Data, miniblockBytes)
	}

	// No miniblocks, no other (to be defined) content
	if len(receiptsBatch.Data) == 0 {
		return make([]byte, 0), nil
	}

	receiptsBytes, err := marshaller.Marshal(receiptsBatch)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errCannotMarshalReceipts, err)
	}

	return receiptsBytes, nil
}

func unmarshalReceiptsHolder(receiptsBytes []byte, marshaller marshal.Marshalizer) (*process.ReceiptsHolder, error) {
	holder := newReceiptsHolder()

	if len(receiptsBytes) == 0 {
		return holder, nil
	}

	receiptsBatch := &batch.Batch{}
	err := marshaller.Unmarshal(receiptsBatch, receiptsBytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errCannotUnmarshalReceipts, err)
	}

	for _, miniblockBytes := range receiptsBatch.Data {
		miniBlock := &block.MiniBlock{}
		err := marshaller.Unmarshal(miniBlock, miniblockBytes)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", errCannotUnmarshalReceipts, err)
		}

		holder.Miniblocks = append(holder.Miniblocks, miniBlock)
	}

	return holder, nil
}
