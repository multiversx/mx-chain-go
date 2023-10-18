int mBufferNew();
int mBufferGetArgument(int id, int mBufferHandle);
int mBufferStorageStore(int keyHandle, int mBufferHandle);
int mBufferStorageLoad(int keyHandle, int mBufferHandle);

void init()
{
}

void trieStore()
{
    int argIndex = 0;
    int key = mBufferNew();
    int data = mBufferNew();
    mBufferGetArgument(argIndex++, key);
    mBufferGetArgument(argIndex++, data);
    mBufferStorageStore(key, data);
}

void trieLoad()
{
    int key = mBufferNew();
    int data = mBufferNew();
    mBufferGetArgument(0, key);
    mBufferStorageLoad(key, data);
}
