@0xa6e50837d4563fc2;
using Go = import "/go.capnp";
$Go.package("capnp");
$Go.import("_");

struct ReceiptCapn {
   value      @0:   Data;
   sndAddr    @1:   Data;
   data       @2:   Data;
   txHash     @3:   Data;
}

##compile with:
##
##
##   capnpc  -I$GOPATH/src/github.com/glycerine/go-capnproto -ogo $GOPATH/src/github.com/ElrondNetwork/elrond-go/data/receipt/capnp/schema.capnp