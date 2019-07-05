@0xff99b03cb6309633;
using Go = import "/go.capnp";
$Go.package("capnp");
$Go.import("_");


struct FeeTxCapn {
   nonce      @0:   UInt64; 
   value      @1:   Data;
   rcvAddr    @2:   Data;
   txHash     @3:   Data;
} 

##compile with:

##
##
##   capnpc  -I$GOPATH/src/github.com/glycerine/go-capnproto -ogo $GOPATH/src/github.com/ElrondNetwork/elrond-go/data/feeTx/capnp/schema.capnp

