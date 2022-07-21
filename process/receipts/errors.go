package receipts

import "errors"

var errCannotCreateReceiptsRepository = errors.New("cannot create receipts repository")
var errCannotMarshalReceipts = errors.New("cannot create marshalized receipts")
var errCannotUnmarshalReceipts = errors.New("cannot unmarshal receipts")
var errCannotSaveReceipts = errors.New("cannot save receipts")
var errCannotLoadReceipts = errors.New("cannot load receipts")
