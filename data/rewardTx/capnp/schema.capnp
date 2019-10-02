@0xa6e50837d4563fc2;
using Go = import "/go.capnp";
$Go.package("capnp");
$Go.import("_");

struct RewardTxCapn {
   round      @0:   UInt64;
   epoch      @1:   UInt32;
   value      @2:   Data;
   rcvAddr    @3:   Data;
   shardId    @4:   UInt32;
} 

##compile with:

##
##
##   capnpc  -I$GOPATH/src/github.com/glycerine/go-capnproto -ogo $GOPATH/src/github.com/ElrondNetwork/elrond-go/data/rewardTx/capnp/schema.capnp

