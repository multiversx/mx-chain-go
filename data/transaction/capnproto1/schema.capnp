@0xd97ff83ab03e4823;
using Go = import "/go.capnp";
$Go.package("capnproto1");
$Go.import("_");


struct TransactionCapn { 
   nonce      @0:   UInt64;
   value      @1:   UInt64;
   rcvAddr    @2:   Data;
   sndAddr    @3:   Data;
   gasPrice   @4:   UInt64;
   gasLimit   @5:   UInt64;
   data       @6:   Data;
   signature  @7:   Data;
   challenge  @8:   Data;
} 

##compile with:

##
##
##   capnpc  -I$GOPATH/src/github.com/glycerine/go-capnproto -ogo $GOPATH/src/github.com/ElrondNetwork/elrond-go-sandbox/data/transaction/capnproto1//schema.capnp

