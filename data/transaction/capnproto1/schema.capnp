@0xd97ff83ab03e4823;
using Go = import "/go.capnp";
$Go.package("capnproto1");
$Go.import("_");


struct TransactionCapn { 
   nonce      @0:   List(UInt8); 
   value      @1:   List(UInt8); 
   rcvAddr    @2:   List(UInt8); 
   sndAddr    @3:   List(UInt8); 
   gasPrice   @4:   List(UInt8); 
   gasLimit   @5:   List(UInt8); 
   data       @6:   List(UInt8); 
   signature  @7:   List(UInt8); 
   challenge  @8:   List(UInt8); 
   pubKey     @9:   List(UInt8); 
} 

##compile with:

##
##
##   capnpc  -I$GOPATH/src/github.com/glycerine/go-capnproto -ogo $GOPATH/src/github.com/ElrondNetwork/elrond-go-sandbox/data/transaction/capnproto1//schema.capnp

