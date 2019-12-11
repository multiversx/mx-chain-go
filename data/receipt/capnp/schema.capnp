@0xff99b03cb6309633;
using Go = import "/go.capnp";
$Go.package("capnp");
$Go.import("_");


struct ReceiptCapn {
   value      @1:   Data;
   sndAddr    @2:   Data;
   data       @3:   Data;
   txHash     @4:   Data;
} 

##compile with:

##
##
##   capnp compile -ogo ./schema.capnp

