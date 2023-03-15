#ifndef _ARGS_H_
#define _ARGS_H_

#include "types.h"

static const int MAX_BINARY_ARGS = 10;

typedef struct binaryArgs {
	byte *arguments[MAX_BINARY_ARGS];
	byte lengths[MAX_BINARY_ARGS];
	unsigned int lengthsAsI32[MAX_BINARY_ARGS];
	int numArgs;
	byte *serialized;
	int lenSerialized;
} BinaryArgs;

BinaryArgs NewBinaryArgs();
int AddBinaryArg(BinaryArgs *args, byte *arg, int arglen);
int TrimLeftZeros(BinaryArgs *args, int argIndex);
int SerializeBinaryArgs(BinaryArgs *args, byte *result);
int SerializeBinaryArgsToDataString(BinaryArgs *args, byte *result);

#endif
