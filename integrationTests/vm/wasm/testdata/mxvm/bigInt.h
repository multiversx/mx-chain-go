#ifndef _BIGINT_H_
#define _BIGINT_H_

#include "types.h"

typedef unsigned int bigInt;

bigInt    bigIntNew(long long value);

void      bigIntGetUnsignedArgument(int argumentIndex, bigInt argument);
void      bigIntGetSignedArgument(int argumentIndex, bigInt argument);

int       bigIntStorageLoadUnsigned(byte *key, int keyLength, bigInt value);
int       bigIntStorageStoreUnsigned(byte *key, int keyLength, bigInt value);

void      bigIntAdd(bigInt destinationHandle, bigInt op1, bigInt op2);
void      bigIntSub(bigInt destinationHandle, bigInt op1, bigInt op2);
void      bigIntMul(bigInt destinationHandle, bigInt op1, bigInt op2);
int       bigIntCmp(bigInt op1, bigInt op2);

int       bigIntIsInt64(bigInt bigIntHandle);
long long bigIntGetInt64(bigInt bigIntHandle);
void      bigIntSetInt64(bigInt destinationHandle, long long value);

void      bigIntFinishUnsigned(bigInt bigIntHandle);
void      bigIntFinishSigned(bigInt bigIntHandle);
void      bigIntGetCallValue(bigInt destinationHandle);
void      bigIntgetExternalBalance(byte *address, bigInt result);

int       bigIntUnsignedByteLength(bigInt bigIntHandle);
int       bigIntSignedByteLength(bigInt bigIntHandle);
int       bigIntGetUnsignedBytes(bigInt bigIntHandle, byte *byte);
int       bigIntGetSignedBytes(bigInt bigIntHandle, byte *byte);
void      bigIntSetUnsignedBytes(bigInt destinationHandle, byte *byte, int byteLength);
void      bigIntSetSignedBytes(bigInt destinationHandle, byte *byte, int byteLength);

void      bigIntGetESDTCallValue(bigInt destinationHandle);
void      bigIntGetESDTExternalBalance(byte *addressOffset, byte *tokenIDOffset, unsigned int tokenIDLen, long long nonce, bigInt result);

#endif
