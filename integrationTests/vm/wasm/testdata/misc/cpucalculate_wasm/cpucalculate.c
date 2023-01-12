typedef unsigned char byte;
typedef unsigned int bigInt;

int getCallValue(byte *resultOffset);
void finish(byte *data, int length);

bigInt bigIntNew(long long value);
void bigIntGetCallValue(bigInt destination);
void bigIntSetInt64(bigInt destination, long long value);
long long bigIntGetInt64(bigInt reference);
void bigIntFinishUnsigned(bigInt reference);

long long calculate(long long cycles)
{
   long long rs = 0, i = 0;
   for (i = 0; i < cycles; ++i)
   {
      rs = rs + i * i * i * i * i;
    }
    return rs;
}

void cpuCalculate(void) {
	bigInt bcv = bigIntNew(0);
	bigIntGetCallValue(bcv);
	long long cv = bigIntGetInt64(bcv);
	long long result = 0;

	result = calculate(cv);

	bigInt bresult = bigIntNew(result);
	bigIntFinishUnsigned(bresult);
}
