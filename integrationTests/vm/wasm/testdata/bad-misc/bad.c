#include "chain/context.h"
#include "chain/bigInt.h"

byte sender[32] = {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};
byte recipient[32] = {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};
byte caller[32] = {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};
byte currentKey[32] = {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};

byte approveEvent[32] = {0x71, 0x34, 0x69, 0x2B, 0x23, 0x0B, 0x9E, 0x1F, 0xFA, 0x39, 0x09, 0x89, 0x04, 0x72, 0x21, 0x34, 0x15, 0x96, 0x52, 0xB0, 0x9C, 0x5B, 0xC4, 0x1D, 0x88, 0xD6, 0x69, 0x87, 0x79, 0xD2, 0x28, 0xFF};
byte transferEvent[32] = {0xF0, 0x99, 0xCD, 0x8B, 0xDE, 0x55, 0x78, 0x14, 0x84, 0x2A, 0x31, 0x21, 0xE8, 0xDD, 0xFD, 0x43, 0x3A, 0x53, 0x9B, 0x8C, 0x9F, 0x14, 0xBF, 0x31, 0xEB, 0xF1, 0x08, 0xD1, 0x2E, 0x61, 0x96, 0xE9};

byte currentTopics[96] = {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                          0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                          0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};
byte currentLogVal[32] = {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};

void init()
{
}

void memoryFault()
{
  sender[2147483647] = 42;
}

void divideByZero()
{
  int x = 1;
  int y = 0;
  int z = x / y;
  writeLog(z, z, z, z);
}

void badGetOwner1()
{
  byte *pointer = 2147483647;
  getSCAddress(pointer);
}

void badGetBlockHash1()
{
  byte *pointer = 2147483647;
  getBlockHash(0, pointer);
}

void badGetBlockHash2()
{
  byte *pointer = 0;
  getBlockHash(2147483647, pointer);
}

void badGetBlockHash3()
{
  byte *pointer = 2147483647;
  getBlockHash(1, pointer);
}

void badWriteLog1()
{
  byte *pointer = 0;
  byte *topicPtr = 0;
  writeLog(pointer, 1, topicPtr, -1);
}

void badWriteLog2()
{
  byte *pointer = 0;
  byte *topicPtr = 0;
  writeLog(pointer, -1, topicPtr, 0);
}

void badWriteLog3()
{
  byte *pointer = 2147483647;
  byte *topicPtr = 0;
  writeLog(pointer, 0, topicPtr, 0);
}

void badWriteLog4()
{
  byte *pointer = 0;
  byte *topicPtr = 2147483647;
  writeLog(pointer, 0, topicPtr, 500);
}

void badBigIntStorageStore1()
{
  bigInt number = bigIntNew(100);
  bigIntStorageStoreUnsigned("test", 32, number + 42);
}

i64 doStackoverflow(i64 a) {
  if (a % 2 == 0) {
    return 42;
  }

  i64 x = doStackoverflow(a*8+1);
  i64 y = doStackoverflow(a*2+1);
  return x + y + a;
}

void badRecursive()
{
  i64 result = doStackoverflow(1);
  bigInt resultBig = bigIntNew(result);
  bigIntFinishUnsigned(resultBig);
}

void _main(void)
{
}
