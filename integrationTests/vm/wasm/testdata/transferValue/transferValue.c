typedef unsigned char byte;
typedef unsigned int i32;
typedef unsigned long long i64;

int getArgument(int argumentIndex, byte *argument);
int transferValueExecute(byte *destination, byte *value, long long gas, byte *function, int functionLength, int numArguments, byte *argumentsLengths, byte *arguments);
void getCaller(byte *callerAddress);
i32 createAsyncCall(byte *destination, byte *value, byte *data, int dataLength, byte *success, int successLength, byte *error, int errorLength, long long gas, long long extraGasForCallback);

byte zero32_a[32] = {0};
byte zero32_b[32] = {0};
byte zero32_c[32] = {0};

byte oneAtomOfEGLD[32] = {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1};
byte functionNameAskMoney[] = "askMoney";
byte functionNameMyCallback[] = "myCallback";

void init()
{
}

void upgrade()
{
}

void fund()
{
}

void forwardAskMoney()
{
    byte *otherContract = zero32_a;
    getArgument(0, otherContract);

    createAsyncCall(
        otherContract,
        0,
        functionNameAskMoney,
        sizeof(functionNameAskMoney) - 1,
        functionNameMyCallback,
        sizeof(functionNameMyCallback) - 1,
        functionNameMyCallback,
        sizeof(functionNameMyCallback) - 1,
        15000000,
        0);
}

void askMoney()
{
    byte *caller = zero32_a;

    getCaller(caller);
    transferValueExecute(caller, oneAtomOfEGLD, 0, 0, 0, 0, 0, 0);
}

void myCallback()
{
}
