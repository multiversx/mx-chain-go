@0xff99b03cb6309633;
using Go = import "/go.capnp";
$Go.package("capnproto1");
$Go.import("_");


struct TransactionCapn { 
   nonce      @0:   UInt64; 
   value      @1:   Data;
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
##   capnp compile -ogo ./schema.capnp

