
// to avoid includes from libc, just hard-code things
typedef unsigned char uint8_t;
typedef int uint32_t;
typedef unsigned long long uint64_t;
​
​
// types used for Ethereum stuff
typedef uint8_t* bytes; // an array of bytes with unrestricted length
typedef uint8_t bytes32[32]; // an array of 32 bytes
typedef uint8_t address[32]; // an array of 20 bytes
typedef unsigned __int128 u128; // a 128 bit number, represented as a 16 bytes long little endian unsigned integer in memory
typedef uint32_t i32; // same as i32 in WebAssembly
typedef uint32_t i32ptr; // same as i32 in WebAssembly, but treated as a pointer to a WebAssembly memory offset
typedef uint64_t i64; // same as i64 in WebAssembly
​
// functions for ethereum stuff
void useGas(i64 amount);
void getCaller(i32ptr* resultOffset);
   // memory offset to load the address into (address)
i32 getCallDataSize();
void callDataCopy(i32ptr* resultOffset, i32 dataOffset, i32 length);
   // memory offset to load data into (bytes), the offset in the input data, the length of data to copy
void revert(i32ptr* dataOffset, i32 dataLength);
void finish(i32ptr* dataOffset, i32 dataLength);
void storageStore(i32ptr* pathOffset, i32ptr* resultOffset);
void storageLoad(i32ptr* pathOffset, i32ptr* resultOffset);
   //the memory offset to load the path from (bytes32), the memory offset to store/load the result at (bytes32)
void printMemHex(i32ptr* offset, i32 length);
void printStorageHex(i32ptr* key);
​
​
i64 reverse_bytes(i64 a){
  i64 b = 0;
  b += (a & 0xff00000000000000)>>56;
  b += (a & 0x00ff000000000000)>>40;
  b += (a & 0x0000ff0000000000)>>24;
  b += (a & 0x000000ff00000000)>>8;
  b += (a & 0x00000000ff000000)<<8;
  b += (a & 0x0000000000ff0000)<<24;
  b += (a & 0x000000000000ff00)<<40;
  b += (a & 0x00000000000000ff)<<56;
  return b;
}
​
​
// global data used in next function, will be allocated to WebAssembly memory
bytes32 addy[1] = {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0};
bytes32 balance[1] = {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0};
​
void do_balance() {
  if (getCallDataSize() != 24)
    revert(0, 0);
​
  // get address to check balance of, note: padded to 32 bytes since used as key
  callDataCopy((i32ptr*)addy, 4, 20);
​
  // get balance
  storageLoad((i32ptr*)addy, (i32ptr*)balance);
​
  // return balance
  finish((i32ptr*)balance, 32);
}
​
​
// global data used in next function, will be allocated to WebAssembly memory
bytes32 sender[1] = {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0};
bytes32 recipient[1] = {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0};
bytes32 value[1] = {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0};
bytes32 recipient_balance[1] = {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0};
bytes32 sender_balance[1] = {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0};
​
void do_transfer() {
  if (getCallDataSize() != 32)
    revert(0, 0);
​
  // get caller
  getCaller((i32ptr*)sender);
​
  // get recipient from message at byte 4, length 20
  callDataCopy((i32ptr*)recipient, 4, 20);
​
  // get amount to transfer from message at byte 24, length 8
  callDataCopy((i32ptr*)value, 24, 8);
  *(i64*)value = reverse_bytes(*(i64*)value);
​
  // get balances
  storageLoad((i32ptr*)sender, (i32ptr*)sender_balance);
  storageLoad((i32ptr*)recipient, (i32ptr*)recipient_balance);
  *(i64*)sender_balance = reverse_bytes(*(i64*)sender_balance);
  *(i64*)recipient_balance = reverse_bytes(*(i64*)recipient_balance);
​
  // make sure sender has enough
  if (*(i64*)sender_balance < *(i64*)value)
    revert(0, 0);
​
  // adjust balances
  * (i64*)sender_balance -= * (i64*)value;
  * (i64*)recipient_balance += * (i64*)value;
​
  // store results
  *(i64*)sender_balance = reverse_bytes(*(i64*)sender_balance);
  *(i64*)recipient_balance = reverse_bytes(*(i64*)recipient_balance);
  storageStore((i32ptr*)sender, (i32ptr*)sender_balance);
  storageStore((i32ptr*)recipient, (i32ptr*)recipient_balance);
​
}
​
​
// global data used in next function, will be allocated to WebAssembly memory
i32 selector[1] = {0};
​
void _main(void) {
  if (getCallDataSize() < 4)
    revert(0, 0);
​
  // get first four bytes of message, use for calling
  callDataCopy((i32ptr*)selector, 0, 4); //from byte 0, length 4
​
  // call function based on selector value
  switch (*selector) {
    case 0x1a029399:
      do_balance();
      break;
    case 0xbd9f355d:
      do_transfer();
      break;
    default:
      revert(0, 0);
  }
}