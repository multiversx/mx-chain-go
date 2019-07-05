@0xff99b03cb6309633;
using Go = import "/go.capnp";
$Go.package("capnp");
$Go.import("_");


struct SmartContractResultCapn {
   nonce      @0:   UInt64; 
   value      @1:   Data;
   rcvAddr    @2:   Data;
   txHash     @6:   Data;
} 

##compile with:

##
##
##   capnp compile -ogo ./schema.capnp

