typedef unsigned char byte;
typedef unsigned int i32;
typedef unsigned long long i64;

void getSCAddress(byte *address);
int transferValue(byte *destination, byte *value, byte *data, int length);
void getCaller(byte *callerAddress);
int getCallValue(byte *result);
i32 createAsyncCall(byte *destination, byte *value, byte *data, int dataLength, byte *success, int successLength, byte *error, int errorLength, long long gas, long long extraGasForCallback);
void finish(byte *data, int length);

byte zero32_a[32] = {0};
byte zero32_b[32] = {0};
byte zero32_c[32] = {0};

byte functionNameEchoValue[] = "echoValue";
byte strThankYouButNo[] = "thank you, but no";

void init()
{
}

void upgrade()
{
}

void receive()
{
    byte *selfAddress = zero32_a;
    byte *callValue = zero32_b;

    getSCAddress(selfAddress);
    getCallValue(callValue);

    createAsyncCall(
        selfAddress,
        callValue,
        functionNameEchoValue,
        sizeof(functionNameEchoValue) - 1,
        0,
        0,
        0,
        0,
        15000000,
        0);
}

void echoValue()
{
    byte *caller = zero32_a;
    byte *callValue = zero32_b;

    getCaller(caller);
    getCallValue(callValue);

    transferValue(caller, callValue, 0, 0);
    finish(strThankYouButNo, sizeof(strThankYouButNo) - 1);
}
