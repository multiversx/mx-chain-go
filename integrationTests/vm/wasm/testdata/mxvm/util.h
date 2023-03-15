#include "context.h"
#include "bigInt.h"

#define STORAGE_KEY(key) \
    const int key##_KEY_LEN = sizeof #key - 1;\
    byte key##_KEY[key##_KEY_LEN + 1] = #key;

#define ERROR_MSG(var, str) \
    const int var##_LEN = sizeof str - 1;\
    byte var[var##_LEN + 1] = str;

#define SIGNAL_ERROR(var) signalError(var, var##_LEN);

ERROR_MSG(ERR_NUM_ARGS, "wrong number of arguments")
#define CHECK_NUM_ARGS(expected) {\
    if (getNumArguments() != expected) {\
        SIGNAL_ERROR(ERR_NUM_ARGS);\
        return;\
    }\
}

ERROR_MSG(ERR_NOT_PAYABLE, "attempted to transfer funds via a non-payable function")
#define CHECK_NOT_PAYABLE() {\
    int callValue = bigIntNew(0);\
    bigIntGetCallValue(callValue);\
    int zero = bigIntNew(0);\
    if (bigIntCmp(callValue, zero) > 0) {\
        SIGNAL_ERROR(ERR_NOT_PAYABLE);\
        return;\
    }\
}
