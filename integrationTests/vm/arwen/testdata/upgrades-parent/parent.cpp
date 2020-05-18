typedef unsigned char byte;
typedef unsigned int u32;
typedef unsigned long long u64;

extern "C"
{
    void int64finish(long long value);
    void finish(byte *data, int length);
    int createContract(byte *value, byte *code, int length, byte *result, int numArguments, byte *argumentsLengths, byte *data);
    int getNumArguments();
    int getArgument(int argumentIndex, byte *argument);
    int getArgumentLength(int argumentIndex);
    int storageStore(byte *key, int keyLength, byte *data, int dataLength);
    int storageLoad(byte *key, int keyLength, byte *data);
    void signalError(byte *message, int length);
    void asyncCall(byte *destination, byte *value, byte *data, int length);
}

void memcpy(void *dest, void *src, int n);

char *childContractAddressKey = "child000000000000000000000000000";

class Foo
{
public:
    Foo()
    {
        this->answer = 45;
    }

    long long GetAnswer()
    {
        return this->answer;
    }

private:
    long long answer;
};

extern "C" void getUltimateAnswer()
{
    Foo foo;
    auto answer = foo.GetAnswer();
    int64finish(answer);
}

extern "C" void getChildAddress()
{
    byte childAddress[32];
    storageLoad((byte *)childContractAddressKey, 32, childAddress);
    finish(childAddress, 32);
}

extern "C" void createChild()
{
    int codeLength = getArgumentLength(0);
    byte code[codeLength];
    getArgument(0, code);
    byte childAddress[32];
    createContract(nullptr, code, codeLength, childAddress, 0, nullptr, nullptr);
    storageStore((byte *)childContractAddressKey, 32, childAddress, 32);
}

extern "C" void upgradeChild()
{
    int codeLength = getArgumentLength(0);
    byte code[codeLength];
    getArgument(0, code);

    byte childAddress[32];
    storageLoad((byte *)childContractAddressKey, 32, childAddress);

    // "upgradeContract@code@0100"
    int dataLength = 15 + 1 + codeLength + 1 + 4;
    byte data[dataLength];
    memcpy(data, (void *)"upgradeContract", 15);
    memcpy(data + 15, (void *)"@", 1);
    memcpy(data + 15 + 1, (void *)code, codeLength);
    memcpy(data + 15 + 1 + codeLength, (void *)"@", 1);
    memcpy(data + 15 + 1 + codeLength + 1, (void *)"0100", 4);
    asyncCall(childAddress, nullptr, data, dataLength);
}

void memcpy(void *dest, void *src, int n)
{
    char *csrc = (char *)src;
    char *cdest = (char *)dest;

    for (int i = 0; i < n; i++)
    {
        cdest[i] = csrc[i];
    }
}
