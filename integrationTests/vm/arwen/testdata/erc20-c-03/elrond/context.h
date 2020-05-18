#ifndef _CONTEXT_H_
#define _CONTEXT_H_

#include "types.h"

void getOwner(byte *ownerAddress);

// Call-related functions
void getCaller(byte *callerAddress);
int getFunction(byte *function);
int getCallValue(byte *result);
long long getGasLeft();
void finish(byte *data, int length);
void int64finish(long long value);
void writeLog(byte *pointer, int length, byte *topicPtr, int numTopics);
void asyncCall(byte *destination, byte *value, byte *data, int length);
void signalError(byte *message, int length);

int executeOnSameContext(long long gas, byte *address, byte *value, byte *function, int functionLength, int numArguments, byte *argumentsLengths, byte *arguments);
int executeOnDestContext(long long gas, byte *address, byte *value, byte *function, int functionLength, int numArguments, byte *argumentsLengths, byte *arguments);

// Blockchain-related functions
long long getBlockTimestamp();
int getBlockHash(long long nonce, byte *hash);

// Argument-related functions
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

#endif
