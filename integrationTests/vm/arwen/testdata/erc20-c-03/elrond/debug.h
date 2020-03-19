#include "types.h"
#include "bigInt.h"

// Debug-related functions.
void debugPrintBigInt(bigInt argument);
void debugPrintInt64(i64 value);
void debugPrintInt32(i32 value);
void debugPrintBytes(byte *data, int length);
void debugPrintString(byte *data, int length);
void debugPrintLineNumber(i32 value);

int strlen(const byte *str)
{
    const byte *s;
    for (s = str; *s; ++s);
    return (s - str);
}

// A helper function.
void myDebugPrintString(byte *data) {
    debugPrintString(data, strlen(data));
}
