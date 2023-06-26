#ifndef _TEST_UTILS_H_
#define _TEST_UTILS_H_

#include "../mxvm/context.h"
#include "../mxvm/bigInt.h"
#include "../mxvm/types.h"
#include "../mxvm/bigFloat.h"

void intTo3String(int value, byte *string, int startPos) {
  string[startPos + 2] = (byte)('0' + value % 10);
  string[startPos + 1] = (byte)('0' + (value / 10) % 10);
  string[startPos + 0] = (byte)('0' + (value / 100) % 10);
}

void storeIterationNumber(byte iteration, byte prefix) {
  byte keyIter[] = "XkeyNNN.........................";
  byte valueIter[] = "XvalueNNN";
  intTo3String(iteration, keyIter, 4);
  intTo3String(iteration, valueIter, 6);
  keyIter[0] = prefix;
  valueIter[0] = prefix;
  storageStore(keyIter, 32, valueIter, 9);
}

void finishIterationNumber(byte iteration, byte prefix) {
  byte finishIter[] = "XfinishNNN";
  intTo3String(iteration, finishIter, 7);
  finishIter[0] = prefix;
  finish(finishIter, 10);
}

void incrementBigIntCounter(bigInt counterID) {
  bigIntSetInt64(42, 1);
  bigIntAdd(counterID, counterID, 42);
}

void didCallerPay(u64 value) {
	bigInt bigInt_payment = bigIntNew(0);
	bigIntGetCallValue(bigInt_payment);

	long long payment = bigIntGetInt64(bigInt_payment);
	if (payment != value) {
		byte message[] = "child execution requires tx value";
		signalError(message, 33);
	}
}

u32 reverseU32(u32 value) {
	u32 lastByteMask = 0x00000000000000FF;
	u32 result = 0;
	int size = sizeof(value);
	for (int i = 0; i < size; i++) {
		byte lastByte = value & lastByteMask;
		value >>= 8;

		result <<= 8;
		result += lastByte;
	}
	return result;
}

void not_ok() {
	byte msg[] = "not ok";
	finish(msg, 6);
}

void incrementIterCounter(byte *storageKey) {
  byte counterValue;
  int len = storageLoadLength(storageKey, 32);
  if (len == 0) {
    counterValue = 1;
    storageStore(storageKey, 32, &counterValue, 1);
  } else {
    storageLoad(storageKey, 32, &counterValue); 
    counterValue = counterValue + 1;
    storageStore(storageKey, 32, &counterValue, 1);
  }
}

void finishResult(int result) {
	if (result == 0) {
		byte message[] = "succ";
		finish(message, 4);
	}
	if (result == 1) {
		byte message[] = "fail";
		finish(message, 4);
	}
	if (result != 0 && result != 1) {
		byte message[] = "unkn";
		finish(message, 4);
	}
}

void bigFloatGetArgument(int argumentId, int bigFloatHandle) {
    int manBufHandle = mBufferNew();
    mBufferGetArgument(argumentId ,manBufHandle);
    mBufferToBigFloat(manBufHandle,bigFloatHandle);
}

void bigFloatFinish(int bigFloatHandle) {
    int manBufHandle = mBufferNew();
    mBufferFromBigFloat(manBufHandle,bigFloatHandle);
    mBufferFinish(manBufHandle);
}

#endif
