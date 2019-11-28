typedef unsigned char byte;
typedef unsigned int i32;
typedef unsigned long long i64;

typedef unsigned int bigInt;

int int64storageStore(byte *key, i64 value);
i64 int64storageLoad(byte *key);
i64 int64getArgument(int argumentIndex);
int getCallValue(byte *result);
int transferValue(byte *destination, byte *value, byte *data, int length);
int getArgument(int argumentIndex, byte *argument);

byte totalStakeKey[32] = {42, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42};
byte totalStakeBytes[32] = {42, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};
byte stakingSc[32] = {0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 255};
byte callValue[32] = {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};
byte data[6 + 64] = { 's', 't', 'a', 'k', 'e', '@', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a', 'a' };

void delegate()
{
    i64 stake = int64getArgument(0);
    i64 totalStake = int64storageLoad(totalStakeKey);
    totalStake += stake;
    int64storageStore(totalStakeKey, totalStake);
}

void sendToStaking()
{
    getCallValue(callValue);
    transferValue(stakingSc, callValue, data, 72);
}

void _main(void)
{
}