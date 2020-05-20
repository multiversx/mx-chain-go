typedef unsigned char byte;
typedef unsigned int i32;
typedef unsigned long long i64;

typedef unsigned int bigInt;

int int64storageStore(byte *key, int keyLength, long long value);
long long int64storageLoad(byte *key, int keyLength);
i64 int64getArgument(int argumentIndex);
int getCallValue(byte *result);
void asyncCall(byte *destination, byte *value, byte *data, int length);
int getArgument(int argumentIndex, byte *argument);

byte totalStakeKey[32] = {42, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42};
byte totalStakeBytes[32] = {42, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};
byte stakingSc[32] = {0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 255, 255};
byte callValue[32] = {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};
byte data[270] = "stake@01@aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa@ffff";

void delegate()
{
    i64 stake = int64getArgument(0);
    i64 totalStake = int64storageLoad(totalStakeKey, 32);
    totalStake += stake;
    int64storageStore(totalStakeKey, 32, totalStake);
}

void sendToStaking()
{
    getCallValue(callValue);
    asyncCall(stakingSc, callValue, data, 270);
}

void callBack() {
}

void _main(void)
{
}
