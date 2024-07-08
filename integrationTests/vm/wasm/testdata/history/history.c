typedef unsigned char byte;
typedef unsigned int i32;
typedef unsigned long long i64;

int getArgument(int argumentIndex, byte *argument);
long long int64getArgument(int argumentIndex);
long long getBlockNonce();
long long getBlockEpoch();
void getStateRootHash(byte *hash);

int int64storageStore(byte *key, int keyLength, long long value);
long long int64storageLoad(byte *key, int keyLength);

void finish(byte *data, int length);
void int64finish(long long value);

byte zero32_buffer_a[32] = {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};
byte zero32_buffer_b[32] = {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};
byte zero32_buffer_c[32] = {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};
byte storageKey[] = "state";

void init()
{
}

void upgrade()
{
}

void setState()
{
    i64 state = int64getArgument(0);
    int64storageStore(storageKey, sizeof(storageKey) - 1, state);
}

void getState()
{
    i64 state = int64storageLoad(storageKey, sizeof(storageKey) - 1);
    int64finish(state);
}

void getNow()
{
    i64 nonce = getBlockNonce();

    byte *stateRootHash = zero32_buffer_a;
    getStateRootHash(stateRootHash);

    int64finish(nonce);
    finish(stateRootHash, 32);
}
