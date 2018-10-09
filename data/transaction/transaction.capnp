@0x933ccff820d2c436;
using Go = import "/go.capnp";
$Go.package("transaction");
$Go.import("_");

struct Transaction $Go.doc("The Transaction class implements the transaction used for moving assets"){
    nonce @0:List(UInt8);
    value @1:List(UInt8);
    rcvAddr @2:List(UInt8);
    sndAddr @3:List(UInt8);
    gasPrice @4:List(UInt8);
    gasLimit @5:List(UInt8);
    data @6:List(UInt8);
    signature @7:List(UInt8);
    challenge @8:List(UInt8);
    pubKey @9:List(UInt8);
}
