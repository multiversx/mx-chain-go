#ifndef _BIGFLOAT_H_
#define _BIGFLOAT_H_

#include "types.h"

int bigFloatNewFromParts(int intBase, int subIntBase, int exponent);
int bigFloatNewFromFrac(long long numerator, long long denominator);
int bigFloatNewFromSci(long long significand, long long exponent);

void bigFloatAdd(int destinationHandle, int op1Handle, int op2Handle);
void bigFloatSub(int destinationHandle, int op1Handle, int op2Handle);
void bigFloatMul(int destinationHandle, int op1Handle, int op2Handle);
void bigFloatDiv(int destinationHandle, int op1Handle, int op2Handle);
void bigFloatTruncate(int opHandle, int bigIntHandle);

void bigFloatAbs(int destinationHandle, int opHandle);
void bigFloatNeg(int destinationHandle, int opHandle);
int bigFloatCmp(int op1Handle, int op2Handle);
int bigFloatSign(int opHandle);
void bigFloatClone(int destinationHandle, int opHandle);
void bigFloatSqrt(int destinationHandle, int opHandle);
void bigFloatPow(int destinationHandle, int op1Handle, int smallValue);

void bigFloatFloor(int destBigIntHandle, int opHandle);
void bigFloatCeil(int destBigIntHandle, int opHandle);

int bigFloatIsInt(int opHandle);
void bigFloatSetInt64(int destinationHandle, long long value);
void bigFloatSetBigInt(int destinationHandle, int bigIntHandle);

void bigFloatGetConstPi(int destinationHandle);
void bigFloatGetConstE(int destinationHandle);

#endif
