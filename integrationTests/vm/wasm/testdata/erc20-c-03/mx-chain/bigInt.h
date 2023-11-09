#ifndef _BIGINT_H_
#define _BIGINT_H_

#include "types.h"

typedef unsigned int bigInt;

bigInt    bigIntNew(long long value);

void      bigIntGetUnsignedArgument(int argumentIndex, bigInt argument);
void      bigIntGetSignedArgument(int argumentIndex, bigInt argument);

int       bigIntStorageLoadUnsigned(byte *key, int keyLength, bigInt value);
int       bigIntStorageStoreUnsigned(byte *key, int keyLength, bigInt value);

void      bigIntAdd(bigInt destination, bigInt op1, bigInt op2);
void      bigIntSub(bigInt destination, bigInt op1, bigInt op2);
void      bigIntMul(bigInt destination, bigInt op1, bigInt op2);
int       bigIntCmp(bigInt op1, bigInt op2);

int       bigIntIsInt64(bigInt reference);
long long bigIntGetInt64(bigInt reference);
void      bigIntSetInt64(bigInt destination, long long value);

void      bigIntFinishUnsigned(bigInt reference);
void      bigIntFinishSigned(bigInt reference);
void      bigIntGetCallValue(bigInt destination);
void      bigIntGetExternalBalance(byte *address, bigInt result);

int       bigIntByteLength(bigInt reference);
int       bigIntGetUnsignedBytes(bigInt reference, byte *byte);
int       bigIntGetSignedBytes(bigInt reference, byte *byte);
void      bigIntSetUnsignedBytes(bigInt destination, byte *byte, int byteLength);
void      bigIntSetSignedBytes(bigInt destination, byte *byte, int byteLength);

#endif
