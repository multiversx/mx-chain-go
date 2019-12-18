@0xff99b03cb6309633;
using Go = import "/go.capnp";
$Go.package("capnp");
$Go.import("_");


struct AuthMessageCapn {
   message      @0: Data;
   sig          @1: Data;
   pubkey       @2: Data;
   timestamp    @3: Int64;
} 

##compile with:

##
##
##   capnpc  -I$GOPATH/src/github.com/glycerine/go-capnproto -ogo $GOPATH/src/github.com/ElrondNetwork/elrond-go/p2p/data/capnp/authMessage.capnp

