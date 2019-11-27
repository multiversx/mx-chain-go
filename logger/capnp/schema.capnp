@0xff99b03cb6309633;
using Go = import "/go.capnp";
$Go.package("capnp");
$Go.import("_");


struct LogLineMessageCapn {
   message      @0: Text;
   logLevel     @1: Int32;
   args         @2: List(Text);
   timestamp    @3: Int64;
} 

##compile with:

##
##
##   capnpc  -I$GOPATH/src/github.com/glycerine/go-capnproto -ogo $GOPATH/src/github.com/ElrondNetwork/elrond-go/logger/capnp/schema.capnp

