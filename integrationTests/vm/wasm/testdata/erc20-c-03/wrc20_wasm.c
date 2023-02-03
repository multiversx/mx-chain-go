#include "chain/context.h"
#include "chain/bigInt.h"
#include "chain/crypto.h"
#include "chain/util.h"

// global data used in functions, will be statically allocated to WebAssembly memory
byte sender[32]          = {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0};
byte recipient[32]       = {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0};
byte caller[32]          = {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0};
byte currentKey[32]      = {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0};
byte balanceKeyRaw[33]   = {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0};
byte allowanceKeyRaw[65] = {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0};

byte approveEvent[32]  = {0x71,0x34,0x69,0x2B,0x23,0x0B,0x9E,0x1F,0xFA,0x39,0x09,0x89,0x04,0x72,0x21,0x34,0x15,0x96,0x52,0xB0,0x9C,0x5B,0xC4,0x1D,0x88,0xD6,0x69,0x87,0x79,0xD2,0x28,0xFF};
byte transferEvent[32] = {0xF0,0x99,0xCD,0x8B,0xDE,0x55,0x78,0x14,0x84,0x2A,0x31,0x21,0xE8,0xDD,0xFD,0x43,0x3A,0x53,0x9B,0x8C,0x9F,0x14,0xBF,0x31,0xEB,0xF1,0x08,0xD1,0x2E,0x61,0x96,0xE9};

byte currentTopics[96] = {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
                          0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
                          0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0};
byte currentLogVal[32] = {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0};

// error messages declared statically with the help of macros
ERROR_MSG(ERR_TRANSFER_NEG, "transfer amount cannot be negative")
ERROR_MSG(ERR_ALLOWANCE_NEG, "approve amount cannot be negative")
ERROR_MSG(ERR_ALLOWANCE_EXCEEDED, "allowance exceeded")
ERROR_MSG(ERR_INSUFFICIENT_FUNDS, "insufficient funds")

void computeTotalSupplyKey(byte *destination) {
  // only the total supply key starts with byte "0"
  for (int i = 0; i < 32; i++) {
    destination[i] = 0;
  }
}

void computeBalanceKey(byte *destination, byte *address) {
  // "1" is for balance keys
  balanceKeyRaw[0] = 1;

  // append address
  for (int i = 0; i < 32; i++) {
    balanceKeyRaw[1+i] = address[i];
  }
  keccak256(balanceKeyRaw, 33, destination);
}

void computeAllowanceKey(byte *destination, byte *from, byte* to) {
  // "2" is for allowance keys
  allowanceKeyRaw[0] = 2;

  // append "from"
  for (int i = 0; i < 32; i++) {
    allowanceKeyRaw[1+i] = from[i];
  }
  // append "to"
  for (int i = 0; i < 32; i++) {
    allowanceKeyRaw[33+i] = to[i];
  }
  keccak256(allowanceKeyRaw, 65, destination);
}

// both transfer and approve have 3 topics (event identifier, sender, recipient)
// so both prepare the log the same way
void saveLogWith3Topics(byte *topic1, byte *topic2, byte *topic3, bigInt value) {
  // copy all topics to currentTopics
  for (int i = 0; i < 32; i++) {
    currentTopics[i] = topic1[i];
  }
  for (int i = 0; i < 32; i++) {
    currentTopics[32+i] = topic2[i];
  }
  for (int i = 0; i < 32; i++) {
    currentTopics[64+i] = topic3[i];
  }

  // extract value bytes to memory
  int valueLen = bigIntGetUnsignedBytes(value, currentLogVal);

  // call api
  writeLog(currentLogVal, valueLen, currentTopics, 3);
}

// constructor function
// is called immediately after the contract is created
// will set the fixed global token supply and give all the supply to the creator
void init() {
  CHECK_NUM_ARGS(1);
  CHECK_NOT_PAYABLE();

  getCaller(sender);
  bigInt totalSupply = bigIntNew(0);
  bigIntGetSignedArgument(0, totalSupply);

  // set total supply
  computeTotalSupplyKey(currentKey);
  bigIntStorageStoreUnsigned(currentKey, 32, totalSupply);

  // sender balance <- total supply
  computeBalanceKey(currentKey, sender);
  bigIntStorageStoreUnsigned(currentKey, 32, totalSupply);
}

// getter function: retrieves total token supply
void totalSupply() {
  CHECK_NUM_ARGS(0);
  CHECK_NOT_PAYABLE();
  
  // load total supply from storage
  computeTotalSupplyKey(currentKey);
  bigInt totalSupply = bigIntNew(0);
  bigIntStorageLoadUnsigned(currentKey, 32, totalSupply);

  // return total supply as big int
  bigIntFinishUnsigned(totalSupply);
}

// getter function: retrieves balance for an account
void balanceOf() {
  CHECK_NUM_ARGS(1);
  CHECK_NOT_PAYABLE();

  // argument: account to get the balance for
  getArgument(0, caller); 

  // load balance
  computeBalanceKey(currentKey, caller);
  bigInt balance = bigIntNew(0);
  bigIntStorageLoadUnsigned(currentKey, 32, balance);

  // return balance as big int
  bigIntFinishUnsigned(balance);
}

// getter function: retrieves allowance granted from one account to another
void allowance() {
  CHECK_NUM_ARGS(2);
  CHECK_NOT_PAYABLE();

  // 1st argument: owner
  getArgument(0, sender);

  // 2nd argument: spender
  getArgument(1, recipient);

  // get allowance
  computeAllowanceKey(currentKey, sender, recipient);
  bigInt allowance = bigIntNew(0);
  bigIntStorageLoadUnsigned(currentKey, 32, allowance);

  // return allowance as big int
  bigIntFinishUnsigned(allowance);
}

// transfers tokens from sender to another account
void transferToken() {
  CHECK_NUM_ARGS(2);
  CHECK_NOT_PAYABLE();

  // sender is the caller
  getCaller(sender);

  // 1st argument: recipient
  getArgument(0, recipient);

  // 2nd argument: amount (should not be negative)
  bigInt amount = bigIntNew(0);
  bigIntGetSignedArgument(1, amount);
  if (bigIntCmp(amount, bigIntNew(0)) < 0) {
    SIGNAL_ERROR(ERR_TRANSFER_NEG);
    return;
  }

  // load sender balance
  computeBalanceKey(currentKey, sender);
  bigInt senderBalance = bigIntNew(0);
  bigIntStorageLoadUnsigned(currentKey, 32, senderBalance);

  // check if enough funds
  if (bigIntCmp(amount, senderBalance) > 0) {
    SIGNAL_ERROR(ERR_INSUFFICIENT_FUNDS);
    return;
  }

  // update sender balance
  bigIntSub(senderBalance, senderBalance, amount);
  bigIntStorageStoreUnsigned(currentKey, 32, senderBalance);

  // load & update receiver balance
  computeBalanceKey(currentKey, recipient);
  bigInt receiverBalance = bigIntNew(0);
  bigIntStorageLoadUnsigned(currentKey, 32, receiverBalance);
  bigIntAdd(receiverBalance, receiverBalance, amount);
  bigIntStorageStoreUnsigned(currentKey, 32, receiverBalance);

  // log operation
  saveLogWith3Topics(transferEvent, sender, recipient, amount);
}

// sender allows beneficiary to use given amount of tokens from sender's balance
// it will completely overwrite any previously existing allowance from sender to beneficiary
void approve() {
  CHECK_NUM_ARGS(2);
  CHECK_NOT_PAYABLE();

  // sender is the caller
  getCaller(sender);

  // 1st argument: spender (beneficiary)
  getArgument(0, recipient);

  // 2nd argument: amount (should not be negative)
  bigInt amount = bigIntNew(0);
  bigIntGetSignedArgument(1, amount);
  if (bigIntCmp(amount, bigIntNew(0)) < 0) {
    SIGNAL_ERROR(ERR_ALLOWANCE_NEG);
    return;
  }

  // store allowance
  computeAllowanceKey(currentKey, sender, recipient);
  bigIntStorageStoreUnsigned(currentKey, 32, amount);

  // log operation
  saveLogWith3Topics(approveEvent, sender, recipient, amount);
}


// caller uses allowance to transfer funds between 2 other accounts
void transferFrom() {
  CHECK_NUM_ARGS(3);
  CHECK_NOT_PAYABLE();

  // save caller
  getCaller(caller);

  // 1st argument: sender
  getArgument(0, sender);

  // 2nd argument: recipient
  getArgument(1, recipient);

  // 3rd argument: amount
  bigInt amount = bigIntNew(0);
  bigIntGetSignedArgument(2, amount);
  if (bigIntCmp(amount, bigIntNew(0)) < 0) {
    SIGNAL_ERROR(ERR_TRANSFER_NEG);
    return;
  }

  // load allowance
  computeAllowanceKey(currentKey, sender, caller);
  bigInt allowance = bigIntNew(0);
  bigIntStorageLoadUnsigned(currentKey, 32, allowance);

  // amount should not exceed allowance
  if (bigIntCmp(amount, allowance) > 0) {
    SIGNAL_ERROR(ERR_ALLOWANCE_EXCEEDED);
    return;
  }

  // update allowance
  bigIntSub(allowance, allowance, amount);
  bigIntStorageStoreUnsigned(currentKey, 32, allowance);

  // load sender balance
  computeBalanceKey(currentKey, sender);
  bigInt senderBalance = bigIntNew(0);
  bigIntStorageLoadUnsigned(currentKey, 32, senderBalance);

  // check if enough funds
  if (bigIntCmp(amount, senderBalance) > 0) {
    SIGNAL_ERROR(ERR_INSUFFICIENT_FUNDS);
    return;
  }

  // update sender balance
  bigIntSub(senderBalance, senderBalance, amount);
  bigIntStorageStoreUnsigned(currentKey, 32, senderBalance);

  // load & update receiver balance
  computeBalanceKey(currentKey, recipient);
  bigInt receiverBalance = bigIntNew(0);
  bigIntStorageLoadUnsigned(currentKey, 32, receiverBalance);
  bigIntAdd(receiverBalance, receiverBalance, amount);
  bigIntStorageStoreUnsigned(currentKey, 32, receiverBalance);

  // log operation
  saveLogWith3Topics(transferEvent, sender, recipient, amount);
}

// global data used in next function, will be allocated to WebAssembly memory
i32 selector[1] = {0};
void _main(void) {
}
