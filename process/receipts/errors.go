package receipts

import "errors"

var errCannotCreateReceiptsRepository = errors.New("cannot create receipts repository")
var errCannotMarshalReceipts = errors.New("cannot marshal receipts")
var errCannotUnmarshalReceipts = errors.New("cannot unmarshal receipts")
var errCannotSaveReceipts = errors.New("cannot save receipts")
var errCannotLoadReceipts = errors.New("cannot load receipts")
var errNilReceiptsHolder = errors.New("nil receipts holder")
var errNilBlockHeader = errors.New("nil block header")
var errEmptyBlockHash = errors.New("empty block hash")
