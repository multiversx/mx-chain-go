@0x933ccff820d2c436;
using Go = import "/go.capnp";
$Go.package("capnproto2");
$Go.import("_");

struct TxCapnp $Go.doc("The Transaction class implements the transaction used for moving assets"){
    nonce @0:UInt64;
    value @1:UInt64;
    rcvAddr @2:Data;
    sndAddr @3:Data;
    gasPrice @4:UInt64;
    gasLimit @5:UInt64;
    data @6:Data;
    signature @7:Data;
    challenge @8:Data;
}
