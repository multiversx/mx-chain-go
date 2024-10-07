typedef unsigned char byte;
typedef unsigned int i32;
typedef unsigned long long i64;

void getSCAddress(byte *address);
int storageStore(byte *key, int keyLength, byte *data, int dataLength);
int storageLoad(byte *key, int keyLength, byte *data);
void finish(byte *data, int length);

int deployFromSourceContract(
    long long gas,
    byte *value,
    byte *sourceContractAddress,
    byte *codeMetadata,
    byte *newAddress,
    int numInitArgs,
    byte *initArgLengths,
    byte *initArgs);

i32 createAsyncCall(
    byte *destination,
    byte *value,
    byte *data,
    int dataLength,
    byte *success,
    int successLength,
    byte *error,
    int errorLength,
    long long gas,
    long long extraGasForCallback);

static const i32 ADDRESS_LENGTH = 32;

byte zero32_red[32] = {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};
byte zero32_green[32] = {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};

byte zeroEGLD[] = {
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};

byte codeMetadataUpgradeableReadable[2] = {5, 0};

byte emptyArguments[0] = {};
int emptyArgumentsLengths[0] = {};
int gasLimitDeploySelf = 20000000;
int gasLimitClaimDeveloperRewards = 6000000;

byte functionNameClaimDeveloperRewards[] = "ClaimDeveloperRewards";
byte functionNameDoSomething[] = "doSomething";
byte storageKeyChildAddress[] = "child";
byte something[] = "something";

void init()
{
}

void upgrade()
{
}

void doSomething()
{
    finish(something, sizeof(something) - 1);
}

void deployChild()
{
    byte *selfAddress = zero32_red;
    byte *newAddress = zero32_green;

    getSCAddress(selfAddress);

    deployFromSourceContract(
        gasLimitDeploySelf,
        zeroEGLD,
        selfAddress,
        codeMetadataUpgradeableReadable,
        newAddress,
        0,
        (byte *)emptyArgumentsLengths,
        emptyArguments);

    storageStore(storageKeyChildAddress, sizeof(storageKeyChildAddress) - 1, newAddress, ADDRESS_LENGTH);
}

void getChildAddress()
{
    byte *childAddress = zero32_red;
    storageLoad(storageKeyChildAddress, sizeof(storageKeyChildAddress) - 1, childAddress);
    finish(childAddress, ADDRESS_LENGTH);
}

void claimDeveloperRewardsOnChild()
{
    byte *childAddress = zero32_red;
    storageLoad(storageKeyChildAddress, sizeof(storageKeyChildAddress) - 1, childAddress);

    createAsyncCall(
        childAddress,
        zeroEGLD,
        functionNameClaimDeveloperRewards,
        sizeof(functionNameClaimDeveloperRewards) - 1,
        0,
        0,
        0,
        0,
        gasLimitClaimDeveloperRewards,
        0);
}
