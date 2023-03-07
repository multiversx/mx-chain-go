#include "args.h"

const byte Base16Figures[] = {
	'0', '1', '2', '3',
	'4', '5', '6', '7',
	'8', '9', 'A', 'B',
	'C', 'D', 'E', 'F'
};

BinaryArgs NewBinaryArgs() {
	BinaryArgs args;
	args.numArgs = 0;
	args.serialized = 0;
	args.lenSerialized = 0;

	for (int i = 0; i < MAX_BINARY_ARGS; i++) {
		args.arguments[i] = 0;
	}

	for (int i = 0; i < MAX_BINARY_ARGS; i++) {
		args.lengths[i] = 0;
	}

	for (int i = 0; i < MAX_BINARY_ARGS; i++) {
		args.lengthsAsI32[i] = 0;
	}

	return args;
}

int AddBinaryArg(BinaryArgs *args, byte *arg, int arglen) {
	int n = args->numArgs;
	args->arguments[n] = arg;
	args->lengths[n] = arglen;
	args->numArgs = n + 1;

	return n;
}

int TrimLeftZeros(BinaryArgs *args, int argIndex) {
	if (argIndex >= args->numArgs) {
		return -1;
	}

	byte *cursor = args->arguments[argIndex];
	int argLen = args->lengths[argIndex];

	while (*cursor == 0 && argLen > 0) {
		cursor++;
		argLen--;
	}

	int trimmedZeros = args->lengths[argIndex] - argLen;
	args->arguments[argIndex] = cursor;
	args->lengths[argIndex] = argLen;

	return trimmedZeros;
}

int SerializeBinaryArgs(BinaryArgs *args, byte *serializedBuffer) {
	int cursor = 0;
	for (int i = 0; i < args->numArgs; i++) {
		byte argLen = args->lengths[i];

		for (int j = 0; j < argLen; j++) {
			byte b = args->arguments[i][j];
			serializedBuffer[cursor] = b;
			cursor += 1;
		}

		args->lengthsAsI32[i] = (unsigned int) argLen;
	}

	args->serialized = serializedBuffer;
	args->lenSerialized = cursor;

	return cursor;
}

int SerializeBinaryArgsToDataString(BinaryArgs *args, byte *serializedBuffer) {
	int cursor = 0;
	for (int i = 0; i < args->numArgs; i++) {
		byte argLen = args->lengths[i];

		for (int j = 0; j < argLen; j++) {
			byte b = args->arguments[i][j];

			serializedBuffer[cursor] = Base16Figures[b / 16];
			cursor += 1;
			serializedBuffer[cursor] = Base16Figures[b % 16];
			cursor += 1;
		}

		args->lengthsAsI32[i] = argLen >> 24;

		if (i < args->numArgs - 1) {
			serializedBuffer[cursor] = '@';
			cursor += 1;
		}
	}

	args->serialized = serializedBuffer;
	args->lenSerialized = cursor;

	return cursor;
}
