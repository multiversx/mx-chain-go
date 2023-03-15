#ifndef _CONTEXT_H_
#define _CONTEXT_H_

#include "types.h"

void getSCAddress(byte *address);
void getOwnerAddress(byte *address);
int getShardOfAddress(byte *address);
int isSmartContract(byte *address);

// EllipticCurve-Related functions

void addEC(int xResultHandle, int yResultHandle, int ecHandle, int fstPointXHandle, int fstPointYHandle, int sndPointXHandle, int sndPointYHandle);
void doubleEC(int xResultHandle, int yResultHandle, int ecHandle, int pointXHandle, int pointYHandle);
int isOnCurveEC(int ecHandle, int pointXHandle, int pointYHandle);
int scalarBaseMultEC(int xResultHandle, int yResultHandle, int ecHandle, byte *dataOffset, int length);
int scalarMultEC(int xResultHandle, int yResultHandle, int ecHandle, int pointXHandle, int pointYHandle, byte *dataOffset, int length);
int marshalEC(int xPairHandle, int yPairHandle, int ecHandle, byte *resultOffset);
int unmarshalEC(int xResultHandle, int yResultHandle, int ecHandle, byte *dataOffset, int length);
int marshalCompressedEC(int xPairHandle, int yPairHandle, int ecHandle, byte *resultOffset);
int unmarshalCompressedEC(int xResultHandle, int yResultHandle, int ecHandle, byte *dataOffset, int length);
int generateKeyEC(int xPubKeyHandle, int yPubKeyHandle, int ecHandle, byte *resultOffset);
int getCurveLengthEC(int ecHandle);
int getPrivKeyByteLengthEC(int ecHandle);
int createEC(byte *dataOffset, int dataLength);
int ellipticCurveGetValues(int ecHandle, int fieldOrderHandle, int basePointOrderHandle, int eqConstantHandle, int xBasePointHandle, int yBasePointHandle);

// Managed Buffers
int	mBufferNew();
int mBufferNewFromBytes(byte*dataOffset, int dataLength);
int mBufferSetRandom(int mBufferHandle, int length);
int	mBufferSetBytes(int mBufferHandle, byte *dataOffset, int dataLength);
int	mBufferSetByteSlice(int mBufferHandle, int startingPosition, int dataLength, byte *dataOffset);
int mBufferGetLength(int mBufferHandle);
int	mBufferGetBytes(int mBufferHandle, byte *resultOffset);
int	mBufferAppend(int mBufferHandle, int otherHandle);
int	mBufferAppendBytes(int mBufferHandle, byte *dataOffset, int dataLength);
int	mBufferToBigIntUnsigned(int mBufferHandle, int bigIntHandle);
int mBufferToBigIntSigned(int mBufferHandle, int bigIntHandle);
int	mBufferFromBigIntUnsigned(int mBufferHandle, int bigIntHandle);
int	mBufferFromBigIntSigned(int mBufferHandle, int bigIntHandle);
int	mBufferStorageStore(int keyHandle, int mBufferHandle);
int	mBufferStorageLoad(int keyHandle, int mBufferHandle);
int	mBufferGetArgument(int id, int mBufferHandle);
int	mBufferFinish(int mBufferHandle);
int	mBufferToBigFloat(int mBufferHandle, int bigFloatHandle);
int	mBufferFromBigFloat(int mBufferHandle, int bigFloatHandle);

// Call-related functions
void getCaller(byte *callerAddress);
int getFunction(byte *function);
int getCallValue(byte *result);
long long getGasLeft();
void writeLog(byte *pointer, int length, byte *topicPtr, int numTopics);
void asyncCall(byte *destination, byte *value, byte *data, int length);
int createAsyncCall(byte *destination, byte *value, byte *data, int length,  byte *successCallback, int successLength, byte *errorCallback, int errorLength, long long gas, long long extraGasForCallback);
void signalError(byte *message, int length);

int executeOnSameContext(
		long long gas,
		byte *address,
		byte *value,
		byte *function,
		int functionLength,
		int numArguments,
		byte *argumentsLengths,
		byte *arguments);

int executeOnDestContext(
		long long gas,
		byte *address,
		byte *value,
		byte *function,
		int functionLength,
		int numArguments,
		byte *argumentsLengths,
		byte *arguments);

int executeOnDestContextByCaller(
		long long gas,
		byte *address,
		byte *value,
		byte *function,
		int functionLength,
		int numArguments,
		byte *argumentsLengths,
		byte *arguments);

int createContract(
		long long gas,
		byte *value,
		byte *code,
		byte *codeMetadata,
		int codeSize,
		byte *newAddress,
		int numInitArgs,
		byte *initArgLengths,
		byte *initArgs);

int deployFromSourceContract(
		long long gas,
		byte *value,
		byte *sourceContractAddress,
		byte *codeMetadata,
		byte *newAddress,
		int numInitArgs,
		byte *initArgLengths,
		byte *initArgs);	

void upgradeFromSourceContract(
		byte *destContractAddress,
		long long gas,
		byte *value,
		byte *sourceContractAddress,
		byte *codeMetadata,
		int numInitArgs,
		byte *initArgLengths,
		byte *initArgs);	

// Return-related functions
void finish(byte *data, int length);
void int64finish(long long value);
int getNumReturnData();
int getReturnDataSize(int index);
int getReturnData(int index, byte *data);

// Blockchain-related functions
long long getBlockTimestamp();
int getBlockHash(long long nonce, byte *hash);

// Argument-related functions
long long smallIntGetUnsignedArgument(int argumentIndex);
int getNumArguments();
int getArgument(int argumentIndex, byte *argument);
long long int64getArgument(int argumentIndex);
int getArgumentLength(int argumentIndex);

// Account-related functions
void getExternalBalance(byte *address, byte *balance);
int transferValue(byte *destination, byte *value, byte *data, int length);

// Storage-related functions
int storageLoadLength(byte *key, int keyLength);
int storageStore(byte *key, int keyLength, byte *data, int dataLength);
int storageLoad(byte *key, int keyLength, byte *data);
int int64storageStore(byte *key, int keyLength, long long value);
long long int64storageLoad(byte *key, int keyLength);

// Timelocks-related functions
int setStorageLock(byte *key, int keyLen, long long timeLock);
long long getStorageLock(byte *key, int keyLen);
int isStorageLocked(byte *key, int keyLen);
int clearStorageLock(byte *key, int keyLen);

// ESDT-related functions
int getESDTTokenName(byte *name);
int getESDTValue(byte *value);
int getESDTBalance(
		byte *address,
		byte *tokenName,
		int tokenNameLen,
		long long nonce,
		byte *result);

#endif
