typedef unsigned char byte;
typedef unsigned int i32;
typedef unsigned long long i64;

int int64storageStore(byte *key, i64 value);
i64 int64storageLoad(byte *key);
void int64finish(i64 value);

byte counterKey[32] = {42, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42};

void init()
{
    int64storageStore(counterKey, 0);
}

void callMe()
{
    i64 counter = int64storageLoad(counterKey);
    counter++;
    int64storageStore(counterKey, counter);
}

void numCalled()
{
    i64 counter = int64storageLoad(counterKey);
    int64finish(counter);
}

void _main(void)
{
}