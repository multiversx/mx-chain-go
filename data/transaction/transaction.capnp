@0x933ccff820d2c436;
using Go = import "/go.capnp";
$Go.package("transaction");
$Go.import("_");

struct Transaction $Go.doc("The Transaction class implements the transaction used for moving assets"){
    nonce @0:Data;
    value @1:Data;
    rcvAddr @2:Data;
    sndAddr @3:Data;
    gasPrice @4:Data;
    gasLimit @5:Data;
    data @6:Data;
    signature @7:Data;
    challenge @8:Data;
    pubKey @9:Data;
}
