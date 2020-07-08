typedef unsigned char byte;
typedef unsigned int i32;
typedef unsigned long long i64;

int int64storageStore(byte *key, int keyLength, long long value);
long long int64storageLoad(byte *key, int keyLength);
void int64finish(i64 value);

byte counterKey[32] = {42, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42};

void init()
{
    int64storageStore(counterKey, 32, 0);
}

void callBack() {
}

void callMe()
{
    i64 counter = int64storageLoad(counterKey, 32);
    counter++;
    int64storageStore(counterKey, 32, counter);
}

void numCalled()
{
    i64 counter = int64storageLoad(counterKey, 32);
    int64finish(counter);
}

void _main(void)
{
}