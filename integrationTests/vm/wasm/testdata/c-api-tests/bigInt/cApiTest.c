#include "../context.h"
#include "../util.h"

STORAGE_KEY(VALUE_IN_STORAGE);

const int LIMIT = 1000;

bigInt values[LIMIT] = { 0 };
byte buffer[100] = { 0 };
byte valueA[32] = {1,2,3,4,5,6,7,8,9,0,1,2,3,4,5,6,7,8,9,0,1,2,3,4,5,6,7,8,9,0,1,2};
byte valueB[32] = {9,9,9,9,9,1,3,5,7,9,2,4,6,8,0,2,4,6,8,0,9,7,5,3,1,3,5,7,9,1,3,5};

void init() 
{

}

void bigIntNewTest()
{
    int i;

    for (i = 0; i < LIMIT; i++)
    {
        values[i] = bigIntNew(i);
    }
}

void bigIntGetUnsignedBytesTest()
{
    int i;

    for (i = 0; i < LIMIT; i++)
    {
        bigIntGetUnsignedBytes(values[i], buffer);
    }
}

void bigIntGetSignedBytesTest()
{
    int i;

    for (i = 0; i < LIMIT; i++)
    {
        bigIntGetSignedBytes(values[i], buffer);
    }
}

void bigIntSetUnsignedBytesTest()
{
    int i;
    byte val[3] = { 42, 127, 255 };

    for (i = 0; i < LIMIT; i++)
    {
        bigIntSetUnsignedBytes(values[i], val, 3);
    }
}

void bigIntSetSignedBytesTest()
{
    int i;
    // 255 will be -1 when stored
    byte val[3] = { 42, 127, 255 };

    for (i = 0; i < LIMIT; i++)
    {
        bigIntSetSignedBytes(values[i], val, 3);
    }
}

void bigIntIsInt64Test()
{
    int i;

    for (i = 0; i < LIMIT; i++)
    {
        bigIntIsInt64(values[i]);
    }
}

void bigIntGetInt64Test()
{
    int i;

    for (i = 0; i < LIMIT; i++)
    {
        bigIntGetInt64(values[i]);   
    }
}

void bigIntSetInt64Test()
{
    int i;

    for (i = 0; i < LIMIT; i++)
    {
        bigIntSetInt64(values[i], i);   
    }
}

void bigIntAddTest()
{
    int i;
    bigIntSetUnsignedBytes(values[0], valueA, 32);
    for (i = 1; i < LIMIT; i++)
    {
        bigIntAdd(values[i], values[i], values[i - 1]);   
    }
}

void bigIntSubTest()
{
    int i;
    bigIntSetUnsignedBytes(values[0], valueA, 32);
    for (i = 1; i < LIMIT; i++)
    {
        bigIntSub(values[i], values[i], values[i - 1]);
    }
}

void bigIntMulTest()
{
    int i;
    values[0] = bigIntNew(2);
    for (i = 1; i < 20; i++)
    {
        bigIntMul(values[i], values[i], values[i - 1]);
    }
}

void bigIntMul25Test()
{
    int i;
    values[0] = bigIntNew(2);
    for (i = 1; i < 25; i++)
    {
        bigIntMul(values[i], values[i], values[i - 1]);
    }
}

void bigIntMul32Test()
{
    bigIntNewTest();
    int i, k;

    for (i = 0; i < LIMIT-3; i+=3)
    {
        bigIntSetUnsignedBytes(values[i], valueA, 32);
	    bigIntSetUnsignedBytes(values[i+1], valueB, 32);
    }

    for (k = 0; k < 1000; k++) {
        for (i = 2; i < LIMIT; i+=3)
        {
            bigIntMul(values[i], values[i-1], values[i-2]);
        }
    }
}

void bigIntTDivTest()
{
    bigIntNewTest();
    int i, k;

    for (i = 0; i < LIMIT-3; i+=3)
    {
        bigIntSetUnsignedBytes(values[i], valueA, 32);
	    bigIntSetUnsignedBytes(values[i+1], valueB, 32);
    }

    for (k = 0; k < 1000; k++) {
        for (i = 2; i < LIMIT; i+=3)
        {
            bigIntTDiv(values[i], values[i-1], values[i-2]);
        }
    }
}

void bigIntTModTest()
{
    bigIntNewTest();
    int i, k;

    for (i = 0; i < LIMIT-3; i+=3)
    {
        bigIntSetUnsignedBytes(values[i], valueA, 32);
	    bigIntSetUnsignedBytes(values[i+1], valueB, 32);
    }

    for (k = 0; k < 1000; k++) {
        for (i = 2; i < LIMIT; i+=3)
        {
            bigIntTMod(values[i], values[i-1], values[i-2]);
        }
    }
}

void bigIntEDivTest()
{
    bigIntNewTest();
    int i, k;

    for (i = 0; i < LIMIT-3; i+=3)
    {
        bigIntSetUnsignedBytes(values[i], valueA, 32);
	    bigIntSetUnsignedBytes(values[i+1], valueB, 32);
    }

    for (k = 0; k < 1000; k++) {
        for (i = 2; i < LIMIT; i+=3)
        {
            bigIntEDiv(values[i], values[i-1], values[i-2]);
        }
    }
}

void bigIntEModTest()
{
    bigIntNewTest();
    int i, k;

    for (i = 0; i < LIMIT-3; i+=3)
    {
        bigIntSetUnsignedBytes(values[i], valueA, 32);
	    bigIntSetUnsignedBytes(values[i+1], valueB, 32);
    }

    for (k = 0; k < 1000; k++) {
        for (i = 2; i < LIMIT; i+=3)
        {
            bigIntEMod(values[i], values[i-1], values[i-2]);
        }
    }
}

void bigIntInitSetup() {
    bigIntNewTest();
    int i;

    for (i = 0; i < LIMIT-3; i+=3)
    {
        bigIntSetUnsignedBytes(values[i], valueA, 32);
	    bigIntSetUnsignedBytes(values[i+1], valueB, 32);
    }
}

void bigIntCmpTest()
{
    int i;

    for (i = 0; i < LIMIT - 1; i++)
    {
        bigIntCmp(values[i], values[i + 1]);
    }
}

void bigIntFinishUnsignedTest()
{
    int i;

    for (i = 0; i < LIMIT; i++)
    {
        // unsure if the process is killed after the first call
        bigIntFinishUnsigned(values[i]);
    }
}

void bigIntFinishSignedTest()
{
    int i;

    for (i = 0; i < LIMIT; i++)
    {
        // unsure if the process is killed after the first call
        bigIntFinishSigned(values[i]);
    }
}

void bigIntGetUnsignedArgumentTest()
{
    int i;
    int args = getNumArguments();

    for (i = 0; i < args; i++)
    {
        bigIntGetUnsignedArgument(i, values[i]);
    }
}

void bigIntGetSignedArgumentTest()
{
    int i;
    int args = getNumArguments();

    for (i = 0; i < args; i++)
    {
        bigIntGetSignedArgument(i, values[i]);
    }
}

void bigIntGetCallValueTest()
{
    int i;

    for (i = 0; i < LIMIT; i++)
    {
        bigIntGetCallValue(values[i]);
    }
}

void bigIntStorageLoadUnsignedTest()
{
    int i;

    for (i = 0; i < LIMIT; i++)
    {
        bigIntStorageLoadUnsigned(VALUE_IN_STORAGE_KEY, VALUE_IN_STORAGE_KEY_LEN, values[i]);
    }
}

void bigIntStorageStoreUnsignedTest()
{
    int i;

    for (i = 0; i < LIMIT; i++)
    {
        bigIntStorageStoreUnsigned(VALUE_IN_STORAGE_KEY, VALUE_IN_STORAGE_KEY_LEN, values[i]);
    }
}

void bigIntAbsTest()
{
    int i;

    for (i = 0; i < LIMIT; i++)
    {
        bigIntAbs(values[i], values[i]);
    }
}

void bigIntNedTest()
{
    int i;

    for (i = 0; i < LIMIT; i++)
    {
        bigIntNeg(values[i], values[i]);
    }
}

void bigIntAndTest()
{
    int i;

    for (i = 0; i < LIMIT - 1; i++)
    {
        bigIntAnd(values[i], values[i], values[i + 1]);
    }
}

void bigIntShrTest()
{
    int i;
    int shiftBy = 5;

    for (i = 0; i < LIMIT; i++)
    {
        bigIntShr(values[i], values[i], shiftBy);
    }
}