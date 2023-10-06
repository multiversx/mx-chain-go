typedef unsigned char byte;
typedef unsigned int bigInt;

int getNumArguments();
long long int64getArgument(int argumentIndex);
void int64finish(long long value);
bigInt bigIntNew(long long value);
int mBufferNew();
int mBufferGetArgument(int id, int mBufferHandle);
int mBufferAppendBytes(int mBufferHandle, byte *dataOffset, int dataLength);
int mBufferFinish(int mBufferHandle);
void bigIntGetUnsignedArgument(int argumentIndex, bigInt argument);
void managedAsyncCall(int addressBuffer, int valueBuffer, int functionBuffer, int argumentsBuffer);
void signalError(byte *message, int length);

int readArgumentsAsVectorOfBuffers(int *argIndex);

void init()
{
}

void doAsyncCall()
{
    int address = mBufferNew();
    bigInt value = bigIntNew(0);
    int function = mBufferNew();

    int argIndex = 0;

    mBufferGetArgument(argIndex++, address);
    bigIntGetUnsignedArgument(argIndex++, value);
    mBufferGetArgument(argIndex++, function);

    int arguments = readArgumentsAsVectorOfBuffers(&argIndex);

    managedAsyncCall(address, value, function, arguments);
}

void callBack()
{
    int numArguments = getNumArguments();

    int64finish(0xCA11BAC3);

    for (int i = 0; i < numArguments; i++)
    {
        int dump = mBufferNew();
        mBufferGetArgument(i, dump);
        mBufferFinish(dump);
    }

    int64finish(0xCA11BAC3);
}

int readArgumentsAsVectorOfBuffers(int *argIndex)
{
    int resultVector = mBufferNew();
    int numItems = int64getArgument((*argIndex)++);

    for (int j = 0; j < numItems; j++)
    {
        int item = mBufferNew();
        mBufferGetArgument((*argIndex)++, item);
        // Append as big endian.
        mBufferAppendBytes(resultVector, &item + 3, 1);
        mBufferAppendBytes(resultVector, &item + 2, 1);
        mBufferAppendBytes(resultVector, &item + 1, 1);
        mBufferAppendBytes(resultVector, &item + 0, 1);
    }

    return resultVector;
}
